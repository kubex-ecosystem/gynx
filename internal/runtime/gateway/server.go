// Package gateway provides the gateway server functionality for the gnyx.
package gateway

import (
	"context"
	"net"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/gin-gonic/gin"
	"github.com/kubex-ecosystem/gnyx/internal/api/routes"
	"github.com/kubex-ecosystem/gnyx/internal/app"
	"github.com/kubex-ecosystem/gnyx/internal/config"
	"github.com/kubex-ecosystem/gnyx/internal/features/providers/registry"
	"github.com/kubex-ecosystem/gnyx/internal/features/ui"
	"github.com/kubex-ecosystem/gnyx/internal/runtime/middlewares"
	"github.com/kubex-ecosystem/gnyx/internal/web"

	kbxMod "github.com/kubex-ecosystem/gnyx/internal/module/kbx"
	kbxGet "github.com/kubex-ecosystem/kbx/get"
	kbxTypes "github.com/kubex-ecosystem/kbx/types"

	gl "github.com/kubex-ecosystem/logz"
)

// Server represents the gateway server
type Server struct {
	*gin.Engine

	cfg        *config.ServerConfig
	registry   *registry.Registry
	middleware *middlewares.ProductionMiddleware
	handler    http.Handler
	once       sync.Once
	container  *app.Container
}

// NewServer creates a new gateway server instance
func NewServer(cfg *config.ServerConfig) (*Server, error) {
	engine := gin.New()
	var err error

	// Initialize production middleware
	prodConfig := middlewares.DefaultProductionConfig()
	prodMiddleware := middlewares.NewProductionMiddleware(prodConfig)

	// Load providers registry for gateway, if specified
	reg, err := registry.LoadResolved(cfg)
	if err != nil {
		return nil, gl.Errorf("failed to load gateway providers registry: %v", err)
	}
	if reg == nil {
		gl.Warn("Provider registry is nil after loading - this may cause issues with provider resolution in the gateway")
	} else {
		for _, providerName := range reg.ListProviders() {
			prodMiddleware.RegisterProvider(providerName)
		}
	}

	// Bootstrap application container
	container, err := app.NewContainer(context.Background(), config.LoadConfig())
	if err != nil {
		return nil, gl.Errorf("failed to bootstrap application container: %v", err)
	}
	if container.GetConfig().ServerConfig == nil {
		sCfg := container.GetConfig()
		sCfg.ServerConfig = config.NewServerConfig()
		sCfg.ServerConfig.SrvConfig = kbxTypes.NewSrvConfig()
		container, err = app.NewContainer(context.Background(), sCfg)
	}

	container.GetConfig().ServerConfig.Runtime.PubCertKeyPath = os.ExpandEnv(kbxGet.EnvOr("KUBEX_GNYX_PUBLIC_KEY_PATH", kbxMod.DefaultGNyxCertPath))
	container.GetConfig().ServerConfig.Runtime.PrivKeyPath = os.ExpandEnv(kbxGet.EnvOr("KUBEX_GNYX_PRIVATE_KEY_PATH", kbxMod.DefaultGNyxKeyPath))

	if container.GetConfig().ServerConfig.Runtime.PrivKeyPath == "" || container.GetConfig().ServerConfig.Runtime.PubCertKeyPath == "" {
		return nil, gl.Errorf("JWT certificate paths are not set in server config")
	}

	if err := container.Bootstrap(context.Background()); err != nil {
		return nil, gl.Errorf("failed to bootstrap application: %v", err)
	}
	container.GetConfig().ServerConfig.ProvidersConfig = cfg.ProvidersConfig
	container.GetConfig().ServerConfig.Files.ProvidersConfig = cfg.Files.ProvidersConfig

	return &Server{
		Engine:     engine,
		cfg:        cfg,
		registry:   reg,
		middleware: prodMiddleware,
		container:  container,
	}, nil
}

// Start starts the gateway server
func (s *Server) Start() error {
	serverCfg := s.container.GetConfig().ServerConfig
	if serverCfg.Basic.Debug && !serverCfg.Basic.ReleaseMode {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}
	// Setup graceful shutdown
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		gl.Log("info", "🛑 Shutting down gracefully...")
		s.middleware.Stop()
		os.Exit(0)
	}()

	serverCfg.Runtime.Bind = kbxGet.ValueOrIf(len(serverCfg.Runtime.Bind) == 0, kbxGet.EnvOr("KUBEX_GNYX_BIND", kbxMod.DefaultServerBind), serverCfg.Runtime.Bind)
	serverCfg.Runtime.Port = kbxGet.ValueOrIf(len(serverCfg.Runtime.Port) == 0, kbxGet.EnvOr("KUBEX_GNYX_PORT", kbxMod.DefaultServerPort), serverCfg.Runtime.Port)

	gl.Debugf("binding address: %s", serverCfg.Runtime.Bind)
	gl.Debugf("binding port: %s", serverCfg.Runtime.Port)

	srvAddr := net.JoinHostPort(
		serverCfg.Runtime.Bind, serverCfg.Runtime.Port,
	)
	swm := s.middleware
	if swm == nil {
		swm = middlewares.NewProductionMiddleware(middlewares.DefaultProductionConfig())
		s.middleware = swm
	}

	reg := s.registry
	if reg == nil {
		var err error
		reg, err = registry.LoadResolved(serverCfg)
		if err != nil {
			return gl.Errorf("failed to load provider registry: %v", err)
		}
		s.registry = reg
	}
	if reg != nil {
		for _, providerName := range reg.ListProviders() {
			swm.RegisterProvider(providerName)
		}
	}
	s.logProviderRegistry()

	// Recover middleware
	s.Use(gin.Recovery())

	// Enable logging middleware in release mode
	s.Use(swm.Logger(gl.GetLoggerZ("gnyx"))) //  Por ora, tudo igual

	if kbxGet.ValOrType[bool](serverCfg.Basic.CORSEnabled, kbxGet.EnvOrType("KUBEX_GNYX_ENABLE_CORS", true)) {
		if err := swm.SecureServerInit(s.Engine, srvAddr); err != nil {
			return gl.Errorf("failed to initialize security middleware: %v", err)
		}
	} else {
		gl.Log("warn", "CORS is disabled. This is not recommended for production environments.")
		s.Use(func(c *gin.Context) {
			origin := c.GetHeader("Origin")
			if origin != "" {
				c.Header("Access-Control-Allow-Origin", origin)
			}
			c.Header("Access-Control-Allow-Credentials", "true")
			c.Header("Access-Control-Allow-Methods", "*")
			c.Header("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
			// c.Header("Content-Security-Policy", "default-src * 'unsafe-inline' 'unsafe-eval' data: blob:;")
			c.Header("Referrer-Policy", "no-referrer")

			if c.Request.Method == "OPTIONS" {
				c.AbortWithStatus(200)
				return
			}

			c.Next()
		})
	}

	// Register routes
	s.GET("/auth/v1/callback", func(c *gin.Context) {
		redirectTo := "/api/v1/auth/v1/callback"
		if rawQuery := c.Request.URL.RawQuery; rawQuery != "" {
			redirectTo = redirectTo + "?" + rawQuery
		}
		c.Redirect(http.StatusTemporaryRedirect, redirectTo)
	})
	s.GET("/auth/google/callback", func(c *gin.Context) {
		redirectTo := "/api/v1/auth/google/callback"
		if rawQuery := c.Request.URL.RawQuery; rawQuery != "" {
			redirectTo = redirectTo + "?" + rawQuery
		}
		c.Redirect(http.StatusTemporaryRedirect, redirectTo)
	})
	routes.RegisterRoutes(s.Group("/api/v1"), s.container)

	if uiroutes, err := web.NewHandler(ui.NewUIService().GetWebFS()); err != nil {
		gl.Log("warn", "Failed to initialize web handler for UI routes: %v", err)
	} else {
		s.NoRoute(func(c *gin.Context) {
			uiroutes.ServeHTTP(c.Writer, c.Request)
		})
		gl.Info("UI routes registered successfully")
	}

	// Start server
	gl.Successf("GNyx listening on %s (Enterprise features enabled)", srvAddr)
	return s.Run(srvAddr)
}

func (s *Server) logProviderRegistry() {
	if s.registry == nil {
		gl.Warn("Provider registry is nil - this may cause issues with provider resolution")
	} else if len(s.registry.ListProviders()) == 0 {
		gl.Warn("Provider registry is empty - no providers available for resolution")
	} else {
		gl.Noticef("Provider registry contains %d providers", len(s.registry.ListProviders()))
		for _, p := range s.registry.ListProviders() {
			r := s.registry.ResolveProvider(p)
			if r == nil {
				gl.Debugf(" - Provider: %s, Available: false", p)
				continue
			}
			gl.Debugf(" - Provider: %s, Available: %v", r.Name(), r.Available() == nil)
		}
	}
}

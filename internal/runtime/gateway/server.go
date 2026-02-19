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
	"github.com/kubex-ecosystem/gnyx/internal/runtime/middlewares"

	"github.com/kubex-ecosystem/gnyx/internal/module/kbx"

	kbxGet "github.com/kubex-ecosystem/kbx/get"

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

	var reg *registry.Registry
	var err error
	if len(cfg.ProvidersConfig) > 0 {
		// Load providers registry for gateway, if specified
		reg, err = registry.Load(cfg.ProvidersConfig)
		if err != nil {
			return nil, gl.Errorf("failed to load gateway providers registry: %v", err)
		}
	} else {
		// Create an empty registry if no config is provided
		reg = &registry.Registry{}
	}

	// Initialize production middleware
	prodConfig := middlewares.DefaultProductionConfig()
	prodMiddleware := middlewares.NewProductionMiddleware(prodConfig)

	// Register all providers with production middleware
	for _, providerName := range reg.ListProviders() {
		prodMiddleware.RegisterProvider(providerName)
	}

	// config.Config

	container, err := app.NewContainer(context.Background(), config.LoadConfig())
	if err != nil {
		return nil, gl.Errorf("failed to bootstrap application container: %v", err)
	}

	if container.GetConfig().ServerConfig == nil {
		return nil, gl.Errorf("server config is nil in application container")
	}

	container.GetConfig().ServerConfig.Runtime.PubCertKeyPath = os.ExpandEnv(kbxGet.EnvOr("KUBEX_GNYX_PUBLIC_KEY_PATH", kbx.DefaultGNyxCertPath))
	container.GetConfig().ServerConfig.Runtime.PrivKeyPath = os.ExpandEnv(kbxGet.EnvOr("KUBEX_GNYX_PRIVATE_KEY_PATH", kbx.DefaultGNyxKeyPath))

	if container.GetConfig().ServerConfig.Runtime.PrivKeyPath == "" || container.GetConfig().ServerConfig.Runtime.PubCertKeyPath == "" {
		return nil, gl.Errorf("JWT certificate paths are not set in server config")
	}

	if err := container.Bootstrap(context.Background()); err != nil {
		return nil, gl.Errorf("failed to bootstrap application: %v", err)
	}

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

	serverCfg.Runtime.Bind = kbxGet.ValueOrIf(len(serverCfg.Runtime.Bind) == 0, kbxGet.EnvOr("KUBEX_GNYX_BIND", kbx.DefaultServerBind), serverCfg.Runtime.Bind)
	serverCfg.Runtime.Port = kbxGet.ValueOrIf(len(serverCfg.Runtime.Port) == 0, kbxGet.EnvOr("KUBEX_GNYX_PORT", kbx.DefaultServerPort), serverCfg.Runtime.Port)

	gl.Debugf("binding address: %s", serverCfg.Runtime.Bind)
	gl.Debugf("binding port: %s", serverCfg.Runtime.Port)

	srvAddr := net.JoinHostPort(
		serverCfg.Runtime.Bind, serverCfg.Runtime.Port,
	)
	swm := middlewares.NewProductionMiddleware(middlewares.DefaultProductionConfig())
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
			c.Header("Content-Security-Policy", "default-src * 'unsafe-inline' 'unsafe-eval' data: blob:;")
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

	// Start server
	gl.Successf("GNyx listening on %s (Enterprise features enabled)", srvAddr)
	return s.Run(srvAddr)
}

// Package wire provides runtime wiring for HTTP, SSE, and WebSocket servers.
// This is the REAL entrypoint for all network protocols, sitting ABOVE the API layer.
package wire

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	config "github.com/kubex-ecosystem/gnyx/internal/config"
	"github.com/kubex-ecosystem/gnyx/internal/features/providers/registry"
	"github.com/kubex-ecosystem/gnyx/internal/module/kbx"
	"github.com/kubex-ecosystem/gnyx/internal/runtime/middlewares"
	kbxGet "github.com/kubex-ecosystem/kbx/get"
	gl "github.com/kubex-ecosystem/logz"
)

type Config = config.Config

// HTTPWire represents the HTTP protocol wiring layer.
// It sits ABOVE Gin and orchestrates the HTTP server setup.
type HTTPWire struct {
	container          Container
	engine             *gin.Engine
	registry           *registry.Registry
	prodMiddleware     *middlewares.ProductionMiddleware
	routeRegistrar     RouteRegistrar
	middlewaresApplied bool
}

// NewHTTPWire creates a new HTTP wire instance.
// This function extracts the HTTP wiring logic from gateway/server.go.
func NewHTTPWire(container Container, routeRegistrar RouteRegistrar) (*HTTPWire, error) {
	if container == nil {
		return nil, gl.Errorf("container cannot be nil")
	}

	if container.Config() == nil {
		return nil, gl.Errorf("container config cannot be nil")
	}

	cf, ok := container.Config().(*config.Config)
	if !ok {
		return nil, gl.Errorf("invalid config type in container")
	}

	// Set Gin mode before creating the engine to avoid Gin debug logs
	if cf.ServerConfig.Basic.Debug && !cf.ServerConfig.Basic.ReleaseMode {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	engine := gin.New()

	// Type assert Config to *config.Config
	cfg, ok := container.Config().(*config.Config)
	if !ok {
		return nil, gl.Errorf("invalid config type in container")
	}

	// Load providers registry
	reg, err := registry.LoadResolved(cfg.ServerConfig)
	if err != nil {
		return nil, gl.Errorf("failed to load providers registry: %v", err)
	}

	// Initialize production middleware
	prodConfig := middlewares.DefaultProductionConfig()
	prodMiddleware := middlewares.NewProductionMiddleware(prodConfig)

	// Register all providers with production middleware
	if reg != nil {
		for _, providerName := range reg.ListProviders() {
			prodMiddleware.RegisterProvider(providerName)
		}
	}

	return &HTTPWire{
		container:      container,
		engine:         engine,
		registry:       reg,
		prodMiddleware: prodMiddleware,
		routeRegistrar: routeRegistrar,
	}, nil
}

// Wire wires up the HTTP server with middlewares and routes.
// Returns the configured Gin engine ready to be started.
func (w *HTTPWire) Wire() (*gin.Engine, error) {
	if w.middlewaresApplied {
		gl.Warn("HTTP middlewares already applied, skipping")
		return w.engine, nil
	}

	// Type assert Config to *config.Config
	configInterface := w.container.Config()
	cfg, ok := configInterface.(*Config)
	if !ok {
		return nil, gl.Errorf("invalid config type in container")
	}
	serverCfg := cfg.ServerConfig

	// Recovery middleware (always first)
	w.engine.Use(gin.Recovery())

	// OAuth callbacks at root -> API handlers
	w.engine.GET("/auth/v1/callback", func(c *gin.Context) {
		target := "/api/v1/auth/v1/callback"
		if c.Request.URL.RawQuery != "" {
			target = target + "?" + c.Request.URL.RawQuery
		}
		c.Redirect(302, target)
	})
	w.engine.GET("/auth/google/callback", func(c *gin.Context) {
		target := "/api/v1/auth/google/callback"
		if c.Request.URL.RawQuery != "" {
			target = target + "?" + c.Request.URL.RawQuery
		}
		c.Redirect(302, target)
	})

	// Fallback for providers that still hit /auth/* but route isn't registered for any reason.
	w.engine.NoRoute(func(c *gin.Context) {
		if c.Request.Method == http.MethodGet {
			path := c.Request.URL.Path
			if path == "/auth/v1/callback" || path == "/auth/google/callback" {
				target := "/api/v1" + path
				if c.Request.URL.RawQuery != "" {
					target = target + "?" + c.Request.URL.RawQuery
				}
				c.Redirect(302, target)
				return
			}
		}
		if strings.HasPrefix(c.Request.URL.Path, "/api/") {
			c.JSON(http.StatusNotFound, gin.H{
				"error":   "route not found",
				"path":    c.Request.URL.Path,
				"method":  c.Request.Method,
				"service": "gnyx-gateway",
			})
			return
		}
		c.Status(http.StatusNotFound)
	})

	// Set Gin mode
	if serverCfg.Basic.Debug && !serverCfg.Basic.ReleaseMode {
		gin.SetMode(gin.DebugMode)
		w.engine.Use(w.prodMiddleware.Logger(gl.GetLoggerZ("github.com/kubex-ecosystem/gnyx")))
	} else {
		gin.SetMode(gin.ReleaseMode)
		w.engine.Use(w.prodMiddleware.Logger(gl.GetLoggerZ("github.com/kubex-ecosystem/gnyx")))
	}

	// CORS middleware
	if kbxGet.ValOrType(serverCfg.Basic.CORSEnabled, kbxGet.EnvOrType("KUBEX_GNYX_ENABLE_CORS", true)) {
		srvRightHost := kbx.GetValueOrDefaultIf(
			(containsAny(serverCfg.Runtime.Bind, ":", "0.0.0.0") && !containsString(serverCfg.Runtime.Bind, "::")),
			"localhost",
			serverCfg.Runtime.Bind,
		)
		srvAddr := srvRightHost + ":" + serverCfg.Runtime.Port
		if err := w.prodMiddleware.SecureServerInit(w.engine, srvAddr); err != nil {
			return nil, gl.Errorf("failed to initialize security middleware: %v", err)
		}
	} else {
		gl.Warn("CORS is disabled. This is not recommended for production environments.")
		w.engine.Use(w.allowAllCORS())
	}

	// Register routes using the provided registrar function
	if w.routeRegistrar != nil {
		w.routeRegistrar(w.engine.Group("/api/v1"), w.container, w.registry, w.prodMiddleware)
	} else {
		gl.Warn("No route registrar provided, routes not registered")
	}

	w.middlewaresApplied = true
	gl.Info("HTTP wire configured successfully")

	return w.engine, nil
}

// Engine returns the underlying Gin engine.
func (w *HTTPWire) Engine() *gin.Engine {
	return w.engine
}

// Shutdown performs graceful shutdown of the HTTP wire.
func (w *HTTPWire) Shutdown() error {
	gl.Info("Shutting down HTTP wire...")
	w.prodMiddleware.Stop()
	return nil
}

// allowAllCORS is a permissive CORS middleware (NOT recommended for production).
func (w *HTTPWire) allowAllCORS() gin.HandlerFunc {
	return func(c *gin.Context) {
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
	}
}

// Helper functions

func containsAny(s string, substrs ...string) bool {
	for _, substr := range substrs {
		if containsString(s, substr) {
			return true
		}
	}
	return false
}

func containsString(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && stringContains(s, substr))
}

func stringContains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// Package app contains the application server and lifecycle management.
package app

import (
	"context"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/gin-gonic/gin"
	"github.com/kubex-ecosystem/gnyx/internal/api/routes"
	"github.com/kubex-ecosystem/gnyx/internal/config"
	"github.com/kubex-ecosystem/gnyx/internal/features/providers/registry"
	"github.com/kubex-ecosystem/gnyx/internal/module/kbx"
	"github.com/kubex-ecosystem/gnyx/internal/runtime/middlewares"
	"github.com/kubex-ecosystem/gnyx/internal/runtime/wire"

	kbxGet "github.com/kubex-ecosystem/kbx/get"
	gl "github.com/kubex-ecosystem/logz"
)

// Server represents the main application server.
// It manages the application lifecycle and coordinates all runtime components.
type Server struct {
	*gin.Engine

	cfg        *config.ServerConfig              `json:"-" yaml:"-" xml:"-" toml:"-" mapstructure:"-"`
	registry   *registry.Registry                `json:"-" yaml:"-" xml:"-" toml:"-" mapstructure:"-"`
	middleware *middlewares.ProductionMiddleware `json:"-" yaml:"-" xml:"-" toml:"-" mapstructure:"-"`
	once       sync.Once                         `json:"-" yaml:"-" xml:"-" toml:"-" mapstructure:"-"`
	container  *Container                        `json:"-" yaml:"-" xml:"-" toml:"-" mapstructure:"-"`
	httpWire   *wire.HTTPWire                    `json:"-" yaml:"-" xml:"-" toml:"-" mapstructure:"-"`
	engine     *gin.Engine                       `json:"-" yaml:"-" xml:"-" toml:"-" mapstructure:"-"`
	config     *config.Config                    `json:"-" yaml:"-" xml:"-" toml:"-" mapstructure:"-"`

	// handler    http.Handler                      `json:"-" yaml:"-" xml:"-" toml:"-" mapstructure:"-"`
}

// NewServer creates a new application server instance.
// This is the main entry point for starting the Kubex BE application.
func NewServer(cfg *config.Config) (*Server, error) {
	gl.Info("Initializing Kubex BE Server...")

	// Create and bootstrap container
	container, err := NewContainer(context.Background(), cfg)
	if err != nil {
		return nil, gl.Errorf("failed to create application container: %v", err)
	}

	// Validate server config
	if container.GetConfig().ServerConfig == nil {
		return nil, gl.Errorf("server config is nil in application container")
	}

	// Setup JWT certificate paths
	if err := setupJWTCertificates(container); err != nil {
		return nil, err
	}

	// Bootstrap the container
	if err := container.Bootstrap(context.Background()); err != nil {
		return nil, gl.Errorf("failed to bootstrap application: %v", err)
	}

	// Load provider registry
	providerRegistry, err := registry.Load(container.GetConfig().ServerConfig.Files.ProvidersConfig)
	if err != nil {
		return nil, gl.Errorf("failed to load provider registry: %v", err)
	}
	if providerRegistry == nil {
		return nil, gl.Errorf("provider registry is nil after loading")
	}

	// Create route registrar function
	routeRegistrar := func(r *gin.RouterGroup, cont any) gin.IRoutes {
		// Type assertion to *Container
		appContainer, ok := cont.(*Container)
		if !ok {
			gl.Fatal("route registrar received invalid container type")
		}
		return routes.RegisterRoutes(r, appContainer)
	}

	// Wire HTTP protocol
	httpWire, err := wire.NewHTTPWire(container, routeRegistrar)
	if err != nil {
		return nil, gl.Errorf("failed to create HTTP wire: %v", err)
	}

	// Wire up the engine
	engine, err := httpWire.Wire()
	if err != nil {
		return nil, gl.Errorf("failed to wire HTTP: %v", err)
	}

	gl.Info("Kubex BE Server initialized successfully")

	return &Server{
		container: container,
		httpWire:  httpWire,
		engine:    engine,
		registry:  providerRegistry,
		config:    cfg,
	}, nil
}

// Start starts the application server.
// It sets up graceful shutdown and starts listening for requests.
func (s *Server) Start() error {
	if s.config.ServerConfig == nil {
		return gl.Errorf("server config is nil")
	}

	// Set Gin mode based on debug config
	if s.config.ServerConfig.Basic.Debug && !s.config.ServerConfig.Basic.ReleaseMode {
		gl.Info("Running in debug mode")
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	// Setup graceful shutdown
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		gl.Infof("🛑 Shutting down gracefully...")
		if err := s.Shutdown(); err != nil {
			gl.Errorf("Error during shutdown: %v", err)
		}
		os.Exit(0)
	}()

	// Determine the correct host to bind
	srvRightHost := kbx.GetValueOrDefaultIf(
		(containsAny(s.config.ServerConfig.Runtime.Bind, ":", "0.0.0.0") &&
			!containsString(s.config.ServerConfig.Runtime.Bind, "::")),
		"localhost",
		s.config.ServerConfig.Runtime.Bind,
	)
	srvAddr := net.JoinHostPort(srvRightHost, s.config.ServerConfig.Runtime.Port)

	swm := middlewares.NewProductionMiddleware(middlewares.DefaultProductionConfig())

	// Register all providers with production middleware
	for _, providerName := range s.registry.ListProviders() {
		swm.RegisterProvider(providerName)
	}

	// Recover middleware
	s.Use(gin.Recovery())

	// Enable logging middleware in release mode
	s.Use(swm.Logger(gl.GetLoggerZ("github.com/kubex-ecosystem/gnyx"))) //  Por ora, tudo igual

	// Start server
	gl.Successf("GNyx listening on %s (Enterprise features enabled)", srvAddr)
	return s.engine.Run(srvAddr)
}

// Shutdown performs graceful shutdown of the server.
func (s *Server) Shutdown() error {
	gl.Notice("Shutting down server components...")

	if err := s.httpWire.Shutdown(); err != nil {
		gl.Errorf("Error shutting down HTTP wire: %v", err)
	}

	// TODO: Add other shutdown logic here (close DB connections, etc.)

	gl.Success("Server shutdown complete")
	return nil
}

// setupJWTCertificates configures JWT certificate paths with fallbacks.
func setupJWTCertificates(container *Container) error {
	cfg := container.GetConfig().ServerConfig.Runtime

	cfg.PubCertKeyPath = os.ExpandEnv(kbxGet.ValOrType(cfg.PubCertKeyPath, kbx.GetEnvOrDefault("KUBEX_GNYX_PUBLIC_KEY_PATH", kbx.DefaultGNyxCertPath)))
	cfg.PrivKeyPath = os.ExpandEnv(kbxGet.ValOrType(cfg.PrivKeyPath, kbx.GetEnvOrDefault("KUBEX_GNYX_PRIVATE_KEY_PATH", kbx.DefaultGNyxKeyPath)))

	return nil
}

func (s *Server) logProviderRegistry() {
	if s.registry == nil {
		gl.Warn("Provider registry is nil - this may cause issues with provider resolution")
	} else if len(s.registry.ListProviders()) == 0 {
		gl.Warn("Provider registry is empty - no providers available for resolution")
	} else {
		gl.Noticef("Provider registry contains %d providers", len(s.registry.ListProviders()))
		for _, p := range s.registry.ListProviders() {
			r := s.registry.Resolve(p)
			gl.Debugf(" - Provider: %s, Available: %v", r.Name(), r.Available() == nil)
		}
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
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// Package gateway provides a stable public API for the GNyx gateway server.
package gateway

import (
	"github.com/kubex-ecosystem/gnyx/internal/config"
	"github.com/kubex-ecosystem/gnyx/internal/runtime/gateway"
)

// ServerConfig re-exports the internal gateway server configuration.
type ServerConfig = config.ServerConfig

// Server wraps the internal gateway server, exposing a stable public API.
type Server struct {
	inner *gateway.Server
}

// NewServer constructs a new GNyx gateway server using the provided configuration.
func NewServer(cfg *ServerConfig) (*Server, error) {
	srv, err := gateway.NewServer(cfg)
	if err != nil {
		return nil, err
	}
	return &Server{inner: srv}, nil
}

// Start delegates to the internal server's Start method.
func (s *Server) Start() error {
	return s.inner.Start()
}

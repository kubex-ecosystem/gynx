// Package gateway provides a stable public API for the GNyx gateway server.
package gateway

import (
	"github.com/kubex-ecosystem/gnyx/internal/config"
	"github.com/kubex-ecosystem/gnyx/internal/runtime/gateway"
)

// ServerConfig exports the internal gateway server configuration.
type ServerConfig = config.ServerConfig

// Server wraps the internal gateway server, exposing a stable public API.
type Server = gateway.Server

// NewServer constructs a new GNyx gateway server using the provided configuration.
func NewServer(cfg *ServerConfig) (*Server, error) { return gateway.NewServer(cfg) }

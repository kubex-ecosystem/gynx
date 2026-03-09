package wire

import (
	"github.com/gin-gonic/gin"
	"github.com/kubex-ecosystem/gnyx/internal/features/providers/registry"
	"github.com/kubex-ecosystem/gnyx/internal/runtime/middlewares"
)

// Container is an interface to avoid import cycles with internal/app.
// Any type that implements these methods can be used with HTTPWire.
type Container interface {
	Config() interface{}
}

// RouteRegistrar is a function type for registering routes.
// This allows HTTPWire to be decoupled from specific route implementations.
type RouteRegistrar func(r *gin.RouterGroup, container any, reg *registry.Registry, prod *middlewares.ProductionMiddleware) gin.IRoutes

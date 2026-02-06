package wire

import "github.com/gin-gonic/gin"

// Container is an interface to avoid import cycles with internal/app.
// Any type that implements these methods can be used with HTTPWire.
type Container interface {
	Config() interface{}
}

// RouteRegistrar is a function type for registering routes.
// This allows HTTPWire to be decoupled from specific route implementations.
type RouteRegistrar func(r *gin.RouterGroup, container any) gin.IRoutes

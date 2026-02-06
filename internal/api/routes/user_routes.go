// Package routes contém a definição das rotas relacionadas aos usuários.
package routes

import (
	"github.com/gin-gonic/gin"
	genericapi "github.com/kubex-ecosystem/gnyx/internal/api"
	"github.com/kubex-ecosystem/gnyx/internal/dsclient"
	"github.com/kubex-ecosystem/gnyx/internal/types"
)

// RegisterUserRoutes registra rotas CRUD genéricas para usuários.
//
// O controller já está criado no Container com adapter unificado (Store + ORM).
// Não é necessário criar driver/executor/store manualmente.
func RegisterUserRoutes(r *gin.RouterGroup, container types.IContainer) (gin.IRoutes, error) {
	m := r.Group("/users")

	ctrl := container.GetUserController().(*genericapi.Controller[dsclient.User])

	m.GET("/", ctrl.GetAll)
	m.GET("/:id", ctrl.GetByID)
	m.POST("/", ctrl.Create)
	m.PUT("/:id", ctrl.Update)
	m.DELETE("/:id", ctrl.Delete)

	return m, nil
}

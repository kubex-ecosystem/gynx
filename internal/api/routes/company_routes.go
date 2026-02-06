// Package routes contém a definição das rotas relacionadas às empresas.
package routes

import (
	"github.com/gin-gonic/gin"
	genericapi "github.com/kubex-ecosystem/gnyx/internal/api"
	"github.com/kubex-ecosystem/gnyx/internal/dsclient"
	"github.com/kubex-ecosystem/gnyx/internal/types"
)

// RegisterCompanyRoutes registra rotas CRUD genéricas para empresas.
//
// O controller já está criado no Container com adapter unificado (Store + ORM).
// Não é necessário criar driver/executor/store manualmente.
func RegisterCompanyRoutes(r *gin.Engine, container types.IContainer) {
	ctrl := container.GetCompanyController().(*genericapi.Controller[dsclient.Company])

	r.GET("/companies", ctrl.GetAll)
	r.GET("/companies/:id", ctrl.GetByID)
	r.POST("/companies", ctrl.Create)
	r.PUT("/companies/:id", ctrl.Update)
	r.DELETE("/companies/:id", ctrl.Delete)
}

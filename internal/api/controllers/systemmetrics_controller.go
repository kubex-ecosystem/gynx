// Package systemmetrics provides the controller for managing system metrics.
package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	mdl "github.com/kubex-ecosystem/gnyx/internal/dsclient"
	svc "github.com/kubex-ecosystem/gnyx/internal/dsclient/adapters"
	t "github.com/kubex-ecosystem/gnyx/internal/types"
)

type SystemMetricController struct {
	systemmetricService svc.Service[mdl.SystemMetrics]
	APIWrapper          *t.APIWrapper[mdl.SystemMetrics]
}

func NewSystemMetricController(bridge svc.Service[mdl.SystemMetrics]) *SystemMetricController {
	return &SystemMetricController{
		systemmetricService: nil,
		APIWrapper:          t.NewAPIWrapper[mdl.SystemMetrics](),
	}
}

// GetAllSystemMetrics retorna todas as métricas de sistema.
//
// @Summary     Listar métricas de sistema
// @Description Recupera a lista de métricas de sistema.
// @Tags        systemmetrics
// @Security    BearerAuth
// @Produce     json
// @Success     200 {array} mdl.SystemMetrics
// @Failure     401 {object} ErrorResponse
// @Failure     500 {object} ErrorResponse
// @Router      /api/v1/systemmetrics [get]
func (pc *SystemMetricController) GetAllSystemMetrics(c *gin.Context) {
	// items, err := pc.systemmetricService.List()
	// if err != nil {
	// 	c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	// 	return
	// }
	// c.JSON(http.StatusOK, items)
	c.JSON(http.StatusOK, []string{})
}

// GetSystemMetricByID retorna uma métrica de sistema pelo ID.
//
// @Summary     Obter métrica de sistema
// @Description Busca uma métrica de sistema específica pelo ID.
// @Tags        systemmetrics
// @Security    BearerAuth
// @Produce     json
// @Param       id path string true "ID da Métrica de Sistema"
// @Success     200 {object} mdl.SystemMetrics
// @Failure     401 {object} ErrorResponse
// @Failure     404 {object} ErrorResponse
// @Router      /api/v1/systemmetrics/{id} [get]
func (pc *SystemMetricController) GetSystemMetricByID(c *gin.Context) {
	id := c.Param("id")
	item, err := pc.systemmetricService.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "system metric not found"})
		return
	}
	c.JSON(http.StatusOK, item)
}

// CreateSystemMetric cria uma nova métrica de sistema.
//
// @Summary     Criar métrica de sistema
// @Description Adiciona uma nova métrica de sistema.
// @Tags        systemmetrics
// @Security    BearerAuth
// @Accept      json
// @Produce     json
// @Param       payload body mdl.SystemMetrics true "Dados da Métrica de Sistema"
// @Success     201 {object} mdl.SystemMetrics
// @Failure     400 {object} ErrorResponse
// @Failure     401 {object} ErrorResponse
// @Failure     500 {object} ErrorResponse
// @Router      /api/v1/systemmetrics [post]
func (pc *SystemMetricController) CreateSystemMetric(c *gin.Context) {
	var request mdl.SystemMetrics
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	err := pc.systemmetricService.Create(c.Request.Context(), &request)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, request)
}

// UpdateSystemMetric atualiza uma métrica de sistema.
//
// @Summary     Atualizar métrica de sistema
// @Description Atualiza os dados de uma métrica de sistema.
// @Tags        systemmetrics
// @Security    BearerAuth
// @Accept      json
// @Produce     json
// @Param       id      path string            true "ID da Métrica de Sistema"
// @Param       payload body mdl.SystemMetrics true "Dados atualizados"
// @Success     200 {object} mdl.SystemMetrics
// @Failure     400 {object} ErrorResponse
// @Failure     401 {object} ErrorResponse
// @Failure     404 {object} ErrorResponse
// @Failure     500 {object} ErrorResponse
// @Router      /api/v1/systemmetrics/{id} [put]
func (pc *SystemMetricController) UpdateSystemMetric(c *gin.Context) {
	var request mdl.SystemMetrics
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	err := pc.systemmetricService.Update(c.Request.Context(), &request)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, request)
}

// DeleteSystemMetric remove uma métrica de sistema.
//
// @Summary     Remover métrica de sistema
// @Description Remove uma métrica de sistema.
// @Tags        systemmetrics
// @Security    BearerAuth
// @Produce     json
// @Param       id path string true "ID da Métrica de Sistema"
// @Success     204
// @Failure     401 {object} ErrorResponse
// @Failure     404 {object} ErrorResponse
// @Failure     500 {object} ErrorResponse
// @Router      /api/v1/systemmetrics/{id} [delete]
func (pc *SystemMetricController) DeleteSystemMetric(c *gin.Context) {
	id := c.Param("id")
	if err := pc.systemmetricService.Delete(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

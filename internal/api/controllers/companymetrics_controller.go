// Package companymetrics provides the controller for managing company metrics.
package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	mdl "github.com/kubex-ecosystem/gnyx/internal/dsclient"
	svc "github.com/kubex-ecosystem/gnyx/internal/dsclient/adapters"
	t "github.com/kubex-ecosystem/gnyx/internal/types"
)

type CompanyMetricController struct {
	companymetricService svc.Service[mdl.CompanyMetrics]
	APIWrapper           *t.APIWrapper[mdl.CompanyMetrics]
}

func NewCompanyMetricController(bridge svc.Service[mdl.CompanyMetrics]) *CompanyMetricController {
	// repo := svc.Repository{}
	// service := svc.Service{}
	return &CompanyMetricController{
		companymetricService: nil,
		APIWrapper:           t.NewAPIWrapper[mdl.CompanyMetrics](),
	}
}

// GetAllCompanyMetrics retorna todas as métricas de empresas.
//
// @Summary     Listar métricas de empresas
// @Description Recupera a lista de métricas de empresas.
// @Tags        companymetrics
// @Security    BearerAuth
// @Produce     json
// @Success     200 {array} mdl.CompanyMetrics
// @Failure     401 {object} ErrorResponse
// @Failure     500 {object} ErrorResponse
// @Router      /api/v1/companymetrics [get]
func (pc *CompanyMetricController) GetAllCompanyMetrics(c *gin.Context) {
	items, err := pc.companymetricService.List(c.Request.Context(), nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, items)
}

// GetCompanyMetricByID retorna uma métrica de empresa pelo ID.
//
// @Summary     Obter métrica de empresa
// @Description Busca uma métrica de empresa específica pelo ID.
// @Tags        companymetrics
// @Security    BearerAuth
// @Produce     json
// @Param       id path string true "ID da Métrica de Empresa"
// @Success     200 {object} mdl.CompanyMetrics
// @Failure     401 {object} ErrorResponse
// @Failure     404 {object} ErrorResponse
// @Router      /api/v1/companymetrics/{id} [get]
func (pc *CompanyMetricController) GetCompanyMetricByID(c *gin.Context) {
	id := c.Param("id")
	item, err := pc.companymetricService.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "company metric not found"})
		return
	}
	c.JSON(http.StatusOK, item)
}

// CreateCompanyMetric cria uma nova métrica de empresa.
//
// @Summary     Criar métrica de empresa
// @Description Adiciona uma nova métrica de empresa.
// @Tags        companymetrics
// @Security    BearerAuth
// @Accept      json
// @Produce     json
// @Param       payload body mdl.CompanyMetrics true "Dados da Métrica de Empresa"
// @Success     201 {object} mdl.CompanyMetrics
// @Failure     400 {object} ErrorResponse
// @Failure     401 {object} ErrorResponse
// @Failure     500 {object} ErrorResponse
// @Router      /api/v1/companymetrics [post]
func (pc *CompanyMetricController) CreateCompanyMetric(c *gin.Context) {
	var request mdl.CompanyMetrics
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	err := pc.companymetricService.Create(c.Request.Context(), &request)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, request)
}

// UpdateCompanyMetric atualiza uma métrica de empresa.
//
// @Summary     Atualizar métrica de empresa
// @Description Atualiza os dados de uma métrica de empresa.
// @Tags        companymetrics
// @Security    BearerAuth
// @Accept      json
// @Produce     json
// @Param       id      path string            true "ID da Métrica de Empresa"
// @Param       payload body mdl.CompanyMetrics true "Dados atualizados"
// @Success     200 {object} mdl.CompanyMetrics
// @Failure     400 {object} ErrorResponse
// @Failure     401 {object} ErrorResponse
// @Failure     404 {object} ErrorResponse
// @Failure     500 {object} ErrorResponse
// @Router      /api/v1/companymetrics/{id} [put]
func (pc *CompanyMetricController) UpdateCompanyMetric(c *gin.Context) {
	var request mdl.CompanyMetrics
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	err := pc.companymetricService.Update(c.Request.Context(), &request)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, request)
}

// DeleteCompanyMetric remove uma métrica de empresa.
//
// @Summary     Remover métrica de empresa
// @Description Remove uma métrica de empresa.
// @Tags        companymetrics
// @Security    BearerAuth
// @Produce     json
// @Param       id path string true "ID da Métrica de Empresa"
// @Success     204
// @Failure     401 {object} ErrorResponse
// @Failure     404 {object} ErrorResponse
// @Failure     500 {object} ErrorResponse
// @Router      /api/v1/companymetrics/{id} [delete]
func (pc *CompanyMetricController) DeleteCompanyMetric(c *gin.Context) {
	id := c.Param("id")
	if err := pc.companymetricService.Delete(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

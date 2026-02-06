// Package pipelines provides the controller for managing pipelines.
package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	mdl "github.com/kubex-ecosystem/gnyx/internal/dsclient"
	svc "github.com/kubex-ecosystem/gnyx/internal/dsclient/adapters"
	t "github.com/kubex-ecosystem/gnyx/internal/types"
)

type PipelineController struct {
	pipelineService svc.Service[mdl.Pipelines]
	APIWrapper      *t.APIWrapper[mdl.Pipelines]
}

func NewPipelineController(bridge svc.Service[mdl.Pipelines]) *PipelineController {
	// repo := svc.Repository{}
	// service := svc.Service{}
	return &PipelineController{
		pipelineService: nil,
		APIWrapper:      t.NewAPIWrapper[mdl.Pipelines](),
	}
}

// GetAllPipelines retorna todos os pipelines.
//
// @Summary     Listar pipelines
// @Description Recupera a lista de pipelines.
// @Tags        pipelines
// @Security    BearerAuth
// @Produce     json
// @Success     200 {array} mdl.Pipelines
// @Failure     401 {object} ErrorResponse
// @Failure     500 {object} ErrorResponse
// @Router      /api/v1/pipelines [get]
func (pc *PipelineController) GetAllPipelines(c *gin.Context) {
	items, err := pc.pipelineService.List(c.Request.Context(), nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, items)
}

// GetPipelineByID retorna um pipeline pelo ID.
//
// @Summary     Obter pipeline
// @Description Busca um pipeline específico pelo ID.
// @Tags        pipelines
// @Security    BearerAuth
// @Produce     json
// @Param       id path string true "ID do Pipeline"
// @Success     200 {object} mdl.Pipelines
// @Failure     401 {object} ErrorResponse
// @Failure     404 {object} ErrorResponse
// @Router      /api/v1/pipelines/{id} [get]
func (pc *PipelineController) GetPipelineByID(c *gin.Context) {
	id := c.Param("id")
	item, err := pc.pipelineService.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "pipeline not found"})
		return
	}
	c.JSON(http.StatusOK, item)
}

// CreatePipeline cria um novo pipeline.
//
// @Summary     Criar pipeline
// @Description Adiciona um novo pipeline.
// @Tags        pipelines
// @Security    BearerAuth
// @Accept      json
// @Produce     json
// @Param       payload body mdl.Pipelines true "Dados do Pipeline"
// @Success     201 {object} mdl.Pipelines
// @Failure     400 {object} ErrorResponse
// @Failure     401 {object} ErrorResponse
// @Failure     500 {object} ErrorResponse
// @Router      /api/v1/pipelines [post]
func (pc *PipelineController) CreatePipeline(c *gin.Context) {
	var request mdl.Pipelines
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	err := pc.pipelineService.Create(c.Request.Context(), &request)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, request)
}

// UpdatePipeline atualiza um pipeline.
//
// @Summary     Atualizar pipeline
// @Description Atualiza os dados de um pipeline.
// @Tags        pipelines
// @Security    BearerAuth
// @Accept      json
// @Produce     json
// @Param       id      path string            true "ID do Pipeline"
// @Param       payload body mdl.Pipelines true "Dados atualizados"
// @Success     200 {object} mdl.Pipelines
// @Failure     400 {object} ErrorResponse
// @Failure     401 {object} ErrorResponse
// @Failure     404 {object} ErrorResponse
// @Failure     500 {object} ErrorResponse
// @Router      /api/v1/pipelines/{id} [put]
func (pc *PipelineController) UpdatePipeline(c *gin.Context) {
	var request mdl.Pipelines
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	err := pc.pipelineService.Update(c.Request.Context(), &request)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, request)
}

// DeletePipeline remove um pipeline.
//
// @Summary     Remover pipeline
// @Description Remove um pipeline.
// @Tags        pipelines
// @Security    BearerAuth
// @Produce     json
// @Param       id path string true "ID do Pipeline"
// @Success     204
// @Failure     401 {object} ErrorResponse
// @Failure     404 {object} ErrorResponse
// @Failure     500 {object} ErrorResponse
// @Router      /api/v1/pipelines/{id} [delete]
func (pc *PipelineController) DeletePipeline(c *gin.Context) {
	id := c.Param("id")
	if err := pc.pipelineService.Delete(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

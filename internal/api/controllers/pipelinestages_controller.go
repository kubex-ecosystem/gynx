// Package pipelinestages provides the controller for managing pipeline stages.
package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	mdl "github.com/kubex-ecosystem/gnyx/internal/dsclient"
	svc "github.com/kubex-ecosystem/gnyx/internal/dsclient/adapters"
	t "github.com/kubex-ecosystem/gnyx/internal/types"
)

type PipelineStageController struct {
	pipelinestageService svc.Service[mdl.PipelineStages]
	APIWrapper           *t.APIWrapper[mdl.PipelineStages]
}

func NewPipelineStageController(bridge svc.Service[mdl.PipelineStages]) *PipelineStageController {
	// repo := svc.Repository{}
	// service := svc.Service{}
	return &PipelineStageController{
		pipelinestageService: nil,
		APIWrapper:           t.NewAPIWrapper[mdl.PipelineStages](),
	}
}

// GetAllPipelineStages retorna todos os estágios de pipeline.
//
// @Summary     Listar estágios de pipeline
// @Description Recupera a lista de estágios de pipeline.
// @Tags        pipelinestages
// @Security    BearerAuth
// @Produce     json
// @Success     200 {array} mdl.PipelineStages
// @Failure     401 {object} ErrorResponse
// @Failure     500 {object} ErrorResponse
// @Router      /api/v1/pipelinestages [get]
func (pc *PipelineStageController) GetAllPipelineStages(c *gin.Context) {
	items, err := pc.pipelinestageService.List(c.Request.Context(), nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, items)
}

// GetPipelineStageByID retorna um estágio de pipeline pelo ID.
//
// @Summary     Obter estágio de pipeline
// @Description Busca um estágio de pipeline específico pelo ID.
// @Tags        pipelinestages
// @Security    BearerAuth
// @Produce     json
// @Param       id path string true "ID do Estágio de Pipeline"
// @Success     200 {object} mdl.PipelineStages
// @Failure     401 {object} ErrorResponse
// @Failure     404 {object} ErrorResponse
// @Router      /api/v1/pipelinestages/{id} [get]
func (pc *PipelineStageController) GetPipelineStageByID(c *gin.Context) {
	id := c.Param("id")
	item, err := pc.pipelinestageService.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "pipeline stage not found"})
		return
	}
	c.JSON(http.StatusOK, item)
}

// CreatePipelineStage cria um novo estágio de pipeline.
//
// @Summary     Criar estágio de pipeline
// @Description Adiciona um novo estágio de pipeline.
// @Tags        pipelinestages
// @Security    BearerAuth
// @Accept      json
// @Produce     json
// @Param       payload body mdl.PipelineStages true "Dados do Estágio de Pipeline"
// @Success     201 {object} mdl.PipelineStages
// @Failure     400 {object} ErrorResponse
// @Failure     401 {object} ErrorResponse
// @Failure     500 {object} ErrorResponse
// @Router      /api/v1/pipelinestages [post]
func (pc *PipelineStageController) CreatePipelineStage(c *gin.Context) {
	var request mdl.PipelineStages
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	err := pc.pipelinestageService.Create(c.Request.Context(), &request)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, request)
}

// UpdatePipelineStage atualiza um estágio de pipeline.
//
// @Summary     Atualizar estágio de pipeline
// @Description Atualiza os dados de um estágio de pipeline.
// @Tags        pipelinestages
// @Security    BearerAuth
// @Accept      json
// @Produce     json
// @Param       id      path string            true "ID do Estágio de Pipeline"
// @Param       payload body mdl.PipelineStages true "Dados atualizados"
// @Success     200 {object} mdl.PipelineStages
// @Failure     400 {object} ErrorResponse
// @Failure     401 {object} ErrorResponse
// @Failure     404 {object} ErrorResponse
// @Failure     500 {object} ErrorResponse
// @Router      /api/v1/pipelinestages/{id} [put]
func (pc *PipelineStageController) UpdatePipelineStage(c *gin.Context) {
	var request mdl.PipelineStages
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	err := pc.pipelinestageService.Update(c.Request.Context(), &request)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, request)
}

// DeletePipelineStage remove um estágio de pipeline.
//
// @Summary     Remover estágio de pipeline
// @Description Remove um estágio de pipeline.
// @Tags        pipelinestages
// @Security    BearerAuth
// @Produce     json
// @Param       id path string true "ID do Estágio de Pipeline"
// @Success     204
// @Failure     401 {object} ErrorResponse
// @Failure     404 {object} ErrorResponse
// @Failure     500 {object} ErrorResponse
// @Router      /api/v1/pipelinestages/{id} [delete]
func (pc *PipelineStageController) DeletePipelineStage(c *gin.Context) {
	id := c.Param("id")
	if err := pc.pipelinestageService.Delete(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

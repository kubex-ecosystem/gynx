// Package systemlogs provides the controller for managing system logs.
package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	mdl "github.com/kubex-ecosystem/gnyx/internal/dsclient"
	svc "github.com/kubex-ecosystem/gnyx/internal/dsclient/adapters"
	t "github.com/kubex-ecosystem/gnyx/internal/types"
)

type SystemLogController struct {
	systemlogService svc.Service[mdl.SystemLogs]
	APIWrapper       *t.APIWrapper[mdl.SystemLogs]
}

func NewSystemLogController(bridge svc.Service[mdl.SystemLogs]) *SystemLogController {
	// repo := svc.Repository{}
	// service := svc.Service{}
	return &SystemLogController{
		systemlogService: nil,
		APIWrapper:       t.NewAPIWrapper[mdl.SystemLogs](),
	}
}

// GetAllSystemLogs retorna todos os logs de sistema.
//
// @Summary     Listar logs de sistema
// @Description Recupera a lista de logs de sistema.
// @Tags        systemlogs
// @Security    BearerAuth
// @Produce     json
// @Success     200 {array} mdl.SystemLogs
// @Failure     401 {object} ErrorResponse
// @Failure     500 {object} ErrorResponse
// @Router      /api/v1/systemlogs [get]
func (pc *SystemLogController) GetAllSystemLogs(c *gin.Context) {
	items, err := pc.systemlogService.List(c.Request.Context(), nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, items)
}

// GetSystemLogByID retorna um log de sistema pelo ID.
//
// @Summary     Obter log de sistema
// @Description Busca um log de sistema específico pelo ID.
// @Tags        systemlogs
// @Security    BearerAuth
// @Produce     json
// @Param       id path string true "ID do Log de Sistema"
// @Success     200 {object} mdl.SystemLogs
// @Failure     401 {object} ErrorResponse
// @Failure     404 {object} ErrorResponse
// @Router      /api/v1/systemlogs/{id} [get]
func (pc *SystemLogController) GetSystemLogByID(c *gin.Context) {
	id := c.Param("id")
	item, err := pc.systemlogService.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "system log not found"})
		return
	}
	c.JSON(http.StatusOK, item)
}

// CreateSystemLog cria um novo log de sistema.
//
// @Summary     Criar log de sistema
// @Description Adiciona um novo log de sistema.
// @Tags        systemlogs
// @Security    BearerAuth
// @Accept      json
// @Produce     json
// @Param       payload body mdl.SystemLogs true "Dados do Log de Sistema"
// @Success     201 {object} mdl.SystemLogs
// @Failure     400 {object} ErrorResponse
// @Failure     401 {object} ErrorResponse
// @Failure     500 {object} ErrorResponse
// @Router      /api/v1/systemlogs [post]
func (pc *SystemLogController) CreateSystemLog(c *gin.Context) {
	var request mdl.SystemLogs
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	err := pc.systemlogService.Create(c.Request.Context(), &request)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, request)
}

// UpdateSystemLog atualiza um log de sistema.
//
// @Summary     Atualizar log de sistema
// @Description Atualiza os dados de um log de sistema.
// @Tags        systemlogs
// @Security    BearerAuth
// @Accept      json
// @Produce     json
// @Param       id      path string            true "ID do Log de Sistema"
// @Param       payload body mdl.SystemLogs true "Dados atualizados"
// @Success     200 {object} mdl.SystemLogs
// @Failure     400 {object} ErrorResponse
// @Failure     401 {object} ErrorResponse
// @Failure     404 {object} ErrorResponse
// @Failure     500 {object} ErrorResponse
// @Router      /api/v1/systemlogs/{id} [put]
func (pc *SystemLogController) UpdateSystemLog(c *gin.Context) {
	var request mdl.SystemLogs
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	err := pc.systemlogService.Update(c.Request.Context(), &request)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, request)
}

// DeleteSystemLog remove um log de sistema.
//
// @Summary     Remover log de sistema
// @Description Remove um log de sistema.
// @Tags        systemlogs
// @Security    BearerAuth
// @Produce     json
// @Param       id path string true "ID do Log de Sistema"
// @Success     204
// @Failure     401 {object} ErrorResponse
// @Failure     404 {object} ErrorResponse
// @Failure     500 {object} ErrorResponse
// @Router      /api/v1/systemlogs/{id} [delete]
func (pc *SystemLogController) DeleteSystemLog(c *gin.Context) {
	id := c.Param("id")
	if err := pc.systemlogService.Delete(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

// Package auditlogs provides the controller for managing audit logs.
package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	mdl "github.com/kubex-ecosystem/gnyx/internal/dsclient"
	svc "github.com/kubex-ecosystem/gnyx/internal/dsclient/adapters"
	t "github.com/kubex-ecosystem/gnyx/internal/types"
)

type AuditLogController struct {
	auditlogService svc.Service[mdl.AuditLogs]
	APIWrapper      *t.APIWrapper[mdl.AuditLogs]
}

func NewAuditLogController(bridge svc.Service[mdl.AuditLogs]) *AuditLogController {
	// repo := svc.Repository{}
	// service := svc.Service{}
	return &AuditLogController{
		auditlogService: nil,
		APIWrapper:      t.NewAPIWrapper[mdl.AuditLogs](),
	}
}

// GetAllAuditLogs retorna todos os logs de auditoria.
//
// @Summary     Listar logs de auditoria
// @Description Recupera a lista de logs de auditoria.
// @Tags        auditlogs
// @Security    BearerAuth
// @Produce     json
// @Success     200 {array} mdl.AuditLogs
// @Failure     401 {object} ErrorResponse
// @Failure     500 {object} ErrorResponse
// @Router      /api/v1/auditlogs [get]
func (pc *AuditLogController) GetAllAuditLogs(c *gin.Context) {
	items, err := pc.auditlogService.List(c.Request.Context(), nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, items)
}

// GetAuditLogByID retorna um log de auditoria pelo ID.
//
// @Summary     Obter log de auditoria
// @Description Busca um log de auditoria específico pelo ID.
// @Tags        auditlogs
// @Security    BearerAuth
// @Produce     json
// @Param       id path string true "ID do Log de Auditoria"
// @Success     200 {object} mdl.AuditLogs
// @Failure     401 {object} ErrorResponse
// @Failure     404 {object} ErrorResponse
// @Router      /api/v1/auditlogs/{id} [get]
func (pc *AuditLogController) GetAuditLogByID(c *gin.Context) {
	id := c.Param("id")
	item, err := pc.auditlogService.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "audit log not found"})
		return
	}
	c.JSON(http.StatusOK, item)
}

// CreateAuditLog cria um novo log de auditoria.
//
// @Summary     Criar log de auditoria
// @Description Adiciona um novo log de auditoria.
// @Tags        auditlogs
// @Security    BearerAuth
// @Accept      json
// @Produce     json
// @Param       payload body mdl.AuditLogs true "Dados do Log de Auditoria"
// @Success     201 {object} mdl.AuditLogs
// @Failure     400 {object} ErrorResponse
// @Failure     401 {object} ErrorResponse
// @Failure     500 {object} ErrorResponse
// @Router      /api/v1/auditlogs [post]
func (pc *AuditLogController) CreateAuditLog(c *gin.Context) {
	var request mdl.AuditLogs
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	err := pc.auditlogService.Create(c.Request.Context(), &request)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, request)
}

// UpdateAuditLog atualiza um log de auditoria.
//
// @Summary     Atualizar log de auditoria
// @Description Atualiza os dados de um log de auditoria.
// @Tags        auditlogs
// @Security    BearerAuth
// @Accept      json
// @Produce     json
// @Param       id      path string            true "ID do Log de Auditoria"
// @Param       payload body mdl.AuditLogs true "Dados atualizados"
// @Success     200 {object} mdl.AuditLogs
// @Failure     400 {object} ErrorResponse
// @Failure     401 {object} ErrorResponse
// @Failure     404 {object} ErrorResponse
// @Failure     500 {object} ErrorResponse
// @Router      /api/v1/auditlogs/{id} [put]
func (pc *AuditLogController) UpdateAuditLog(c *gin.Context) {
	var request mdl.AuditLogs
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	err := pc.auditlogService.Update(c.Request.Context(), &request)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, request)
}

// DeleteAuditLog remove um log de auditoria.
//
// @Summary     Remover log de auditoria
// @Description Remove um log de auditoria.
// @Tags        auditlogs
// @Security    BearerAuth
// @Produce     json
// @Param       id path string true "ID do Log de Auditoria"
// @Success     204
// @Failure     401 {object} ErrorResponse
// @Failure     404 {object} ErrorResponse
// @Failure     500 {object} ErrorResponse
// @Router      /api/v1/auditlogs/{id} [delete]
func (pc *AuditLogController) DeleteAuditLog(c *gin.Context) {
	id := c.Param("id")
	if err := pc.auditlogService.Delete(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

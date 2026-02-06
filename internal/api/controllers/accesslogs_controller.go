// Package accesslogs provides the controller for managing access logs in the application.
package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	mdl "github.com/kubex-ecosystem/gnyx/internal/dsclient"
	svc "github.com/kubex-ecosystem/gnyx/internal/dsclient/adapters"
	t "github.com/kubex-ecosystem/gnyx/internal/types"
)

type AccessLogController struct {
	accesslogService svc.Service[mdl.AccessLogs]
	APIWrapper       *t.APIWrapper[mdl.AccessLogs]
}

func NewAccessLogController(bridge svc.Service[mdl.AccessLogs]) *AccessLogController {
	return &AccessLogController{
		accesslogService: bridge,
		APIWrapper:       t.NewAPIWrapper[mdl.AccessLogs](),
	}
}

// GetAllAccessLogs retorna todos os logs de acesso.
//
// @Summary     Listar logs de acesso
// @Description Recupera a lista de logs de acesso registrados na base.
// @Tags        accesslogs
// @Security    BearerAuth
// @Produce     json
// @Success     200 {array} mdl.AccessLogs
// @Failure     401 {object} ErrorResponse
// @Failure     500 {object} ErrorResponse
// @Router      /api/v1/accesslogs [get]
func (pc *AccessLogController) GetAllAccessLogs(c *gin.Context) {
	accesslogs, err := pc.accesslogService.List(c.Request.Context(), nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, accesslogs)
}

// GetAccessLogByID retorna um log de acesso pelo ID.
//
// @Summary     Obter log de acesso
// @Description Busca um log de acesso específico pelo ID.
// @Tags        accesslogs
// @Security    BearerAuth
// @Produce     json
// @Param       id path string true "ID do Log de Acesso"
// @Success     200 {object} mdl.AccessLogs
// @Failure     401 {object} ErrorResponse
// @Failure     404 {object} ErrorResponse
// @Router      /api/v1/accesslogs/{id} [get]
func (pc *AccessLogController) GetAccessLogByID(c *gin.Context) {
	id := c.Param("id")
	accesslog, err := pc.accesslogService.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "access log not found"})
		return
	}
	c.JSON(http.StatusOK, accesslog)
}

// CreateAccessLog cria um novo log de acesso.
//
// @Summary     Criar log de acesso
// @Description Persiste um novo log de acesso com os dados enviados.
// @Tags        accesslogs
// @Security    BearerAuth
// @Accept      json
// @Produce     json
// @Param       payload body mdl.AccessLogs true "Dados do Log de Acesso"
// @Success     201 {object} mdl.AccessLogs
// @Failure     400 {object} ErrorResponse
// @Failure     401 {object} ErrorResponse
// @Failure     500 {object} ErrorResponse
// @Router      /api/v1/accesslogs [post]
func (pc *AccessLogController) CreateAccessLog(c *gin.Context) {
	var accesslogRequest mdl.AccessLogs
	if err := c.ShouldBindJSON(&accesslogRequest); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	err := pc.accesslogService.Create(c.Request.Context(), &accesslogRequest)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, accesslogRequest)
}

// UpdateAccessLog atualiza um log de acesso existente.
//
// @Summary     Atualizar log de acesso
// @Description Atualiza os dados de um log de acesso identificado por ID.
// @Tags        accesslogs
// @Security    BearerAuth
// @Accept      json
// @Produce     json
// @Param       id      path string            true "ID do Log de Acesso"
// @Param       payload body mdl.AccessLogs true "Dados atualizados"
// @Success     200 {object} mdl.AccessLogs
// @Failure     400 {object} ErrorResponse
// @Failure     401 {object} ErrorResponse
// @Failure     404 {object} ErrorResponse
// @Failure     500 {object} ErrorResponse
// @Router      /api/v1/accesslogs/{id} [put]
func (pc *AccessLogController) UpdateAccessLog(c *gin.Context) {
	var accesslogRequest mdl.AccessLogs
	if err := c.ShouldBindJSON(&accesslogRequest); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	err := pc.accesslogService.Update(c.Request.Context(), &accesslogRequest)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, accesslogRequest)
}

// DeleteAccessLog remove um log de acesso.
//
// @Summary     Remover log de acesso
// @Description Exclui um log de acesso identificado pelo ID.
// @Tags        accesslogs
// @Security    BearerAuth
// @Produce     json
// @Param       id path string true "ID do Log de Acesso"
// @Success     204
// @Failure     401 {object} ErrorResponse
// @Failure     404 {object} ErrorResponse
// @Failure     500 {object} ErrorResponse
// @Router      /api/v1/accesslogs/{id} [delete]
func (pc *AccessLogController) DeleteAccessLog(c *gin.Context) {
	id := c.Param("id")
	if err := pc.accesslogService.Delete(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

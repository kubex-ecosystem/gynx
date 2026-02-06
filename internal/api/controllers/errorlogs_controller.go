// Package errorlogs provides the controller for managing error logs.
package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	mdl "github.com/kubex-ecosystem/gnyx/internal/dsclient"
	svc "github.com/kubex-ecosystem/gnyx/internal/dsclient/adapters"
	t "github.com/kubex-ecosystem/gnyx/internal/types"
)

type ErrorLogController struct {
	errorlogService svc.Service[mdl.ErrorLogs]
	APIWrapper      *t.APIWrapper[mdl.ErrorLogs]
}

func NewErrorLogController(bridge svc.Service[mdl.ErrorLogs]) *ErrorLogController {
	// repo := svc.Repository{}
	// service := svc.Service{}
	return &ErrorLogController{
		errorlogService: nil,
		APIWrapper:      t.NewAPIWrapper[mdl.ErrorLogs](),
	}
}

// GetAllErrorLogs retorna todos os logs de erro.
//
// @Summary     Listar logs de erro
// @Description Recupera a lista de logs de erro.
// @Tags        errorlogs
// @Security    BearerAuth
// @Produce     json
// @Success     200 {array} mdl.ErrorLogs
// @Failure     401 {object} ErrorResponse
// @Failure     500 {object} ErrorResponse
// @Router      /api/v1/errorlogs [get]
func (pc *ErrorLogController) GetAllErrorLogs(c *gin.Context) {
	items, err := pc.errorlogService.List(c.Request.Context(), nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, items)
}

// GetErrorLogByID retorna um log de erro pelo ID.
//
// @Summary     Obter log de erro
// @Description Busca um log de erro específico pelo ID.
// @Tags        errorlogs
// @Security    BearerAuth
// @Produce     json
// @Param       id path string true "ID do Log de Erro"
// @Success     200 {object} mdl.ErrorLogs
// @Failure     401 {object} ErrorResponse
// @Failure     404 {object} ErrorResponse
// @Router      /api/v1/errorlogs/{id} [get]
func (pc *ErrorLogController) GetErrorLogByID(c *gin.Context) {
	id := c.Param("id")
	item, err := pc.errorlogService.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "error log not found"})
		return
	}
	c.JSON(http.StatusOK, item)
}

// CreateErrorLog cria um novo log de erro.
//
// @Summary     Criar log de erro
// @Description Adiciona um novo log de erro.
// @Tags        errorlogs
// @Security    BearerAuth
// @Accept      json
// @Produce     json
// @Param       payload body mdl.ErrorLogs true "Dados do Log de Erro"
// @Success     201 {object} mdl.ErrorLogs
// @Failure     400 {object} ErrorResponse
// @Failure     401 {object} ErrorResponse
// @Failure     500 {object} ErrorResponse
// @Router      /api/v1/errorlogs [post]
func (pc *ErrorLogController) CreateErrorLog(c *gin.Context) {
	var request mdl.ErrorLogs
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	err := pc.errorlogService.Create(c.Request.Context(), &request)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, request)
}

// UpdateErrorLog atualiza um log de erro.
//
// @Summary     Atualizar log de erro
// @Description Atualiza os dados de um log de erro.
// @Tags        errorlogs
// @Security    BearerAuth
// @Accept      json
// @Produce     json
// @Param       id      path string            true "ID do Log de Erro"
// @Param       payload body mdl.ErrorLogs true "Dados atualizados"
// @Success     200 {object} mdl.ErrorLogs
// @Failure     400 {object} ErrorResponse
// @Failure     401 {object} ErrorResponse
// @Failure     404 {object} ErrorResponse
// @Failure     500 {object} ErrorResponse
// @Router      /api/v1/errorlogs/{id} [put]
func (pc *ErrorLogController) UpdateErrorLog(c *gin.Context) {
	var request mdl.ErrorLogs
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	err := pc.errorlogService.Update(c.Request.Context(), &request)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, request)
}

// DeleteErrorLog remove um log de erro.
//
// @Summary     Remover log de erro
// @Description Remove um log de erro.
// @Tags        errorlogs
// @Security    BearerAuth
// @Produce     json
// @Param       id path string true "ID do Log de Erro"
// @Success     204
// @Failure     401 {object} ErrorResponse
// @Failure     404 {object} ErrorResponse
// @Failure     500 {object} ErrorResponse
// @Router      /api/v1/errorlogs/{id} [delete]
func (pc *ErrorLogController) DeleteErrorLog(c *gin.Context) {
	id := c.Param("id")
	if err := pc.errorlogService.Delete(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

// Package permissions provides the controller for managing permissions.
package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	mdl "github.com/kubex-ecosystem/gnyx/internal/dsclient"
	svc "github.com/kubex-ecosystem/gnyx/internal/dsclient/adapters"
	t "github.com/kubex-ecosystem/gnyx/internal/types"
)

type PermissionController struct {
	permissionService svc.Service[mdl.Permissions]
	APIWrapper        *t.APIWrapper[mdl.Permissions]
}

func NewPermissionController(bridge svc.Service[mdl.Permissions]) *PermissionController {
	// repo := svc.Repository{}
	// service := svc.Service{}
	return &PermissionController{
		permissionService: nil,
		APIWrapper:        t.NewAPIWrapper[mdl.Permissions](),
	}
}

// GetAllPermissions retorna todas as permissões.
//
// @Summary     Listar permissões
// @Description Recupera a lista de permissões.
// @Tags        permissions
// @Security    BearerAuth
// @Produce     json
// @Success     200 {array} mdl.Permissions
// @Failure     401 {object} ErrorResponse
// @Failure     500 {object} ErrorResponse
// @Router      /api/v1/permissions [get]
func (pc *PermissionController) GetAllPermissions(c *gin.Context) {
	items, err := pc.permissionService.List(c.Request.Context(), nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, items)
}

// GetPermissionByID retorna uma permissão pelo ID.
//
// @Summary     Obter permissão
// @Description Busca uma permissão específica pelo ID.
// @Tags        permissions
// @Security    BearerAuth
// @Produce     json
// @Param       id path string true "ID da Permissão"
// @Success     200 {object} mdl.Permissions
// @Failure     401 {object} ErrorResponse
// @Failure     404 {object} ErrorResponse
// @Router      /api/v1/permissions/{id} [get]
func (pc *PermissionController) GetPermissionByID(c *gin.Context) {
	id := c.Param("id")
	item, err := pc.permissionService.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "permission not found"})
		return
	}
	c.JSON(http.StatusOK, item)
}

// CreatePermission cria uma nova permissão.
//
// @Summary     Criar permissão
// @Description Adiciona uma nova permissão.
// @Tags        permissions
// @Security    BearerAuth
// @Accept      json
// @Produce     json
// @Param       payload body mdl.Permissions true "Dados da Permissão"
// @Success     201 {object} mdl.Permissions
// @Failure     400 {object} ErrorResponse
// @Failure     401 {object} ErrorResponse
// @Failure     500 {object} ErrorResponse
// @Router      /api/v1/permissions [post]
func (pc *PermissionController) CreatePermission(c *gin.Context) {
	var request mdl.Permissions
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	err := pc.permissionService.Create(c.Request.Context(), &request)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, request)
}

// UpdatePermission atualiza uma permissão.
//
// @Summary     Atualizar permissão
// @Description Atualiza os dados de uma permissão.
// @Tags        permissions
// @Security    BearerAuth
// @Accept      json
// @Produce     json
// @Param       id      path string            true "ID da Permissão"
// @Param       payload body mdl.Permissions true "Dados atualizados"
// @Success     200 {object} mdl.Permissions
// @Failure     400 {object} ErrorResponse
// @Failure     401 {object} ErrorResponse
// @Failure     404 {object} ErrorResponse
// @Failure     500 {object} ErrorResponse
// @Router      /api/v1/permissions/{id} [put]
func (pc *PermissionController) UpdatePermission(c *gin.Context) {
	var request mdl.Permissions
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	err := pc.permissionService.Update(c.Request.Context(), &request)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, request)
}

// DeletePermission remove uma permissão.
//
// @Summary     Remover permissão
// @Description Remove uma permissão.
// @Tags        permissions
// @Security    BearerAuth
// @Produce     json
// @Param       id path string true "ID da Permissão"
// @Success     204
// @Failure     401 {object} ErrorResponse
// @Failure     404 {object} ErrorResponse
// @Failure     500 {object} ErrorResponse
// @Router      /api/v1/permissions/{id} [delete]
func (pc *PermissionController) DeletePermission(c *gin.Context) {
	id := c.Param("id")
	if err := pc.permissionService.Delete(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

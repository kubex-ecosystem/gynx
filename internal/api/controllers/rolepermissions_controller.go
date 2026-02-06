// Package rolepermissions provides the controller for managing role permissions.
package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	mdl "github.com/kubex-ecosystem/gnyx/internal/dsclient"
	svc "github.com/kubex-ecosystem/gnyx/internal/dsclient/adapters"
	t "github.com/kubex-ecosystem/gnyx/internal/types"
)

type RolePermissionController struct {
	rolepermissionService svc.Service[mdl.RolePermissions]
	APIWrapper            *t.APIWrapper[mdl.RolePermissions]
}

func NewRolePermissionController(bridge svc.Service[mdl.RolePermissions]) *RolePermissionController {
	// repo := svc.Repository{}
	// service := svc.Service{}
	return &RolePermissionController{
		rolepermissionService: nil,
		APIWrapper:            t.NewAPIWrapper[mdl.RolePermissions](),
	}
}

// GetAllRolePermissions retorna todas as permissões de role.
//
// @Summary     Listar permissões de role
// @Description Recupera a lista de permissões de role.
// @Tags        rolepermissions
// @Security    BearerAuth
// @Produce     json
// @Success     200 {array} mdl.RolePermissions
// @Failure     401 {object} ErrorResponse
// @Failure     500 {object} ErrorResponse
// @Router      /api/v1/rolepermissions [get]
func (pc *RolePermissionController) GetAllRolePermissions(c *gin.Context) {
	items, err := pc.rolepermissionService.List(c.Request.Context(), nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, items)
}

// GetRolePermissionByID retorna uma permissão de role pelo ID.
//
// @Summary     Obter permissão de role
// @Description Busca uma permissão de role específica pelo ID.
// @Tags        rolepermissions
// @Security    BearerAuth
// @Produce     json
// @Param       id path string true "ID da Permissão de Role"
// @Success     200 {object} mdl.RolePermissions
// @Failure     401 {object} ErrorResponse
// @Failure     404 {object} ErrorResponse
// @Router      /api/v1/rolepermissions/{id} [get]
func (pc *RolePermissionController) GetRolePermissionByID(c *gin.Context) {
	id := c.Param("id")
	item, err := pc.rolepermissionService.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "role permission not found"})
		return
	}
	c.JSON(http.StatusOK, item)
}

// CreateRolePermission cria uma nova permissão de role.
//
// @Summary     Criar permissão de role
// @Description Adiciona uma nova permissão de role.
// @Tags        rolepermissions
// @Security    BearerAuth
// @Accept      json
// @Produce     json
// @Param       payload body mdl.RolePermissions true "Dados da Permissão de Role"
// @Success     201 {object} mdl.RolePermissions
// @Failure     400 {object} ErrorResponse
// @Failure     401 {object} ErrorResponse
// @Failure     500 {object} ErrorResponse
// @Router      /api/v1/rolepermissions [post]
func (pc *RolePermissionController) CreateRolePermission(c *gin.Context) {
	var request mdl.RolePermissions
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	err := pc.rolepermissionService.Create(c.Request.Context(), &request)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, request)
}

// UpdateRolePermission atualiza uma permissão de role.
//
// @Summary     Atualizar permissão de role
// @Description Atualiza os dados de uma permissão de role.
// @Tags        rolepermissions
// @Security    BearerAuth
// @Accept      json
// @Produce     json
// @Param       id      path string            true "ID da Permissão de Role"
// @Param       payload body mdl.RolePermissions true "Dados atualizados"
// @Success     200 {object} mdl.RolePermissions
// @Failure     400 {object} ErrorResponse
// @Failure     401 {object} ErrorResponse
// @Failure     404 {object} ErrorResponse
// @Failure     500 {object} ErrorResponse
// @Router      /api/v1/rolepermissions/{id} [put]
func (pc *RolePermissionController) UpdateRolePermission(c *gin.Context) {
	var request mdl.RolePermissions
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	err := pc.rolepermissionService.Update(c.Request.Context(), &request)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, request)
}

// DeleteRolePermission remove uma permissão de role.
//
// @Summary     Remover permissão de role
// @Description Remove uma permissão de role.
// @Tags        rolepermissions
// @Security    BearerAuth
// @Produce     json
// @Param       id path string true "ID da Permissão de Role"
// @Success     204
// @Failure     401 {object} ErrorResponse
// @Failure     404 {object} ErrorResponse
// @Failure     500 {object} ErrorResponse
// @Router      /api/v1/rolepermissions/{id} [delete]
func (pc *RolePermissionController) DeleteRolePermission(c *gin.Context) {
	id := c.Param("id")
	if err := pc.rolepermissionService.Delete(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

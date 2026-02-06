// Package roles provides the controller for managing roles.
package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	mdl "github.com/kubex-ecosystem/gnyx/internal/dsclient"
	svc "github.com/kubex-ecosystem/gnyx/internal/dsclient/adapters"
	t "github.com/kubex-ecosystem/gnyx/internal/types"
)

type RoleController struct {
	roleService svc.Service[mdl.Roles]
	APIWrapper  *t.APIWrapper[mdl.Roles]
}

func NewRoleController(bridge svc.Service[mdl.Roles]) *RoleController {
	// repo := svc.Repository{}
	// service := svc.Service{}
	return &RoleController{
		roleService: nil,
		APIWrapper:  t.NewAPIWrapper[mdl.Roles](),
	}
}

// GetAllRoles retorna todos os roles.
//
// @Summary     Listar roles
// @Description Recupera a lista de roles.
// @Tags        roles
// @Security    BearerAuth
// @Produce     json
// @Success     200 {array} mdl.Roles
// @Failure     401 {object} ErrorResponse
// @Failure     500 {object} ErrorResponse
// @Router      /api/v1/roles [get]
func (pc *RoleController) GetAllRoles(c *gin.Context) {
	items, err := pc.roleService.List(c.Request.Context(), nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, items)
}

// GetRoleByID retorna um role pelo ID.
//
// @Summary     Obter role
// @Description Busca um role específico pelo ID.
// @Tags        roles
// @Security    BearerAuth
// @Produce     json
// @Param       id path string true "ID do Role"
// @Success     200 {object} mdl.Roles
// @Failure     401 {object} ErrorResponse
// @Failure     404 {object} ErrorResponse
// @Router      /api/v1/roles/{id} [get]
func (pc *RoleController) GetRoleByID(c *gin.Context) {
	id := c.Param("id")
	item, err := pc.roleService.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "role not found"})
		return
	}
	c.JSON(http.StatusOK, item)
}

// CreateRole cria um novo role.
//
// @Summary     Criar role
// @Description Adiciona um novo role.
// @Tags        roles
// @Security    BearerAuth
// @Accept      json
// @Produce     json
// @Param       payload body mdl.Roles true "Dados do Role"
// @Success     201 {object} mdl.Roles
// @Failure     400 {object} ErrorResponse
// @Failure     401 {object} ErrorResponse
// @Failure     500 {object} ErrorResponse
// @Router      /api/v1/roles [post]
func (pc *RoleController) CreateRole(c *gin.Context) {
	var request mdl.Roles
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	err := pc.roleService.Create(c.Request.Context(), &request)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, request)
}

// UpdateRole atualiza um role.
//
// @Summary     Atualizar role
// @Description Atualiza os dados de um role.
// @Tags        roles
// @Security    BearerAuth
// @Accept      json
// @Produce     json
// @Param       id      path string            true "ID do Role"
// @Param       payload body mdl.Roles true "Dados atualizados"
// @Success     200 {object} mdl.Roles
// @Failure     400 {object} ErrorResponse
// @Failure     401 {object} ErrorResponse
// @Failure     404 {object} ErrorResponse
// @Failure     500 {object} ErrorResponse
// @Router      /api/v1/roles/{id} [put]
func (pc *RoleController) UpdateRole(c *gin.Context) {
	var request mdl.Roles
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	err := pc.roleService.Update(c.Request.Context(), &request)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, request)
}

// DeleteRole remove um role.
//
// @Summary     Remover role
// @Description Remove um role.
// @Tags        roles
// @Security    BearerAuth
// @Produce     json
// @Param       id path string true "ID do Role"
// @Success     204
// @Failure     401 {object} ErrorResponse
// @Failure     404 {object} ErrorResponse
// @Failure     500 {object} ErrorResponse
// @Router      /api/v1/roles/{id} [delete]
func (pc *RoleController) DeleteRole(c *gin.Context) {
	id := c.Param("id")
	if err := pc.roleService.Delete(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

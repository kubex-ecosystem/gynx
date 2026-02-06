// Package roleconfig provides the controller for managing role configurations.
package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	mdl "github.com/kubex-ecosystem/gnyx/internal/dsclient"
	svc "github.com/kubex-ecosystem/gnyx/internal/dsclient/adapters"
	t "github.com/kubex-ecosystem/gnyx/internal/types"
)

type RoleConfigController struct {
	roleconfigService svc.Service[mdl.RoleConfig]
	APIWrapper        *t.APIWrapper[mdl.RoleConfig]
}

func NewRoleConfigController(bridge svc.Service[mdl.RoleConfig]) *RoleConfigController {
	// repo := svc.Repository{}
	// service := svc.Service{}
	return &RoleConfigController{
		roleconfigService: nil,
		APIWrapper:        t.NewAPIWrapper[mdl.RoleConfig](),
	}
}

// GetAllRoleConfigs retorna todas as configurações de role.
//
// @Summary     Listar configurações de role
// @Description Recupera a lista de configurações de role.
// @Tags        roleconfig
// @Security    BearerAuth
// @Produce     json
// @Success     200 {array} mdl.RoleConfig
// @Failure     401 {object} ErrorResponse
// @Failure     500 {object} ErrorResponse
// @Router      /api/v1/roleconfigs [get]
func (pc *RoleConfigController) GetAllRoleConfigs(c *gin.Context) {
	items, err := pc.roleconfigService.List(c.Request.Context(), nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, items)
}

// GetRoleConfigByID retorna uma configuração de role pelo ID.
//
// @Summary     Obter configuração de role
// @Description Busca uma configuração de role específica pelo ID.
// @Tags        roleconfig
// @Security    BearerAuth
// @Produce     json
// @Param       id path string true "ID da Configuração de Role"
// @Success     200 {object} mdl.RoleConfig
// @Failure     401 {object} ErrorResponse
// @Failure     404 {object} ErrorResponse
// @Router      /api/v1/roleconfigs/{id} [get]
func (pc *RoleConfigController) GetRoleConfigByID(c *gin.Context) {
	id := c.Param("id")
	item, err := pc.roleconfigService.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "role config not found"})
		return
	}
	c.JSON(http.StatusOK, item)
}

// CreateRoleConfig cria uma nova configuração de role.
//
// @Summary     Criar configuração de role
// @Description Adiciona uma nova configuração de role.
// @Tags        roleconfig
// @Security    BearerAuth
// @Accept      json
// @Produce     json
// @Param       payload body mdl.RoleConfig true "Dados da Configuração de Role"
// @Success     201 {object} mdl.RoleConfig
// @Failure     400 {object} ErrorResponse
// @Failure     401 {object} ErrorResponse
// @Failure     500 {object} ErrorResponse
// @Router      /api/v1/roleconfigs [post]
func (pc *RoleConfigController) CreateRoleConfig(c *gin.Context) {
	var request mdl.RoleConfig
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	err := pc.roleconfigService.Create(c.Request.Context(), &request)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, request)
}

// UpdateRoleConfig atualiza uma configuração de role.
//
// @Summary     Atualizar configuração de role
// @Description Atualiza os dados de uma configuração de role.
// @Tags        roleconfig
// @Security    BearerAuth
// @Accept      json
// @Produce     json
// @Param       id      path string            true "ID da Configuração de Role"
// @Param       payload body mdl.RoleConfig true "Dados atualizados"
// @Success     200 {object} mdl.RoleConfig
// @Failure     400 {object} ErrorResponse
// @Failure     401 {object} ErrorResponse
// @Failure     404 {object} ErrorResponse
// @Failure     500 {object} ErrorResponse
// @Router      /api/v1/roleconfigs/{id} [put]
func (pc *RoleConfigController) UpdateRoleConfig(c *gin.Context) {
	var request mdl.RoleConfig
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	err := pc.roleconfigService.Update(c.Request.Context(), &request)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, request)
}

// DeleteRoleConfig remove uma configuração de role.
//
// @Summary     Remover configuração de role
// @Description Remove uma configuração de role.
// @Tags        roleconfig
// @Security    BearerAuth
// @Produce     json
// @Param       id path string true "ID da Configuração de Role"
// @Success     204
// @Failure     401 {object} ErrorResponse
// @Failure     404 {object} ErrorResponse
// @Failure     500 {object} ErrorResponse
// @Router      /api/v1/roleconfigs/{id} [delete]
func (pc *RoleConfigController) DeleteRoleConfig(c *gin.Context) {
	id := c.Param("id")
	if err := pc.roleconfigService.Delete(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

// Package tenants provides the controller for managing tenants.
package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	mdl "github.com/kubex-ecosystem/gnyx/internal/dsclient"
	svc "github.com/kubex-ecosystem/gnyx/internal/dsclient/adapters"
	t "github.com/kubex-ecosystem/gnyx/internal/types"
)

type TenantController struct {
	tenantService svc.Service[mdl.Tenants]
	APIWrapper    *t.APIWrapper[mdl.Tenants]
}

func NewTenantController(bridge svc.Service[mdl.Tenants]) *TenantController {
	// repo := svc.Repository{}
	// service := svc.Service{}
	return &TenantController{
		tenantService: nil,
		APIWrapper:    t.NewAPIWrapper[mdl.Tenants](),
	}
}

// GetAllTenants retorna todos os tenants.
//
// @Summary     Listar tenants
// @Description Recupera a lista de tenants.
// @Tags        tenants
// @Security    BearerAuth
// @Produce     json
// @Success     200 {array} mdl.Tenants
// @Failure     401 {object} ErrorResponse
// @Failure     500 {object} ErrorResponse
// @Router      /api/v1/tenants [get]
func (pc *TenantController) GetAllTenants(c *gin.Context) {
	items, err := pc.tenantService.List(c.Request.Context(), nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, items)
}

// GetTenantByID retorna um tenant pelo ID.
//
// @Summary     Obter tenant
// @Description Busca um tenant específico pelo ID.
// @Tags        tenants
// @Security    BearerAuth
// @Produce     json
// @Param       id path string true "ID do Tenant"
// @Success     200 {object} mdl.Tenants
// @Failure     401 {object} ErrorResponse
// @Failure     404 {object} ErrorResponse
// @Router      /api/v1/tenants/{id} [get]
func (pc *TenantController) GetTenantByID(c *gin.Context) {
	id := c.Param("id")
	item, err := pc.tenantService.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "tenant not found"})
		return
	}
	c.JSON(http.StatusOK, item)
}

// CreateTenant cria um novo tenant.
//
// @Summary     Criar tenant
// @Description Adiciona um novo tenant.
// @Tags        tenants
// @Security    BearerAuth
// @Accept      json
// @Produce     json
// @Param       payload body mdl.Tenants true "Dados do Tenant"
// @Success     201 {object} mdl.Tenants
// @Failure     400 {object} ErrorResponse
// @Failure     401 {object} ErrorResponse
// @Failure     500 {object} ErrorResponse
// @Router      /api/v1/tenants [post]
func (pc *TenantController) CreateTenant(c *gin.Context) {
	var request mdl.Tenants
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	err := pc.tenantService.Create(c.Request.Context(), &request)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, request)
}

// UpdateTenant atualiza um tenant.
//
// @Summary     Atualizar tenant
// @Description Atualiza os dados de um tenant.
// @Tags        tenants
// @Security    BearerAuth
// @Accept      json
// @Produce     json
// @Param       id      path string            true "ID do Tenant"
// @Param       payload body mdl.Tenants true "Dados atualizados"
// @Success     200 {object} mdl.Tenants
// @Failure     400 {object} ErrorResponse
// @Failure     401 {object} ErrorResponse
// @Failure     404 {object} ErrorResponse
// @Failure     500 {object} ErrorResponse
// @Router      /api/v1/tenants/{id} [put]
func (pc *TenantController) UpdateTenant(c *gin.Context) {
	var request mdl.Tenants
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	err := pc.tenantService.Update(c.Request.Context(), &request)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, request)
}

// DeleteTenant remove um tenant.
//
// @Summary     Remover tenant
// @Description Remove um tenant.
// @Tags        tenants
// @Security    BearerAuth
// @Produce     json
// @Param       id path string true "ID do Tenant"
// @Success     204
// @Failure     401 {object} ErrorResponse
// @Failure     404 {object} ErrorResponse
// @Failure     500 {object} ErrorResponse
// @Router      /api/v1/tenants/{id} [delete]
func (pc *TenantController) DeleteTenant(c *gin.Context) {
	id := c.Param("id")
	if err := pc.tenantService.Delete(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

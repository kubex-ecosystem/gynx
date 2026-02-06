// Package tenantsubscriptions provides the controller for managing tenant subscriptions.
package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	mdl "github.com/kubex-ecosystem/gnyx/internal/dsclient"
	svc "github.com/kubex-ecosystem/gnyx/internal/dsclient/adapters"
	t "github.com/kubex-ecosystem/gnyx/internal/types"
)

type TenantSubscriptionController struct {
	tenantsubscriptionService svc.Service[mdl.TenantSubscriptions]
	APIWrapper                *t.APIWrapper[mdl.TenantSubscriptions]
}

func NewTenantSubscriptionController(bridge svc.Service[mdl.TenantSubscriptions]) *TenantSubscriptionController {
	// repo := svc.Repository{}
	// service := svc.Service{}
	return &TenantSubscriptionController{
		tenantsubscriptionService: nil,
		APIWrapper:                t.NewAPIWrapper[mdl.TenantSubscriptions](),
	}
}

// GetAllTenantSubscriptions retorna todas as assinaturas de tenant.
//
// @Summary     Listar assinaturas de tenant
// @Description Recupera a lista de assinaturas de tenant.
// @Tags        tenantsubscriptions
// @Security    BearerAuth
// @Produce     json
// @Success     200 {array} mdl.TenantSubscriptions
// @Failure     401 {object} ErrorResponse
// @Failure     500 {object} ErrorResponse
// @Router      /api/v1/tenantsubscriptions [get]
func (pc *TenantSubscriptionController) GetAllTenantSubscriptions(c *gin.Context) {
	items, err := pc.tenantsubscriptionService.List(c.Request.Context(), nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, items)
}

// GetTenantSubscriptionByID retorna uma assinatura de tenant pelo ID.
//
// @Summary     Obter assinatura de tenant
// @Description Busca uma assinatura de tenant específica pelo ID.
// @Tags        tenantsubscriptions
// @Security    BearerAuth
// @Produce     json
// @Param       id path string true "ID da Assinatura de Tenant"
// @Success     200 {object} mdl.TenantSubscriptions
// @Failure     401 {object} ErrorResponse
// @Failure     404 {object} ErrorResponse
// @Router      /api/v1/tenantsubscriptions/{id} [get]
func (pc *TenantSubscriptionController) GetTenantSubscriptionByID(c *gin.Context) {
	id := c.Param("id")
	item, err := pc.tenantsubscriptionService.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "tenant subscription not found"})
		return
	}
	c.JSON(http.StatusOK, item)
}

// CreateTenantSubscription cria uma nova assinatura de tenant.
//
// @Summary     Criar assinatura de tenant
// @Description Adiciona uma nova assinatura de tenant.
// @Tags        tenantsubscriptions
// @Security    BearerAuth
// @Accept      json
// @Produce     json
// @Param       payload body mdl.TenantSubscriptions true "Dados da Assinatura de Tenant"
// @Success     201 {object} mdl.TenantSubscriptions
// @Failure     400 {object} ErrorResponse
// @Failure     401 {object} ErrorResponse
// @Failure     500 {object} ErrorResponse
// @Router      /api/v1/tenantsubscriptions [post]
func (pc *TenantSubscriptionController) CreateTenantSubscription(c *gin.Context) {
	var request mdl.TenantSubscriptions
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	err := pc.tenantsubscriptionService.Create(c.Request.Context(), &request)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, request)
}

// UpdateTenantSubscription atualiza uma assinatura de tenant.
//
// @Summary     Atualizar assinatura de tenant
// @Description Atualiza os dados de uma assinatura de tenant.
// @Tags        tenantsubscriptions
// @Security    BearerAuth
// @Accept      json
// @Produce     json
// @Param       id      path string            true "ID da Assinatura de Tenant"
// @Param       payload body mdl.TenantSubscriptions true "Dados atualizados"
// @Success     200 {object} mdl.TenantSubscriptions
// @Failure     400 {object} ErrorResponse
// @Failure     401 {object} ErrorResponse
// @Failure     404 {object} ErrorResponse
// @Failure     500 {object} ErrorResponse
// @Router      /api/v1/tenantsubscriptions/{id} [put]
func (pc *TenantSubscriptionController) UpdateTenantSubscription(c *gin.Context) {
	var request mdl.TenantSubscriptions
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	err := pc.tenantsubscriptionService.Update(c.Request.Context(), &request)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, request)
}

// DeleteTenantSubscription remove uma assinatura de tenant.
//
// @Summary     Remover assinatura de tenant
// @Description Remove uma assinatura de tenant.
// @Tags        tenantsubscriptions
// @Security    BearerAuth
// @Produce     json
// @Param       id path string true "ID da Assinatura de Tenant"
// @Success     204
// @Failure     401 {object} ErrorResponse
// @Failure     404 {object} ErrorResponse
// @Failure     500 {object} ErrorResponse
// @Router      /api/v1/tenantsubscriptions/{id} [delete]
func (pc *TenantSubscriptionController) DeleteTenantSubscription(c *gin.Context) {
	id := c.Param("id")
	if err := pc.tenantsubscriptionService.Delete(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

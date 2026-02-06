// Package subscriptionplans provides the controller for managing subscription plans.
package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	mdl "github.com/kubex-ecosystem/gnyx/internal/dsclient"
	svc "github.com/kubex-ecosystem/gnyx/internal/dsclient/adapters"
	t "github.com/kubex-ecosystem/gnyx/internal/types"
)

type SubscriptionPlanController struct {
	subscriptionplanService svc.Service[mdl.SubscriptionPlans]
	APIWrapper              *t.APIWrapper[mdl.SubscriptionPlans]
}

type (
	// ErrorResponse padroniza a documentação de erros dos endpoints.
	ErrorResponse = t.ErrorResponse
)

func NewSubscriptionPlanController(bridge svc.Service[mdl.SubscriptionPlans]) *SubscriptionPlanController {
	// repo := svc.Repository{}
	// service := svc.Service{}
	return &SubscriptionPlanController{
		subscriptionplanService: nil,
		APIWrapper:              t.NewAPIWrapper[mdl.SubscriptionPlans](),
	}
}

// GetAllSubscriptionPlans retorna todos os planos de assinatura.
//
// @Summary     Listar planos de assinatura
// @Description Recupera a lista de planos de assinatura.
// @Tags        subscriptionplans
// @Security    BearerAuth
// @Produce     json
// @Success     200 {array} mdl.SubscriptionPlans
// @Failure     401 {object} ErrorResponse
// @Failure     500 {object} ErrorResponse
// @Router      /api/v1/subscriptionplans [get]
func (pc *SubscriptionPlanController) GetAllSubscriptionPlans(c *gin.Context) {
	items, err := pc.subscriptionplanService.List(c.Request.Context(), nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, items)
}

// GetSubscriptionPlanByID retorna um plano de assinatura pelo ID.
//
// @Summary     Obter plano de assinatura
// @Description Busca um plano de assinatura específico pelo ID.
// @Tags        subscriptionplans
// @Security    BearerAuth
// @Produce     json
// @Param       id path string true "ID do Plano de Assinatura"
// @Success     200 {object} mdl.SubscriptionPlans
// @Failure     401 {object} ErrorResponse
// @Failure     404 {object} ErrorResponse
// @Router      /api/v1/subscriptionplans/{id} [get]
func (pc *SubscriptionPlanController) GetSubscriptionPlanByID(c *gin.Context) {
	id := c.Param("id")
	item, err := pc.subscriptionplanService.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "subscription plan not found"})
		return
	}
	c.JSON(http.StatusOK, item)
}

// CreateSubscriptionPlan cria um novo plano de assinatura.
//
// @Summary     Criar plano de assinatura
// @Description Adiciona um novo plano de assinatura.
// @Tags        subscriptionplans
// @Security    BearerAuth
// @Accept      json
// @Produce     json
// @Param       payload body mdl.SubscriptionPlans true "Dados do Plano de Assinatura"
// @Success     201 {object} mdl.SubscriptionPlans
// @Failure     400 {object} ErrorResponse
// @Failure     401 {object} ErrorResponse
// @Failure     500 {object} ErrorResponse
// @Router      /api/v1/subscriptionplans [post]
func (pc *SubscriptionPlanController) CreateSubscriptionPlan(c *gin.Context) {
	var request mdl.SubscriptionPlans
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	err := pc.subscriptionplanService.Create(c.Request.Context(), &request)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, request)
}

// UpdateSubscriptionPlan atualiza um plano de assinatura.
//
// @Summary     Atualizar plano de assinatura
// @Description Atualiza os dados de um plano de assinatura.
// @Tags        subscriptionplans
// @Security    BearerAuth
// @Accept      json
// @Produce     json
// @Param       id      path string            true "ID do Plano de Assinatura"
// @Param       payload body mdl.SubscriptionPlans true "Dados atualizados"
// @Success     200 {object} mdl.SubscriptionPlans
// @Failure     400 {object} ErrorResponse
// @Failure     401 {object} ErrorResponse
// @Failure     404 {object} ErrorResponse
// @Failure     500 {object} ErrorResponse
// @Router      /api/v1/subscriptionplans/{id} [put]
func (pc *SubscriptionPlanController) UpdateSubscriptionPlan(c *gin.Context) {
	var request mdl.SubscriptionPlans
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	err := pc.subscriptionplanService.Update(c.Request.Context(), &request)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, request)
}

// DeleteSubscriptionPlan remove um plano de assinatura.
//
// @Summary     Remover plano de assinatura
// @Description Remove um plano de assinatura.
// @Tags        subscriptionplans
// @Security    BearerAuth
// @Produce     json
// @Param       id path string true "ID do Plano de Assinatura"
// @Success     204
// @Failure     401 {object} ErrorResponse
// @Failure     404 {object} ErrorResponse
// @Failure     500 {object} ErrorResponse
// @Router      /api/v1/subscriptionplans/{id} [delete]
func (pc *SubscriptionPlanController) DeleteSubscriptionPlan(c *gin.Context) {
	id := c.Param("id")
	if err := pc.subscriptionplanService.Delete(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

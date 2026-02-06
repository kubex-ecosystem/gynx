// Package paymentmethods provides the controller for managing payment methods.
package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	mdl "github.com/kubex-ecosystem/gnyx/internal/dsclient"
	svc "github.com/kubex-ecosystem/gnyx/internal/dsclient/adapters"
	t "github.com/kubex-ecosystem/gnyx/internal/types"
)

type PaymentMethodController struct {
	paymentmethodService svc.Service[mdl.PaymentMethods]
	APIWrapper           *t.APIWrapper[mdl.PaymentMethods]
}

func NewPaymentMethodController(bridge svc.Service[mdl.PaymentMethods]) *PaymentMethodController {
	// repo := svc.Repository{}
	// service := svc.Service{}
	return &PaymentMethodController{
		// paymentmethodService: nil,
		// APIWrapper:           t.NewAPIWrapper[mdl.PaymentMethod](),
	}
}

// GetAllPaymentMethods retorna todos os métodos de pagamento.
//
// @Summary     Listar métodos de pagamento
// @Description Recupera a lista de métodos de pagamento.
// @Tags        paymentmethods
// @Security    BearerAuth
// @Produce     json
// @Success     200 {array} mdl.PaymentMethods
// @Failure     401 {object} ErrorResponse
// @Failure     500 {object} ErrorResponse
// @Router      /api/v1/paymentmethods [get]
func (pc *PaymentMethodController) GetAllPaymentMethods(c *gin.Context) {
	items, err := pc.paymentmethodService.List(c.Request.Context(), nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, items)
}

// GetPaymentMethodByID retorna um método de pagamento pelo ID.
//
// @Summary     Obter método de pagamento
// @Description Busca um método de pagamento específico pelo ID.
// @Tags        paymentmethods
// @Security    BearerAuth
// @Produce     json
// @Param       id path string true "ID do Método de Pagamento"
// @Success     200 {object} mdl.PaymentMethods
// @Failure     401 {object} ErrorResponse
// @Failure     404 {object} ErrorResponse
// @Router      /api/v1/paymentmethods/{id} [get]
func (pc *PaymentMethodController) GetPaymentMethodByID(c *gin.Context) {
	id := c.Param("id")
	item, err := pc.paymentmethodService.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "payment method not found"})
		return
	}
	c.JSON(http.StatusOK, item)
}

// CreatePaymentMethod cria um novo método de pagamento.
//
// @Summary     Criar método de pagamento
// @Description Adiciona um novo método de pagamento.
// @Tags        paymentmethods
// @Security    BearerAuth
// @Accept      json
// @Produce     json
// @Param       payload body mdl.PaymentMethods true "Dados do Método de Pagamento"
// @Success     201 {object} mdl.PaymentMethods
// @Failure     400 {object} ErrorResponse
// @Failure     401 {object} ErrorResponse
// @Failure     500 {object} ErrorResponse
// @Router      /api/v1/paymentmethods [post]
func (pc *PaymentMethodController) CreatePaymentMethod(c *gin.Context) {
	var request mdl.PaymentMethods
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	err := pc.paymentmethodService.Create(c.Request.Context(), &request)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, request)
}

// UpdatePaymentMethod atualiza um método de pagamento.
//
// @Summary     Atualizar método de pagamento
// @Description Atualiza os dados de um método de pagamento.
// @Tags        paymentmethods
// @Security    BearerAuth
// @Accept      json
// @Produce     json
// @Param       id      path string            true "ID do Método de Pagamento"
// @Param       payload body mdl.PaymentMethods true "Dados atualizados"
// @Success     200 {object} mdl.PaymentMethods
// @Failure     400 {object} ErrorResponse
// @Failure     401 {object} ErrorResponse
// @Failure     404 {object} ErrorResponse
// @Failure     500 {object} ErrorResponse
// @Router      /api/v1/paymentmethods/{id} [put]
func (pc *PaymentMethodController) UpdatePaymentMethod(c *gin.Context) {
	var request mdl.PaymentMethods
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	err := pc.paymentmethodService.Update(c.Request.Context(), &request)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, request)
}

// DeletePaymentMethod remove um método de pagamento.
//
// @Summary     Remover método de pagamento
// @Description Remove um método de pagamento.
// @Tags        paymentmethods
// @Security    BearerAuth
// @Produce     json
// @Param       id path string true "ID do Método de Pagamento"
// @Success     204
// @Failure     401 {object} ErrorResponse
// @Failure     404 {object} ErrorResponse
// @Failure     500 {object} ErrorResponse
// @Router      /api/v1/paymentmethods/{id} [delete]
func (pc *PaymentMethodController) DeletePaymentMethod(c *gin.Context) {
	id := c.Param("id")
	if err := pc.paymentmethodService.Delete(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

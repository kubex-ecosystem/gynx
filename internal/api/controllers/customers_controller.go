// Package customers provides the controller for managing customers.
package controllers

import (
	svc "github.com/kubex-ecosystem/gnyx/internal/dsclient/adapters"
	// mdl "github.com/kubex-ecosystem/gnyx/internal/dsclient"
	t "github.com/kubex-ecosystem/gnyx/internal/types"
)

type CustomerController struct {
	customerService svc.Service[any]
	APIWrapper      *t.APIWrapper[any]
}

// type (
// 	// ErrorResponse padroniza a documentação de erros dos endpoints.
// 	ErrorResponse = t.ErrorResponse
// )

// func NewCustomerController(bridge svc.Service[mdl.Customers]) *CustomerController {
// 	// repo := svc.Repository{}
// 	// service := svc.Service{}
// 	return &CustomerController{
// 		customerService: nil,
// 		APIWrapper:      t.NewAPIWrapper[mdl.Customer](),
// 	}
// }

// // GetAllCustomers retorna todos os clientes.
// //
// // @Summary     Listar clientes
// // @Description Recupera a lista de clientes.
// // @Tags        customers
// // @Security    BearerAuth
// // @Produce     json
// // @Success     200 {array} mdl.Customers
// // @Failure     401 {object} ErrorResponse
// // @Failure     500 {object} ErrorResponse
// // @Router      /api/v1/customers [get]
// func (pc *CustomerController) GetAllCustomers(c *gin.Context) {
// 	items, err := pc.customerService.List()
// 	if err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
// 		return
// 	}
// 	c.JSON(http.StatusOK, items)
// }

// // GetCustomerByID retorna um cliente pelo ID.
// //
// // @Summary     Obter cliente
// // @Description Busca um cliente específico pelo ID.
// // @Tags        customers
// // @Security    BearerAuth
// // @Produce     json
// // @Param       id path string true "ID do Cliente"
// // @Success     200 {object} mdl.Customers
// // @Failure     401 {object} ErrorResponse
// // @Failure     404 {object} ErrorResponse
// // @Router      /api/v1/customers/{id} [get]
// func (pc *CustomerController) GetCustomerByID(c *gin.Context) {
// 	id := c.Param("id")
// 	item, err := pc.customerService.GetByID(id)
// 	if err != nil {
// 		c.JSON(http.StatusNotFound, gin.H{"error": "customer not found"})
// 		return
// 	}
// 	c.JSON(http.StatusOK, item)
// }

// // CreateCustomer cria um novo cliente.
// //
// // @Summary     Criar cliente
// // @Description Adiciona um novo cliente.
// // @Tags        customers
// // @Security    BearerAuth
// // @Accept      json
// // @Produce     json
// // @Param       payload body mdl.Customers true "Dados do Cliente"
// // @Success     201 {object} mdl.Customers
// // @Failure     400 {object} ErrorResponse
// // @Failure     401 {object} ErrorResponse
// // @Failure     500 {object} ErrorResponse
// // @Router      /api/v1/customers [post]
// func (pc *CustomerController) CreateCustomer(c *gin.Context) {
// 	var request mdl.Customers
// 	if err := c.ShouldBindJSON(&request); err != nil {
// 		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
// 		return
// 	}
// 	createdItem, err := pc.customerService.Create(&request)
// 	if err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
// 		return
// 	}
// 	c.JSON(http.StatusCreated, createdItem)
// }

// // UpdateCustomer atualiza um cliente.
// //
// // @Summary     Atualizar cliente
// // @Description Atualiza os dados de um cliente.
// // @Tags        customers
// // @Security    BearerAuth
// // @Accept      json
// // @Produce     json
// // @Param       id      path string            true "ID do Cliente"
// // @Param       payload body mdl.Customers true "Dados atualizados"
// // @Success     200 {object} mdl.Customers
// // @Failure     400 {object} ErrorResponse
// // @Failure     401 {object} ErrorResponse
// // @Failure     404 {object} ErrorResponse
// // @Failure     500 {object} ErrorResponse
// // @Router      /api/v1/customers/{id} [put]
// func (pc *CustomerController) UpdateCustomer(c *gin.Context) {
// 	var request mdl.customers
// 	if err := c.ShouldBindJSON(&request); err != nil {
// 		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
// 		return
// 	}
// 	updatedItem, err := pc.customerService.Update(&request)
// 	if err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
// 		return
// 	}
// 	c.JSON(http.StatusOK, updatedItem)
// }

// // DeleteCustomer remove um cliente.
// //
// // @Summary     Remover cliente
// // @Description Remove um cliente.
// // @Tags        customers
// // @Security    BearerAuth
// // @Produce     json
// // @Param       id path string true "ID do Cliente"
// // @Success     204
// // @Failure     401 {object} ErrorResponse
// // @Failure     404 {object} ErrorResponse
// // @Failure     500 {object} ErrorResponse
// // @Router      /api/v1/customers/{id} [delete]
// func (pc *CustomerController) DeleteCustomer(c *gin.Context) {
// 	id := c.Param("id")
// 	if err := pc.customerService.Delete(id); err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
// 		return
// 	}
// 	c.Status(http.StatusNoContent)
// }

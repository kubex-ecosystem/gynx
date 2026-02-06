// Package addresses provides the controller for managing addresses.
package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	mdl "github.com/kubex-ecosystem/gnyx/internal/dsclient"
	svc "github.com/kubex-ecosystem/gnyx/internal/dsclient/adapters"
	t "github.com/kubex-ecosystem/gnyx/internal/types"
)

type AddressController struct {
	addressService svc.Service[mdl.Addresses]
	APIWrapper     *t.APIWrapper[mdl.Addresses]
}

func NewAddressController(bridge svc.Service[mdl.Addresses]) *AddressController {
	// repo := svc.Repository{}
	// service := svc.Service{}
	return &AddressController{
		addressService: nil,
		APIWrapper:     t.NewAPIWrapper[mdl.Addresses](),
	}
}

// GetAllAddresses retorna todos os endereços.
//
// @Summary     Listar endereços
// @Description Recupera a lista de endereços.
// @Tags        addresses
// @Security    BearerAuth
// @Produce     json
// @Success     200 {array} mdl.Addresses
// @Failure     401 {object} ErrorResponse
// @Failure     500 {object} ErrorResponse
// @Router      /api/v1/addresses [get]
func (pc *AddressController) GetAllAddresses(c *gin.Context) {
	items, err := pc.addressService.List(c.Request.Context(), nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, items)
}

// GetAddressByID retorna um endereço pelo ID.
//
// @Summary     Obter endereço
// @Description Busca um endereço específico pelo ID.
// @Tags        addresses
// @Security    BearerAuth
// @Produce     json
// @Param       id path string true "ID do Endereço"
// @Success     200 {object} mdl.Addresses
// @Failure     401 {object} ErrorResponse
// @Failure     404 {object} ErrorResponse
// @Router      /api/v1/addresses/{id} [get]
func (pc *AddressController) GetAddressByID(c *gin.Context) {
	id := c.Param("id")
	item, err := pc.addressService.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "address not found"})
		return
	}
	c.JSON(http.StatusOK, item)
}

// CreateAddress cria um novo endereço.
//
// @Summary     Criar endereço
// @Description Adiciona um novo endereço.
// @Tags        addresses
// @Security    BearerAuth
// @Accept      json
// @Produce     json
// @Param       payload body mdl.Addresses true "Dados do Endereço"
// @Success     201 {object} mdl.Addresses
// @Failure     400 {object} ErrorResponse
// @Failure     401 {object} ErrorResponse
// @Failure     500 {object} ErrorResponse
// @Router      /api/v1/addresses [post]
func (pc *AddressController) CreateAddress(c *gin.Context) {
	var request mdl.Addresses
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	err := pc.addressService.Create(c.Request.Context(), &request)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, request)
}

// UpdateAddress atualiza um endereço.
//
// @Summary     Atualizar endereço
// @Description Atualiza os dados de um endereço.
// @Tags        addresses
// @Security    BearerAuth
// @Accept      json
// @Produce     json
// @Param       id      path string            true "ID do Endereço"
// @Param       payload body mdl.Addresses true "Dados atualizados"
// @Success     200 {object} mdl.Addresses
// @Failure     400 {object} ErrorResponse
// @Failure     401 {object} ErrorResponse
// @Failure     404 {object} ErrorResponse
// @Failure     500 {object} ErrorResponse
// @Router      /api/v1/addresses/{id} [put]
func (pc *AddressController) UpdateAddress(c *gin.Context) {
	var request mdl.Addresses
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	err := pc.addressService.Update(c.Request.Context(), &request)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, request)
}

// DeleteAddress remove um endereço.
//
// @Summary     Remover endereço
// @Description Remove um endereço.
// @Tags        addresses
// @Security    BearerAuth
// @Produce     json
// @Param       id path string true "ID do Endereço"
// @Success     204
// @Failure     401 {object} ErrorResponse
// @Failure     404 {object} ErrorResponse
// @Failure     500 {object} ErrorResponse
// @Router      /api/v1/addresses/{id} [delete]
func (pc *AddressController) DeleteAddress(c *gin.Context) {
	id := c.Param("id")
	if err := pc.addressService.Delete(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

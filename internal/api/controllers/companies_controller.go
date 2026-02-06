// Package companies provides the controller for managing companies.
package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	mdl "github.com/kubex-ecosystem/gnyx/internal/dsclient"
	svc "github.com/kubex-ecosystem/gnyx/internal/dsclient/adapters"
	t "github.com/kubex-ecosystem/gnyx/internal/types"
)

type CompanyController struct {
	companyService svc.Service[mdl.Companies]
	APIWrapper     *t.APIWrapper[mdl.Companies]
}

func NewCompanyController(bridge svc.Service[mdl.Companies]) *CompanyController {
	// repo := svc.Repository{}
	// service := svc.Service{}
	return &CompanyController{
		companyService: nil,
		APIWrapper:     t.NewAPIWrapper[mdl.Companies](),
	}
}

// GetAllCompanies retorna todas as empresas.
//
// @Summary     Listar empresas
// @Description Recupera a lista de empresas.
// @Tags        companies
// @Security    BearerAuth
// @Produce     json
// @Success     200 {array} mdl.Companies
// @Failure     401 {object} ErrorResponse
// @Failure     500 {object} ErrorResponse
// @Router      /api/v1/companies [get]
func (pc *CompanyController) GetAllCompanies(c *gin.Context) {
	items, err := pc.companyService.List(c.Request.Context(), nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, items)
}

// GetCompanyByID retorna uma empresa pelo ID.
//
// @Summary     Obter empresa
// @Description Busca uma empresa específica pelo ID.
// @Tags        companies
// @Security    BearerAuth
// @Produce     json
// @Param       id path string true "ID da Empresa"
// @Success     200 {object} mdl.Companies
// @Failure     401 {object} ErrorResponse
// @Failure     404 {object} ErrorResponse
// @Router      /api/v1/companies/{id} [get]
func (pc *CompanyController) GetCompanyByID(c *gin.Context) {
	id := c.Param("id")
	item, err := pc.companyService.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "company not found"})
		return
	}
	c.JSON(http.StatusOK, item)
}

// CreateCompany cria uma nova empresa.
//
// @Summary     Criar empresa
// @Description Adiciona uma nova empresa.
// @Tags        companies
// @Security    BearerAuth
// @Accept      json
// @Produce     json
// @Param       payload body mdl.Companies true "Dados da Empresa"
// @Success     201 {object} mdl.Companies
// @Failure     400 {object} ErrorResponse
// @Failure     401 {object} ErrorResponse
// @Failure     500 {object} ErrorResponse
// @Router      /api/v1/companies [post]
func (pc *CompanyController) CreateCompany(c *gin.Context) {
	var request mdl.Companies
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	err := pc.companyService.Create(c.Request.Context(), &request)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, request)
}

// UpdateCompany atualiza uma empresa.
//
// @Summary     Atualizar empresa
// @Description Atualiza os dados de uma empresa.
// @Tags        companies
// @Security    BearerAuth
// @Accept      json
// @Produce     json
// @Param       id      path string            true "ID da Empresa"
// @Param       payload body mdl.Companies true "Dados atualizados"
// @Success     200 {object} mdl.Companies
// @Failure     400 {object} ErrorResponse
// @Failure     401 {object} ErrorResponse
// @Failure     404 {object} ErrorResponse
// @Failure     500 {object} ErrorResponse
// @Router      /api/v1/companies/{id} [put]
func (pc *CompanyController) UpdateCompany(c *gin.Context) {
	var request mdl.Companies
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	err := pc.companyService.Update(c.Request.Context(), &request)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, request)
}

// DeleteCompany remove uma empresa.
//
// @Summary     Remover empresa
// @Description Remove uma empresa.
// @Tags        companies
// @Security    BearerAuth
// @Produce     json
// @Param       id path string true "ID da Empresa"
// @Success     204
// @Failure     401 {object} ErrorResponse
// @Failure     404 {object} ErrorResponse
// @Failure     500 {object} ErrorResponse
// @Router      /api/v1/companies/{id} [delete]
func (pc *CompanyController) DeleteCompany(c *gin.Context) {
	id := c.Param("id")
	if err := pc.companyService.Delete(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

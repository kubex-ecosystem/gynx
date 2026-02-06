// Package profiles provides the controller for managing profiles.
package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	mdl "github.com/kubex-ecosystem/gnyx/internal/dsclient"
	svc "github.com/kubex-ecosystem/gnyx/internal/dsclient/adapters"
	t "github.com/kubex-ecosystem/gnyx/internal/types"
)

type ProfileController struct {
	profileService svc.Service[mdl.Profiles]
	APIWrapper     *t.APIWrapper[mdl.Profiles]
}

func NewProfileController(bridge svc.Service[mdl.Profiles]) *ProfileController {
	// repo := svc.Repository{}
	// service := svc.Service{}
	return &ProfileController{
		profileService: nil,
		APIWrapper:     t.NewAPIWrapper[mdl.Profiles](),
	}
}

// GetAllProfiles retorna todos os perfis.
//
// @Summary     Listar perfis
// @Description Recupera a lista de perfis.
// @Tags        profiles
// @Security    BearerAuth
// @Produce     json
// @Success     200 {array} mdl.Profiles
// @Failure     401 {object} ErrorResponse
// @Failure     500 {object} ErrorResponse
// @Router      /api/v1/profiles [get]
func (pc *ProfileController) GetAllProfiles(c *gin.Context) {
	items, err := pc.profileService.List(c.Request.Context(), nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, items)
}

// GetProfileByID retorna um perfil pelo ID.
//
// @Summary     Obter perfil
// @Description Busca um perfil específico pelo ID.
// @Tags        profiles
// @Security    BearerAuth
// @Produce     json
// @Param       id path string true "ID do Perfil"
// @Success     200 {object} mdl.Profiles
// @Failure     401 {object} ErrorResponse
// @Failure     404 {object} ErrorResponse
// @Router      /api/v1/profiles/{id} [get]
func (pc *ProfileController) GetProfileByID(c *gin.Context) {
	id := c.Param("id")
	item, err := pc.profileService.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "profile not found"})
		return
	}
	c.JSON(http.StatusOK, item)
}

// CreateProfile cria um novo perfil.
//
// @Summary     Criar perfil
// @Description Adiciona um novo perfil.
// @Tags        profiles
// @Security    BearerAuth
// @Accept      json
// @Produce     json
// @Param       payload body mdl.Profiles true "Dados do Perfil"
// @Success     201 {object} mdl.Profiles
// @Failure     400 {object} ErrorResponse
// @Failure     401 {object} ErrorResponse
// @Failure     500 {object} ErrorResponse
// @Router      /api/v1/profiles [post]
func (pc *ProfileController) CreateProfile(c *gin.Context) {
	var request mdl.Profiles
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	err := pc.profileService.Create(c.Request.Context(), &request)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, request)
}

// UpdateProfile atualiza um perfil.
//
// @Summary     Atualizar perfil
// @Description Atualiza os dados de um perfil.
// @Tags        profiles
// @Security    BearerAuth
// @Accept      json
// @Produce     json
// @Param       id      path string            true "ID do Perfil"
// @Param       payload body mdl.Profiles true "Dados atualizados"
// @Success     200 {object} mdl.Profiles
// @Failure     400 {object} ErrorResponse
// @Failure     401 {object} ErrorResponse
// @Failure     404 {object} ErrorResponse
// @Failure     500 {object} ErrorResponse
// @Router      /api/v1/profiles/{id} [put]
func (pc *ProfileController) UpdateProfile(c *gin.Context) {
	var request mdl.Profiles
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	err := pc.profileService.Update(c.Request.Context(), &request)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, request)
}

// DeleteProfile remove um perfil.
//
// @Summary     Remover perfil
// @Description Remove um perfil.
// @Tags        profiles
// @Security    BearerAuth
// @Produce     json
// @Param       id path string true "ID do Perfil"
// @Success     204
// @Failure     401 {object} ErrorResponse
// @Failure     404 {object} ErrorResponse
// @Failure     500 {object} ErrorResponse
// @Router      /api/v1/profiles/{id} [delete]
func (pc *ProfileController) DeleteProfile(c *gin.Context) {
	id := c.Param("id")
	if err := pc.profileService.Delete(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

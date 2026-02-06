// Package refreshtokens provides the controller for managing refresh tokens.
package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	mdl "github.com/kubex-ecosystem/gnyx/internal/dsclient"
	svc "github.com/kubex-ecosystem/gnyx/internal/dsclient/adapters"
	t "github.com/kubex-ecosystem/gnyx/internal/types"
)

type RefreshTokenController struct {
	refreshtokenService svc.Service[mdl.RefreshTokens]
	APIWrapper          *t.APIWrapper[mdl.RefreshTokens]
}

func NewRefreshTokenController(bridge svc.Service[mdl.RefreshTokens]) *RefreshTokenController {
	// repo := svc.Repository{}
	// service := svc.Service{}
	return &RefreshTokenController{
		refreshtokenService: nil,
		APIWrapper:          t.NewAPIWrapper[mdl.RefreshTokens](),
	}
}

// GetAllRefreshTokens retorna todos os refresh tokens.
//
// @Summary     Listar refresh tokens
// @Description Recupera a lista de refresh tokens.
// @Tags        refreshtokens
// @Security    BearerAuth
// @Produce     json
// @Success     200 {array} mdl.RefreshTokens
// @Failure     401 {object} ErrorResponse
// @Failure     500 {object} ErrorResponse
// @Router      /api/v1/refreshtokens [get]
func (pc *RefreshTokenController) GetAllRefreshTokens(c *gin.Context) {
	items, err := pc.refreshtokenService.List(c.Request.Context(), nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, items)
}

// GetRefreshTokenByID retorna um refresh token pelo ID.
//
// @Summary     Obter refresh token
// @Description Busca um refresh token específico pelo ID.
// @Tags        refreshtokens
// @Security    BearerAuth
// @Produce     json
// @Param       id path string true "ID do Refresh Token"
// @Success     200 {object} mdl.RefreshTokens
// @Failure     401 {object} ErrorResponse
// @Failure     404 {object} ErrorResponse
// @Router      /api/v1/refreshtokens/{id} [get]
func (pc *RefreshTokenController) GetRefreshTokenByID(c *gin.Context) {
	id := c.Param("id")
	item, err := pc.refreshtokenService.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "refresh token not found"})
		return
	}
	c.JSON(http.StatusOK, item)
}

// CreateRefreshToken cria um novo refresh token.
//
// @Summary     Criar refresh token
// @Description Adiciona um novo refresh token.
// @Tags        refreshtokens
// @Security    BearerAuth
// @Accept      json
// @Produce     json
// @Param       payload body mdl.RefreshTokens true "Dados do Refresh Token"
// @Success     201 {object} mdl.RefreshTokens
// @Failure     400 {object} ErrorResponse
// @Failure     401 {object} ErrorResponse
// @Failure     500 {object} ErrorResponse
// @Router      /api/v1/refreshtokens [post]
func (pc *RefreshTokenController) CreateRefreshToken(c *gin.Context) {
	var request mdl.RefreshTokens
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	err := pc.refreshtokenService.Create(c.Request.Context(), &request)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, request)
}

// UpdateRefreshToken atualiza um refresh token.
//
// @Summary     Atualizar refresh token
// @Description Atualiza os dados de um refresh token.
// @Tags        refreshtokens
// @Security    BearerAuth
// @Accept      json
// @Produce     json
// @Param       id      path string            true "ID do Refresh Token"
// @Param       payload body mdl.RefreshTokens true "Dados atualizados"
// @Success     200 {object} mdl.RefreshTokens
// @Failure     400 {object} ErrorResponse
// @Failure     401 {object} ErrorResponse
// @Failure     404 {object} ErrorResponse
// @Failure     500 {object} ErrorResponse
// @Router      /api/v1/refreshtokens/{id} [put]
func (pc *RefreshTokenController) UpdateRefreshToken(c *gin.Context) {
	var request mdl.RefreshTokens
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	err := pc.refreshtokenService.Update(c.Request.Context(), &request)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, request)
}

// DeleteRefreshToken remove um refresh token.
//
// @Summary     Remover refresh token
// @Description Remove um refresh token.
// @Tags        refreshtokens
// @Security    BearerAuth
// @Produce     json
// @Param       id path string true "ID do Refresh Token"
// @Success     204
// @Failure     401 {object} ErrorResponse
// @Failure     404 {object} ErrorResponse
// @Failure     500 {object} ErrorResponse
// @Router      /api/v1/refreshtokens/{id} [delete]
func (pc *RefreshTokenController) DeleteRefreshToken(c *gin.Context) {
	id := c.Param("id")
	if err := pc.refreshtokenService.Delete(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

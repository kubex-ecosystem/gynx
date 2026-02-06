// Package activityparticipants provides the controller for managing activity participants.
package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	mdl "github.com/kubex-ecosystem/gnyx/internal/dsclient"
	svc "github.com/kubex-ecosystem/gnyx/internal/dsclient/adapters"
	t "github.com/kubex-ecosystem/gnyx/internal/types"
)

type ActivityParticipantController struct {
	activityparticipantService svc.Service[mdl.ActivityParticipants]
	APIWrapper                 *t.APIWrapper[mdl.ActivityParticipants]
}

func NewActivityParticipantController(bridge svc.Service[mdl.ActivityParticipants]) *ActivityParticipantController {
	// service := svc.Service[mdl.ActivityParticipants]{nil}
	return &ActivityParticipantController{
		activityparticipantService: nil,
		APIWrapper:                 t.NewAPIWrapper[mdl.ActivityParticipants](),
	}
}

// GetAllActivityParticipants retorna todos os participantes de atividades.
//
// @Summary     Listar participantes de atividades
// @Description Recupera a lista de participantes de atividades.
// @Tags        activityparticipants
// @Security    BearerAuth
// @Produce     json
// @Success     200 {array} mdl.ActivityParticipants
// @Failure     401 {object} ErrorResponse
// @Failure     500 {object} ErrorResponse
// @Router      /api/v1/activityparticipants [get]
func (pc *ActivityParticipantController) GetAllActivityParticipants(c *gin.Context) {
	items, err := pc.activityparticipantService.List(c.Request.Context(), nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, items)
}

// GetActivityParticipantByID retorna um participante pelo ID.
//
// @Summary     Obter participante de atividade
// @Description Busca um participante de atividade específico pelo ID.
// @Tags        activityparticipants
// @Security    BearerAuth
// @Produce     json
// @Param       id path string true "ID do Participante"
// @Success     200 {object} mdl.ActivityParticipants
// @Failure     401 {object} ErrorResponse
// @Failure     404 {object} ErrorResponse
// @Router      /api/v1/activityparticipants/{id} [get]
func (pc *ActivityParticipantController) GetActivityParticipantByID(c *gin.Context) {
	id := c.Param("id")
	item, err := pc.activityparticipantService.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "activity participant not found"})
		return
	}
	c.JSON(http.StatusOK, item)
}

// CreateActivityParticipant cria um novo participante.
//
// @Summary     Criar participante de atividade
// @Description Adiciona um novo participante a uma atividade.
// @Tags        activityparticipants
// @Security    BearerAuth
// @Accept      json
// @Produce     json
// @Param       payload body mdl.ActivityParticipants true "Dados do Participante"
// @Success     201 {object} mdl.ActivityParticipants
// @Failure     400 {object} ErrorResponse
// @Failure     401 {object} ErrorResponse
// @Failure     500 {object} ErrorResponse
// @Router      /api/v1/activityparticipants [post]
func (pc *ActivityParticipantController) CreateActivityParticipant(c *gin.Context) {
	var request mdl.ActivityParticipants
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	err := pc.activityparticipantService.Create(c.Request.Context(), &request)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, request)
}

// UpdateActivityParticipant atualiza um participante.
//
// @Summary     Atualizar participante de atividade
// @Description Atualiza os dados de um participante de atividade.
// @Tags        activityparticipants
// @Security    BearerAuth
// @Accept      json
// @Produce     json
// @Param       id      path string            true "ID do Participante"
// @Param       payload body mdl.ActivityParticipants true "Dados atualizados"
// @Success     200 {object} mdl.ActivityParticipants
// @Failure     400 {object} ErrorResponse
// @Failure     401 {object} ErrorResponse
// @Failure     404 {object} ErrorResponse
// @Failure     500 {object} ErrorResponse
// @Router      /api/v1/activityparticipants/{id} [put]
func (pc *ActivityParticipantController) UpdateActivityParticipant(c *gin.Context) {
	var request mdl.ActivityParticipants
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	err := pc.activityparticipantService.Update(c.Request.Context(), &request)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, request)
}

// DeleteActivityParticipant remove um participante.
//
// @Summary     Remover participante de atividade
// @Description Remove um participante de uma atividade.
// @Tags        activityparticipants
// @Security    BearerAuth
// @Produce     json
// @Param       id path string true "ID do Participante"
// @Success     204
// @Failure     401 {object} ErrorResponse
// @Failure     404 {object} ErrorResponse
// @Failure     500 {object} ErrorResponse
// @Router      /api/v1/activityparticipants/{id} [delete]
func (pc *ActivityParticipantController) DeleteActivityParticipant(c *gin.Context) {
	id := c.Param("id")
	if err := pc.activityparticipantService.Delete(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

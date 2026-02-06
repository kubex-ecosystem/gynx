// Package activities provides the controller for managing activities in the application.
package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	mdl "github.com/kubex-ecosystem/gnyx/internal/dsclient"
	svc "github.com/kubex-ecosystem/gnyx/internal/dsclient/adapters"
	t "github.com/kubex-ecosystem/gnyx/internal/types"
)

type ActivityController struct {
	activityService svc.Service[mdl.Activities]
	APIWrapper      *t.APIWrapper[mdl.Activities]
}

func NewActivityController(bridge svc.Service[mdl.Activities]) *ActivityController {
	// Acessa o repo e o service através da bridge
	// activityRepo := svc.Repo[mdl.Activities]{}
	// activityService := svc.Service[mdl.Activities]{}
	return &ActivityController{
		// activityService: activityService,
		// APIWrapper:      t.NewAPIWrapper[mdl.Activities](),
	}
}

// GetAllActivities retorna todas as atividades.
//
// @Summary     Listar atividades
// @Description Recupera a lista de atividades registradas na base.
// @Tags        activities
// @Security    BearerAuth
// @Produce     json
// @Success     200 {array} mdl.Activities
// @Failure     401 {object} ErrorResponse
// @Failure     500 {object} ErrorResponse
// @Router      /api/v1/activities [get]
func (pc *ActivityController) GetAllActivities(c *gin.Context) {
	// activities, err := pc.activityService.List()
	// if err != nil {
	// 	c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	// 	return
	// }
	// c.JSON(http.StatusOK, activities)
	c.JSON(http.StatusOK, gin.H{"message": "GetAllActivities not implemented yet"})
}

// GetActivityByID retorna uma atividade pelo ID.
//
// @Summary     Obter atividade
// @Description Busca uma atividade específica pelo ID.
// @Tags        activities
// @Security    BearerAuth
// @Produce     json
// @Param       id path string true "ID da Atividade"
// @Success     200 {object} mdl.Activities
// @Failure     401 {object} ErrorResponse
// @Failure     404 {object} ErrorResponse
// @Router      /api/v1/activities/{id} [get]
func (pc *ActivityController) GetActivityByID(c *gin.Context) {
	// id := c.Param("id")
	// activity, err := pc.activityService.GetByID(id)
	// if err != nil {
	// 	c.JSON(http.StatusNotFound, gin.H{"error": "activity not found"})
	// 	return
	// }
	// c.JSON(http.StatusOK, activity)
	c.JSON(http.StatusOK, gin.H{"message": "GetActivityByID not implemented yet"})
}

// CreateActivity cria uma nova atividade.
//
// @Summary     Criar atividade
// @Description Persiste uma nova atividade com os dados enviados.
// @Tags        activities
// @Security    BearerAuth
// @Accept      json
// @Produce     json
// @Param       payload body mdl.Activities true "Dados da Atividade"
// @Success     201 {object} mdl.Activities
// @Failure     400 {object} ErrorResponse
// @Failure     401 {object} ErrorResponse
// @Failure     500 {object} ErrorResponse
// @Router      /api/v1/activities [post]
func (pc *ActivityController) CreateActivity(c *gin.Context) {
	// var activityRequest mdl.Activities
	// if err := c.ShouldBindJSON(&activityRequest); err != nil {
	// 	c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	// 	return
	// }
	// createdActivity, err := pc.activityService.CreateActivity(&activityRequest)
	// if err != nil {
	// 	c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	// 	return
	// }
	// c.JSON(http.StatusCreated, createdActivity)
	c.JSON(http.StatusOK, gin.H{"message": "CreateActivity not implemented yet"})
}

// UpdateActivity atualiza uma atividade existente.
//
// @Summary     Atualizar atividade
// @Description Atualiza os dados de uma atividade identificada por ID.
// @Tags        activities
// @Security    BearerAuth
// @Accept      json
// @Produce     json
// @Param       id      path string            true "ID da Atividade"
// @Param       payload body mdl.Activities true "Dados atualizados"
// @Success     200 {object} mdl.Activities
// @Failure     400 {object} ErrorResponse
// @Failure     401 {object} ErrorResponse
// @Failure     404 {object} ErrorResponse
// @Failure     500 {object} ErrorResponse
// @Router      /api/v1/activities/{id} [put]
func (pc *ActivityController) UpdateActivity(c *gin.Context) {
	// var activityRequest mdl.Activities
	// if err := c.ShouldBindJSON(&activityRequest); err != nil {
	// 	c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	// 	return
	// }
	// updatedActivity, err := pc.activityService.UpdateActivity(&activityRequest)
	// if err != nil {
	// 	c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	// 	return
	// }
	// c.JSON(http.StatusOK, updatedActivity)
	c.JSON(http.StatusOK, gin.H{"message": "UpdateActivity not implemented yet"})
}

// DeleteActivity remove uma atividade.
//
// @Summary     Remover atividade
// @Description Exclui uma atividade identificada pelo ID.
// @Tags        activities
// @Security    BearerAuth
// @Produce     json
// @Param       id path string true "ID da Atividade"
// @Success     204
// @Failure     401 {object} ErrorResponse
// @Failure     404 {object} ErrorResponse
// @Failure     500 {object} ErrorResponse
// @Router      /api/v1/activities/{id} [delete]
func (pc *ActivityController) DeleteActivity(c *gin.Context) {
	// id := c.Param("id")
	// if err := pc.activityService.Delete(id); err != nil {
	// 	c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	// 	return
	// }
	// c.Status(http.StatusNoContent)
	c.JSON(http.StatusOK, gin.H{"message": "DeleteActivity not implemented yet"})
}

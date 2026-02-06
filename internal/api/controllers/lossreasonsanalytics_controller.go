// Package lossreasonsanalytics provides the controller for managing loss reason analytics.
package controllers

// import (
// 	"net/http"

// 	svc "github.com/kubex-ecosystem/gnyx/internal/dsclient/adapters"
// 	// mdl "github.com/kubex-ecosystem/gnyx/internal/dsclient"
// 	t "github.com/kubex-ecosystem/gnyx/internal/types"
// 	"github.com/gin-gonic/gin"
// )

// type LossReasonAnalyticController struct {
// 	// lossreasonanalyticService svc.Service[mdl.LossReasonAnalytics]
// 	// APIWrapper                *t.APIWrapper[mdl.LossReasonAnalytics]
// }

// type (
// 	// ErrorResponse padroniza a documentação de erros dos endpoints.
// 	ErrorResponse = t.ErrorResponse
// )

// func NewLossReasonAnalyticController(bridge svc.Service[mdl.LossReasonAnalytics]) *LossReasonAnalyticController {
// 	// repo := svc.Repository{}
// 	// service := svc.Service{}
// 	return &LossReasonAnalyticController{
// 		lossreasonanalyticService: nil,
// 		APIWrapper:                t.NewAPIWrapper[mdl.LossReasonAnalytic](),
// 	}
// }

// // GetAllLossReasonAnalytics retorna todos os analytics de motivo de perda.
// //
// // @Summary     Listar analytics de motivo de perda
// // @Description Recupera a lista de analytics de motivo de perda.
// // @Tags        lossreasonsanalytics
// // @Security    BearerAuth
// // @Produce     json
// // @Success     200 {array} mdl.LossReasonAnalytics
// // @Failure     401 {object} ErrorResponse
// // @Failure     500 {object} ErrorResponse
// // @Router      /api/v1/lossreasonsanalytics [get]
// func (pc *LossReasonAnalyticController) GetAllLossReasonAnalytics(c *gin.Context) {
// 	items, err := pc.lossreasonanalyticService.List()
// 	if err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
// 		return
// 	}
// 	c.JSON(http.StatusOK, items)
// }

// // GetLossReasonAnalyticByID retorna um analytic pelo ID.
// //
// // @Summary     Obter analytic de motivo de perda
// // @Description Busca um analytic de motivo de perda específico pelo ID.
// // @Tags        lossreasonsanalytics
// // @Security    BearerAuth
// // @Produce     json
// // @Param       id path string true "ID do Analytic"
// // @Success     200 {object} mdl.LossReasonAnalytics
// // @Failure     401 {object} ErrorResponse
// // @Failure     404 {object} ErrorResponse
// // @Router      /api/v1/lossreasonsanalytics/{id} [get]
// func (pc *LossReasonAnalyticController) GetLossReasonAnalyticByID(c *gin.Context) {
// 	id := c.Param("id")
// 	item, err := pc.lossreasonanalyticService.GetByID(id)
// 	if err != nil {
// 		c.JSON(http.StatusNotFound, gin.H{"error": "loss reason analytic not found"})
// 		return
// 	}
// 	c.JSON(http.StatusOK, item)
// }

// // CreateLossReasonAnalytic cria um novo analytic.
// //
// // @Summary     Criar analytic de motivo de perda
// // @Description Adiciona um novo analytic de motivo de perda.
// // @Tags        lossreasonsanalytics
// // @Security    BearerAuth
// // @Accept      json
// // @Produce     json
// // @Param       payload body mdl.LossReasonAnalytics true "Dados do Analytic"
// // @Success     201 {object} mdl.LossReasonAnalytics
// // @Failure     400 {object} ErrorResponse
// // @Failure     401 {object} ErrorResponse
// // @Failure     500 {object} ErrorResponse
// // @Router      /api/v1/lossreasonsanalytics [post]
// func (pc *LossReasonAnalyticController) CreateLossReasonAnalytic(c *gin.Context) {
// 	var request mdl.LossReasonAnalytics
// 	if err := c.ShouldBindJSON(&request); err != nil {
// 		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
// 		return
// 	}
// 	createdItem, err := pc.lossreasonanalyticService.Create(&request)
// 	if err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
// 		return
// 	}
// 	c.JSON(http.StatusCreated, createdItem)
// }

// // UpdateLossReasonAnalytic atualiza um analytic.
// //
// // @Summary     Atualizar analytic de motivo de perda
// // @Description Atualiza os dados de um analytic de motivo de perda.
// // @Tags        lossreasonsanalytics
// // @Security    BearerAuth
// // @Accept      json
// // @Produce     json
// // @Param       id      path string            true "ID do Analytic"
// // @Param       payload body mdl.LossReasonAnalytics true "Dados atualizados"
// // @Success     200 {object} mdl.LossReasonAnalytics
// // @Failure     400 {object} ErrorResponse
// // @Failure     401 {object} ErrorResponse
// // @Failure     404 {object} ErrorResponse
// // @Failure     500 {object} ErrorResponse
// // @Router      /api/v1/lossreasonsanalytics/{id} [put]
// func (pc *LossReasonAnalyticController) UpdateLossReasonAnalytic(c *gin.Context) {
// 	var request mdl.LossReasonAnalytics
// 	if err := c.ShouldBindJSON(&request); err != nil {
// 		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
// 		return
// 	}
// 	updatedItem, err := pc.lossreasonanalyticService.Update(&request)
// 	if err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
// 		return
// 	}
// 	c.JSON(http.StatusOK, updatedItem)
// }

// // DeleteLossReasonAnalytic remove um analytic.
// //
// // @Summary     Remover analytic de motivo de perda
// // @Description Remove um analytic de motivo de perda.
// // @Tags        lossreasonsanalytics
// // @Security    BearerAuth
// // @Produce     json
// // @Param       id path string true "ID do Analytic"
// // @Success     204
// // @Failure     401 {object} ErrorResponse
// // @Failure     404 {object} ErrorResponse
// // @Failure     500 {object} ErrorResponse
// // @Router      /api/v1/lossreasonsanalytics/{id} [delete]
// func (pc *LossReasonAnalyticController) DeleteLossReasonAnalytic(c *gin.Context) {
// 	id := c.Param("id")
// 	if err := pc.lossreasonanalyticService.Delete(id); err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
// 		return
// 	}
// 	c.Status(http.StatusNoContent)
// }

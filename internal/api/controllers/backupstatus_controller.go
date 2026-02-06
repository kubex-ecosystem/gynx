// Package backupstatus provides the controller for managing backup statuses.
package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	mdl "github.com/kubex-ecosystem/gnyx/internal/dsclient"
	svc "github.com/kubex-ecosystem/gnyx/internal/dsclient/adapters"
	t "github.com/kubex-ecosystem/gnyx/internal/types"
)

type BackupStatusController struct {
	backupstatusService svc.Service[mdl.BackupStatus]
	APIWrapper          *t.APIWrapper[mdl.BackupStatus]
}

func NewBackupStatusController(bridge svc.Service[mdl.BackupStatus]) *BackupStatusController {
	// repo := svc.Repository{}
	// service := svc.Service{}
	return &BackupStatusController{
		backupstatusService: nil,
		APIWrapper:          t.NewAPIWrapper[mdl.BackupStatus](),
	}
}

// GetAllBackupStatuses retorna todos os status de backup.
//
// @Summary     Listar status de backups
// @Description Recupera a lista de status de backups.
// @Tags        backupstatus
// @Security    BearerAuth
// @Produce     json
// @Success     200 {array} mdl.BackupStatus
// @Failure     401 {object} ErrorResponse
// @Failure     500 {object} ErrorResponse
// @Router      /api/v1/backupstatus [get]
func (pc *BackupStatusController) GetAllBackupStatuses(c *gin.Context) {
	items, err := pc.backupstatusService.List(c.Request.Context(), nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, items)
}

// GetBackupStatusByID retorna um status de backup pelo ID.
//
// @Summary     Obter status de backup
// @Description Busca um status de backup específico pelo ID.
// @Tags        backupstatus
// @Security    BearerAuth
// @Produce     json
// @Param       id path string true "ID do Status de Backup"
// @Success     200 {object} mdl.BackupStatus
// @Failure     401 {object} ErrorResponse
// @Failure     404 {object} ErrorResponse
// @Router      /api/v1/backupstatus/{id} [get]
func (pc *BackupStatusController) GetBackupStatusByID(c *gin.Context) {
	id := c.Param("id")
	item, err := pc.backupstatusService.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "backup status not found"})
		return
	}
	c.JSON(http.StatusOK, item)
}

// CreateBackupStatus cria um novo status de backup.
//
// @Summary     Criar status de backup
// @Description Adiciona um novo status de backup.
// @Tags        backupstatus
// @Security    BearerAuth
// @Accept      json
// @Produce     json
// @Param       payload body mdl.BackupStatus true "Dados do Status de Backup"
// @Success     201 {object} mdl.BackupStatus
// @Failure     400 {object} ErrorResponse
// @Failure     401 {object} ErrorResponse
// @Failure     500 {object} ErrorResponse
// @Router      /api/v1/backupstatus [post]
func (pc *BackupStatusController) CreateBackupStatus(c *gin.Context) {
	var request mdl.BackupStatus
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	err := pc.backupstatusService.Create(c.Request.Context(), &request)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, request)
}

// UpdateBackupStatus atualiza um status de backup.
//
// @Summary     Atualizar status de backup
// @Description Atualiza os dados de um status de backup.
// @Tags        backupstatus
// @Security    BearerAuth
// @Accept      json
// @Produce     json
// @Param       id      path string            true "ID do Status de Backup"
// @Param       payload body mdl.BackupStatus true "Dados atualizados"
// @Success     200 {object} mdl.BackupStatus
// @Failure     400 {object} ErrorResponse
// @Failure     401 {object} ErrorResponse
// @Failure     404 {object} ErrorResponse
// @Failure     500 {object} ErrorResponse
// @Router      /api/v1/backupstatus/{id} [put]
func (pc *BackupStatusController) UpdateBackupStatus(c *gin.Context) {
	var request mdl.BackupStatus
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	err := pc.backupstatusService.Update(c.Request.Context(), &request)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, request)
}

// DeleteBackupStatus remove um status de backup.
//
// @Summary     Remover status de backup
// @Description Remove um status de backup.
// @Tags        backupstatus
// @Security    BearerAuth
// @Produce     json
// @Param       id path string true "ID do Status de Backup"
// @Success     204
// @Failure     401 {object} ErrorResponse
// @Failure     404 {object} ErrorResponse
// @Failure     500 {object} ErrorResponse
// @Router      /api/v1/backupstatus/{id} [delete]
func (pc *BackupStatusController) DeleteBackupStatus(c *gin.Context) {
	id := c.Param("id")
	if err := pc.backupstatusService.Delete(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

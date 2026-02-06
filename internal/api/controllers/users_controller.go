// Package userscontroller implements the Users controller for handling HTTP requests.
package controllers

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/kubex-ecosystem/gnyx/internal/dsclient"
	"github.com/kubex-ecosystem/gnyx/internal/dsclient/adapters"
	"gorm.io/gorm"
)

type UsersController[T any] struct {
	userStore dsclient.UserStore
	executor  dsclient.PGExecutor
	service   adapters.Service[T]
}

func NewUsersController[T any](modelStore dsclient.StoreType, service adapters.Service[T]) Controller[T] {
	if us, ok := modelStore.(dsclient.UserStore); ok {
		return &UsersController[T]{userStore: us, service: service}
	}
	return &UsersController[T]{userStore: nil, service: service}
}

func (ctrl *UsersController[T]) GetAll(c *gin.Context) {
	items, err := ctrl.userStore.List(c.Request.Context(), nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, items)
}

func (ctrl *UsersController[T]) GetByID(c *gin.Context) {
	id := c.Param("id")
	item, err := ctrl.userStore.GetByID(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Users not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, item)
}

func (ctrl *UsersController[T]) Create(c *gin.Context) {
	var payload dsclient.CreateUserInput
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if user, err := ctrl.userStore.Create(c.Request.Context(), &payload); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	} else {
		c.JSON(http.StatusCreated, user)
		return
	}
}

func (ctrl *UsersController[T]) Update(c *gin.Context) {
	// id := c.Param("id")
	var payload dsclient.UpdateUserInput
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	// payload.ID = &id // Ensure the ID from the URL is used for the update
	user, err := ctrl.userStore.Update(c.Request.Context(), &payload)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Users not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, user)
}

func (ctrl *UsersController[T]) Delete(c *gin.Context) {
	id := c.Param("id")
	if err := ctrl.userStore.Delete(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

package controllers

import (
	mdl "github.com/kubex-ecosystem/gnyx/internal/dsclient"
	svc "github.com/kubex-ecosystem/gnyx/internal/dsclient/adapters"
)

type UserProfilesController struct {
	service svc.Service[mdl.TrainingBadges]
}

func NewUserProfilesController(service svc.Service[mdl.UserProfiles]) Controller[mdl.UserProfiles] {
	return NewController[mdl.UserProfiles](nil, service)
}

// func (ctrl *Controller) GetAll(c *gin.Context) {
// 	items, err := ctrl.service.List(c.Request.Context(), nil)
// 	if err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
// 		return
// 	}
// 	c.JSON(http.StatusOK, items)
// }

// func (ctrl *Controller) GetByID(c *gin.Context) {
// 	id := c.Param("id")
// 	item, err := ctrl.service.GetByID(c.Request.Context(), id)
// 	if err != nil {
// 		if errors.Is(err, gorm.ErrRecordNotFound) {
// 			c.JSON(http.StatusNotFound, gin.H{"error": "UserProfiles not found"})
// 			return
// 		}
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
// 		return
// 	}
// 	c.JSON(http.StatusOK, item)
// }

// func (ctrl *Controller) Create(c *gin.Context) {
// 	var payload mdl.UserProfiles
// 	if err := c.ShouldBindJSON(&payload); err != nil {
// 		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
// 		return
// 	}
// 	if err := ctrl.service.Create(c.Request.Context(), &payload); err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
// 		return
// 	}
// 	c.JSON(http.StatusCreated, payload)
// }

// func (ctrl *Controller) Update(c *gin.Context) {
// 	id := c.Param("id")
// 	var payload mdl.UserProfiles
// 	if err := c.ShouldBindJSON(&payload); err != nil {
// 		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
// 		return
// 	}
// 	payload.ID = &id // Ensure the ID from the URL is used for the update
// 	if err := ctrl.service.Update(c.Request.Context(), &payload); err != nil {
// 		if errors.Is(err, gorm.ErrRecordNotFound) {
// 			c.JSON(http.StatusNotFound, gin.H{"error": "UserProfiles not found"})
// 			return
// 		}
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
// 		return
// 	}
// 	c.JSON(http.StatusOK, payload)
// }

// func (ctrl *Controller) Delete(c *gin.Context) {
// 	id := c.Param("id")
// 	if err := ctrl.service.Delete(c.Request.Context(), id); err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
// 		return
// 	}
// 	c.Status(http.StatusNoContent)
// }

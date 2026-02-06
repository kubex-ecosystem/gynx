// Package invite defines the response helpers for the Invite service.
package invite

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type envelope struct {
	Status  string      `json:"status"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

func ok(c *gin.Context, data any) {
	c.JSON(http.StatusOK, envelope{Status: "ok", Data: data})
}

func created(c *gin.Context, data any) {
	c.JSON(http.StatusCreated, envelope{Status: "ok", Data: data})
}

func fail(c *gin.Context, code int, msg string) {
	c.JSON(code, envelope{Status: "error", Message: msg})
}

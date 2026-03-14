// Package main provides a standalone Swagger UI preview for local documentation review.
// @title       GNyx API
// @version     1.0.0
// @description Backend modular API for GNyx. (Powered by Kubex Ecosystem)
// @termsOfService  https://gnyx.kubex.world/terms
// @contact.name   GNyx API Support
// @contact.url    https://gnyx.kubex.world
// @contact.email  contact@kubex.world
// @host      localhost:5001
// @BasePath  /
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Provide "Bearer <token>"
// @externalDocs.description  OpenAPI
// @externalDocs.url          https://swagger.io/resources/open-api/
package main

import (
	"github.com/gin-gonic/gin"
	docs "github.com/kubex-ecosystem/gnyx/docs"
	"github.com/kubex-ecosystem/gnyx/internal/config"
	swaggerruntime "github.com/kubex-ecosystem/gnyx/internal/features/swagger"
	gl "github.com/kubex-ecosystem/logz"
)

func main() {
	gin.SetMode(gin.ReleaseMode)
	engine := gin.New()
	engine.Use(gin.Recovery())

	cfg := config.LoadConfig()
	swaggerruntime.Register(engine, cfg)
	docs.SwaggerInfo.Title = "GNyx API"

	gl.Log("info", "Swagger preview available at: http://localhost:5001/swagger/index.html")
	if err := engine.Run("0.0.0.0:5001"); err != nil {
		gl.Log("fatal", "Failed to start standalone Swagger preview:", err)
	}
}

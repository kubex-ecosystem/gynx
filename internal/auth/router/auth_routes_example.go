// Package router implements the authentication routes.
package router

import (
	"context"
	"crypto/rsa"

	"github.com/kubex-ecosystem/gnyx/internal/app"
	"github.com/kubex-ecosystem/gnyx/internal/auth/controllers"
	"github.com/kubex-ecosystem/gnyx/internal/auth/middlewares"
	"github.com/kubex-ecosystem/gnyx/internal/auth/tokens"
	"github.com/kubex-ecosystem/gnyx/internal/config"
	services "github.com/kubex-ecosystem/gnyx/internal/dsclient/datastore/session_store"
	repositories "github.com/kubex-ecosystem/gnyx/internal/dsclient/datastore/user_store"
	gl "github.com/kubex-ecosystem/logz"

	"github.com/gin-gonic/gin"
)

// Codex: esse arquivo é só EXEMPLO de composição.
// A injeção real deve ser feita onde o router é montado (internal/app/container.go?).

func RegisterAuthRoutes(r *gin.Engine, priv *rsa.PrivateKey, pub *rsa.PublicKey, container *app.Container) {
	cfg := container.GetConfig()
	if cfg == nil {
		cfg = config.LoadConfig()
	}

	userRepo, err := repositories.NewUserRepository()
	if err != nil {
		gl.Log("error", "failed to init user repo: %v", err)
		return
	}
	sessRepo, err := repositories.NewSessionRepository()
	if err != nil {
		gl.Log("error", "failed to init session repo: %v", err)
		return
	}
	jwtSvc := tokens.NewJWTService(cfg, priv, pub)
	logger := gl.GetLoggerZ("auth_service")

	userBridge, err := repositories.NewUserBridge(context.Background())
	if err != nil {
		gl.Log("error", "failed to init user bridge: %v", err)
		return
	}

	authSvc := services.NewAuthService(userRepo, sessRepo, jwtSvc, logger)
	controller := controllers.NewAuthController(authSvc, userBridge)
	authMW := middlewares.NewAuthMiddleware(jwtSvc)

	v1 := r.Group("/api/v1")
	authGroup := v1.Group("/auth")
	{
		authGroup.POST("/sign-up", controller.SignUp)
		authGroup.POST("/sign-in", controller.SignIn)
		authGroup.POST("/refresh", controller.Refresh)
		authGroup.POST("/sign-out", controller.SignOut)
	}

	protected := v1.Group("")
	protected.Use(authMW.Handle())
	{
		// Exemplo: rota protegida básica
		protected.GET("/me", func(c *gin.Context) {
			userID, _ := c.Get(middlewares.ContextUserIDKey)
			c.JSON(200, gin.H{"user_id": userID})
		})
	}
}

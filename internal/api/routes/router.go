package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/kubex-ecosystem/gnyx/internal/api/email"
	invite "github.com/kubex-ecosystem/gnyx/internal/api/invite"
	"github.com/kubex-ecosystem/gnyx/internal/auth/middlewares"
	"github.com/kubex-ecosystem/gnyx/internal/auth/tokens"
	"github.com/kubex-ecosystem/gnyx/internal/config"
	"github.com/kubex-ecosystem/gnyx/internal/features/providers/registry"
	runtimeMW "github.com/kubex-ecosystem/gnyx/internal/runtime/middlewares"
	"github.com/kubex-ecosystem/gnyx/internal/services/mailer"
	"github.com/kubex-ecosystem/gnyx/internal/types"

	kbxIs "github.com/kubex-ecosystem/kbx/is"
	gl "github.com/kubex-ecosystem/logz"
)

func RegisterRoutes(r *gin.RouterGroup, container types.IContainer) gin.IRoutes {
	return RegisterRoutesWithProviders(r, container, nil, nil)
}

func RegisterRoutesWithProviders(
	r *gin.RouterGroup,
	container types.IContainer,
	reg *registry.Registry,
	prod *runtimeMW.ProductionMiddleware,
) gin.IRoutes {
	if reg != nil {
		registerRuntimeAIRoutes(r, container, reg, prod)
	}

	if _, err := RegisterAuthHTTP(r, container); err != nil {
		gl.Fatalf("failed to register auth routes: %v", err)
	}

	if _, err := RegisterUserRoutes(r, container); err != nil {
		gl.Fatalf("failed to register user routes: %v", err)
	}

	if _, err := RegisterContactRoutes(r, container); err != nil {
		gl.Warnf("failed to register contact routes: %v", err)
	}

	if kbxIs.Safe(container, false) && kbxIs.Safe(container.IMAPService(), false) {
		emailCtl := email.NewController(container.IMAPService().(*mailer.IMAPService))
		email.Register(r, emailCtl)
	} else {
		gl.Log("info", "IMAP service not configured; skipping /email endpoints")
	}

	// Invite routes (public accept + admin create)
	if kbxIs.Safe(container, false) && kbxIs.Safe(container.InviteService(), false) {
		if svc, ok := container.InviteService().(invite.Service); ok {
			inviteCtl := invite.NewController(svc)

			// Auth middleware para rotas protegidas (create/list)
			cfg, ok := container.Config().(*config.Config)
			if !ok {
				gl.Warn("missing config for auth middleware; skipping invite protection")
				return r
			}
			priv, pub, err := loadOrGenerateKeys(cfg)
			if err != nil {
				gl.Warnf("failed to init auth keys for invites: %v", err)
			}
			jwtSvc := tokens.NewJWTService(cfg, priv, pub)
			authMW := middlewares.NewAuthMiddleware(jwtSvc)

			protected := r.Group("/invites")
			protected.Use(authMW.Handle())
			protected.GET("", inviteCtl.ListInvites)
			protected.POST("", inviteCtl.CreateInvite)

			// Alias compatível com FE
			r.POST("/users/invite", authMW.Handle(), inviteCtl.CreateInvite)

			public := r.Group("/invites")
			public.GET("/:token", inviteCtl.ValidateToken)
			public.POST("/:token/accept", inviteCtl.AcceptInvite)

			gl.Debug("Invite routes registered")
		} else {
			gl.Log("warn", "InviteService missing expected interface, skipping invite routes")
		}
	} else {
		gl.Log("warn", "InviteService not available, skipping invite routes")
	}

	return r
}

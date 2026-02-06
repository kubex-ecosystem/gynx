// Package invite contém as rotas e handlers para o serviço de convites.
package invite

import "github.com/gin-gonic/gin"

func Register(rg *gin.RouterGroup, ctl *Controller) {
	g := rg.Group("/invites")
	{
		g.GET("", ctl.ListInvites)                 // admin/tenant scoped
		g.POST("", ctl.CreateInvite)               // admin-only
		g.GET("/:token", ctl.ValidateToken)        // público
		g.POST("/:token/accept", ctl.AcceptInvite) // público (token scoped)
	}

	// Alias de compatibilidade: o FE chama /api/v1/users/invite para criar convite
	rg.POST("/users/invite", ctl.CreateInvite)
}

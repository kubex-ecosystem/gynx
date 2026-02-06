// Package invite defines the controller for handling invite-related API requests.
package invite

import (
	"context"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/kubex-ecosystem/gnyx/internal/auth/middlewares"
)

type Controller struct {
	svc Service
}

func NewController(svc Service) *Controller { return &Controller{svc: svc} }

// ListInvites handles GET /api/v1/invites
func (ctl *Controller) ListInvites(c *gin.Context) {
	var q InviteListFilters
	if err := c.ShouldBindQuery(&q); err != nil {
		fail(c, 400, "invalid query params")
		return
	}
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	res, err := ctl.svc.ListInvites(ctx, q)
	if err != nil {
		fail(c, 422, err.Error())
		return
	}
	ok(c, res)
}

// CreateInvite handles POST /api/v1/invites
//
// @Summary     Create an invite
// @Description Creates a new invite and sends an invitation email.
// @Tags        Invites
// @Security    BearerAuth
// @Accept      json
// @Produce     json
// @Param       invite body CreateInviteReq true "Invite Request"
// @Success     200 {object} MessageResponse
// @Failure     400 {object} ErrorResponse
// @Failure     401 {object} ErrorResponse
// @Failure     403 {object} ErrorResponse
// @Failure     500 {object} ErrorResponse
// @Router      /api/v1/invites [post]
func (ctl *Controller) CreateInvite(c *gin.Context) {
	var req CreateInviteReq
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, 400, "invalid payload")
		return
	}

	// Server-side attribution of invited_by: take from auth context/header if not provided
	if strings.TrimSpace(req.InvitedBy) == "" {
		if v, ok := c.Get(middlewares.ContextUserIDKey); ok {
			if s, ok := v.(string); ok {
				req.InvitedBy = s
			}
		}
	}
	if strings.TrimSpace(req.InvitedBy) == "" {
		if v, ok := c.Get("user_id"); ok {
			if s, ok := v.(string); ok {
				req.InvitedBy = s
			}
		}
	}
	if strings.TrimSpace(req.InvitedBy) == "" {
		req.InvitedBy = strings.TrimSpace(c.GetHeader("X-User-ID"))
	}

	if strings.TrimSpace(req.InvitedBy) == "" {
		fail(c, 401, "invited_by is required (missing user context)")
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	dto, err := ctl.svc.CreateInvite(ctx, req)
	if err != nil {
		fail(c, 422, err.Error())
		return
	}
	created(c, dto)
}

// ValidateToken handles GET /api/v1/invites/:token
//
// @Summary     Validate an invite token
// @Description Validates the invite token and retrieves associated invite details.
// @Tags        Invites
// @Accept      json
// @Produce     json
// @Param       token path string true "Invite Token"
// @Success     200 {object} InviteDetailsDTO
// @Failure     400 {object} ErrorResponse
// @Failure     404 {object} ErrorResponse
// @Router      /api/v1/invites/{token} [get]
func (ctl *Controller) ValidateToken(c *gin.Context) {
	token := c.Param("token")
	if token == "" {
		fail(c, 400, "missing token")
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	dto, err := ctl.svc.ValidateToken(ctx, token)
	if err != nil {
		fail(c, 404, err.Error())
		return
	}
	ok(c, dto)
}

// AcceptInvite handles POST /api/v1/invites/:token/accept
//
// @Summary     Accept an invite
// @Description Accepts the invite associated with the token and creates a user account.
// @Tags        Invites
// @Accept      json
// @Produce     json
// @Param       token path string true "Invite Token"
// @Param       acceptInvite body AcceptInviteReq true "Accept Invite Request"
// @Success     200 {object} UserDTO
// @Failure     400 {object} ErrorResponse
// @Failure     422 {object} ErrorResponse
// @Router      /api/v1/invites/{token}/accept [post]
func (ctl *Controller) AcceptInvite(c *gin.Context) {
	token := c.Param("token")
	if token == "" {
		fail(c, 400, "missing token")
		return
	}

	var req AcceptInviteReq
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, 400, "invalid payload")
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 15*time.Second)
	defer cancel()

	res, err := ctl.svc.AcceptInvite(ctx, token, req)
	if err != nil {
		fail(c, 422, err.Error())
		return
	}
	ok(c, res)
}

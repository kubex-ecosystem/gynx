package routes

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"cloud.google.com/go/auth/credentials/idtoken"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/kubex-ecosystem/gnyx/internal/auth/tokens"
	"github.com/kubex-ecosystem/gnyx/internal/config"
	auth "github.com/kubex-ecosystem/gnyx/internal/domain/auth"
	"github.com/kubex-ecosystem/gnyx/internal/dsclient"
	"github.com/kubex-ecosystem/gnyx/internal/dsclient/datastore"
	"github.com/kubex-ecosystem/gnyx/internal/features/cookies"
	"github.com/kubex-ecosystem/gnyx/internal/types"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"

	services "github.com/kubex-ecosystem/gnyx/internal/dsclient/datastore/session_store"
	userstore "github.com/kubex-ecosystem/gnyx/internal/dsclient/datastore/user_store"

	crt "github.com/kubex-ecosystem/gnyx/internal/services/security/certificates"
	"github.com/kubex-ecosystem/kbx"

	defaults "github.com/kubex-ecosystem/gnyx/internal/module/kbx"
	kbxGet "github.com/kubex-ecosystem/kbx/get"
	kbxIs "github.com/kubex-ecosystem/kbx/is"
	kbxLoad "github.com/kubex-ecosystem/kbx/load"
	kbxTool "github.com/kubex-ecosystem/kbx/tools"
	gl "github.com/kubex-ecosystem/logz"
)

type authHTTP struct {
	authSvc  services.AuthService
	jwt      tokens.JWTService
	userRepo userstore.UserRepository
	authCfg  *config.Config
}

type accessMemberPayload struct {
	ID          string    `json:"id"`
	Email       string    `json:"email"`
	Name        string    `json:"name"`
	LastName    string    `json:"last_name,omitempty"`
	Status      string    `json:"status"`
	TenantID    string    `json:"tenant_id"`
	RoleID      string    `json:"role_id"`
	RoleCode    string    `json:"role_code,omitempty"`
	RoleName    string    `json:"role_name,omitempty"`
	IsActive    bool      `json:"is_active"`
	CreatedAt   time.Time `json:"created_at"`
	Permissions []string  `json:"permissions"`
}

type accessRolePayload struct {
	ID              string `json:"id"`
	Code            string `json:"code"`
	DisplayName     string `json:"display_name"`
	Description     string `json:"description,omitempty"`
	IsSystemRole    bool   `json:"is_system_role"`
	PermissionCount int    `json:"permission_count"`
}

type accessMembersResponse struct {
	TenantID           string                `json:"tenant_id"`
	CurrentUserID      string                `json:"current_user_id"`
	CurrentRoleCode    string                `json:"current_role_code,omitempty"`
	CurrentPermissions []string              `json:"current_permissions"`
	Members            []accessMemberPayload `json:"members"`
	Roles              []accessRolePayload   `json:"roles"`
}

type accessMemberMutationResponse struct {
	Message string              `json:"message"`
	Member  accessMemberPayload `json:"member"`
}

type accessPendingRequestPayload struct {
	ID               string  `json:"id"`
	Email            string  `json:"email"`
	Provider         string  `json:"provider"`
	Name             string  `json:"name,omitempty"`
	Company          string  `json:"company,omitempty"`
	UseCase          string  `json:"use_case,omitempty"`
	Source           string  `json:"source,omitempty"`
	AvatarURL        string  `json:"avatar_url,omitempty"`
	Status           string  `json:"status"`
	TenantID         *string `json:"tenant_id,omitempty"`
	RoleCode         *string `json:"role_code,omitempty"`
	CreatedAt        string  `json:"created_at"`
	ReviewedAt       string  `json:"reviewed_at,omitempty"`
	UserID           string  `json:"user_id,omitempty"`
	UserExists       bool    `json:"user_exists"`
	HasAnyMembership bool    `json:"has_any_membership"`
}

type accessPendingListResponse struct {
	TenantID           string                        `json:"tenant_id"`
	CurrentUserID      string                        `json:"current_user_id"`
	CurrentRoleCode    string                        `json:"current_role_code,omitempty"`
	CurrentPermissions []string                      `json:"current_permissions"`
	Requests           []accessPendingRequestPayload `json:"requests"`
}

type accessPendingReviewResponse struct {
	Message string                      `json:"message"`
	Request accessPendingRequestPayload `json:"request"`
}

type accessScopePayload struct {
	HasAccess            bool                  `json:"has_access"`
	HasPendingAccess     bool                  `json:"has_pending_access"`
	ActiveTenantID       string                `json:"active_tenant_id,omitempty"`
	ActiveTenantName     string                `json:"active_tenant_name,omitempty"`
	ActiveTenantSlug     string                `json:"active_tenant_slug,omitempty"`
	ActiveRoleCode       string                `json:"active_role_code,omitempty"`
	ActiveRoleName       string                `json:"active_role_name,omitempty"`
	EffectivePermissions []string              `json:"effective_permissions,omitempty"`
	TeamMemberships      int                   `json:"team_memberships"`
	PendingAccess        *pendingAccessPayload `json:"pending_access,omitempty"`
}

type pendingAccessPayload struct {
	ID         string  `json:"id"`
	Provider   string  `json:"provider"`
	Status     string  `json:"status"`
	TenantID   *string `json:"tenant_id,omitempty"`
	RoleCode   *string `json:"role_code,omitempty"`
	CreatedAt  string  `json:"created_at"`
	ReviewedAt string  `json:"reviewed_at,omitempty"`
}

type publicAccessRequestPayload struct {
	Name        string  `json:"name"`
	Email       string  `json:"email"`
	Company     string  `json:"company,omitempty"`
	UseCase     string  `json:"use_case,omitempty"`
	Source      string  `json:"source,omitempty"`
	PageURL     string  `json:"page_url,omitempty"`
	Language    string  `json:"language,omitempty"`
	UTMSource   string  `json:"utm_source,omitempty"`
	UTMMedium   string  `json:"utm_medium,omitempty"`
	UTMCampaign string  `json:"utm_campaign,omitempty"`
	TenantID    *string `json:"tenant_id,omitempty"`
	RoleCode    *string `json:"role_code,omitempty"`
}

type publicAccessRequestResponse struct {
	Status    string `json:"status"`
	Message   string `json:"message"`
	Email     string `json:"email"`
	Action    string `json:"action"`
	RequestID string `json:"request_id,omitempty"`
}

type signUpRequest struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type signUpResponse struct {
	ID    string `json:"id"`
	Email string `json:"email"`
	Name  string `json:"name"`
}

type signInRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type refreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

type authTokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
}

type authErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message,omitempty"`
}

type pendingAccessReviewRequest struct {
	Action   string `json:"action"`
	TenantID string `json:"tenant_id"`
	RoleCode string `json:"role_code"`
}

type accessMemberRoleUpdateRequest struct {
	TenantID string `json:"tenant_id"`
	RoleCode string `json:"role_code"`
}

type mePayload struct {
	ID              string                `json:"id"`
	Email           string                `json:"email"`
	Name            string                `json:"name"`
	LastName        string                `json:"last_name,omitempty"`
	Status          string                `json:"status"`
	Memberships     []auth.Membership     `json:"memberships"`
	TeamMemberships []auth.TeamMembership `json:"team_memberships"`
	AccessScope     accessScopePayload    `json:"access_scope"`
	Phone           string                `json:"phone,omitempty"`
	AvatarURL       string                `json:"avatar_url,omitempty"`
	LastLogin       *time.Time            `json:"last_login,omitempty"`
	CreatedAt       time.Time             `json:"created_at"`
	UpdatedAt       time.Time             `json:"updated_at"`
}

type input struct {
	config.AuthOAuthClientConfig `json:",inline,omitempty" yaml:",inline,omitempty" toml:",inline,omitempty" mapstructure:"squash,omitempty" binding:"-"`
	Scopes                       []string           `json:"scopes" binding:"required"`
	Endpoint                     oauth2.Endpoint    `json:"endpoint" binding:"-"`
	Code                         oauth2.TokenSource `json:"code" binding:"required"`
}

func RegisterAuthHTTP(r *gin.RouterGroup, container types.IContainer) (gin.IRoutes, error) {
	cfg, ok := container.Config().(*config.Config)
	if !ok {
		return nil, gl.Errorf("invalid config type")
	}
	priv, pub, err := loadOrGenerateKeys(cfg)
	if err != nil {
		return nil, err
	}
	jwtSvc := tokens.NewJWTService(cfg, priv, pub)
	userRepo, err := userstore.NewUserRepository()
	if err != nil {
		return nil, err
	}
	sessRepo, err := userstore.NewSessionRepository()
	if err != nil {
		return nil, err
	}
	var refreshTTL time.Duration
	if cfg.AuthConfig != nil {
		refreshTTL = cfg.AuthConfig.RefreshTokenTTL
	}
	authSvc := services.NewAuthService(userRepo, sessRepo, jwtSvc, gl.GetLoggerZ("auth.service"), refreshTTL)
	h := &authHTTP{
		authSvc:  authSvc,
		jwt:      jwtSvc,
		userRepo: userRepo,
		authCfg:  cfg,
	}
	// a := controllers.NewAuthController(authSvc, userRepo)

	routesMap := map[string]gin.HandlerFunc{
		"POST /auth/sign-up":                  h.signUp,
		"POST /auth/sign-in":                  h.signIn,
		"POST /auth/refresh":                  h.refresh,
		"POST /public/access-request":         h.publicAccessRequest,
		"POST /sign-out":                      h.signOut,
		"GET /me":                             h.me,
		"GET /auth/me":                        h.me,
		"GET /access/members":                 h.accessMembers,
		"GET /access/pending":                 h.accessPending,
		"PATCH /access/pending/:request_id":   h.reviewPendingAccess,
		"PATCH /access/members/:user_id/role": h.updateMemberRole,
		"GET /auth/google/start":              h.googleStart,
		"GET /auth/v1/callback":               h.handleGoogleCallback,
		// "GET /auth/google/oauth2/callback": h.googleCallback,
	}
	// Register routes
	for route, handler := range routesMap {
		parts := strings.SplitN(route, " ", 2)
		if len(parts) != 2 {
			return nil, gl.Errorf("invalid route format: %s", route)
		}
		method := parts[0]
		path := parts[1]
		switch method {
		case "GET":
			r.GET(path, handler)
		case "POST":
			r.POST(path, handler)
		case "PUT":
			r.PUT(path, handler)
		case "DELETE":
			r.DELETE(path, handler)
		case "PATCH":
			r.PATCH(path, handler)
		default:
			return nil, gl.Errorf("unsupported HTTP method: %s", method)
		}
	}
	return r, nil
}

// --- Handlers ---------------------------------------------------------------

// signUp godoc
// @Summary      Sign up a local user
// @Description  Creates a local user with email and password.
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        payload  body      signUpRequest      true  "Sign up payload"
// @Success      201      {object}  signUpResponse
// @Failure      400      {object}  authErrorResponse
// @Router       /api/v1/auth/sign-up [post]
func (h *authHTTP) signUp(c *gin.Context) {
	var req signUpRequest
	if err := c.ShouldBindJSON(&req); err != nil || strings.TrimSpace(req.Email) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
		return
	}
	user, err := h.authSvc.Register(c.Request.Context(), req.Name, req.Email, req.Password)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "could not register user"})
		return
	}
	c.JSON(http.StatusCreated, gin.H{
		"id":    user.ID.String(),
		"email": user.Email,
		"name":  user.Name,
	})
}

// publicAccessRequest godoc
// @Summary      Register a public access request
// @Description  Receives a public onboarding or prospect request and stores it in the pending access queue for manual review.
// @Tags         Public Access
// @Accept       json
// @Produce      json
// @Param        payload  body      publicAccessRequestPayload  true  "Access request payload"
// @Success      202      {object}  publicAccessRequestResponse
// @Success      200      {object}  publicAccessRequestResponse
// @Failure      400      {object}  authErrorResponse
// @Failure      500      {object}  authErrorResponse
// @Router       /api/v1/public/access-request [post]
func (h *authHTTP) publicAccessRequest(c *gin.Context) {
	var req publicAccessRequestPayload
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
		return
	}

	req.Email = strings.ToLower(strings.TrimSpace(req.Email))
	req.Name = strings.TrimSpace(req.Name)
	req.Company = strings.TrimSpace(req.Company)
	req.UseCase = strings.TrimSpace(req.UseCase)
	req.Source = firstNonEmpty(strings.TrimSpace(req.Source), "kubex.world-landing")
	req.PageURL = strings.TrimSpace(req.PageURL)
	req.Language = strings.TrimSpace(req.Language)
	req.UTMSource = strings.TrimSpace(req.UTMSource)
	req.UTMMedium = strings.TrimSpace(req.UTMMedium)
	req.UTMCampaign = strings.TrimSpace(req.UTMCampaign)

	if req.Name == "" || req.Email == "" || !strings.Contains(req.Email, "@") {
		c.JSON(http.StatusBadRequest, gin.H{"error": "name and a valid email are required"})
		return
	}

	if h.userRepo != nil {
		user, err := h.userRepo.FindByEmail(c.Request.Context(), req.Email)
		if err == nil && user != nil {
			memberships, listErr := h.userRepo.ListMemberships(c.Request.Context(), user.ID)
			if listErr == nil && len(memberships) > 0 {
				writeJSON(c.Writer, http.StatusOK, publicAccessRequestResponse{
					Status:  "already_has_access",
					Message: "An active account already exists for this email. Use the workspace sign-in flow.",
					Email:   req.Email,
					Action:  "sign_in",
				})
				return
			}
		}
	}

	store, err := datastore.PendingAccessStore(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "pending access store not available"})
		return
	}

	metadata, _ := json.Marshal(map[string]any{
		"company":      req.Company,
		"use_case":     req.UseCase,
		"source":       req.Source,
		"page_url":     req.PageURL,
		"language":     req.Language,
		"utm_source":   req.UTMSource,
		"utm_medium":   req.UTMMedium,
		"utm_campaign": req.UTMCampaign,
	})

	input := &dsclient.CreatePendingAccessRequestInput{
		Email:              req.Email,
		Provider:           "landing",
		Name:               optionalString(req.Name),
		RequesterIP:        optionalString(c.ClientIP()),
		RequesterUserAgent: optionalString(c.Request.UserAgent()),
		TenantID:           req.TenantID,
		RoleCode:           req.RoleCode,
		Metadata:           metadata,
	}

	item, err := store.Create(c.Request.Context(), input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to register access request"})
		return
	}

	writeJSON(c.Writer, http.StatusAccepted, publicAccessRequestResponse{
		Status:    "pending_review",
		Message:   "Your access request was registered and is waiting for manual review.",
		Email:     req.Email,
		Action:    "wait_for_review",
		RequestID: item.ID,
	})
}

// signIn godoc
// @Summary      Sign in
// @Description  Authenticates a user and returns access and refresh tokens.
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        payload  body      signInRequest      true  "Sign in payload"
// @Success      200      {object}  authTokenResponse
// @Failure      400      {object}  authErrorResponse
// @Failure      401      {object}  authErrorResponse
// @Router       /api/v1/auth/sign-in [post]
func (h *authHTTP) signIn(c *gin.Context) {
	var req signInRequest
	if err := c.ShouldBindJSON(&req); err != nil || strings.TrimSpace(req.Email) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
		return
	}
	ua := c.Request.UserAgent()
	ip := c.Request.RemoteAddr
	access, accessExp, refresh, refreshExp, err := h.authSvc.Login(c.Request.Context(), req.Email, req.Password, ua, ip)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}

	// Set cookies for browser flow
	cookies.SetAuthCookie(c.Writer, cookies.CookieAccessToken, access, accessExp)
	cookies.SetAuthCookie(c.Writer, cookies.CookieRefreshToken, refresh, refreshExp)

	c.JSON(http.StatusOK, gin.H{
		"access_token":  access,
		"refresh_token": refresh,
		"token_type":    "bearer",
	})
}

// googleStart godoc
// @Summary      Start Google OAuth
// @Description  Redirects the browser to the Google consent screen.
// @Tags         Auth
// @Produce      plain
// @Success      302  {string}  string  "Redirect to Google OAuth"
// @Failure      500  {object}  authErrorResponse
// @Router       /api/v1/auth/google/start [get]
func (h *authHTTP) googleStart(c *gin.Context) {
	oauthConf, err := generateOauthConfState(c, h)
	if err != nil {
		return
	}

	next := c.Query("next")
	if strings.TrimSpace(next) == "" || !strings.HasPrefix(next, "/") {
		next = "/"
	}

	googleCfg := h.authCfg.AuthConfig.AuthProvidersConfig.Google.Web

	gl.Debugf("Using Google OAuth2 redirect URL: %s", googleCfg.RedirectURL)

	authURL := oauthConf.AuthCodeURL(next,
		oauth2.AccessTypeOffline,
		oauth2.SetAuthURLParam("prompt", "select_account"),
	)

	c.Redirect(http.StatusFound, authURL)
}

// handleGoogleCallback godoc
// @Summary      Handle Google OAuth callback
// @Description  Exchanges the Google authorization code, issues local tokens, and redirects back to the frontend.
// @Tags         Auth
// @Produce      plain
// @Param        code   query     string  true   "OAuth authorization code"
// @Param        state  query     string  false  "Frontend redirect path"
// @Success      302    {string}  string  "Redirect back to frontend"
// @Failure      400    {object}  authErrorResponse
// @Failure      401    {object}  authErrorResponse
// @Failure      500    {object}  authErrorResponse
// @Router       /api/v1/auth/v1/callback [get]
func (h *authHTTP) handleGoogleCallback(c *gin.Context) {
	oauthConf, err := generateOauthConfState(c, h)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "auth config not loaded"})
		return
	}

	code := c.Query("code")
	if code == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing code"})
		return
	}

	// 1. Troca o Code pelo Token
	tk, err := oauthConf.Exchange(c.Request.Context(), code)
	if err != nil {
		gl.Errorf("Google Exchange Failed: %v", err)
		c.JSON(401, gin.H{"error": "Failed to exchange token"})
		return
	}

	// 2. Valida o ID Token (Segurança Real)
	// O id_token garante que o token foi emitido pelo Google PRA VOCÊ
	rawIDToken, ok := tk.Extra("id_token").(string)
	if !ok {
		c.JSON(500, gin.H{"error": "No id_token field in oauth2 token."})
		return
	}

	email, name, avatarURL, providerUserID, err := extractGoogleIDTokenClaims(c.Request.Context(), rawIDToken, h.authCfg.AuthConfig.AuthProvidersConfig.Google.Web.ClientID)
	if err != nil {
		gl.Debugf("google id_token claims parse failed: %v", err)
	}

	// 3. Emite o SEU Token da Kubex
	accessTk, accessExp, refreshTk, refreshExp, err := h.authSvc.LoginWithOAuth2(c, "google", rawIDToken, c.Request.UserAgent(), c.ClientIP())
	if err != nil {
		gl.Error("Token Issuance Failed", "err", err)
		c.JSON(500, gin.H{"error": "Failed to issue token"})
		return
	}

	cookies.SetAuthCookie(c.Writer, cookies.CookieAccessToken, accessTk, accessExp)
	cookies.SetAuthCookie(c.Writer, cookies.CookieRefreshToken, refreshTk, refreshExp)

	h.ensurePendingAccessRequest(c, "google", email, name, avatarURL, providerUserID)

	next := c.Query("state")
	if strings.TrimSpace(next) == "" || !strings.HasPrefix(next, "/") {
		next = "/"
	}

	// Redirect back to the frontend (cookies already set)
	c.Redirect(http.StatusFound, buildFrontendRedirect(next))
}

// refresh godoc
// @Summary      Refresh tokens
// @Description  Exchanges a valid refresh token for a new access and refresh token pair.
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        payload  body      refreshRequest     false  "Refresh payload"
// @Success      200      {object}  authTokenResponse
// @Failure      400      {object}  authErrorResponse
// @Failure      401      {object}  authErrorResponse
// @Router       /api/v1/auth/refresh [post]
func (h *authHTTP) refresh(c *gin.Context) {
	if c.Request.Method != http.MethodPost {
		c.JSON(http.StatusMethodNotAllowed, gin.H{"error": "method not allowed"})
		return
	}
	var req struct {
		RefreshToken string `json:"refresh_token"`
	}
	_ = c.ShouldBindJSON(&req)
	if rt, ok := cookies.GetCookieValue(c.Request, cookies.CookieRefreshToken); ok && rt != "" {
		req.RefreshToken = rt
	}
	if strings.TrimSpace(req.RefreshToken) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
		return
	}
	ua := c.Request.UserAgent()
	ip := c.Request.RemoteAddr
	access, accessExp, refresh, refreshExp, err := h.authSvc.Refresh(c.Request.Context(), req.RefreshToken, ua, ip)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid refresh token"})
		return
	}
	cookies.SetAuthCookie(c.Writer, cookies.CookieAccessToken, access, accessExp)
	cookies.SetAuthCookie(c.Writer, cookies.CookieRefreshToken, refresh, refreshExp)
	c.JSON(http.StatusOK, gin.H{
		"access_token":  access,
		"refresh_token": refresh,
		"token_type":    "bearer",
	})
}

// signOut godoc
// @Summary      Sign out
// @Description  Revokes the current refresh token and clears auth cookies.
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        payload  body      refreshRequest     false  "Sign out payload"
// @Success      204      {string}  string  "No Content"
// @Failure      400      {object}  authErrorResponse
// @Failure      401      {object}  authErrorResponse
// @Router       /api/v1/sign-out [post]
func (h *authHTTP) signOut(c *gin.Context) {
	if c.Request.Method != http.MethodPost {
		c.JSON(http.StatusMethodNotAllowed, gin.H{"error": "method not allowed"})
		return
	}
	var req struct {
		RefreshToken string `json:"refresh_token"`
	}
	_ = c.ShouldBindJSON(&req)
	if rt, ok := cookies.GetCookieValue(c.Request, cookies.CookieRefreshToken); ok && rt != "" {
		req.RefreshToken = rt
	}
	if strings.TrimSpace(req.RefreshToken) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
		return
	}
	if err := h.authSvc.Logout(c.Request.Context(), req.RefreshToken); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid refresh token"})
		return
	}
	// Clear cookies on logout
	cookies.ClearAuthCookie(c.Writer, cookies.CookieAccessToken)
	cookies.ClearAuthCookie(c.Writer, cookies.CookieRefreshToken)
	c.Status(http.StatusNoContent)
}

// me godoc
// @Summary      Get current user
// @Description  Returns the authenticated user profile plus memberships, team memberships, and access scope.
// @Tags         Auth
// @Security     BearerAuth
// @Produce      json
// @Success      200  {object}  map[string]interface{}
// @Failure      401  {object}  authErrorResponse
// @Router       /api/v1/me [get]
// @Router       /api/v1/auth/me [get]
func (h *authHTTP) me(c *gin.Context) {
	claims, err := h.validateAuthHeader(c.Request)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	userID, err := uuid.Parse(claims.Sub)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	user, err := h.userRepo.FindByID(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	memberships, _, _ := h.loadMembershipsWithPermissions(c.Request.Context(), user.ID)
	teamMemberships, _ := h.userRepo.ListTeamMemberships(c.Request.Context(), user.ID)
	scope := h.resolveAccessScope(c.Request.Context(), user, memberships, teamMemberships)

	writeJSON(c.Writer, http.StatusOK, mePayload{
		ID:              user.ID.String(),
		Email:           user.Email,
		Name:            user.Name,
		LastName:        user.LastName,
		Status:          user.Status,
		Memberships:     memberships,
		TeamMemberships: teamMemberships,
		AccessScope:     scope,
		Phone:           user.Phone,
		AvatarURL:       user.AvatarURL,
		LastLogin:       user.LastLogin,
		CreatedAt:       user.CreatedAt,
		UpdatedAt:       user.UpdatedAt,
	})
}

// accessMembers godoc
// @Summary      List tenant access members
// @Description  Returns members, effective permissions, and role catalog for the active tenant scope.
// @Tags         Access
// @Security     BearerAuth
// @Produce      json
// @Param        tenant_id  query     string  false  "Tenant ID"
// @Success      200        {object}  accessMembersResponse
// @Failure      400        {object}  authErrorResponse
// @Failure      401        {object}  authErrorResponse
// @Failure      403        {object}  authErrorResponse
// @Failure      500        {object}  authErrorResponse
// @Router       /api/v1/access/members [get]
func (h *authHTTP) accessMembers(c *gin.Context) {
	claims, err := h.validateAuthHeader(c.Request)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	userID, err := uuid.Parse(claims.Sub)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	memberships, currentMembership, err := h.loadMembershipsWithPermissions(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	tenantID := strings.TrimSpace(c.Query("tenant_id"))
	if tenantID == "" && currentMembership != nil {
		tenantID = currentMembership.TenantID.String()
	}
	if tenantID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing tenant scope"})
		return
	}

	requestedMembership := pickMembershipByTenantID(memberships, tenantID)
	if requestedMembership == nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "tenant scope not allowed"})
		return
	}
	if !hasPermission(requestedMembership.Permissions, "user.read") {
		c.JSON(http.StatusForbidden, gin.H{"error": "insufficient permissions"})
		return
	}

	members, err := loadTenantMembers(c.Request.Context(), tenantID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load tenant members"})
		return
	}

	roles, err := loadAccessRoles(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load roles"})
		return
	}

	writeJSON(c.Writer, http.StatusOK, accessMembersResponse{
		TenantID:           tenantID,
		CurrentUserID:      userID.String(),
		CurrentRoleCode:    requestedMembership.RoleCode,
		CurrentPermissions: requestedMembership.Permissions,
		Members:            members,
		Roles:              roles,
	})
}

// accessPending godoc
// @Summary      List pending access requests
// @Description  Returns the pending access queue for the active tenant scope.
// @Tags         Access
// @Security     BearerAuth
// @Produce      json
// @Param        tenant_id  query     string  false  "Tenant ID"
// @Success      200        {object}  accessPendingListResponse
// @Failure      400        {object}  authErrorResponse
// @Failure      401        {object}  authErrorResponse
// @Failure      403        {object}  authErrorResponse
// @Failure      500        {object}  authErrorResponse
// @Router       /api/v1/access/pending [get]
func (h *authHTTP) accessPending(c *gin.Context) {
	claims, err := h.validateAuthHeader(c.Request)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	userID, err := uuid.Parse(claims.Sub)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	memberships, currentMembership, err := h.loadMembershipsWithPermissions(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	tenantID := strings.TrimSpace(c.Query("tenant_id"))
	if tenantID == "" && currentMembership != nil {
		tenantID = currentMembership.TenantID.String()
	}
	if tenantID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing tenant scope"})
		return
	}

	requestedMembership := pickMembershipByTenantID(memberships, tenantID)
	if requestedMembership == nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "tenant scope not allowed"})
		return
	}
	if !hasPermission(requestedMembership.Permissions, "user.read") {
		c.JSON(http.StatusForbidden, gin.H{"error": "insufficient permissions"})
		return
	}

	requests, err := loadPendingAccessRequests(c.Request.Context(), h.userRepo, tenantID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load pending access requests"})
		return
	}

	writeJSON(c.Writer, http.StatusOK, accessPendingListResponse{
		TenantID:           tenantID,
		CurrentUserID:      userID.String(),
		CurrentRoleCode:    requestedMembership.RoleCode,
		CurrentPermissions: requestedMembership.Permissions,
		Requests:           requests,
	})
}

// reviewPendingAccess godoc
// @Summary      Review a pending access request
// @Description  Approves or rejects a pending access request for the target tenant.
// @Tags         Access
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        request_id  path      string                       true   "Pending request ID"
// @Param        payload     body      pendingAccessReviewRequest   true   "Review payload"
// @Success      200         {object}  accessPendingReviewResponse
// @Failure      400         {object}  authErrorResponse
// @Failure      401         {object}  authErrorResponse
// @Failure      403         {object}  authErrorResponse
// @Failure      404         {object}  authErrorResponse
// @Failure      500         {object}  authErrorResponse
// @Router       /api/v1/access/pending/{request_id} [patch]
func (h *authHTTP) reviewPendingAccess(c *gin.Context) {
	claims, err := h.validateAuthHeader(c.Request)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	actorUserID, err := uuid.Parse(claims.Sub)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	requestID := strings.TrimSpace(c.Param("request_id"))
	if requestID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing request id"})
		return
	}

	var req pendingAccessReviewRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
		return
	}

	req.Action = strings.ToLower(strings.TrimSpace(req.Action))
	req.TenantID = strings.TrimSpace(req.TenantID)
	req.RoleCode = strings.TrimSpace(req.RoleCode)
	if req.Action != "approve" && req.Action != "reject" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "action must be approve or reject"})
		return
	}
	if req.TenantID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "tenant_id is required"})
		return
	}

	memberships, _, err := h.loadMembershipsWithPermissions(c.Request.Context(), actorUserID)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	requestedMembership := pickMembershipByTenantID(memberships, req.TenantID)
	if requestedMembership == nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "tenant scope not allowed"})
		return
	}
	if !hasPermission(requestedMembership.Permissions, "user.invite") {
		c.JSON(http.StatusForbidden, gin.H{"error": "insufficient permissions"})
		return
	}

	store, err := datastore.PendingAccessStore(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "pending access store not available"})
		return
	}

	item, err := store.GetByID(c.Request.Context(), requestID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load pending access request"})
		return
	}
	if item == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "pending access request not found"})
		return
	}
	if string(item.Status) != "pending" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "pending access request is no longer pending"})
		return
	}

	message := "pending access request rejected"
	if req.Action == "approve" {
		roleCode := firstNonEmpty(req.RoleCode, stringPtrValue(item.RoleCode), "viewer")
		user, findErr := h.userRepo.FindByEmail(c.Request.Context(), strings.TrimSpace(item.Email))
		if findErr != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to resolve pending access user"})
			return
		}
		if user == nil {
			c.JSON(http.StatusUnprocessableEntity, gin.H{"error": "pending access request is not linked to a user yet; issue an invite instead"})
			return
		}
		if err := ensureTenantMemberAccess(c.Request.Context(), user.ID.String(), req.TenantID, roleCode); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to grant tenant access"})
			return
		}
		message = "pending access request approved"
	}

	if err := updatePendingAccessRequestReview(c.Request.Context(), item.ID, req.Action, actorUserID.String()); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update pending access request"})
		return
	}

	item, err = store.GetByID(c.Request.Context(), requestID)
	if err != nil || item == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "request updated but reload failed"})
		return
	}

	payload, err := loadPendingAccessRequestPayload(c.Request.Context(), h.userRepo, item)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "request updated but payload reload failed"})
		return
	}

	writeJSON(c.Writer, http.StatusOK, accessPendingReviewResponse{
		Message: message,
		Request: payload,
	})
}

// updateMemberRole godoc
// @Summary      Update a tenant member role
// @Description  Reassigns the role of a member inside the active tenant scope.
// @Tags         Access
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        user_id  path      string                         true  "Member user ID"
// @Param        payload  body      accessMemberRoleUpdateRequest  true  "Role mutation payload"
// @Success      200      {object}  accessMemberMutationResponse
// @Failure      400      {object}  authErrorResponse
// @Failure      401      {object}  authErrorResponse
// @Failure      403      {object}  authErrorResponse
// @Failure      404      {object}  authErrorResponse
// @Failure      500      {object}  authErrorResponse
// @Router       /api/v1/access/members/{user_id}/role [patch]
func (h *authHTTP) updateMemberRole(c *gin.Context) {
	claims, err := h.validateAuthHeader(c.Request)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	actorUserID, err := uuid.Parse(claims.Sub)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	targetUserID := strings.TrimSpace(c.Param("user_id"))
	if targetUserID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing target user"})
		return
	}

	var req accessMemberRoleUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
		return
	}

	req.TenantID = strings.TrimSpace(req.TenantID)
	req.RoleCode = strings.TrimSpace(req.RoleCode)
	if req.TenantID == "" || req.RoleCode == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "tenant_id and role_code are required"})
		return
	}
	if actorUserID.String() == targetUserID {
		c.JSON(http.StatusBadRequest, gin.H{"error": "self role changes are not allowed in this slice"})
		return
	}

	memberships, _, err := h.loadMembershipsWithPermissions(c.Request.Context(), actorUserID)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	requestedMembership := pickMembershipByTenantID(memberships, req.TenantID)
	if requestedMembership == nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "tenant scope not allowed"})
		return
	}
	if !hasPermission(requestedMembership.Permissions, "role.manage") {
		c.JSON(http.StatusForbidden, gin.H{"error": "insufficient permissions"})
		return
	}

	if err := updateTenantMemberRole(c.Request.Context(), targetUserID, req.TenantID, req.RoleCode); err != nil {
		switch {
		case strings.Contains(err.Error(), "role not found"):
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		case strings.Contains(err.Error(), "membership not found"):
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update tenant member role"})
		}
		return
	}

	members, err := loadTenantMembers(c.Request.Context(), req.TenantID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "role updated but member reload failed"})
		return
	}

	for _, member := range members {
		if member.ID == targetUserID {
			writeJSON(c.Writer, http.StatusOK, accessMemberMutationResponse{
				Message: "member role updated",
				Member:  member,
			})
			return
		}
	}

	c.JSON(http.StatusNotFound, gin.H{"error": "updated member not found"})
}

func (h *authHTTP) resolveAccessScope(
	ctx context.Context,
	user *auth.User,
	memberships []auth.Membership,
	teamMemberships []auth.TeamMembership,
) accessScopePayload {
	scope := accessScopePayload{
		HasAccess:       len(memberships) > 0,
		TeamMemberships: len(teamMemberships),
	}

	if active := pickActiveMembership(memberships); active != nil {
		scope.ActiveTenantID = active.TenantID.String()
		scope.ActiveTenantName = active.TenantName
		scope.ActiveTenantSlug = active.TenantSlug
		scope.ActiveRoleCode = active.RoleCode
		scope.ActiveRoleName = active.RoleName
		scope.EffectivePermissions = active.Permissions
	}

	if scope.HasAccess || user == nil || strings.TrimSpace(user.Email) == "" {
		return scope
	}

	pending, err := loadPendingAccessByEmail(ctx, strings.TrimSpace(user.Email))
	if err != nil || pending == nil {
		return scope
	}

	scope.HasPendingAccess = true
	scope.PendingAccess = pending
	return scope
}

func pickActiveMembership(memberships []auth.Membership) *auth.Membership {
	for i := range memberships {
		if memberships[i].IsActive {
			return &memberships[i]
		}
	}
	if len(memberships) == 0 {
		return nil
	}
	return &memberships[0]
}

func pickMembershipByTenantID(memberships []auth.Membership, tenantID string) *auth.Membership {
	for i := range memberships {
		if memberships[i].TenantID.String() == tenantID {
			return &memberships[i]
		}
	}
	return nil
}

func (h *authHTTP) loadMembershipsWithPermissions(
	ctx context.Context,
	userID uuid.UUID,
) ([]auth.Membership, *auth.Membership, error) {
	memberships, err := h.userRepo.ListMemberships(ctx, userID)
	if err != nil {
		return nil, nil, err
	}

	if permissionMap, err := h.userRepo.ListMembershipPermissions(ctx, userID); err == nil {
		for i := range memberships {
			memberships[i].Permissions = permissionMap[memberships[i].TenantID]
		}
	}

	return memberships, pickActiveMembership(memberships), nil
}

func hasPermission(permissions []string, permission string) bool {
	for _, current := range permissions {
		if current == "*" || current == permission {
			return true
		}
	}
	return false
}

func loadPendingAccessByEmail(ctx context.Context, email string) (*pendingAccessPayload, error) {
	store, err := datastore.PendingAccessStore(ctx)
	if err != nil {
		return nil, err
	}

	res, err := store.List(ctx, &dsclient.PendingAccessFilters{
		Email: optionalString(strings.ToLower(strings.TrimSpace(email))),
		Page:  1,
		Limit: 10,
	})
	if err != nil || res == nil || len(res.Data) == 0 {
		return nil, err
	}

	var item *dsclient.PendingAccessRequest
	for i := range res.Data {
		if string(res.Data[i].Status) == "pending" {
			item = &res.Data[i]
			break
		}
	}
	if item == nil {
		return nil, nil
	}

	payload := &pendingAccessPayload{
		ID:        item.ID,
		Provider:  item.Provider,
		Status:    string(item.Status),
		TenantID:  item.TenantID,
		RoleCode:  item.RoleCode,
		CreatedAt: item.CreatedAt.UTC().Format(time.RFC3339),
	}
	if item.ReviewedAt != nil {
		payload.ReviewedAt = item.ReviewedAt.UTC().Format(time.RFC3339)
	}
	return payload, nil
}

func loadPendingAccessRequests(ctx context.Context, userRepo userstore.UserRepository, tenantID string) ([]accessPendingRequestPayload, error) {
	store, err := datastore.PendingAccessStore(ctx)
	if err != nil {
		return nil, err
	}

	res, err := store.List(ctx, &dsclient.PendingAccessFilters{
		Page:  1,
		Limit: 50,
	})
	if err != nil {
		return nil, err
	}
	if res == nil || len(res.Data) == 0 {
		return []accessPendingRequestPayload{}, nil
	}

	requests := make([]accessPendingRequestPayload, 0, len(res.Data))
	for i := range res.Data {
		item := res.Data[i]
		if string(item.Status) != "pending" {
			continue
		}
		if item.TenantID != nil && strings.TrimSpace(stringPtrValue(item.TenantID)) != "" && strings.TrimSpace(stringPtrValue(item.TenantID)) != tenantID {
			continue
		}
		payload, err := loadPendingAccessRequestPayload(ctx, userRepo, &item)
		if err != nil {
			return nil, err
		}
		requests = append(requests, payload)
	}

	return requests, nil
}

func loadPendingAccessRequestPayload(ctx context.Context, userRepo userstore.UserRepository, item *dsclient.PendingAccessRequest) (accessPendingRequestPayload, error) {
	payload := accessPendingRequestPayload{
		ID:        item.ID,
		Email:     item.Email,
		Provider:  item.Provider,
		Name:      stringPtrValue(item.Name),
		AvatarURL: stringPtrValue(item.AvatarURL),
		Status:    string(item.Status),
		TenantID:  item.TenantID,
		RoleCode:  item.RoleCode,
		CreatedAt: item.CreatedAt.UTC().Format(time.RFC3339),
	}
	if item.ReviewedAt != nil {
		payload.ReviewedAt = item.ReviewedAt.UTC().Format(time.RFC3339)
	}
	if len(item.Metadata) > 0 {
		var metadata map[string]any
		if err := json.Unmarshal(item.Metadata, &metadata); err == nil {
			if value, ok := metadata["company"].(string); ok {
				payload.Company = strings.TrimSpace(value)
			}
			if value, ok := metadata["use_case"].(string); ok {
				payload.UseCase = strings.TrimSpace(value)
			}
			if value, ok := metadata["source"].(string); ok {
				payload.Source = strings.TrimSpace(value)
			}
		}
	}

	if userRepo == nil {
		return payload, nil
	}

	user, err := userRepo.FindByEmail(ctx, strings.TrimSpace(item.Email))
	if err != nil {
		return payload, nil
	}
	if user == nil {
		return payload, nil
	}

	payload.UserID = user.ID.String()
	payload.UserExists = true
	memberships, err := userRepo.ListMemberships(ctx, user.ID)
	if err == nil && len(memberships) > 0 {
		payload.HasAnyMembership = true
	}
	return payload, nil
}

func loadTenantMembers(ctx context.Context, tenantID string) ([]accessMemberPayload, error) {
	conn, err := datastore.Connection(ctx)
	if err != nil {
		return nil, err
	}
	pgExec, err := dsclient.GetPGExecutor(ctx, conn)
	if err != nil {
		return nil, err
	}

	const q = `
		SELECT
			u.id,
			u.email,
			COALESCE(u.name, ''),
			COALESCE(u.last_name, ''),
			COALESCE(u.status, ''),
			tm.tenant_id,
			tm.role_id,
			COALESCE(r.code, ''),
			COALESCE(r.display_name, ''),
			tm.is_active,
			tm.created_at,
			COALESCE(
				array_agg(DISTINCT p.code ORDER BY p.code)
					FILTER (WHERE rp.value = true AND p.code IS NOT NULL),
				ARRAY[]::text[]
			) AS permissions
		FROM tenant_membership tm
		JOIN "user" u ON u.id = tm.user_id
		JOIN role r ON r.id = tm.role_id
		LEFT JOIN role_permission rp ON rp.role_id = r.id
		LEFT JOIN permission p ON p.id = rp.permission_id
		WHERE tm.tenant_id = $1
		GROUP BY
			u.id, u.email, u.name, u.last_name, u.status,
			tm.tenant_id, tm.role_id, r.code, r.display_name, tm.is_active, tm.created_at
		ORDER BY tm.created_at ASC, u.email ASC`

	rows, err := pgExec.Query(ctx, q, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []accessMemberPayload
	for rows.Next() {
		var item accessMemberPayload
		if err := rows.Scan(
			&item.ID,
			&item.Email,
			&item.Name,
			&item.LastName,
			&item.Status,
			&item.TenantID,
			&item.RoleID,
			&item.RoleCode,
			&item.RoleName,
			&item.IsActive,
			&item.CreatedAt,
			&item.Permissions,
		); err != nil {
			return nil, err
		}
		result = append(result, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return result, nil
}

func loadAccessRoles(ctx context.Context) ([]accessRolePayload, error) {
	conn, err := datastore.Connection(ctx)
	if err != nil {
		return nil, err
	}
	pgExec, err := dsclient.GetPGExecutor(ctx, conn)
	if err != nil {
		return nil, err
	}

	const q = `
		SELECT
			r.id,
			r.code,
			r.display_name,
			COALESCE(r.description, ''),
			COALESCE(r.is_system_role, false),
			COUNT(rp.permission_id) FILTER (WHERE rp.value = true) AS permission_count
		FROM role r
		LEFT JOIN role_permission rp ON rp.role_id = r.id
		GROUP BY r.id, r.code, r.display_name, r.description, r.is_system_role
		ORDER BY r.code ASC`

	rows, err := pgExec.Query(ctx, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []accessRolePayload
	for rows.Next() {
		var item accessRolePayload
		if err := rows.Scan(
			&item.ID,
			&item.Code,
			&item.DisplayName,
			&item.Description,
			&item.IsSystemRole,
			&item.PermissionCount,
		); err != nil {
			return nil, err
		}
		result = append(result, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return result, nil
}

func updateTenantMemberRole(ctx context.Context, userID, tenantID, roleCode string) error {
	conn, err := datastore.Connection(ctx)
	if err != nil {
		return err
	}
	pgExec, err := dsclient.GetPGExecutor(ctx, conn)
	if err != nil {
		return err
	}

	var roleID string
	if err := pgExec.QueryRow(ctx, `SELECT id FROM role WHERE code = $1`, roleCode).Scan(&roleID); err != nil {
		return gl.Errorf("role not found: %s", roleCode)
	}

	tag, err := pgExec.Exec(ctx, `
		UPDATE tenant_membership
		SET role_id = $1, updated_at = now()
		WHERE user_id = $2 AND tenant_id = $3
	`, roleID, userID, tenantID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return gl.Errorf("membership not found for user %s in tenant %s", userID, tenantID)
	}
	return nil
}

func ensureTenantMemberAccess(ctx context.Context, userID, tenantID, roleCode string) error {
	conn, err := datastore.Connection(ctx)
	if err != nil {
		return err
	}
	pgExec, err := dsclient.GetPGExecutor(ctx, conn)
	if err != nil {
		return err
	}

	roleCode = firstNonEmpty(roleCode, "viewer")
	var roleID string
	if err := pgExec.QueryRow(ctx, `SELECT id FROM role WHERE code = $1`, roleCode).Scan(&roleID); err != nil {
		if err := pgExec.QueryRow(ctx, `SELECT id FROM role WHERE code = $1`, "viewer").Scan(&roleID); err != nil {
			return gl.Errorf("failed to resolve role '%s' and fallback 'viewer': %v", roleCode, err)
		}
	}

	_, err = pgExec.Exec(ctx, `
		INSERT INTO tenant_membership (user_id, tenant_id, role_id, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, true, now(), now())
		ON CONFLICT (user_id, tenant_id)
		DO UPDATE SET role_id = EXCLUDED.role_id, is_active = EXCLUDED.is_active, updated_at = now()
	`, userID, tenantID, roleID)
	return err
}

func updatePendingAccessRequestReview(ctx context.Context, requestID, action, reviewedBy string) error {
	conn, err := datastore.Connection(ctx)
	if err != nil {
		return err
	}
	pgExec, err := dsclient.GetPGExecutor(ctx, conn)
	if err != nil {
		return err
	}

	status := "rejected"
	if strings.EqualFold(strings.TrimSpace(action), "approve") {
		status = "approved"
	}

	_, err = pgExec.Exec(ctx, `
		UPDATE pending_access_requests
		SET status = $1, reviewed_by = $2, reviewed_at = now(), updated_at = now()
		WHERE id = $3
	`, status, firstNonEmpty(reviewedBy), requestID)
	return err
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}

func stringPtrValue(value *string) string {
	if value == nil {
		return ""
	}
	return strings.TrimSpace(*value)
}

// --- Helpers ----------------------------------------------------------------

func (h *authHTTP) validateAuthHeader(r *http.Request) (*tokens.Claims, error) {
	if raw, ok := cookies.GetCookieValue(r, cookies.CookieAccessToken); ok {
		if claims, err := h.jwt.ValidateAccessToken(raw); err == nil {
			return claims, nil
		}
	}
	authz := r.Header.Get("Authorization")
	parts := strings.Fields(authz)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "bearer") {
		return nil, errors.New("missing bearer")
	}
	return h.jwt.ValidateAccessToken(parts[1])
}

func loadOrGenerateKeys(cfg *config.Config) (*rsa.PrivateKey, *rsa.PublicKey, error) {
	var (
		// certservice;
		certService = crt.NewCertServiceType(
			os.ExpandEnv(cfg.ServerConfig.Runtime.PrivKeyPath),
			os.ExpandEnv(cfg.ServerConfig.Runtime.PubCertKeyPath),
		)
	)
	// Descriptografa as chaves se existirem
	rsaPrivateKey, err := certService.DecryptPrivateKey(nil)
	if err == nil && rsaPrivateKey != nil {
		return rsaPrivateKey, &rsaPrivateKey.PublicKey, nil
	}

	// Dev fallback: gera uma chave temporária
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, nil, err
	}
	pub := &priv.PublicKey
	gl.Log("warn", "Auth keys not provided, generated ephemeral RSA keys for dev")
	return priv, pub, nil
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func defineRedirectURL(r *http.Request, redirectURIs []string, vList []string) (string, error) {
	var origin *url.URL
	var err error

	origin, err = url.Parse(
		kbxGet.ValueOrIf(
			len(r.Header.Get("Origin")) > 0,
			r.Header.Get("Origin"),
			"http://"+kbxGet.ValOrType(r.Host, "localhost:5000"),
		),
	)
	if err != nil {
		return "", errors.New("invalid origin header")
	}
	origin.RawQuery = ""
	origin.Fragment = ""
	vList = kbxGet.ValueOrIf(len(vList) == 0, []string{"hostname", "scheme", "port", "path"}, vList)
	for _, uri := range redirectURIs {
		rURL, err := url.Parse(uri)
		if err != nil {
			continue
		}
		rURL.RawQuery = ""
		rURL.Fragment = ""
		if compareURLs(origin, rURL, vList...) {
			return uri, nil
		}
	}
	return "", errors.New("no matching redirect URI found")
}

func compareURLs(u1, u2 *url.URL, opts ...string) bool {
	var ok *bool
	ok = kbxGet.ValueOrIf(kbxIs.ArrayObj("hostname", opts), kbxGet.BlPtr(u1.Hostname() == u2.Hostname()), nil)
	ok = kbxGet.ValueOrIf(kbxIs.ArrayObj("scheme", opts), kbxGet.BlPtr(u1.Scheme == u2.Scheme), ok)
	ok = kbxGet.ValueOrIf(kbxIs.ArrayObj("port", opts), kbxGet.BlPtr(u1.Port() == u2.Port()), ok)
	ok = kbxGet.ValueOrIf(kbxIs.ArrayObj("path", opts), kbxGet.BlPtr(u1.Path == u2.Path), ok)
	ok = kbxGet.ValueOrIf(kbxIs.ArrayObj("request_uri", opts), kbxGet.BlPtr(u1.RequestURI() == u2.RequestURI()), ok)
	if ok != nil {
		return *ok
	}
	// Default: compare all
	return compareURLs(u1, u2, "hostname", "scheme", "port", "path", "request_uri")
}

func generateOauthConfState(c *gin.Context, h *authHTTP) (*oauth2.Config, error) {
	if h.authCfg == nil || h.authCfg.AuthConfig == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "auth config not loaded"})
		return nil, gl.Error("auth config not loaded")
	}

	var data []byte
	googleVdrCfgMpr := kbxTool.NewEmptyMapperType[kbx.VendorAuthConfig](kbxGet.EnvOr("KUBEX_GOOGLE_CREDENTIALS_PATH", os.ExpandEnv(defaults.DefaultGoogleAuthClientPath)))
	googleCfg := h.authCfg.AuthConfig.AuthProvidersConfig.Google.Web
	if googleCfg.ClientID == "" || googleCfg.ClientSecret == "" || len(googleCfg.RedirectURIs) == 0 {
		googleVdrCfgPtr, err := kbx.LoadConfigOrDefault[kbx.VendorAuthConfig](kbxGet.EnvOr("KUBEX_GOOGLE_CREDENTIALS_PATH", os.ExpandEnv(defaults.DefaultGoogleAuthClientPath)), true)
		if err != nil && googleVdrCfgPtr == nil {
			gl.Debugf("google oauth config load failed: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "google oauth not configured"})
			return nil, err
		}
		googleVdrCfgMpr.SetValue(googleVdrCfgPtr)
	} else {
		vdrConfig := kbxLoad.NewVendorAuthConfig("")
		googleVdrCfgMpr.SetValue(&vdrConfig)
	}
	h.authCfg.AuthConfig.AuthProvidersConfig.Google.Web = googleCfg
	if len(googleCfg.RedirectURL) == 0 {
		r, err := defineRedirectURL(c.Request, googleCfg.RedirectURIs, []string{"hostname", "scheme", "port"})
		if err != nil {
			gl.Errorf("google oauth redirect URL define failed: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "google oauth redirect URL not defined"})
			return nil, err
		}
		c.Set("google_oauth_redirect_url", r)
		googleCfg.RedirectURL = r
		googleVdrCfg := googleVdrCfgMpr.GetValue()
		googleVdrCfg.Web = googleCfg
		googleVdrCfgMpr.SetValue(googleVdrCfg)
	}

	if len(googleCfg.Scopes) == 0 {
		googleCfg.Scopes = []string{"openid", "email", "profile"}
	}

	// state param to redirect back after auth
	next := c.Query("next")

	data, _ = googleVdrCfgMpr.Serialize("json")
	oauthConf, err := google.ConfigFromJSON(data, googleCfg.Scopes...)
	if err != nil {
		gl.Errorf("google oauth config parse failed: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "google oauth failed"})
		return nil, err
	}

	stateUUID := uuid.New().String()
	if next != "" {
		stateUUID = stateUUID + "|" + url.QueryEscape(next)
	}

	return oauthConf, nil
}

func extractGoogleIDTokenClaims(ctx context.Context, rawToken, clientID string) (string, string, string, string, error) {
	if strings.TrimSpace(rawToken) == "" || strings.TrimSpace(clientID) == "" {
		return "", "", "", "", errors.New("missing token or client id")
	}
	payload, err := idtoken.Validate(ctx, rawToken, clientID)
	if err != nil {
		return "", "", "", "", err
	}
	email, _ := payload.Claims["email"].(string)
	name, _ := payload.Claims["name"].(string)
	picture, _ := payload.Claims["picture"].(string)
	subject, _ := payload.Claims["sub"].(string)
	return strings.ToLower(strings.TrimSpace(email)), strings.TrimSpace(name), strings.TrimSpace(picture), strings.TrimSpace(subject), nil
}

func (h *authHTTP) ensurePendingAccessRequest(c *gin.Context, provider, email, name, avatarURL, providerUserID string) {
	if h == nil || h.userRepo == nil {
		return
	}
	email = strings.TrimSpace(strings.ToLower(email))
	if email == "" {
		return
	}
	user, err := h.userRepo.FindByEmail(c.Request.Context(), email)
	if err != nil || user == nil {
		return
	}
	memberships, err := h.userRepo.ListMemberships(c.Request.Context(), user.ID)
	if err == nil && len(memberships) > 0 {
		return
	}

	store, err := datastore.PendingAccessStore(c.Request.Context())
	if err != nil {
		gl.Debugf("pending access store not available: %v", err)
		return
	}

	input := &dsclient.CreatePendingAccessRequestInput{
		Email:              email,
		Provider:           strings.TrimSpace(provider),
		ProviderUserID:     optionalString(providerUserID),
		Name:               optionalString(name),
		AvatarURL:          optionalString(avatarURL),
		RequesterIP:        optionalString(c.ClientIP()),
		RequesterUserAgent: optionalString(c.Request.UserAgent()),
	}

	if _, err := store.Create(c.Request.Context(), input); err != nil {
		gl.Debugf("pending access request create failed: %v", err)
	}
}

func buildFrontendRedirect(next string) string {
	base := resolveFrontendBaseURL()
	if base == "" {
		return next
	}
	return base + next
}

func resolveFrontendBaseURL() string {
	env := kbxGet.EnvOr("KUBEX_ENV", kbxGet.EnvOr("ENV", "development"))
	base := strings.TrimSpace(kbxGet.EnvOr("INVITE_BASE_URL", ""))
	if base == "" {
		base = strings.TrimSpace(kbxGet.EnvOr("KUBEX_PUBLIC_URL", ""))
	}
	if base == "" {
		base = strings.TrimSpace(kbxGet.EnvOr("KUBEX_GNYX_PUBLIC_URL", ""))
	}
	if base == "" {
		base = kbxGet.ValueOrIf(env == "production", "https://gnyx.kubex.world", "http://localhost:5000")
	}
	return strings.TrimRight(base, "/")
}

func optionalString(value string) *string {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil
	}
	return &value
}

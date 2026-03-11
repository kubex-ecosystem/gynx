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
		"POST /sign-out":                      h.signOut,
		"GET /me":                             h.me,
		"GET /auth/me":                        h.me,
		"GET /access/members":                 h.accessMembers,
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
		default:
			return nil, gl.Errorf("unsupported HTTP method: %s", method)
		}
	}
	return r, nil
}

// --- Handlers ---------------------------------------------------------------

func (h *authHTTP) signUp(c *gin.Context) {
	var req struct {
		Name     string `json:"name"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}
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

func (h *authHTTP) signIn(c *gin.Context) {
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
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

// GET /auth/google/start -> redireciona para consent
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

// GET /auth/google/callback (redirect_uri)
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

	var req struct {
		TenantID string `json:"tenant_id"`
		RoleCode string `json:"role_code"`
	}
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
		base = kbxGet.ValueOrIf(env == "production", "https://app.kubex.world", "http://localhost:5000")
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

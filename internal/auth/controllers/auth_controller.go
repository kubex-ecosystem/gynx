// Package controllers implements the authentication controller.
package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/kubex-ecosystem/gnyx/internal/config"
	services "github.com/kubex-ecosystem/gnyx/internal/dsclient/datastore/session_store"
	userstore "github.com/kubex-ecosystem/gnyx/internal/dsclient/datastore/user_store"
	"github.com/kubex-ecosystem/gnyx/internal/features/cookies"
	"github.com/kubex-ecosystem/logz"
	"golang.org/x/oauth2"
)

type AuthController struct {
	auth services.AuthService
	user userstore.UserRepository
}

func NewAuthController(auth services.AuthService, user userstore.UserRepository) *AuthController {
	return &AuthController{auth: auth, user: user}
}

type signUpRequest struct {
	Name     string `json:"name" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}

type signInRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type tokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type endpoint = oauth2.Endpoint

// POST /api/v1/auth/sign-up

func (h *AuthController) SignUp(c *gin.Context) {
	var req signUpRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
		return
	}

	user, err := h.auth.Register(c.Request.Context(), req.Name, req.Email, req.Password)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "could not register user"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"id":    user.ID,
		"email": user.Email,
		"name":  user.Name,
	})
}

// POST /api/v1/auth/sign-in

func (h *AuthController) SignIn(c *gin.Context) {
	var req signInRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
		return
	}

	ua := c.Request.UserAgent()
	ip := c.ClientIP()

	access, accessExp, refresh, refreshExp, err := h.auth.Login(c.Request.Context(), req.Email, req.Password, ua, ip)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}

	// Set cookie-based auth
	cookies.SetAuthCookie(c.Writer, cookies.CookieAccessToken, access, accessExp)
	cookies.SetAuthCookie(c.Writer, cookies.CookieRefreshToken, refresh, refreshExp)

	c.JSON(http.StatusOK, tokenResponse{
		AccessToken:  access,
		RefreshToken: refresh,
	})
}

// POST /api/v1/auth/refresh

func (h *AuthController) Refresh(c *gin.Context) {
	var req struct {
		RefreshToken string `json:"refresh_token" binding:"-"`
	}
	_ = c.ShouldBindJSON(&req)

	// Prefer cookie refresh token; fallback to body
	if rt, ok := cookies.GetCookieValue(c.Request, cookies.CookieRefreshToken); ok && rt != "" {
		req.RefreshToken = rt
	}
	if req.RefreshToken == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
		return
	}

	ua := c.Request.UserAgent()
	ip := c.ClientIP()

	access, accessExp, refresh, refreshExp, err := h.auth.Refresh(c.Request.Context(), req.RefreshToken, ua, ip)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid refresh token"})
		return
	}

	cookies.SetAuthCookie(c.Writer, cookies.CookieAccessToken, access, accessExp)
	cookies.SetAuthCookie(c.Writer, cookies.CookieRefreshToken, refresh, refreshExp)

	c.JSON(http.StatusOK, tokenResponse{
		AccessToken:  access,
		RefreshToken: refresh,
	})
}

// POST /api/v1/auth/sign-out

func (h *AuthController) SignOut(c *gin.Context) {
	var req struct {
		RefreshToken string `json:"refresh_token" binding:"-"`
	}
	_ = c.ShouldBindJSON(&req)

	// Prefer cookie refresh token; fallback to body
	if rt, ok := cookies.GetCookieValue(c.Request, cookies.CookieRefreshToken); ok && rt != "" {
		req.RefreshToken = rt
	}
	if req.RefreshToken == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
		return
	}

	if err := h.auth.Logout(c.Request.Context(), req.RefreshToken); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid refresh token"})
		return
	}

	// Clear cookies
	cookies.ClearAuthCookie(c.Writer, cookies.CookieAccessToken)
	cookies.ClearAuthCookie(c.Writer, cookies.CookieRefreshToken)

	c.Status(http.StatusNoContent)
}

// HandleGoogleCallback recebe o code do front
func (h *AuthController) HandleGoogleCallback(c *gin.Context) {
	var input struct {
		config.VendorAuthConfig `json:",inline,omitempty" yaml:",inline,omitempty" toml:",inline,omitempty" mapstructure:"squash,omitempty" binding:"-"`
		Scopes                  []string           `json:"scopes" binding:"required"`
		Endpoint                endpoint           `json:"endpoint" binding:"-"`
		Code                    oauth2.TokenSource `json:"code" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(400, gin.H{"error": "Code is required"})
		return
	}

	// Config do OAuth (puxando do kbx)
	oauthConf := &oauth2.Config{
		ClientID:     input.Web.ClientID,
		ClientSecret: input.Web.ClientSecret,
		RedirectURL:  input.Web.RedirectURL,
		Scopes:       input.Scopes,
		Endpoint:     input.Endpoint,
	}

	// Pega o Code do Input
	tk, err := input.Code.Token()
	if err != nil {
		c.JSON(400, gin.H{"error": "Invalid code"})
		return
	}

	// 1. Troca o Code pelo Token
	token, err := oauthConf.Exchange(c, tk.AccessToken)
	if err != nil {
		logz.Error("Google Exchange Failed", "err", err)
		c.JSON(401, gin.H{"error": "Failed to exchange token"})
		return
	}

	// 2. Valida o ID Token (Segurança Real)
	// O id_token garante que o token foi emitido pelo Google PRA VOCÊ
	rawIDToken, ok := token.Extra("id_token").(string)
	if !ok {
		c.JSON(500, gin.H{"error": "No id_token field in oauth2 token."})
		return
	}

	// 3. Emite o SEU Token da Kubex

	accessToken, accessExp, refreshToken, refreshExp, err := h.auth.LoginWithOAuth2(c.Request.Context(), "google", rawIDToken, c.Request.UserAgent(), c.ClientIP())
	if err != nil {
		logz.Error("Token Issuance Failed", "err", err)
		c.JSON(500, gin.H{"error": "Failed to issue token"})
		return
	}

	// 4. Retorna o Token pro Frontend

	c.JSON(200, gin.H{
		"access_token_exp":  accessExp,
		"refresh_token_exp": refreshExp,
		"access_token":      accessToken,
		"refresh_token":     refreshToken,
	})

}

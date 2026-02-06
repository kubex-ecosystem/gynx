// Package middlewares implements the authentication middleware.
package middlewares

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/kubex-ecosystem/gnyx/internal/auth/tokens"
	"github.com/kubex-ecosystem/gnyx/internal/features/cookies"
)

const (
	ContextUserIDKey = "auth.user_id"
)

type AuthMiddleware struct {
	jwt tokens.JWTService
}

func NewAuthMiddleware(jwt tokens.JWTService) *AuthMiddleware {
	return &AuthMiddleware{jwt: jwt}
}

func (m *AuthMiddleware) Handle() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := ""

		// 1) Cookie primeiro (fluxo web)
		if raw, ok := cookies.GetCookieValue(c.Request, cookies.CookieAccessToken); ok && raw != "" {
			token = raw
		}

		// 2) Authorization header (fallback)
		if token == "" {
			authHeader := c.GetHeader("Authorization")
			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) == 2 && strings.EqualFold(parts[0], "bearer") {
				token = parts[1]
			}
		}

		if strings.TrimSpace(token) == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing authorization"})
			return
		}

		claims, err := m.jwt.ValidateAccessToken(token)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			return
		}

		c.Set(ContextUserIDKey, claims.Sub)
		c.Next()
	}
}

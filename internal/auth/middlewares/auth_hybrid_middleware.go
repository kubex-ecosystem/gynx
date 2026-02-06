package middlewares

import (
	"net/http"
	"strings"

	"github.com/kubex-ecosystem/gnyx/internal/auth/tokens"
	"github.com/kubex-ecosystem/gnyx/internal/features/cookies"

	"github.com/gin-gonic/gin"
)

type AuthHybridMiddleware struct {
	jwt tokens.JWTService
}

func NewAuthHybridMiddleware(jwt tokens.JWTService) *AuthHybridMiddleware {
	return &AuthHybridMiddleware{jwt: jwt}
}

func (m *AuthHybridMiddleware) Handle() gin.HandlerFunc {
	return func(c *gin.Context) {

		// 1) Tenta cookie first
		if raw, ok := cookies.GetCookieValue(c.Request, cookies.CookieAccessToken); ok {
			if claims, err := m.jwt.ValidateAccessToken(raw); err == nil {
				c.Set(ContextUserIDKey, claims.Sub)
				c.Next()
				return
			}
		}

		// 2) Fallback para Authorization Bearer
		authHeader := c.GetHeader("Authorization")
		if authHeader != "" {
			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) == 2 && strings.EqualFold(parts[0], "bearer") {
				if claims, err := m.jwt.ValidateAccessToken(parts[1]); err == nil {
					c.Set(ContextUserIDKey, claims.Sub)
					c.Next()
					return
				}
			}
		}

		// 3) Nenhum válido → 401
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
			"error": "unauthenticated",
		})
	}
}

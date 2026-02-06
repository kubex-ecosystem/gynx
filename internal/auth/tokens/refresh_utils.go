// Package tokens contém utilitários para manipulação de tokens de autenticação.
package tokens

import (
	"crypto/sha256"
	"time"

	"github.com/kubex-ecosystem/gnyx/internal/module/kbx"
)

// GenerateRefreshTokenFromPlain recalcula o hash a partir de um refresh token já existente.
// Útil para Refresh/Logout.
func GenerateRefreshTokenFromPlain(plain string) (string, string, time.Time, error) {
	h := sha256.Sum256([]byte(plain))
	hash := encodeToBase64URL(h[:])
	// A expiração aqui é irrelevante — vem da sessão no DB.
	return plain, hash, kbx.GetEnvOrDefaultWithType("KUBEX_BE_REFRESH_TIMEOUT", time.Time{}), nil
}

package auth

import (
	"time"

	"github.com/google/uuid"
)

// Session representa uma sessão de refresh token persistida no banco.
type Session struct {
	ID               uuid.UUID  `json:"id"`
	UserID           uuid.UUID  `json:"user_id"`
	RefreshTokenHash string     `json:"-"` // hash do refresh token
	UserAgent        string     `json:"user_agent,omitempty"`
	IP               string     `json:"ip,omitempty"`
	ExpiresAt        time.Time  `json:"expires_at"`
	RevokedAt        *time.Time `json:"revoked_at,omitempty"`
	CreatedAt        time.Time  `json:"created_at"`
}

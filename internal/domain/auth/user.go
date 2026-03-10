// Package auth implements the user model for authentication.
package auth

import (
	"time"

	"github.com/google/uuid"
)

// User representa o usuário visto pela camada de autenticação.
// Codex: ajustar nomes de colunas conforme schema real (users.* no Kubex).
type User struct {
	ID                 uuid.UUID  `json:"id"`
	Email              string     `json:"email"`
	Name               string     `json:"name"`
	LastName           string     `json:"last_name,omitempty"`
	PasswordHash       string     `json:"-"`
	Status             string     `json:"status"`
	CreatedAt          time.Time  `json:"created_at"`
	UpdatedAt          time.Time  `json:"updated_at"`
	Phone              string     `json:"phone,omitempty"`
	AvatarURL          string     `json:"avatar_url,omitempty"`
	ForcePasswordReset bool       `json:"force_password_reset,omitempty"`
	LastLogin          *time.Time `json:"last_login,omitempty"`
}

func (u *User) IsActive() bool {
	return u.Status == "active" || u.Status == "ACTIVE"
}

// Membership traz o vínculo do usuário com um tenant e role.
type Membership struct {
	TenantID   uuid.UUID `json:"tenant_id"`
	TenantName string    `json:"tenant_name,omitempty"`
	TenantSlug string    `json:"tenant_slug,omitempty"`
	RoleID     uuid.UUID `json:"role_id"`
	RoleCode   string    `json:"role_code,omitempty"`
	RoleName   string    `json:"role_name,omitempty"`
	IsActive   bool      `json:"is_active"`
	CreatedAt  time.Time `json:"created_at"`
}

// TeamMembership traz o vínculo do usuário com um team e role.
type TeamMembership struct {
	TeamID      uuid.UUID `json:"team_id"`
	TeamName    string    `json:"team_name,omitempty"`
	TenantID    uuid.UUID `json:"tenant_id"`
	TenantName  string    `json:"tenant_name,omitempty"`
	RoleID      uuid.UUID `json:"role_id"`
	RoleCode    string    `json:"role_code,omitempty"`
	RoleName    string    `json:"role_name,omitempty"`
	IsActive    bool      `json:"is_active"`
	IsDefault   bool      `json:"is_default,omitempty"`
	Description string    `json:"description,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
}

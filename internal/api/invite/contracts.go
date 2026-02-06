// Package invite defines the contracts for the Invite service.
package invite

import (
	"context"
	"time"

	t "github.com/kubex-ecosystem/gnyx/internal/types"
)

type Service interface {
	CreateInvite(ctx context.Context, req CreateInviteReq) (*InviteDTO, error)
	ValidateToken(ctx context.Context, token string) (*InviteDTO, error)
	AcceptInvite(ctx context.Context, token string, req AcceptInviteReq) (*AcceptResult, error)
	ListInvites(ctx context.Context, filters InviteListFilters) (*InviteListResponse, error)
}

type CreateInviteReq struct {
	Name          string    `json:"name,omitempty" validate:"required"`
	CompanyName   string    `json:"company_name,omitempty"` // só para convites de parceiros
	Type          string    `json:"type"      validate:"default=internal,oneof=internal partner"`
	TenantID      string    `json:"tenant_id" validate:"required,uuid4"`
	Email         string    `json:"email"     validate:"required,email"`
	Role          string    `json:"role"      validate:"required"` // code ou UUID (decisão v3)
	RoleCode      string    `json:"role_code,omitempty"`           // compat
	RoleID        string    `json:"role_id,omitempty"`             // compat
	TeamID        string    `json:"team_id,omitempty"`
	InvitedBy     string    `json:"invited_by,omitempty"`
	ExpiresInDays int       `json:"expires_in_days,omitempty"` // default 7
	ExpiresAt     time.Time `json:"-"`                         // calculado internamente
}

type InviteDTO struct {
	ID        string `json:"id"`
	Name      string `json:"name,omitempty"`
	Token     string `json:"token,omitempty"` // só na criação
	Email     string `json:"email"`
	Role      string `json:"role"`
	TenantID  string `json:"tenant_id"`
	TeamID    string `json:"team_id,omitempty"`
	Status    string `json:"status"` // pending/accepted/revoked/expired
	ExpiresAt string `json:"expires_at"`
	Type      string `json:"type,omitempty"`
}

type InviteDetailsDTO struct {
	InviteDTO
	InvitedBy string `json:"invited_by,omitempty"`
}

type AcceptInviteReq struct {
	Name     string `json:"name,omitempty"`
	LastName string `json:"last_name,omitempty"`
	Password string `json:"password,omitempty"` // opcional se auth externo
	// Campos futuros: phone, newsletter_optin, etc.
}

type AcceptResult struct {
	UserID     string `json:"user_id"`
	TenantID   string `json:"tenant_id"`
	Membership string `json:"membership"` // role efetiva
}

// InviteListFilters define filtros básicos de listagem para o FE.
type InviteListFilters struct {
	TenantID string `form:"tenant_id" json:"tenant_id"`
	Email    string `form:"email" json:"email"`
	Status   string `form:"status" json:"status"` // pending/accepted/expired/revoked
	Type     string `form:"type" json:"type"`     // partner/internal
	Page     int    `form:"page" json:"page"`
	Limit    int    `form:"limit" json:"limit"`
}

// InviteListResponse padroniza paginação para o FE.
type InviteListResponse struct {
	Data       []*InviteDTO `json:"data"`
	Total      int64        `json:"total"`
	Page       int          `json:"page"`
	Limit      int          `json:"limit"`
	TotalPages int          `json:"total_pages"`
}

type UserDTO struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	LastName  string `json:"last_name"`
	Email     string `json:"email"`
	TenantID  string `json:"tenant_id"`
	CreatedAt string `json:"created_at"`
}

type ErrorResponse = t.ErrorResponse

type MessageResponse = t.MessageResponse

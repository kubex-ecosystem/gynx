// Package invitations defines the database adapter interface for invitation operations.
// The BE is database-agnostic and only knows about this interface.
// The actual implementation (PostgreSQL via kubexdb, Supabase, etc.) is injected at runtime.
package invitations

import (
	"context"
	"time"

	"github.com/kubex-ecosystem/gnyx/internal/dsclient"
)

// InvitationStatus represents the status of an invitation
type InvitationStatus = dsclient.InvitationStatus

const (
	StatusPending  dsclient.InvitationStatus = dsclient.InvitationStatus("pending")
	StatusAccepted dsclient.InvitationStatus = dsclient.InvitationStatus("accepted")
	StatusExpired  dsclient.InvitationStatus = dsclient.InvitationStatus("expired")
	StatusRevoked  dsclient.InvitationStatus = dsclient.InvitationStatus("revoked")
)

// InvitationType represents the type of invitation
type InvitationType = dsclient.InvitationType

const (
	TypePartner  dsclient.InvitationType = dsclient.InvitationType("partner")
	TypeInternal dsclient.InvitationType = dsclient.InvitationType("internal")
)

// // DSClient Invitation is the underlying data structure from the DSClient layer
// type Invitation struct {
// 	ID         string           `json:"id"`
// 	Type       InvitationType   `json:"type"`
// 	Name       string           `json:"name"`
// 	Email      string           `json:"email"`
// 	Role       string           `json:"role"`
// 	Token      string           `json:"token"`
// 	TenantID   string           `json:"tenant_id"`
// 	TeamID     *string          `json:"team_id,omitempty"`
// 	InvitedBy  string           `json:"invited_by"`
// 	Status     InvitationStatus `json:"status"`
// 	ExpiresAt  time.Time        `json:"expires_at"`
// 	AcceptedAt *time.Time       `json:"accepted_at,omitempty"`
// 	CreatedAt  time.Time        `json:"created_at"`
// 	UpdatedAt  *time.Time       `json:"updated_at,omitempty"`
// }

// Invitation represents a generic invitation (partner or internal)
// This is the BE's domain model - database-agnostic
type Invitation struct {
	dsclient.Invitation `json:",inline" mapstructure:"squash"`

	// Additional fields or methods can be added here if needed (For BE domain logic)

	Token       string     `json:"token"`
	Metadata    *string    `json:"metadata,omitempty"`
	CancelledAt *time.Time `json:"cancelled_at,omitempty"`
}

// CreatePartnerInvitationInput represents input for creating a partner invitation
type CreatePartnerInvitationInput struct {
	Token        string     `json:"token"`
	PartnerEmail string     `json:"partner_email"`
	PartnerName  *string    `json:"partner_name,omitempty"`
	Role         string     `json:"role"`
	TenantID     string     `json:"tenant_id"`
	CompanyName  *string    `json:"company_name,omitempty"`
	InvitedBy    string     `json:"invited_by"`
	ExpiresAt    *time.Time `json:"expires_at,omitempty"`
}

// CreateInternalInvitationInput represents input for creating an internal invitation
type CreateInternalInvitationInput struct {
	Token        string     `json:"token"`
	InviteeEmail string     `json:"invitee_email"`
	InviteeName  *string    `json:"invitee_name,omitempty"`
	Role         string     `json:"role"`
	TeamID       *string    `json:"team_id,omitempty"`
	TenantID     string     `json:"tenant_id"`
	InvitedBy    string     `json:"invited_by"`
	ExpiresAt    *time.Time `json:"expires_at,omitempty"`
}

// UpdateInvitationInput represents input for updating an invitation
type UpdateInvitationInput struct {
	Status     *InvitationStatus `json:"status,omitempty"`
	AcceptedAt *time.Time        `json:"accepted_at,omitempty"`
	ExpiresAt  *time.Time        `json:"expires_at,omitempty"`
}

// InvitationFilters represents filters for listing invitations
type InvitationFilters struct {
	Email     *string           `json:"email,omitempty"`
	TenantID  *string           `json:"tenant_id,omitempty"`
	Status    *InvitationStatus `json:"status,omitempty"`
	InvitedBy *string           `json:"invited_by,omitempty"`
	Type      *InvitationType   `json:"type,omitempty"`
	Page      int               `json:"page"`
	Limit     int               `json:"limit"`
}

// PaginatedInvitations represents a paginated result of invitations
type PaginatedInvitations struct {
	Data       []*Invitation `json:"data"`
	Total      int64         `json:"total"`
	Page       int           `json:"page"`
	Limit      int           `json:"limit"`
	TotalPages int           `json:"total_pages"`
}

// Adapter defines the interface for invitation database operations.
// The BE only knows about this interface - the implementation can be:
// - kubexdb (PostgreSQL + GORM)
// - Supabase client
// - MongoDB
// - Any other data source
//
// This makes the BE database-agnostic.
type Adapter interface {
	// GetByToken retrieves an invitation by its token (searches both partner and internal)
	GetByToken(ctx context.Context, token string) (*Invitation, error)

	// GetByID retrieves an invitation by its ID
	GetByID(ctx context.Context, id string, invType InvitationType) (*Invitation, error)

	// CreatePartner creates a new partner invitation
	CreatePartner(ctx context.Context, input *CreatePartnerInvitationInput) (*Invitation, error)

	// CreateInternal creates a new internal invitation
	CreateInternal(ctx context.Context, input *CreateInternalInvitationInput) (*Invitation, error)

	// Update updates an existing invitation
	Update(ctx context.Context, id string, invType InvitationType, input *UpdateInvitationInput) (*Invitation, error)

	// Revoke revokes an invitation
	Revoke(ctx context.Context, id string, invType InvitationType) error

	// Accept accepts an invitation by token
	Accept(ctx context.Context, token string) (*Invitation, error)

	// Delete deletes an invitation
	Delete(ctx context.Context, id string, invType InvitationType) error

	// List lists invitations with optional filtering
	List(ctx context.Context, filters *InvitationFilters) (*PaginatedInvitations, error)
}

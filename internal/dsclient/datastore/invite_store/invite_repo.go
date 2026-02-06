// Package invitestore implements data access layer for invites
package invitestore

import (
	"database/sql"
	"errors"
	"time"
)

var (
	// ErrInviteNotFound is returned when an invite is not found
	ErrInviteNotFound = errors.New("invite not found")
	// ErrInviteExpired is returned when an invite has expired
	ErrInviteExpired = errors.New("invite has expired")
	// ErrInvalidInviteStatus is returned when invite status is invalid
	ErrInvalidInviteStatus = errors.New("invalid invite status")
)

// InviteModel represents the database model for invites
type InviteModel struct {
	ID         string
	Token      string
	Email      string
	Role       string
	Status     string
	ExpiresAt  time.Time
	CreatedAt  time.Time
	UpdatedAt  *time.Time
	TenantID   string
	InvitedBy  string
	InviteType string // "partner" or "internal"
}

// IInviteRepository defines the interface for invite data access
type IInviteRepository interface {
	GetByToken(token string) (*InviteModel, error)
	Create(invite *InviteModel) error
	UpdateStatus(token string, status string) error
}

// InviteRepository is the PostgreSQL implementation of IInviteRepository
type InviteRepository struct {
	db *sql.DB
}

// NewInviteRepository creates a new InviteRepository
func NewInviteRepository(db *sql.DB) IInviteRepository {
	return &InviteRepository{db: db}
}

// GetByToken retrieves an invite by its token
func (r *InviteRepository) GetByToken(token string) (*InviteModel, error) {
	// First, try partner_invitation
	invite, err := r.getPartnerInviteByToken(token)
	if err == nil {
		return invite, nil
	}

	// If not found, try internal_invitation
	invite, err = r.getInternalInviteByToken(token)
	if err == nil {
		return invite, nil
	}

	return nil, ErrInviteNotFound
}

// getPartnerInviteByToken retrieves a partner invitation by token
func (r *InviteRepository) getPartnerInviteByToken(token string) (*InviteModel, error) {
	query := `
		SELECT
			id,
			token,
			partner_email as email,
			role,
			status,
			expires_at,
			created_at,
			updated_at,
			tenant_id,
			invited_by
		FROM partner_invitation
		WHERE token = $1
	`

	var invite InviteModel
	var updatedAt sql.NullTime

	err := r.db.QueryRow(query, token).Scan(
		&invite.ID,
		&invite.Token,
		&invite.Email,
		&invite.Role,
		&invite.Status,
		&invite.ExpiresAt,
		&invite.CreatedAt,
		&updatedAt,
		&invite.TenantID,
		&invite.InvitedBy,
	)

	if err == sql.ErrNoRows {
		return nil, ErrInviteNotFound
	}
	if err != nil {
		return nil, err
	}

	if updatedAt.Valid {
		invite.UpdatedAt = &updatedAt.Time
	}

	invite.InviteType = "partner"

	// Validate invite status
	if invite.Status != "pending" {
		return nil, ErrInvalidInviteStatus
	}

	// Check if expired
	if time.Now().After(invite.ExpiresAt) {
		return nil, ErrInviteExpired
	}

	return &invite, nil
}

// getInternalInviteByToken retrieves an internal invitation by token
func (r *InviteRepository) getInternalInviteByToken(token string) (*InviteModel, error) {
	query := `
		SELECT
			id,
			token,
			invitee_email as email,
			role,
			status,
			expires_at,
			created_at,
			updated_at,
			tenant_id,
			invited_by
		FROM internal_invitation
		WHERE token = $1
	`

	var invite InviteModel
	var updatedAt sql.NullTime

	err := r.db.QueryRow(query, token).Scan(
		&invite.ID,
		&invite.Token,
		&invite.Email,
		&invite.Role,
		&invite.Status,
		&invite.ExpiresAt,
		&invite.CreatedAt,
		&updatedAt,
		&invite.TenantID,
		&invite.InvitedBy,
	)

	if err == sql.ErrNoRows {
		return nil, ErrInviteNotFound
	}
	if err != nil {
		return nil, err
	}

	if updatedAt.Valid {
		invite.UpdatedAt = &updatedAt.Time
	}

	invite.InviteType = "internal"

	// Validate invite status
	if invite.Status != "pending" {
		return nil, ErrInvalidInviteStatus
	}

	// Check if expired
	if time.Now().After(invite.ExpiresAt) {
		return nil, ErrInviteExpired
	}

	return &invite, nil
}

// Create creates a new invite (placeholder for future implementation)
func (r *InviteRepository) Create(invite *InviteModel) error {
	// TODO: Implement create logic
	return errors.New("not implemented")
}

// UpdateStatus updates the status of an invite (placeholder for future implementation)
func (r *InviteRepository) UpdateStatus(token string, status string) error {
	// TODO: Implement update logic
	return errors.New("not implemented")
}

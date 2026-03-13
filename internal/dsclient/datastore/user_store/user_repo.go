package userstore

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	models "github.com/kubex-ecosystem/gnyx/internal/domain/auth"
)

var ErrUserNotFound = errors.New("user not found")

type UserRepository interface {
	FindByEmail(ctx context.Context, email string) (*models.User, error)
	FindByID(ctx context.Context, id uuid.UUID) (*models.User, error)
	Create(ctx context.Context, u *models.User) error
	ListMemberships(ctx context.Context, userID uuid.UUID) ([]models.Membership, error)
	ListMembershipPermissions(ctx context.Context, userID uuid.UUID) (map[uuid.UUID][]string, error)
	ListTeamMemberships(ctx context.Context, userID uuid.UUID) ([]models.TeamMembership, error)
}

type userRepository struct {
	bridge *UserBridge
}

func (r *userRepository) FindByEmail(ctx context.Context, email string) (*models.User, error) {
	u, err := r.bridge.FindByEmail(ctx, email)
	if err != nil {
		return nil, err
	}
	if u == nil {
		return nil, ErrUserNotFound
	}
	return u, nil
}

func (r *userRepository) FindByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	u, err := r.bridge.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if u == nil {
		return nil, ErrUserNotFound
	}
	return u, nil
}

func (r *userRepository) Create(ctx context.Context, u *models.User) error {
	if u.ID == uuid.Nil {
		u.ID = uuid.New()
	}
	now := time.Now().UTC()
	if u.CreatedAt.IsZero() {
		u.CreatedAt = now
	}
	u.UpdatedAt = now

	if err := r.bridge.Create(ctx, u); err != nil {
		return err
	}
	return nil
}

// ListMemberships retorna os vínculos do usuário com tenants e roles.
func (r *userRepository) ListMemberships(ctx context.Context, userID uuid.UUID) ([]models.Membership, error) {
	return r.bridge.ListMemberships(ctx, userID)
}

func (r *userRepository) ListMembershipPermissions(ctx context.Context, userID uuid.UUID) (map[uuid.UUID][]string, error) {
	return r.bridge.ListMembershipPermissions(ctx, userID)
}

func (r *userRepository) ListTeamMemberships(ctx context.Context, userID uuid.UUID) ([]models.TeamMembership, error) {
	return r.bridge.ListTeamMemberships(ctx, userID)
}

// NewUserRepository cria repositório baseado no UserStore do DS.
func NewUserRepository() (UserRepository, error) {
	bridge, err := NewUserBridge(context.Background())
	if err != nil {
		return nil, err
	}
	return &userRepository{bridge: bridge}, nil
}

// Package userstore implementa um bridge para o UserStore do Datastore.
package userstore

import (
	"context"
	"strings"
	"time"

	"github.com/google/uuid"
	models "github.com/kubex-ecosystem/gnyx/internal/domain/auth"
	ds "github.com/kubex-ecosystem/gnyx/internal/dsclient"
	"github.com/kubex-ecosystem/gnyx/internal/dsclient/datastore"
	"github.com/kubex-ecosystem/gnyx/internal/module/kbx"
	gl "github.com/kubex-ecosystem/logz"
)

// UserBridge expõe operações de usuários consumindo o UserStore do DS.
type UserBridge struct {
	store ds.UserStore
}

// NewUserBridge cria um bridge de usuários.
func NewUserBridge(ctx context.Context) (*UserBridge, error) {
	store, err := datastore.UserStore(ctx)
	if err != nil {
		return nil, err
	}
	return &UserBridge{store: store}, nil
}

func (b *UserBridge) FindByEmail(ctx context.Context, email string) (*models.User, error) {
	email = strings.TrimSpace(strings.ToLower(email))
	u, err := b.store.GetByEmail(ctx, email)
	if err != nil {
		return nil, err
	}
	if u == nil {
		return nil, nil
	}
	return toAuthUser(u), nil
}

func (b *UserBridge) FindByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	u, err := b.store.GetByID(ctx, id.String())
	if err != nil {
		return nil, err
	}
	if u == nil {
		return nil, nil
	}
	return toAuthUser(u), nil
}

func (b *UserBridge) Create(ctx context.Context, m *models.User) error {
	if m == nil {
		return gl.Errorf("nil user")
	}

	input := &ds.CreateUserInput{
		Email:              m.Email,
		Name:               strPtrUser(m.Name),
		LastName:           strPtrUser(m.LastName),
		PasswordHash:       strPtrUser(m.PasswordHash),
		Phone:              strPtrUser(m.Phone),
		AvatarURL:          strPtrUser(m.AvatarURL),
		Status:             strPtrUser(m.Status),
		ForcePasswordReset: m.ForcePasswordReset,
	}

	user, err := b.store.Create(ctx, input)
	if err != nil {
		return err
	}
	// Preenche campos retornados
	if user != nil {
		if id, err := uuid.Parse(user.ID); err == nil {
			m.ID = id
		}
		if user.CreatedAt != (time.Time{}) {
			m.CreatedAt = user.CreatedAt
		}
		m.UpdatedAt = time.Now().UTC()
	}
	return nil
}

// ListMemberships mantém compatibilidade com o endpoint /me.
// Ele usa o executor PG do DS para consultar as tabelas de membership.
func (b *UserBridge) ListMemberships(ctx context.Context, userID uuid.UUID) ([]models.Membership, error) {
	conn, err := datastore.Connection(ctx)
	if err != nil {
		return nil, err
	}
	pgExec, err := ds.GetPGExecutor(ctx, conn)
	if err != nil {
		return nil, err
	}

	const q = `
		SELECT
			tm.tenant_id,
			t.name,
			COALESCE(t.slug, ''),
			tm.role_id,
			r.code,
			r.display_name,
			tm.is_active,
			tm.created_at
		FROM tenant_membership tm
		JOIN tenant t ON t.id = tm.tenant_id
		JOIN role r ON r.id = tm.role_id
		WHERE tm.user_id = $1
		ORDER BY tm.created_at DESC`

	rows, err := pgExec.Query(ctx, q, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []models.Membership
	for rows.Next() {
		var m models.Membership
		if err := rows.Scan(
			&m.TenantID,
			&m.TenantName,
			&m.TenantSlug,
			&m.RoleID,
			&m.RoleCode,
			&m.RoleName,
			&m.IsActive,
			&m.CreatedAt,
		); err != nil {
			return nil, err
		}
		result = append(result, m)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return result, nil
}

func (b *UserBridge) Update(ctx context.Context, m *models.User) error {
	if m == nil {
		return gl.Errorf("nil user")
	}

	input := &ds.UpdateUserInput{
		ID:                 m.ID.String(),
		Email:              strPtrUser(m.Email),
		Name:               strPtrUser(m.Name),
		LastName:           strPtrUser(m.LastName),
		PasswordHash:       strPtrUser(m.PasswordHash),
		Phone:              strPtrUser(m.Phone),
		AvatarURL:          strPtrUser(m.AvatarURL),
		Status:             strPtrUser(m.Status),
		ForcePasswordReset: kbx.BoolPtr(m.ForcePasswordReset),
	}

	user, err := b.store.Update(ctx, input)
	if err != nil {
		return err
	}
	// Preenche campos retornados
	if user != nil {
		if id, err := uuid.Parse(user.ID); err == nil {
			m.ID = id
		}
		if user.CreatedAt != (time.Time{}) {
			m.CreatedAt = user.CreatedAt
		}
		m.UpdatedAt = time.Now().UTC()
	}
	return nil
}

func (b *UserBridge) List(ctx context.Context, filters *ds.UserFilters) ([]*models.User, error) {
	users, err := b.store.List(ctx, filters)
	if err != nil {
		return nil, err
	}
	if users == nil {
		return nil, nil
	}
	pageSize := filters.Limit
	if pageSize <= 0 {
		pageSize = 50
	}
	filters.Limit = pageSize
	filters.Page = 1
	pageCount := users.TotalPages
	if filters.Page > pageCount {
		filters.Page = pageCount
	}
	// Preenche campos retornados

	var result []*models.User
	for i := users.Page; i < users.TotalPages; i++ {
		if i > users.Page {
			filters.Page = i
			users, err = b.store.List(ctx, filters)
			if err != nil {
				return nil, err
			}
			if users == nil {
				break
			}
			for _, u := range users.Data {
				result = append(result, toAuthUser(&u))
			}
		}
	}
	return result, nil
}

func toAuthUser(u *ds.User) *models.User {
	if u == nil {
		return nil
	}
	var id uuid.UUID
	if parsed, err := uuid.Parse(strings.TrimSpace(u.ID)); err == nil {
		id = parsed
	}
	return &models.User{
		ID:                 id,
		Email:              strings.TrimSpace(u.Email),
		Name:               deref(u.Name),
		LastName:           deref(u.LastName),
		PasswordHash:       deref(u.PasswordHash),
		Status:             deref(u.Status),
		Phone:              deref(u.Phone),
		AvatarURL:          deref(u.AvatarURL),
		ForcePasswordReset: u.ForcePasswordReset,
		LastLogin:          u.LastLogin,
		CreatedAt:          u.CreatedAt,
		UpdatedAt:          kbx.SafeTime(u.UpdatedAt),
	}
}

func deref(v *string) string {
	if v == nil {
		return ""
	}
	return *v
}

func strPtrUser(s string) *string {
	if strings.TrimSpace(s) == "" {
		return nil
	}
	v := strings.TrimSpace(s)
	return &v
}

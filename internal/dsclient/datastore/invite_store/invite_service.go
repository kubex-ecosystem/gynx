package invitestore

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	invitations "github.com/kubex-ecosystem/gnyx/internal/domain/invites"
	"github.com/kubex-ecosystem/gnyx/internal/dsclient"
	gl "github.com/kubex-ecosystem/logz"
)

// PostgresAdapter implementa o Adapter usando pgxpool contra o schema do DataService.
type PostgresAdapter struct {
	pool       *pgxpool.Pool
	defaultTTL time.Duration
	now        func() time.Time
}

// Option permite personalizar o adapter.
type Option func(*PostgresAdapter)

// WithDefaultTTL configura o TTL padrão para novos convites.
func WithDefaultTTL(ttl time.Duration) Option {
	return func(a *PostgresAdapter) {
		if ttl > 0 {
			a.defaultTTL = ttl
		}
	}
}

// WithNow injeta função de clock (para testes).
func WithNow(fn func() time.Time) Option {
	return func(a *PostgresAdapter) {
		if fn != nil {
			a.now = fn
		}
	}
}

// NewPostgresAdapter cria um novo adapter baseado em um pool pgx.
func NewPostgresAdapter(pool *pgxpool.Pool, opts ...Option) *PostgresAdapter {
	adapter := &PostgresAdapter{
		pool:       pool,
		defaultTTL: 7 * 24 * time.Hour,
		now: func() time.Time {
			return time.Now().UTC()
		},
	}
	for _, opt := range opts {
		opt(adapter)
	}
	return adapter
}

// GetByToken busca convites em ambas as tabelas.
func (a *PostgresAdapter) GetByToken(ctx context.Context, token string) (*invitations.Invitation, error) {
	if strings.TrimSpace(token) == "" {
		return nil, gl.Errorf("token is required")
	}

	if inv, err := a.findByToken(ctx, "partner_invitation", invitations.TypePartner, token); err == nil {
		return inv, nil
	} else if !errors.Is(err, dsclient.ErrNotFound) {
		return nil, err
	}

	return a.findByToken(ctx, "internal_invitation", invitations.TypeInternal, token)
}

// GetByID retorna convite por ID + tipo.
func (a *PostgresAdapter) GetByID(ctx context.Context, id string, invType invitations.InvitationType) (*invitations.Invitation, error) {
	table := tableName(invType)
	return a.findByID(ctx, table, invType, id)
}

// CreatePartner cria um novo convite partner.
func (a *PostgresAdapter) CreatePartner(ctx context.Context, input *invitations.CreatePartnerInvitationInput) (*invitations.Invitation, error) {
	if input == nil {
		return nil, gl.Errorf("input is required")
	}
	if strings.TrimSpace(input.Token) == "" {
		return nil, gl.Errorf("token is required")
	}
	expiresAt := a.computeExpiry(input.ExpiresAt)
	name := normalizeName(input.PartnerName, input.PartnerEmail)

	const q = `
		INSERT INTO partner_invitation (
			name, email, role, token, tenant_id, team_id, invited_by, status, expires_at
		) VALUES ($1,$2,$3,$4,$5,NULL,$6,'pending',$7)
		RETURNING id, name, email, role, token, tenant_id, team_id, invited_by, status, expires_at, accepted_at, created_at, updated_at`

	row := a.pool.QueryRow(ctx, q,
		name,
		input.PartnerEmail,
		input.Role,
		input.Token,
		input.TenantID,
		input.InvitedBy,
		expiresAt,
	)

	inv, err := scanInvitation(row, invitations.TypePartner)
	if err != nil {
		return nil, gl.Errorf("failed to create partner invitation: %v", err)
	}
	return inv, nil
}

// CreateInternal cria um convite interno.
func (a *PostgresAdapter) CreateInternal(ctx context.Context, input *invitations.CreateInternalInvitationInput) (*invitations.Invitation, error) {
	if input == nil {
		return nil, gl.Errorf("input is required")
	}
	if strings.TrimSpace(input.Token) == "" {
		return nil, gl.Errorf("token is required")
	}
	expiresAt := a.computeExpiry(input.ExpiresAt)
	name := normalizeName(input.InviteeName, input.InviteeEmail)

	const q = `
		INSERT INTO internal_invitation (
			name, email, role, token, tenant_id, team_id, invited_by, status, expires_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,'pending',$8)
		RETURNING id, name, email, role, token, tenant_id, team_id, invited_by, status, expires_at, accepted_at, created_at, updated_at`

	row := a.pool.QueryRow(ctx, q,
		name,
		input.InviteeEmail,
		input.Role,
		input.Token,
		input.TenantID,
		input.TeamID,
		input.InvitedBy,
		expiresAt,
	)

	inv, err := scanInvitation(row, invitations.TypeInternal)
	if err != nil {
		return nil, gl.Errorf("failed to create internal invitation: %v", err)
	}
	return inv, nil
}

// Update atualiza um convite específico.
func (a *PostgresAdapter) Update(ctx context.Context, id string, invType invitations.InvitationType, input *invitations.UpdateInvitationInput) (*invitations.Invitation, error) {
	if input == nil {
		return nil, errors.New("update input is required")
	}

	table := tableName(invType)
	updates := make([]string, 0, 4)
	args := make([]any, 0, 5)

	idx := 1
	if input.Status != nil {
		updates = append(updates, fmt.Sprintf("status = $%d", idx))
		args = append(args, *input.Status)
		idx++
		if *input.Status == invitations.StatusAccepted && input.AcceptedAt == nil {
			now := a.now()
			input.AcceptedAt = &now
		}
	}
	if input.AcceptedAt != nil {
		updates = append(updates, fmt.Sprintf("accepted_at = $%d", idx))
		args = append(args, input.AcceptedAt.UTC())
		idx++
	}
	if input.ExpiresAt != nil {
		updates = append(updates, fmt.Sprintf("expires_at = $%d", idx))
		args = append(args, input.ExpiresAt.UTC())
		idx++
	}

	if len(updates) == 0 {
		return a.GetByID(ctx, id, invType)
	}

	updates = append(updates, fmt.Sprintf("updated_at = $%d", idx))
	args = append(args, a.now())
	idx++

	args = append(args, id)

	query := fmt.Sprintf(`UPDATE %s SET %s WHERE id = $%d RETURNING id, name, email, role, token, tenant_id, team_id, invited_by, status, expires_at, accepted_at, created_at, updated_at`,
		table, strings.Join(updates, ", "), idx)

	row := a.pool.QueryRow(ctx, query, args...)
	inv, err := scanInvitation(row, invType)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, dsclient.ErrNotFound
		}
		return nil, err
	}
	return inv, nil
}

// Revoke altera o status para revoked.
func (a *PostgresAdapter) Revoke(ctx context.Context, id string, invType invitations.InvitationType) error {
	_, err := a.Update(ctx, id, invType, &invitations.UpdateInvitationInput{
		Status: invitationsStatusPtr(invitations.StatusRevoked),
	})
	return err
}

// Accept marca o convite como aceito usando o token.
func (a *PostgresAdapter) Accept(ctx context.Context, token string) (*invitations.Invitation, error) {
	if inv, err := a.accept(ctx, "partner_invitation", invitations.TypePartner, token); err == nil {
		return inv, nil
	} else if !errors.Is(err, dsclient.ErrNotFound) {
		return nil, err
	}
	return a.accept(ctx, "internal_invitation", invitations.TypeInternal, token)
}

// Delete remove o convite totalmente.
func (a *PostgresAdapter) Delete(ctx context.Context, id string, invType invitations.InvitationType) error {
	table := tableName(invType)
	cmd, err := a.pool.Exec(ctx, fmt.Sprintf("DELETE FROM %s WHERE id = $1", table), id)
	if err != nil {
		return err
	}
	if cmd.RowsAffected() == 0 {
		return dsclient.ErrNotFound
	}
	return nil
}

// List lista convites filtrando por tipo.
func (a *PostgresAdapter) List(ctx context.Context, filters *invitations.InvitationFilters) (*invitations.PaginatedInvitations, error) {
	if filters == nil || filters.Type == nil {
		return nil, errors.New("filters with type are required")
	}

	table := tableName(*filters.Type)
	where, args := buildFilters(filters)

	limit := max(filters.Limit, 20)
	page := max(filters.Page, 1)
	offset := (page - 1) * limit

	query := fmt.Sprintf(`SELECT id, name, email, role, token, tenant_id, team_id, invited_by, status, expires_at, accepted_at, created_at, updated_at
		FROM %s %s ORDER BY created_at DESC LIMIT $%d OFFSET $%d`, table, where, len(args)+1, len(args)+2)
	args = append(args, limit, offset)

	rows, err := a.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := &invitations.PaginatedInvitations{
		Data:       []*invitations.Invitation{},
		Page:       page,
		Limit:      limit,
		TotalPages: 0,
	}

	for rows.Next() {
		inv, err := scanInvitation(rows, *filters.Type)
		if err != nil {
			return nil, err
		}
		result.Data = append(result.Data, inv)
	}

	countQuery := fmt.Sprintf("SELECT COUNT(1) FROM %s %s", table, where)
	var total int64
	if err := a.pool.QueryRow(ctx, countQuery, args[:len(args)-2]...).Scan(&total); err != nil {
		return nil, err
	}
	result.Total = total
	result.TotalPages = calcTotalPages(total, limit)
	return result, nil
}

// Helpers ----------------------------------------------------------------

func (a *PostgresAdapter) computeExpiry(explicit *time.Time) time.Time {
	if explicit != nil && !explicit.IsZero() {
		return explicit.UTC()
	}
	return a.now().Add(a.defaultTTL)
}

func (a *PostgresAdapter) validateInvite(inv *invitations.Invitation) error {
	if inv == nil {
		return dsclient.ErrNotFound
	}
	if inv.Status != invitations.StatusPending {
		return dsclient.ErrInvalidStatus
	}
	if a.now().After(inv.ExpiresAt) {
		return dsclient.ErrExpired
	}
	return nil
}

func (a *PostgresAdapter) findByToken(ctx context.Context, table string, invType invitations.InvitationType, token string) (*invitations.Invitation, error) {
	query := fmt.Sprintf(`SELECT id, name, email, role, token, tenant_id, team_id, invited_by, status, expires_at, accepted_at, created_at, updated_at
		FROM %s WHERE token = $1`, table)

	row := a.pool.QueryRow(ctx, query, token)
	inv, err := scanInvitation(row, invType)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, dsclient.ErrNotFound
		}
		return nil, err
	}
	if err := a.validateInvite(inv); err != nil {
		return nil, err
	}
	return inv, nil
}

func (a *PostgresAdapter) findByID(ctx context.Context, table string, invType invitations.InvitationType, id string) (*invitations.Invitation, error) {
	query := fmt.Sprintf(`SELECT id, name, email, role, token, tenant_id, team_id, invited_by, status, expires_at, accepted_at, created_at, updated_at
		FROM %s WHERE id = $1`, table)
	row := a.pool.QueryRow(ctx, query, id)
	inv, err := scanInvitation(row, invType)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, dsclient.ErrNotFound
		}
		return nil, err
	}
	return inv, nil
}

func (a *PostgresAdapter) accept(ctx context.Context, table string, invType invitations.InvitationType, token string) (*invitations.Invitation, error) {
	tx, err := a.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	query := fmt.Sprintf(`SELECT id, name, email, role, token, tenant_id, team_id, invited_by, status, expires_at, accepted_at, created_at, updated_at
		FROM %s WHERE token = $1 FOR UPDATE`, table)

	row := tx.QueryRow(ctx, query, token)
	inv, err := scanInvitation(row, invType)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, dsclient.ErrNotFound
		}
		return nil, err
	}
	if err := a.validateInvite(inv); err != nil {
		return nil, err
	}

	now := a.now()
	updateQuery := fmt.Sprintf(`UPDATE %s SET status = $1, accepted_at = $2, updated_at = $2 WHERE id = $3`, table)
	if _, err := tx.Exec(ctx, updateQuery, invitations.StatusAccepted, now, inv.ID); err != nil {
		return nil, err
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	inv.Status = invitations.StatusAccepted
	inv.AcceptedAt = &now
	inv.UpdatedAt = &now
	return inv, nil
}

func scanInvitation(row pgx.Row, invType invitations.InvitationType) (*invitations.Invitation, error) {
	var (
		invite invitations.Invitation
		teamID *string
		accAt  *time.Time
		upAt   *time.Time
	)
	err := row.Scan(
		&invite.ID,
		&invite.Name,
		&invite.Email,
		&invite.Role,
		&invite.Token,
		&invite.TenantID,
		&teamID,
		&invite.InvitedBy,
		&invite.Status,
		&invite.ExpiresAt,
		&accAt,
		&invite.CreatedAt,
		&upAt,
	)
	if err != nil {
		return nil, err
	}
	invite.Type = invType
	invite.TeamID = teamID
	invite.AcceptedAt = accAt
	invite.UpdatedAt = upAt
	return &invite, nil
}

func tableName(t invitations.InvitationType) string {
	if t == invitations.TypePartner {
		return "partner_invitation"
	}
	return "internal_invitation"
}

func buildFilters(filters *invitations.InvitationFilters) (string, []any) {
	if filters == nil {
		return "", nil
	}
	var where []string
	var args []any
	add := func(cond string, val any) {
		where = append(where, cond)
		args = append(args, val)
	}

	if filters.Email != nil && strings.TrimSpace(*filters.Email) != "" {
		add(fmt.Sprintf("LOWER(email) = $%d", len(args)+1), strings.ToLower(*filters.Email))
	}
	if filters.TenantID != nil {
		add(fmt.Sprintf("tenant_id = $%d", len(args)+1), *filters.TenantID)
	}
	if filters.Status != nil {
		add(fmt.Sprintf("status = $%d", len(args)+1), *filters.Status)
	}
	if filters.InvitedBy != nil {
		add(fmt.Sprintf("invited_by = $%d", len(args)+1), *filters.InvitedBy)
	}

	if len(where) == 0 {
		return "", args
	}
	return "WHERE " + strings.Join(where, " AND "), args
}

func calcTotalPages(total int64, limit int) int {
	if limit <= 0 {
		return 0
	}
	pages := int(total) / limit
	if int(total)%limit != 0 {
		pages++
	}
	return max(pages, 1)
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func invitationsStatusPtr(status invitations.InvitationStatus) *invitations.InvitationStatus {
	return &status
}

func normalizeName(name *string, email string) string {
	if name != nil && strings.TrimSpace(*name) != "" {
		return strings.TrimSpace(*name)
	}
	return email
}

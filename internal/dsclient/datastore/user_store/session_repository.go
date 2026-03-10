package userstore

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	models "github.com/kubex-ecosystem/gnyx/internal/domain/auth"
	"github.com/kubex-ecosystem/gnyx/internal/dsclient/datastore"
)

type SessionRepository interface {
	Create(ctx context.Context, s *models.Session) error
	FindByRefreshHash(ctx context.Context, hash string) (*models.Session, error)
	Revoke(ctx context.Context, id uuid.UUID) error
	RevokeByRefreshHash(ctx context.Context, hash string) error
	RevokeByUser(ctx context.Context, userID uuid.UUID) error
}

type sessionRepository struct {
	exec datastore.PGExecutor
}

// NewSessionRepository cria um repositório de sessões baseado em pgx.
func NewSessionRepository() (SessionRepository, error) {
	conn, err := datastore.Connection(context.Background())
	if err != nil {
		return nil, err
	}
	exec, err := datastore.GetPGExecutor(context.Background(), conn)
	if err != nil {
		return nil, err
	}
	return &sessionRepository{exec: exec}, nil
}

// Codex: ajustar nome da tabela/colunas para schema real (auth_sessions, sessions, etc).
func (r *sessionRepository) Create(ctx context.Context, s *models.Session) error {
	if s.ID == uuid.Nil {
		s.ID = uuid.New()
	}
	if s.CreatedAt.IsZero() {
		s.CreatedAt = time.Now().UTC()
	}

	// Prefer auth_sessions; fallback to refresh_tokens for environments that don't have auth_sessions seeded.
	if r.tableExists(ctx, "auth_sessions") {
		const q = `
			INSERT INTO auth_sessions (
				id, user_id, refresh_token_hash, user_agent, ip, expires_at, revoked_at, created_at
			) VALUES ($1, $2, $3, $4, $5, $6, $7, $8);
		`

		_, err := r.exec.Exec(ctx, q,
			s.ID,
			s.UserID,
			s.RefreshTokenHash,
			s.UserAgent,
			s.IP,
			s.ExpiresAt,
			s.RevokedAt,
			s.CreatedAt,
		)
		return err
	}

	// Fallback: store hash in refresh_tokens.token_id
	const alt = `
		INSERT INTO refresh_tokens (token_id, user_id, expires_at, created_at)
		VALUES ($1, $2, $3, $4);
	`
	_, err := r.exec.Exec(ctx, alt, s.RefreshTokenHash, s.UserID.String(), s.ExpiresAt, s.CreatedAt)
	return err
}

func (r *sessionRepository) FindByRefreshHash(ctx context.Context, hash string) (*models.Session, error) {
	if r.tableExists(ctx, "auth_sessions") {
		const q = `
			SELECT id, user_id, refresh_token_hash, user_agent, ip, expires_at, revoked_at, created_at
			FROM auth_sessions
			WHERE refresh_token_hash = $1
			LIMIT 1;
		`

		row := r.exec.QueryRow(ctx, q, hash)

		var s models.Session
		if err := row.Scan(
			&s.ID,
			&s.UserID,
			&s.RefreshTokenHash,
			&s.UserAgent,
			&s.IP,
			&s.ExpiresAt,
			&s.RevokedAt,
			&s.CreatedAt,
		); err != nil {
			return nil, err
		}

		return &s, nil
	}

	const alt = `
		SELECT token_id, user_id, expires_at, created_at
		FROM refresh_tokens
		WHERE token_id = $1
		LIMIT 1;
	`

	row := r.exec.QueryRow(ctx, alt, hash)
	var (
		tokenID string
		userID  string
		s       models.Session
	)
	if err := row.Scan(&tokenID, &userID, &s.ExpiresAt, &s.CreatedAt); err != nil {
		return nil, err
	}

	s.ID = uuid.Nil
	s.UserID, _ = uuid.Parse(userID)
	s.RefreshTokenHash = tokenID
	return &s, nil
}

func (r *sessionRepository) Revoke(ctx context.Context, id uuid.UUID) error {
	now := time.Now().UTC()
	if r.tableExists(ctx, "auth_sessions") {
		const q = `
			UPDATE auth_sessions
			SET revoked_at = $2
			WHERE id = $1;
		`
		_, err := r.exec.Exec(ctx, q, id, now)
		return err
	}

	const alt = `DELETE FROM refresh_tokens WHERE token_id = $1`
	res, err := r.exec.Exec(ctx, alt, id.String())
	if err != nil {
		return err
	}
	if res.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

func (r *sessionRepository) RevokeByRefreshHash(ctx context.Context, hash string) error {
	if hash == "" {
		return pgx.ErrNoRows
	}

	now := time.Now().UTC()
	if r.tableExists(ctx, "auth_sessions") {
		const q = `
			UPDATE auth_sessions
			SET revoked_at = $2
			WHERE refresh_token_hash = $1 AND revoked_at IS NULL;
		`
		res, err := r.exec.Exec(ctx, q, hash, now)
		if err != nil {
			return err
		}
		if res.RowsAffected() == 0 {
			return pgx.ErrNoRows
		}
		return nil
	}

	const alt = `DELETE FROM refresh_tokens WHERE token_id = $1`
	res, err := r.exec.Exec(ctx, alt, hash)
	if err != nil {
		return err
	}
	if res.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

func (r *sessionRepository) RevokeByUser(ctx context.Context, userID uuid.UUID) error {
	now := time.Now().UTC()
	if r.tableExists(ctx, "auth_sessions") {
		const q = `
			UPDATE auth_sessions
			SET revoked_at = $2
			WHERE user_id = $1 AND revoked_at IS NULL;
		`
		_, err := r.exec.Exec(ctx, q, userID, now)
		return err
	}

	const alt = `DELETE FROM refresh_tokens WHERE user_id = $1`
	_, err := r.exec.Exec(ctx, alt, userID.String())
	return err
}

func (r *sessionRepository) tableExists(ctx context.Context, table string) bool {
	const q = `
		SELECT 1
		FROM information_schema.tables
		WHERE table_name = $1
		LIMIT 1;
	`
	if err := r.exec.QueryRow(ctx, q, table).Scan(new(int)); err != nil {
		return !errors.Is(err, pgx.ErrNoRows)
	}
	return true
}

package sessionstore

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/kubex-ecosystem/gnyx/internal/auth/tokens"
	models "github.com/kubex-ecosystem/gnyx/internal/domain/auth"
	"golang.org/x/crypto/bcrypt"
)

type stubUserRepo struct {
	user *models.User
	err  error
}

func (s *stubUserRepo) FindByEmail(ctx context.Context, email string) (*models.User, error) {
	return s.user, s.err
}

func (s *stubUserRepo) FindByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	if s.err != nil {
		return nil, s.err
	}
	if s.user == nil {
		return nil, errors.New("user not found")
	}
	return s.user, nil
}

func (s *stubUserRepo) Create(ctx context.Context, u *models.User) error { return nil }

func (s *stubUserRepo) ListMemberships(ctx context.Context, userID uuid.UUID) ([]models.Membership, error) {
	return nil, nil
}

func (s *stubUserRepo) ListTeamMemberships(ctx context.Context, userID uuid.UUID) ([]models.TeamMembership, error) {
	return nil, nil
}

func (s *stubUserRepo) ListMembershipPermissions(context.Context, uuid.UUID) (map[uuid.UUID][]string, error) {
	return nil, nil
}

type stubJWT struct{}

func (s *stubJWT) GenerateAccessToken(userID string) (string, time.Time, error) {
	return "token", time.Now().UTC().Add(time.Hour), nil
}

func (s *stubJWT) ValidateAccessToken(token string) (*tokens.Claims, error) { return nil, nil }

func (s *stubJWT) VerifyOAuth2Token(ctx context.Context, providerName, oauthToken string) (string, error) {
	return "", nil
}

type stubSessionRepo struct {
	findResult *models.Session
	findErr    error
	createErr  error
	created    []*models.Session

	revokeIDCalled   bool
	revokeHashCalled bool
	revokeUserCalled bool
	revokeIDErr      error
	revokeHashErr    error
	revokeUserErr    error
}

func (s *stubSessionRepo) Create(ctx context.Context, session *models.Session) error {
	if s.createErr != nil {
		return s.createErr
	}
	cp := *session
	s.created = append(s.created, &cp)
	return nil
}

func (s *stubSessionRepo) FindByRefreshHash(ctx context.Context, hash string) (*models.Session, error) {
	if s.findErr != nil {
		return nil, s.findErr
	}
	return s.findResult, nil
}

func (s *stubSessionRepo) Revoke(ctx context.Context, id uuid.UUID) error {
	s.revokeIDCalled = true
	return s.revokeIDErr
}

func (s *stubSessionRepo) RevokeByRefreshHash(ctx context.Context, hash string) error {
	s.revokeHashCalled = true
	return s.revokeHashErr
}

func (s *stubSessionRepo) RevokeByUser(ctx context.Context, userID uuid.UUID) error {
	s.revokeUserCalled = true
	return s.revokeUserErr
}

func TestRevokeSessionUsesIDWhenAvailable(t *testing.T) {
	sessions := &stubSessionRepo{}
	svc := &authService{users: &stubUserRepo{}, sessions: sessions, jwt: &stubJWT{}, refreshTTL: 24 * time.Hour}

	sess := &models.Session{ID: uuid.New(), RefreshTokenHash: "hash"}
	if err := svc.revokeSession(context.Background(), sess); err != nil {
		t.Fatalf("revokeSession() error = %v", err)
	}
	if !sessions.revokeIDCalled {
		t.Fatalf("expected Revoke to be called")
	}
	if sessions.revokeHashCalled {
		t.Fatalf("did not expect RevokeByRefreshHash to be called")
	}
}

func TestRevokeSessionFallsBackToHashWhenSessionIDIsNil(t *testing.T) {
	sessions := &stubSessionRepo{}
	svc := &authService{users: &stubUserRepo{}, sessions: sessions, jwt: &stubJWT{}, refreshTTL: 24 * time.Hour}

	sess := &models.Session{ID: uuid.Nil, RefreshTokenHash: "hash"}
	if err := svc.revokeSession(context.Background(), sess); err != nil {
		t.Fatalf("revokeSession() error = %v", err)
	}
	if sessions.revokeIDCalled {
		t.Fatalf("did not expect Revoke to be called")
	}
	if !sessions.revokeHashCalled {
		t.Fatalf("expected RevokeByRefreshHash to be called")
	}
}

func TestRevokeSessionRejectsNilSession(t *testing.T) {
	sessions := &stubSessionRepo{}
	svc := &authService{users: &stubUserRepo{}, sessions: sessions, jwt: &stubJWT{}, refreshTTL: 24 * time.Hour}

	err := svc.revokeSession(context.Background(), nil)
	if !errors.Is(err, ErrInvalidToken) {
		t.Fatalf("revokeSession() error = %v, want %v", err, ErrInvalidToken)
	}
}

func TestValidateSessionRejectsRevokedSession(t *testing.T) {
	now := time.Now().UTC()
	svc := &authService{refreshTTL: 24 * time.Hour}

	err := svc.validateSession(&models.Session{
		ID:        uuid.New(),
		UserID:    uuid.New(),
		ExpiresAt: now.Add(time.Hour),
		RevokedAt: &now,
	})
	if !errors.Is(err, ErrInvalidToken) {
		t.Fatalf("validateSession() error = %v, want %v", err, ErrInvalidToken)
	}
}

func TestValidateSessionRejectsExpiredSession(t *testing.T) {
	svc := &authService{refreshTTL: 24 * time.Hour}

	err := svc.validateSession(&models.Session{
		ID:        uuid.New(),
		UserID:    uuid.New(),
		ExpiresAt: time.Now().UTC().Add(-time.Minute),
	})
	if !errors.Is(err, ErrInvalidToken) {
		t.Fatalf("validateSession() error = %v, want %v", err, ErrInvalidToken)
	}
}

func TestRefreshRejectsRevokedSession(t *testing.T) {
	userID := uuid.New()
	plain, hash, _, err := tokens.GenerateRefreshTokenWithTTL(time.Hour)
	if err != nil {
		t.Fatalf("GenerateRefreshTokenWithTTL() error = %v", err)
	}
	now := time.Now().UTC()
	sessions := &stubSessionRepo{
		findResult: &models.Session{
			ID:               uuid.New(),
			UserID:           userID,
			RefreshTokenHash: hash,
			ExpiresAt:        now.Add(time.Hour),
			RevokedAt:        &now,
		},
	}
	users := &stubUserRepo{
		user: &models.User{ID: userID, Status: "active"},
	}
	svc := &authService{users: users, sessions: sessions, jwt: &stubJWT{}, refreshTTL: 2 * time.Hour}

	_, _, _, _, err = svc.Refresh(context.Background(), plain, "ua", "ip")
	if !errors.Is(err, ErrInvalidToken) {
		t.Fatalf("Refresh() error = %v, want %v", err, ErrInvalidToken)
	}
}

func TestRefreshRejectsExpiredSession(t *testing.T) {
	userID := uuid.New()
	plain, hash, _, err := tokens.GenerateRefreshTokenWithTTL(time.Hour)
	if err != nil {
		t.Fatalf("GenerateRefreshTokenWithTTL() error = %v", err)
	}
	sessions := &stubSessionRepo{
		findResult: &models.Session{
			ID:               uuid.New(),
			UserID:           userID,
			RefreshTokenHash: hash,
			ExpiresAt:        time.Now().UTC().Add(-time.Minute),
		},
	}
	users := &stubUserRepo{
		user: &models.User{ID: userID, Status: "active"},
	}
	svc := &authService{users: users, sessions: sessions, jwt: &stubJWT{}, refreshTTL: 2 * time.Hour}

	_, _, _, _, err = svc.Refresh(context.Background(), plain, "ua", "ip")
	if !errors.Is(err, ErrInvalidToken) {
		t.Fatalf("Refresh() error = %v, want %v", err, ErrInvalidToken)
	}
}

func TestLoginUsesConfiguredRefreshTTL(t *testing.T) {
	userID := uuid.New()
	users := &stubUserRepo{
		user: &models.User{
			ID:           userID,
			Email:        "user@example.com",
			PasswordHash: mustHashPassword(t, "secret"),
			Status:       "active",
		},
	}
	sessions := &stubSessionRepo{}
	ttl := 2 * time.Hour
	svc := &authService{users: users, sessions: sessions, jwt: &stubJWT{}, refreshTTL: ttl}

	_, _, _, refreshExp, err := svc.Login(context.Background(), "user@example.com", "secret", "ua", "ip")
	if err != nil {
		t.Fatalf("Login() error = %v", err)
	}
	if len(sessions.created) != 1 {
		t.Fatalf("expected one created session, got %d", len(sessions.created))
	}
	if got := sessions.created[0].ExpiresAt.Sub(time.Now().UTC()); got < ttl-time.Minute || got > ttl+time.Minute {
		t.Fatalf("created session expiry drift = %v, want around %v", got, ttl)
	}
	if diff := refreshExp.Sub(sessions.created[0].ExpiresAt); diff < -time.Second || diff > time.Second {
		t.Fatalf("refreshExp differs from stored session expiry by %v", diff)
	}
}

func mustHashPassword(t *testing.T, password string) string {
	t.Helper()
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("GenerateFromPassword() error = %v", err)
	}
	return string(hash)
}

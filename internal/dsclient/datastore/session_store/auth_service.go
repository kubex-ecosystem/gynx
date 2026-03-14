// Package sessionstore implements the authentication service.
package sessionstore

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/kubex-ecosystem/gnyx/internal/auth/tokens"
	models "github.com/kubex-ecosystem/gnyx/internal/domain/auth"
	userstore "github.com/kubex-ecosystem/gnyx/internal/dsclient/datastore/user_store"
	gl "github.com/kubex-ecosystem/logz"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidToken       = errors.New("invalid token")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUserInactive       = errors.New("user is inactive")
)

type AuthService interface {
	Register(ctx context.Context, name, email, password string) (*models.User, error)
	Login(ctx context.Context, email, password, ua, ip string) (accessToken string, accessExp time.Time, refreshToken string, refreshExp time.Time, err error)
	Refresh(ctx context.Context, refreshToken string, ua, ip string) (newAccess string, accessExp time.Time, newRefresh string, refreshExp time.Time, err error)
	Logout(ctx context.Context, refreshToken string) error
	ListMemberships(ctx context.Context, userID uuid.UUID) ([]models.Membership, error)
	LoginWithOAuth2(ctx context.Context, providerName, oauthToken, ua, ip string) (accessToken string, accessExp time.Time, refreshToken string, refreshExp time.Time, err error)
}

type authService struct {
	users      userstore.UserRepository
	sessions   userstore.SessionRepository
	jwt        tokens.JWTService
	logger     *gl.LoggerZ
	refreshTTL time.Duration
}

func NewAuthService(
	users userstore.UserRepository,
	sessions userstore.SessionRepository,
	jwt tokens.JWTService,
	logger *gl.LoggerZ,
	refreshTTL time.Duration,
) AuthService {
	if logger == nil {
		logger = gl.GetLoggerZ("auth_service")
	}
	if refreshTTL <= 0 {
		refreshTTL = 30 * 24 * time.Hour
	}
	return &authService{
		users:      users,
		sessions:   sessions,
		jwt:        jwt,
		logger:     logger,
		refreshTTL: refreshTTL,
	}
}

func (s *authService) Register(ctx context.Context, name, email, password string) (*models.User, error) {
	email = strings.TrimSpace(strings.ToLower(email))

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	u := &models.User{
		Email:        email,
		Name:         name,
		PasswordHash: string(hash),
		Status:       "active",
	}

	if err := s.users.Create(ctx, u); err != nil {
		return nil, err
	}

	return u, nil
}

func (s *authService) Login(ctx context.Context, email, password, ua, ip string) (string, time.Time, string, time.Time, error) {
	email = strings.TrimSpace(strings.ToLower(email))
	u, err := s.users.FindByEmail(ctx, email)
	if err != nil {
		s.logger.Warn("login failed during user lookup", "email", email, "err", err)
		return "", time.Time{}, "", time.Time{}, ErrInvalidCredentials
	}
	if !u.IsActive() {
		s.logger.Warn("login rejected for inactive user", "email", email, "user_id", u.ID)
		return "", time.Time{}, "", time.Time{}, ErrUserInactive
	}

	if err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password)); err != nil {
		s.logger.Warn("login rejected due to password mismatch", "email", email, "user_id", u.ID)
		return "", time.Time{}, "", time.Time{}, ErrInvalidCredentials
	}

	accessToken, accessExp, err := s.jwt.GenerateAccessToken(u.ID.String())
	if err != nil {
		return "", time.Time{}, "", time.Time{}, err
	}

	plainRefresh, hash, exp, err := tokens.GenerateRefreshTokenWithTTL(s.refreshTTL)
	if err != nil {
		return "", time.Time{}, "", time.Time{}, err
	}

	sess := &models.Session{
		UserID:           u.ID,
		RefreshTokenHash: hash,
		UserAgent:        ua,
		IP:               ip,
		ExpiresAt:        exp,
	}

	if err := s.sessions.Create(ctx, sess); err != nil {
		return "", time.Time{}, "", time.Time{}, err
	}

	return accessToken, accessExp, plainRefresh, exp, nil
}

func (s *authService) Refresh(ctx context.Context, refreshToken string, ua, ip string) (string, time.Time, string, time.Time, error) {
	_, hash, _, err := tokens.GenerateRefreshTokenFromPlain(refreshToken)
	if err != nil {
		return "", time.Time{}, "", time.Time{}, ErrInvalidToken
	}

	sess, err := s.sessions.FindByRefreshHash(ctx, hash)
	if err != nil {
		return "", time.Time{}, "", time.Time{}, ErrInvalidToken
	}
	if err := s.validateSession(sess); err != nil {
		return "", time.Time{}, "", time.Time{}, err
	}
	user, err := s.users.FindByID(ctx, sess.UserID)
	if err != nil {
		return "", time.Time{}, "", time.Time{}, ErrInvalidToken
	}
	if !user.IsActive() {
		return "", time.Time{}, "", time.Time{}, ErrUserInactive
	}

	accessToken, accessExp, err := s.jwt.GenerateAccessToken(user.ID.String())
	if err != nil {
		return "", time.Time{}, "", time.Time{}, err
	}

	plainRefresh, newHash, exp, err := tokens.GenerateRefreshTokenWithTTL(s.refreshTTL)
	if err != nil {
		return "", time.Time{}, "", time.Time{}, err
	}

	// revoga antigo e cria novo
	if err := s.revokeSession(ctx, sess); err != nil {
		s.logger.Warn("failed to revoke old session", "err", err)
		gl.Warn("performing fallback: revoking all sessions for user", "user_id", user.ID)
		// Fallback para ambientes sem auth_sessions: revoga tudo do usuário
		if revokeErr := s.sessions.RevokeByUser(ctx, user.ID); revokeErr != nil {
			s.logger.Warn("failed to revoke sessions by user", "err", revokeErr)
			gl.Warn("failed to revoke sessions by user", "err", revokeErr)
		}
	}

	newSess := &models.Session{
		UserID:           user.ID,
		RefreshTokenHash: newHash,
		UserAgent:        ua,
		IP:               ip,
		ExpiresAt:        exp,
	}
	if err := s.sessions.Create(ctx, newSess); err != nil {
		return "", time.Time{}, "", time.Time{}, err
	}

	return accessToken, accessExp, plainRefresh, exp, nil
}

func (s *authService) Logout(ctx context.Context, refreshToken string) error {
	_, hash, _, err := tokens.GenerateRefreshTokenFromPlain(refreshToken)
	if err != nil {
		return ErrInvalidToken
	}

	sess, err := s.sessions.FindByRefreshHash(ctx, hash)
	if err != nil {
		return ErrInvalidToken
	}
	if err := s.validateSession(sess); err != nil {
		return err
	}

	if err := s.revokeSession(ctx, sess); err != nil {
		// Em fallback, limpa todas as sessões do usuário
		return s.sessions.RevokeByUser(ctx, sess.UserID)
	}

	if err := s.sessions.RevokeByUser(ctx, sess.UserID); err != nil {
		gl.Error("failed to revoke sessions by user", "user_id", sess.UserID, "err", err)
		return s.logger.Errorf("failed to revoke sessions by user %v", err)
	}

	return nil
}

func (s *authService) revokeSession(ctx context.Context, sess *models.Session) error {
	if sess == nil {
		return ErrInvalidToken
	}
	if sess.ID != uuid.Nil {
		return s.sessions.Revoke(ctx, sess.ID)
	}
	return s.sessions.RevokeByRefreshHash(ctx, sess.RefreshTokenHash)
}

func (s *authService) validateSession(sess *models.Session) error {
	if sess == nil {
		return ErrInvalidToken
	}

	now := time.Now().UTC()
	if sess.RevokedAt != nil && !sess.RevokedAt.IsZero() {
		return ErrInvalidToken
	}
	if !sess.ExpiresAt.IsZero() && !now.Before(sess.ExpiresAt) {
		return ErrInvalidToken
	}
	return nil
}

func (s *authService) ListMemberships(ctx context.Context, userID uuid.UUID) ([]models.Membership, error) {
	return s.users.ListMemberships(ctx, userID)
}

func (s *authService) LoginWithOAuth2(ctx context.Context, providerName, oauthToken, ua, ip string) (string, time.Time, string, time.Time, error) {
	// 1. Verifica o token OAuth2 com o provedor (ex: Google)
	email, err := s.jwt.VerifyOAuth2Token(ctx, providerName, oauthToken)
	if err != nil {
		return "", time.Time{}, "", time.Time{}, ErrInvalidToken
	}

	// 2. Busca (ou cria) o usuário com esse email
	user, err := s.users.FindByEmail(ctx, email)
	if err != nil {
		// Usuário não existe, cria um novo
		user, err = s.Register(ctx, "", email, uuid.NewString()) // senha aleatória
		if err != nil {
			return "", time.Time{}, "", time.Time{}, err
		}
	}
	if !user.IsActive() {
		return "", time.Time{}, "", time.Time{}, ErrUserInactive
	}

	// 3. Gera tokens JWT e Refresh Token
	accessToken, accessExp, err := s.jwt.GenerateAccessToken(user.ID.String())
	if err != nil {
		return "", time.Time{}, "", time.Time{}, err
	}

	plainRefresh, hash, exp, err := tokens.GenerateRefreshToken()
	if err != nil {
		return "", time.Time{}, "", time.Time{}, err
	}

	sess := &models.Session{
		UserID:           user.ID,
		RefreshTokenHash: hash,
		UserAgent:        ua,
		IP:               ip,
		ExpiresAt:        exp,
	}

	if err := s.sessions.Create(ctx, sess); err != nil {
		return "", time.Time{}, "", time.Time{}, err
	}

	return accessToken, accessExp, plainRefresh, exp, nil
}

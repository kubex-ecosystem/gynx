package tokens

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"cloud.google.com/go/auth/credentials/idtoken"
	"github.com/golang-jwt/jwt/v5"
	"github.com/kubex-ecosystem/gnyx/internal/config"
)

var (
	ErrInvalidToken = errors.New("invalid token")
)

type Claims struct {
	Sub string `json:"sub"`
	jwt.RegisteredClaims
}

// JWTService define geração/validação de access tokens.
// Refresh token aqui é string aleatória separada (hash no DB).
type JWTService interface {
	GenerateAccessToken(userID string) (string, time.Time, error)
	ValidateAccessToken(token string) (*Claims, error)
	VerifyOAuth2Token(ctx context.Context, providerName, oauthToken string) (string, error)
}

// jwtService implementa JWTService com RSA.
type jwtService struct {
	cfg        *config.MainConfig
	privateKey *rsa.PrivateKey
	publicKey  *rsa.PublicKey
}

// NewJWTService cria um serviço JWT baseado na Config.
// Codex: se as chaves vierem de arquivo, carregar aqui via os.ReadFile.
func NewJWTService(cfg *config.MainConfig, priv *rsa.PrivateKey, pub *rsa.PublicKey) JWTService {
	return &jwtService{
		cfg:        cfg,
		privateKey: priv,
		publicKey:  pub,
	}
}

func (s *jwtService) GenerateAccessToken(userID string) (string, time.Time, error) {
	now := time.Now().UTC()
	ttl := s.cfg.ServerConfig.Runtime.AccessTokenTTL
	if ttl <= 0 {
		ttl = 15 * time.Minute // fallback seguro para dev se não vier config
	}
	exp := now.Add(ttl)

	claims := &Claims{
		Sub: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    s.cfg.ServerConfig.Runtime.Issuer,
			Subject:   userID,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(exp),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	signed, err := token.SignedString(s.privateKey)
	if err != nil {
		return "", time.Time{}, err
	}

	return signed, exp, nil
}

func (s *jwtService) ValidateAccessToken(tokenStr string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, ErrInvalidToken
		}
		return s.publicKey, nil
	})
	if err != nil {
		return nil, ErrInvalidToken
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, ErrInvalidToken
	}

	return claims, nil
}

// GenerateRefreshToken gera um refresh token randômico usando o TTL default legado.
func GenerateRefreshToken() (plain string, hash string, exp time.Time, err error) {
	return GenerateRefreshTokenWithTTL(30 * 24 * time.Hour)
}

// GenerateRefreshTokenWithTTL gera um refresh token randômico + hash SHA-256 com TTL configurável.
func GenerateRefreshTokenWithTTL(ttl time.Duration) (plain string, hash string, exp time.Time, err error) {
	if ttl <= 0 {
		ttl = 30 * 24 * time.Hour
	}
	exp = time.Now().UTC().Add(ttl)
	b := make([]byte, 64)

	if _, err = rand.Read(b); err != nil {
		return "", "", time.Time{}, err
	}

	plain = encodeToBase64URL(b)
	h := sha256.Sum256([]byte(plain))
	hash = encodeToBase64URL(h[:])

	return plain, hash, exp, nil
}

// Implementa base64 URL-safe simples, sem depender de libs extras.
func encodeToBase64URL(b []byte) string {
	const alphabet = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789-_"
	out := make([]byte, len(b))
	for i, v := range b {
		out[i] = alphabet[int(v)%len(alphabet)]
	}
	return string(out)
}

func (s *jwtService) VerifyOAuth2Token(ctx context.Context, providerName, oauthToken string) (string, error) {
	if strings.TrimSpace(oauthToken) == "" {
		return "", ErrInvalidToken
	}

	switch strings.ToLower(providerName) {
	case "microsoft":
		return s.handleMicrosoftOAuth2Token(ctx, oauthToken)
	case "github":
		return s.handleGitHubOAuth2Token(ctx, oauthToken)
	case "google", "":
		return s.handleGoogleOAuth2Token(ctx, oauthToken)
	case "facebook":
		return s.handleFacebookOAuth2Token(ctx, oauthToken)
	case "apple":
		return "", ErrInvalidToken // Apple OAuth2 token verification not implemented yet
	case "linkedin":
		return "", ErrInvalidToken // LinkedIn OAuth2 token verification not implemented yet
	case "pipedrive":
		return "", ErrInvalidToken // Pipedrive OAuth2 token verification not implemented yet
	case "rdstation":
		return "", ErrInvalidToken // RD Station OAuth2 token verification not implemented yet
	case "hubspot":
		return "", ErrInvalidToken // HubSpot OAuth2 token verification not implemented yet
	default:
		return "", ErrInvalidToken
	}
}

func (s *jwtService) handleMicrosoftOAuth2Token(ctx context.Context, oauthToken string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://graph.microsoft.com/v1.0/me", nil)
	if err != nil {
		return "", ErrInvalidToken
	}
	req.Header.Set("Authorization", "Bearer "+oauthToken)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", ErrInvalidToken
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", ErrInvalidToken
	}
	var data struct {
		Mail          string `json:"mail"`
		UserPrincipal string `json:"userPrincipalName"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return "", ErrInvalidToken
	}
	email := data.Mail
	if email == "" {
		email = data.UserPrincipal
	}
	if email == "" {
		return "", ErrInvalidToken
	}
	return strings.ToLower(email), nil
}

func (s *jwtService) handleGitHubOAuth2Token(ctx context.Context, oauthToken string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://api.github.com/user/emails", nil)
	if err != nil {
		return "", ErrInvalidToken
	}
	req.Header.Set("Authorization", "Bearer "+oauthToken)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", ErrInvalidToken
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", ErrInvalidToken
	}
	var emails []struct {
		Email    string `json:"email"`
		Primary  bool   `json:"primary"`
		Verified bool   `json:"verified"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&emails); err != nil || len(emails) == 0 {
		return "", ErrInvalidToken
	}
	for _, e := range emails {
		if e.Primary && e.Verified {
			return strings.ToLower(e.Email), nil
		}
	}
	return "", ErrInvalidToken
}

func (s *jwtService) handleGoogleOAuth2Token(ctx context.Context, oauthToken string) (string, error) {
	var clientID string
	if s.cfg != nil && s.cfg.AuthConfig != nil {
		clientID = s.cfg.AuthConfig.AuthProvidersConfig.Google.Web.ClientID
	}
	if clientID == "" {
		return "", ErrInvalidToken
	}

	payload, err := idtoken.Validate(ctx, oauthToken, clientID)
	if err != nil {
		return "", ErrInvalidToken
	}
	email, _ := payload.Claims["email"].(string)
	if email == "" {
		return "", ErrInvalidToken
	}
	return strings.ToLower(email), nil
}

func (s *jwtService) handleFacebookOAuth2Token(ctx context.Context, oauthToken string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://graph.facebook.com/me?fields=email", nil)
	if err != nil {
		return "", ErrInvalidToken
	}
	req.Header.Set("Authorization", "Bearer "+oauthToken)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", ErrInvalidToken
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", ErrInvalidToken
	}
	var data struct {
		Email string `json:"email"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil || data.Email == "" {
		return "", ErrInvalidToken
	}
	return strings.ToLower(data.Email), nil
}

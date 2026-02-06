package authentication

import (
	"context"
	"crypto/rsa"
	"fmt"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
	userstore "github.com/kubex-ecosystem/gnyx/internal/dsclient/datastore/user_store"
	sci "github.com/kubex-ecosystem/gnyx/internal/services/security/interfaces"
	gl "github.com/kubex-ecosystem/logz"
)

// JWTService handles JWT token generation and validation using kubexdb repositories
type JWTService struct {
	tokenRepo             sci.TokenRepo
	userRepo              userRepository
	privKey               *rsa.PrivateKey
	pubKey                *rsa.PublicKey
	certService           sci.ICertService
	refreshSecret         string
	idExpirationSecs      int64
	refreshExpirationSecs int64
}

type userRepository interface {
	GetByID(ctx context.Context, id string) (*userstore.Users, error)
}

var _ sci.TokenService = (*JWTService)(nil)

// jwtIDTokenClaims represents the claims embedded in ID tokens
type jwtIDTokenClaims struct {
	User *userstore.Users `json:"UserImpl"`
	jwt.RegisteredClaims
}

// jwtRefreshTokenClaims represents the claims embedded in refresh tokens
type jwtRefreshTokenClaims struct {
	UID string `json:"uid"`
	jwt.RegisteredClaims
}

// jwtRefreshTokenData holds refresh token metadata
type jwtRefreshTokenData struct {
	SS        string
	ID        string
	ExpiresIn time.Duration
}

// NewJWTService creates a new JWT service instance using kubexdb token repository
func NewJWTService(
	tokenRepo sci.TokenRepo,
	userRepo userRepository,
	certService sci.ICertService,
	privKey *rsa.PrivateKey,
	pubKey *rsa.PublicKey,
	refreshSecret string,
	idExpSecs,
	refreshExpSecs int64,
) *JWTService {
	if tokenRepo == nil {
		gl.Log("error", "TokenRepo cannot be nil") // pragma: allowlist secret // pragma: allowlist secret
		return nil
	}
	if userRepo == nil {
		gl.Log("error", "User repository cannot be nil")
		return nil
	}
	if privKey == nil && certService != nil { // pragma: allowlist secret
		privKey, err := certService.GetPrivateKey()
		if err != nil {
			gl.Log("error", fmt.Sprintf("Private key cannot be nil: %v", err)) // pragma: allowlist secret
			return nil
		}
		if privKey == nil { // pragma: allowlist secret
			gl.Log("error", "Private key cannot be nil") // pragma: allowlist secret
			return nil
		}
	} else if privKey == nil {
		gl.Log("error", "Private key cannot be nil") // pragma: allowlist secret
		return nil
	}
	if pubKey == nil && certService != nil { // pragma: allowlist secret
		pubKey, err := certService.GetPublicKey()
		if err != nil {
			gl.Log("error", fmt.Sprintf("Public key cannot be nil: %v", err)) // pragma: allowlist secret
			return nil
		}
		if pubKey == nil { // pragma: allowlist secret
			gl.Log("error", "Public key cannot be nil") // pragma: allowlist secret
			return nil
		}
	} else if pubKey == nil {
		gl.Log("error", "Public key cannot be nil") // pragma: allowlist secret
		return nil
	}
	// Set default expiration times
	if idExpSecs == 0 {
		idExpSecs = 3600 // 1 hour
	}
	if refreshExpSecs == 0 {
		refreshExpSecs = 604800 // 7 days
	}

	return &JWTService{
		tokenRepo:             tokenRepo,
		userRepo:              userRepo,
		certService:           certService,
		privKey:               privKey,
		pubKey:                pubKey,
		refreshSecret:         refreshSecret,
		idExpirationSecs:      idExpSecs,
		refreshExpirationSecs: refreshExpSecs,
	}
}

// NewPairFromUser generates a new token pair (ID token + Refresh token) for a user
func (s *JWTService) NewPairFromUser(ctx context.Context, u *userstore.Users, prevTokenID string) (*sci.TokenPair, error) {
	if u == nil || u.ID == nil || *u.ID == "" {
		return nil, gl.Errorf("invalid user payload for token generation")
	}
	// Delete previous refresh token if provided
	if prevTokenID != "" {
		if err := s.tokenRepo.DeleteRefreshToken(ctx, *u.ID, prevTokenID); err != nil {
			gl.Log("error", fmt.Sprintf("could not delete previous refresh token for uid: %v, tokenID: %v: %v", u.ID, prevTokenID, err))
			return nil, gl.Errorf("could not delete previous refresh token: %v", err)
		}
	}

	// Generate ID token
	idToken, err := s.generateIDToken(u)
	if err != nil {
		gl.Log("error", fmt.Sprintf("error generating id token for uid: %v: %v", u.ID, err))
		return nil, gl.Errorf("error generating id token: %v", err)
	}

	// Ensure refresh secret is set
	if s.refreshSecret == "" {
		jwtSecret, jwtSecretErr := s.certService.GetPrivPwd(nil) // pragma: allowlist secret
		if jwtSecretErr != nil {                                 // pragma: allowlist secret
			gl.Log("fatal", fmt.Sprintf("Error retrieving JWT secret key: %v", jwtSecretErr)) // pragma: allowlist secret
			return nil, jwtSecretErr
		}
		s.refreshSecret = jwtSecret // pragma: allowlist secret
	}

	// Generate refresh token
	refreshToken, err := s.generateRefreshToken(*u.ID)
	if err != nil {
		gl.Log("error", fmt.Sprintf("error generating refresh token for uid: %v: %v", u.ID, err))
		return nil, gl.Errorf("error generating refresh token: %v", err)
	}

	// Store refresh token metadata
	if err := s.tokenRepo.SetRefreshToken(ctx, *u.ID, refreshToken.ID, refreshToken.ExpiresIn); err != nil {
		gl.Log("error", fmt.Sprintf("error storing token ID for uid: %v: %v", u.ID, err))
		return nil, gl.Errorf("error storing token: %v", err)
	}

	return &sci.TokenPair{
		IDToken:      sci.IDToken{SS: idToken},
		RefreshToken: sci.RefreshToken{SS: refreshToken.SS, ID: refreshToken.ID, UID: *u.ID},
	}, nil
}

// SignOut revokes all refresh tokens for a user
func (s *JWTService) SignOut(ctx context.Context, uid string) error {
	if uid == "" {
		return gl.Errorf("user id is empty")
	}
	return s.tokenRepo.DeleteUserRefreshTokens(ctx, uid)
}

// ValidateIDToken validates an ID token and returns the user claims
func (s *JWTService) ValidateIDToken(tokenString string) (*userstore.Users, error) {
	claims, err := s.validateIDToken(tokenString)
	if err != nil {
		return nil, gl.Errorf("unable to validate or parse ID token: %v", err)
	}
	return claims.User, nil
}

// ValidateRefreshToken validates a refresh token string
func (s *JWTService) ValidateRefreshToken(tokenString string) (*sci.RefreshToken, error) {
	claims, err := s.validateRefreshToken(tokenString)
	if err != nil {
		return nil, gl.Errorf("unable to validate or parse refresh token: %v", err)
	}

	tokenUUID, err := uuid.Parse(claims.ID)
	if err != nil {
		return nil, gl.Errorf("claims ID could not be parsed as UUID: %v", err)
	}

	return &sci.RefreshToken{
		SS:  tokenString,
		ID:  tokenUUID.String(),
		UID: claims.UID,
	}, nil
}

// RenewToken renews an expired ID token using a valid refresh token
func (s *JWTService) RenewToken(ctx context.Context, refreshToken string) (*sci.TokenPair, error) {
	if len(strings.Split(refreshToken, ".")) != 3 {
		return nil, gl.Errorf("invalid refresh token format")
	}

	claims, err := s.validateRefreshToken(refreshToken)
	if err != nil {
		return nil, gl.Errorf("unable to validate refresh token: %v", err)
	}

	// Delete the old refresh token
	if err := s.tokenRepo.DeleteRefreshToken(ctx, claims.UID, claims.ID); err != nil {
		return nil, gl.Errorf("error deleting refresh token: %v", err)
	}

	user, err := s.userRepo.GetByID(ctx, claims.UID)
	if err != nil {
		return nil, gl.Errorf("error fetching user for uid %s: %v", claims.UID, err)
	}
	if user == nil || user.ID == nil {
		return nil, gl.Errorf("user not found for uid %s", claims.UID)
	}

	return s.NewPairFromUser(ctx, user, claims.ID)
}

// generateIDToken creates a signed ID token for a user
func (s *JWTService) generateIDToken(u *userstore.Users) (string, error) {
	if s.privKey == nil { // pragma: allowlist secret
		return "", gl.Errorf("private key is nil") // pragma: allowlist secret
	}
	if u == nil {
		return "", gl.Errorf("user model is nil")
	}
	if u.ID == nil || *u.ID == "" {
		return "", gl.Errorf("user id is missing")
	}

	unixTime := time.Now().Unix()
	tokenExp := unixTime + s.idExpirationSecs

	claims := jwtIDTokenClaims{
		User: u,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(time.Unix(unixTime, 0)),
			ExpiresAt: jwt.NewNumericDate(time.Unix(tokenExp, 0)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	ss, err := token.SignedString(s.privKey) // pragma: allowlist secret
	if err != nil {
		return "", gl.Errorf("failed to sign ID token: %v", err)
	}

	return ss, nil
}

// generateRefreshToken creates a signed refresh token
func (s *JWTService) generateRefreshToken(uid string) (*jwtRefreshTokenData, error) {
	currentTime := time.Now()
	tokenExp := currentTime.Add(time.Duration(s.refreshExpirationSecs) * time.Second)
	tokenID, err := uuid.NewRandom()
	if err != nil {
		return nil, gl.Errorf("failed to generate refresh token ID: %v", err)
	}

	claims := jwtRefreshTokenClaims{
		UID: uid,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(currentTime),
			ExpiresAt: jwt.NewNumericDate(tokenExp),
			ID:        tokenID.String(),
		},
	}

	if s.refreshSecret == "" {
		return nil, gl.Errorf("refresh token secret key is empty")
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	ss, err := token.SignedString([]byte(s.refreshSecret))
	if err != nil {
		return nil, gl.Errorf("failed to sign refresh token: %v", err)
	}

	return &jwtRefreshTokenData{
		SS:        ss,
		ID:        tokenID.String(),
		ExpiresIn: tokenExp.Sub(currentTime),
	}, nil
}

// validateIDToken validates and parses an ID token
func (s *JWTService) validateIDToken(tokenString string) (*jwtIDTokenClaims, error) {
	claims := &jwtIDTokenClaims{}

	if tokenString == "" {
		return nil, gl.Errorf("token string is empty")
	}
	if s.pubKey == nil {
		return nil, gl.Errorf("public key is nil")
	}
	if len(strings.Split(tokenString, ".")) != 3 {
		return nil, gl.Errorf("invalid token format")
	}
	if !strings.HasPrefix(tokenString, "ey") {
		return nil, gl.Errorf("invalid JWT token")
	}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, gl.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return s.pubKey, nil
	})

	if err != nil {
		return nil, gl.Errorf("error parsing token: %v", err)
	}
	if !token.Valid {
		return nil, gl.Errorf("token is invalid")
	}

	claims, ok := token.Claims.(*jwtIDTokenClaims)
	if !ok {
		return nil, gl.Errorf("token valid but couldn't parse claims")
	}
	if claims.User == nil {
		return nil, gl.Errorf("user claims are nil")
	}
	if claims.ExpiresAt.Time.Unix() < time.Now().Unix() {
		return nil, gl.Errorf("token has expired")
	}

	return claims, nil
}

// validateRefreshToken validates and parses a refresh token
func (s *JWTService) validateRefreshToken(tokenString string) (*jwtRefreshTokenClaims, error) {
	claims := &jwtRefreshTokenClaims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(s.refreshSecret), nil
	})

	if err != nil {
		return nil, err
	}
	if !token.Valid {
		return nil, gl.Errorf("refresh token is invalid")
	}

	claims, ok := token.Claims.(*jwtRefreshTokenClaims)
	if !ok {
		return nil, gl.Errorf("refresh token valid but couldn't parse claims")
	}

	return claims, nil
}

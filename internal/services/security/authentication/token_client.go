package authentication

import (
	"crypto/rsa"
	"fmt"

	svc "github.com/kubex-ecosystem/gnyx/internal/dsclient/adapters"
	userstore "github.com/kubex-ecosystem/gnyx/internal/dsclient/datastore/user_store"

	"github.com/kubex-ecosystem/gnyx/internal/module/kbx"
	crt "github.com/kubex-ecosystem/gnyx/internal/services/security/certificates"
	kri "github.com/kubex-ecosystem/gnyx/internal/services/security/external"
	sci "github.com/kubex-ecosystem/gnyx/internal/services/security/interfaces"

	ci "github.com/kubex-ecosystem/gnyx/interfaces"
	gl "github.com/kubex-ecosystem/logz"
	"gorm.io/gorm"
)

type TokenClientImpl struct {
	mapper                ci.IMapper[*sci.TSConfig]
	db                    *gorm.DB
	crtSrv                sci.ICertService
	KeyService            sci.IKeyService
	TokenService          sci.TokenService
	IDExpirationSecs      int64
	RefreshExpirationSecs int64
	tokenRepo             sci.TokenRepo
}

func (t *TokenClientImpl) LoadPublicKey() *rsa.PublicKey {
	pubKey, err := t.crtSrv.GetPublicKey()
	if err != nil {
		gl.Errorf("Error reading public key file: %v", err)
		return nil
	}
	return pubKey
}

func (t *TokenClientImpl) LoadPrivateKey() (*rsa.PrivateKey, error) {
	return t.crtSrv.GetPrivateKey()
}
func (t *TokenClientImpl) LoadTokenCfg() (sci.TokenService, int64, int64, error) {
	if t == nil {
		gl.Log("error", "TokenClient is nil, trying to create a new one")
		t = &TokenClientImpl{}
	}
	if t.crtSrv == nil {
		gl.Debug("crtService is nil, trying to create a new one")
		t.crtSrv = crt.NewCertService(kbx.DefaultGNyxKeyPath, kbx.DefaultGNyxCertPath) // pragma: allowlist secret
		if t.crtSrv == nil {
			gl.Log("fatal", "crtService is nil, unable to create a new one") // pragma: allowlist secret
		}
	}
	if t.db == nil {
		gl.Debug("database handle is nil, unable to create token repository")
		return nil, 0, 0, gl.Errorf("database handle is nil")
	}

	// Get RSA keys
	privKey, err := t.crtSrv.GetPrivateKey() // pragma: allowlist secret
	if err != nil {
		return nil, 0, 0, gl.Errorf("Error reading private key file: %v", err)
	}
	pubKey, pubKeyErr := t.crtSrv.GetPublicKey() // pragma: allowlist secret
	if pubKeyErr != nil {
		return nil, 0, 0, gl.Errorf("Error reading public key file: %v", pubKeyErr)
	}

	// ctx := context.Background()

	// Garantir valores padrão seguros
	if t.IDExpirationSecs == 0 {
		t.IDExpirationSecs = 3600 // 1 hora
	}
	if t.RefreshExpirationSecs == 0 {
		t.RefreshExpirationSecs = 604800 // 7 dias
	}

	// Setup keyring service (using FileKeyring instead of DBUS-based keyring)
	if t.KeyService == nil { //
		t.KeyService = kri.NewFileKeyService("", fmt.Sprintf("gnyx-%s", "jwt_secret"))
		if t.KeyService == nil {
			return nil, 0, 0, gl.Errorf("Error creating file keyring service: %v", err)
		}
	}

	// Get or generate JWT secret
	jwtSecret, jwtSecretErr := t.crtSrv.GetPrivPwd(nil) // pragma: allowlist secret
	if jwtSecretErr != nil {                            // pragma: allowlist secret
		gl.Fatalf("Error retrieving JWT secret key: %v", jwtSecretErr)
	}

	// Create token repository via kubexdb factory
	if t.tokenRepo == nil {
		t.tokenRepo = NewTokenRepo(t.db)
		if t.tokenRepo == nil {
			return nil, 0, 0, gl.Error("Failed to create token repository")
		}
	}

	userRepo := svc.NewPGRepository[userstore.Users](t.db)
	if userRepo == nil {
		return nil, 0, 0, gl.Error("Failed to create user repository")
	}

	// Create JWT service using new JWTService
	jwtService := NewJWTService(
		t.tokenRepo,
		userRepo,
		t.crtSrv,
		privKey,
		pubKey,
		jwtSecret,
		t.IDExpirationSecs,
		t.RefreshExpirationSecs,
	)

	if jwtService == nil {
		return nil, 0, 0, gl.Error("failed to create JWT service")
	}

	// Wrap JWTService to implement TokenService interface
	tokenService := jwtService

	return tokenService, t.IDExpirationSecs, t.RefreshExpirationSecs, nil
}

func NewTokenClient(crtService sci.ICertService, db *gorm.DB) *TokenClientImpl {
	if crtService == nil {
		gl.Errorf("error reading private key file: %v", "crtService is nil")
		return nil
	}
	if db == nil {
		gl.Error("database handle is nil, cannot initialize token client")
		return nil
	}
	tokenClient := &TokenClientImpl{
		crtSrv: crtService,
		db:     db,
	}

	return tokenClient
}

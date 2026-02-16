// Package auth contém configuração e lógica de autenticação.
package auth

import (
	"os"
	"time"

	"github.com/kubex-ecosystem/gnyx/internal/config"
	"github.com/kubex-ecosystem/gnyx/internal/module/kbx"

	kbxGet "github.com/kubex-ecosystem/kbx/get"
	kbxLoad "github.com/kubex-ecosystem/kbx/load"
)

// LoadConfig carrega configuração a partir de envs mínimas.
func LoadConfig(initArgs *config.ServerConfig) *config.AuthConfig {
	var (
		accessTokenTTL        time.Duration
		refreshTokenTTL       time.Duration
		accessTokenPrivateKey string
		accessTokenPublicKey  string
		issuer                string
		googleClientID        string
		googleClientSecret    string
		googleRedirectURL     string
	)

	if initArgs == nil {
		initArgs = &config.ServerConfig{}
	}

	accessTokenTTL = kbxGet.ValOrType(initArgs.Runtime.AccessTokenTTL, kbxGet.EnvOrType("KUBEX_AUTH_ACCESS_TTL", 15*time.Minute))
	refreshTokenTTL = kbxGet.ValOrType(initArgs.Runtime.RefreshTokenTTL, kbxGet.EnvOrType("KUBEX_AUTH_REFRESH_TTL", 30*24*time.Hour))

	accessTokenPrivateKey = os.ExpandEnv(kbxGet.ValOrType(initArgs.Runtime.PrivKeyPath, kbxGet.EnvOr("KUBEX_AUTH_PRIVATE_KEY", kbx.DefaultGNyxKeyPath)))
	accessTokenPublicKey = os.ExpandEnv(kbxGet.ValOrType(initArgs.Runtime.PubCertKeyPath, kbxGet.EnvOr("KUBEX_AUTH_PUBLIC_KEY", kbx.DefaultGNyxCertPath)))
	issuer = kbxGet.ValOrType(initArgs.Runtime.Issuer, kbxGet.EnvOr("KUBEX_AUTH_ISSUER", "gnyx"))

	googleCfg := kbxLoad.NewVendorAuthConfig(initArgs.ProvidersConfig)
	googleCfg.Web.ClientID = kbxGet.ValOrType(googleClientID, kbxGet.EnvOr("GOOGLE_CLIENT_ID", ""))
	googleCfg.Web.ClientSecret = kbxGet.ValOrType(googleClientSecret, kbxGet.EnvOr("GOOGLE_CLIENT_SECRET", ""))
	googleCfg.Web.RedirectURL = kbxGet.ValOrType(googleRedirectURL, kbxGet.EnvOr("GOOGLE_REDIRECT_URL", ""))
	googleCfg.Web.Scopes = kbxGet.EnvOrType("GOOGLE_OAUTH_SCOPES", []string{"openid", "email", "profile"})

	return &config.AuthConfig{
		AccessTokenTTL:        accessTokenTTL,
		RefreshTokenTTL:       refreshTokenTTL,
		AccessTokenPrivateKey: accessTokenPrivateKey,
		AccessTokenPublicKey:  accessTokenPublicKey,
		Issuer:                issuer,
		AuthProvidersConfig: config.AuthProvidersConfig{
			Google: googleCfg.AuthClientConfig,
		},
	}
}

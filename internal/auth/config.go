// Package auth contém configuração e lógica de autenticação.
package auth

import (
	"os"
	"time"

	"github.com/kubex-ecosystem/gnyx/internal/config"
	"github.com/kubex-ecosystem/gnyx/internal/module/kbx"

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

	accessTokenTTL = kbx.GetValueOrDefaultSimple(initArgs.Runtime.AccessTokenTTL, kbx.GetEnvOrDefaultWithType("KUBEX_AUTH_ACCESS_TTL", 15*time.Minute))
	refreshTokenTTL = kbx.GetValueOrDefaultSimple(initArgs.Runtime.RefreshTokenTTL, kbx.GetEnvOrDefaultWithType("KUBEX_AUTH_REFRESH_TTL", 30*24*time.Hour))
	accessTokenPrivateKey = kbx.GetValueOrDefaultSimple(initArgs.Runtime.PrivKeyPath, kbx.GetEnvOrDefault("KUBEX_AUTH_PRIVATE_KEY", "kubex_dev_rsa"))
	accessTokenPublicKey = kbx.GetValueOrDefaultSimple(initArgs.Runtime.PubCertKeyPath, kbx.GetEnvOrDefault("KUBEX_AUTH_PUBLIC_KEY", kbx.GetValueOrDefaultSimple(os.ExpandEnv("$HOME/.gnyx/github.com/kubex-ecosystem/gnyx/certs/be_rsa.pub"), "")))
	issuer = kbx.GetValueOrDefaultSimple(initArgs.Runtime.Issuer, kbx.GetEnvOrDefault("KUBEX_AUTH_ISSUER", "gnyx"))

	googleCfg := kbxLoad.NewVendorAuthConfig(initArgs.ProvidersConfig)
	googleCfg.Web.ClientID = kbx.GetValueOrDefaultSimple(googleClientID, kbx.GetEnvOrDefault("GOOGLE_CLIENT_ID", ""))
	googleCfg.Web.ClientSecret = kbx.GetValueOrDefaultSimple(googleClientSecret, kbx.GetEnvOrDefault("GOOGLE_CLIENT_SECRET", ""))
	googleCfg.Web.RedirectURL = kbx.GetValueOrDefaultSimple(googleRedirectURL, kbx.GetEnvOrDefault("GOOGLE_REDIRECT_URL", ""))
	googleCfg.Web.Scopes = kbx.GetEnvOrDefaultWithType("GOOGLE_OAUTH_SCOPES", []string{"openid", "email", "profile"})

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

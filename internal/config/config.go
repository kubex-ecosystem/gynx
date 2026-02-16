// Package config fornece a configuração e inicialização dos serviços do backend.
package config

import (
	"os"
	"path/filepath"
	"time"

	"cloud.google.com/go/auth/credentials/idtoken"
	"github.com/kubex-ecosystem/gnyx/internal/types"
	"github.com/kubex-ecosystem/kbx"

	kbxMod "github.com/kubex-ecosystem/gnyx/internal/module/kbx"
	kbxGet "github.com/kubex-ecosystem/kbx/get"
	gl "github.com/kubex-ecosystem/logz"
)

// Config agrega todas as dependências externas necessárias para inicializar os serviços do backend.
type Config struct {
	ServerConfig *ServerConfig `json:"server_config,omitempty" yaml:"server_config,omitempty" toml:"server_config,omitempty" mapstructure:"server_config,omitempty"`
	AuthConfig   *AuthConfig   `json:"auth_config,omitempty" yaml:"auth_config,omitempty" toml:"auth_config,omitempty" mapstructure:"auth_config,omitempty"`

	Database             *PGConfig          `json:"database,omitempty" yaml:"database,omitempty" toml:"database,omitempty" mapstructure:"database,omitempty"`
	DataService          *DataServiceConfig `json:"data_service,omitempty" yaml:"data_service,omitempty" toml:"data_service,omitempty" mapstructure:"data_service,omitempty"`
	MailerConfigFilePath string             `json:"mailer_config_file_path,omitempty" yaml:"mailer_config_file_path,omitempty" toml:"mailer_config_file_path,omitempty" mapstructure:"mailer_config_file_path,omitempty"`
	MailerConfig         *kbx.MailConfig    `json:"mailer_config,omitempty" yaml:"mailer_config,omitempty" toml:"mailer_config,omitempty" mapstructure:"mailer_config,omitempty"`
	TemplatesDir         string             `json:"templates_dir,omitempty" yaml:"templates_dir,omitempty" toml:"templates_dir,omitempty" mapstructure:"templates_dir,omitempty"`
	Invite               *InviteConfig      `json:"invite,omitempty" yaml:"invite,omitempty" toml:"invite,omitempty" mapstructure:"invite,omitempty"`
}

// InviteConfig controla opções de envio e branding.
type InviteConfig struct {
	BaseURL     string        `json:"base_url,omitempty" yaml:"base_url,omitempty" toml:"base_url,omitempty" mapstructure:"base_url,omitempty"`
	SenderName  string        `json:"sender_name,omitempty" yaml:"sender_name,omitempty" toml:"sender_name,omitempty" mapstructure:"sender_name,omitempty"`
	SenderEmail string        `json:"sender_email,omitempty" yaml:"sender_email,omitempty" toml:"sender_email,omitempty" mapstructure:"sender_email,omitempty"`
	CompanyName string        `json:"company_name,omitempty" yaml:"company_name,omitempty" toml:"company_name,omitempty" mapstructure:"company_name,omitempty"`
	DefaultTTL  time.Duration `json:"default_ttl,omitempty" yaml:"default_ttl,omitempty" toml:"default_ttl,omitempty" mapstructure:"default_ttl,omitempty"`
}

type OAuth21Config struct {
	ClientID     string `json:"client_id" env:"OAUTH2_CLIENT_ID"`
	ClientSecret string `json:"client_secret" env:"OAUTH2_CLIENT_SECRET"` // Cuidado com esse log!
	RedirectURL  string `json:"redirect_url" env:"OAUTH2_REDIRECT_URL"`

	ComputeTokenFormat idtoken.Validator  `json:"-"`
	IDTokenVerifier    *idtoken.Validator `json:"-"`

	ProviderName string `json:"provider_name" env:"OAUTH2_PROVIDER_NAME"`
	Issuer       string `json:"issuer" env:"OAUTH2_ISSUER"`
}

type AuthOAuthClientConfig = kbx.AuthOAuthClientConfig
type AuthClientConfig = kbx.AuthClientConfig
type VendorAuthConfig = kbx.VendorAuthConfig
type AuthProvidersConfig = kbx.AuthProvidersConfig

// AuthConfig define parâmetros de autenticação.
// Codex: se já existir config global no Kubex, integrar isso lá depois.
type AuthConfig struct {
	AccessTokenTTL        time.Duration       `json:"access_token_ttl,omitempty" yaml:"access_token_ttl,omitempty" toml:"access_token_ttl,omitempty" mapstructure:"access_token_ttl,omitempty"`
	RefreshTokenTTL       time.Duration       `json:"refresh_token_ttl,omitempty" yaml:"refresh_token_ttl,omitempty" toml:"refresh_token_ttl,omitempty" mapstructure:"refresh_token_ttl,omitempty"`
	AccessTokenPrivateKey string              `json:"access_token_private_key,omitempty" yaml:"access_token_private_key,omitempty" toml:"access_token_private_key,omitempty" mapstructure:"access_token_private_key,omitempty"` // PEM private key (RSA)
	AccessTokenPublicKey  string              `json:"access_token_public_key,omitempty" yaml:"access_token_public_key,omitempty" toml:"access_token_public_key,omitempty" mapstructure:"access_token_public_key,omitempty"`     // PEM public key (RSA)
	Issuer                string              `json:"issuer,omitempty" yaml:"issuer,omitempty" toml:"issuer,omitempty" mapstructure:"issuer,omitempty"`
	AuthProvidersConfig   AuthProvidersConfig `json:"auth_providers_config,omitempty" yaml:"auth_providers_config,omitempty" toml:"auth_providers_config,omitempty" mapstructure:"auth_providers_config,omitempty"`
}

// DataServiceConfig define onde está a config do DS e qual database usar.
type DataServiceConfig struct {
	ConfigPath string `json:"config_path,omitempty" yaml:"config_path,omitempty" toml:"config_path,omitempty" mapstructure:"config_path,omitempty"`
	DBName     string `json:"db_name,omitempty" yaml:"db_name,omitempty" toml:"db_name,omitempty" mapstructure:"db_name,omitempty"`
}

// LoadConfig carrega a configuração a partir das variáveis de ambiente.
func LoadConfig() *Config {
	ref := types.NewReference("gnyx").GetReference()
	// Base URL para links públicos (convites, etc.)
	baseURL := kbxGet.EnvOr("GNYX_PUBLIC_URL", "http://localhost:4000")
	if baseURL == "" {
		baseURL = kbxGet.EnvOr("INVITE_BASE_URL", "")
	}
	if baseURL == "" {
		baseURL = kbxGet.ValueOrIf(kbxGet.EnvOr("KUBEX_ENV", "development") == "production",
			"https://api.gnyx.app",
			"http://localhost:4000",
		)
	}
	defaultTTL := kbxGet.EnvOrType("INVITE_EXPIRATION", 7*24*time.Hour)
	configPath := os.ExpandEnv(kbxGet.EnvOr("GNYX_CONFIG_PATH", kbxMod.DefaultGNyxConfigPath))
	dataServiceConfig := &DataServiceConfig{
		ConfigPath: os.ExpandEnv(kbxGet.EnvOr("KUBEX_DS_CONFIG_PATH", kbxMod.DefaultKubexDomusConfigPath)),
		DBName:     kbxGet.EnvOr("KUBEX_DS_DB_NAME", "postgres"),
	}
	pubKeyPath := os.ExpandEnv(kbxGet.EnvOr("GNYX_PUBLIC_KEY_PATH", kbxMod.DefaultGNyxCertPath))
	privKeyPath := os.ExpandEnv(kbxGet.EnvOr("GNYX_PRIVATE_KEY_PATH", kbxMod.DefaultGNyxKeyPath))
	InitArgs := kbxMod.NewInitArgs(
		os.ExpandEnv(configPath),
		filepath.Ext(configPath)[1:],
		os.ExpandEnv(dataServiceConfig.ConfigPath),
		filepath.Ext(dataServiceConfig.ConfigPath)[1:],
		os.ExpandEnv(kbxGet.EnvOr("GNYX_ENV_PATH", kbxMod.DefaultGNyxEnvPath)),
		os.ExpandEnv(kbxGet.EnvOr("GNYX_LOG_FILE_PATH", kbxMod.DefaultGNyxLogPath)),
		ref.GetName(),
		kbxGet.EnvOrType("GNYX_DEBUG_MODE", false),
		kbxGet.EnvOrType("GNYX_RELEASE_MODE", false),
		kbxGet.EnvOrType("GNYX_CONFIDENCIAL_MODE", false),
		kbxGet.EnvOrType("GNYX_PORT", "4000"),
		kbxGet.EnvOrType("GNYX_HOST", "localhost"),
		pubKeyPath,
		privKeyPath,
		kbxGet.EnvOr("GNYX_PRIVATE_KEY_PASSWORD", ""),
		kbxGet.EnvOr("GNYX_TEMPLATES_DIR", kbxMod.DefaultTemplatesDir),
	)

	glgAuthConfig := loadGoogleAuthConfig(InitArgs)

	authCfg := &AuthConfig{
		AccessTokenTTL:        kbxGet.ValOrType(InitArgs.AccessTokenTTL, kbxGet.EnvOrType("KUBEX_AUTH_ACCESS_TTL", 15*time.Minute)),
		RefreshTokenTTL:       kbxGet.ValOrType(InitArgs.RefreshTokenTTL, kbxGet.EnvOrType("KUBEX_AUTH_REFRESH_TTL", 30*24*time.Hour)),
		AccessTokenPrivateKey: kbxGet.ValOrType(InitArgs.PrivKeyPath, kbxGet.EnvOr("KUBEX_AUTH_PRIVATE_KEY", "kubex_dev_rsa")),
		AccessTokenPublicKey:  kbxGet.ValOrType(InitArgs.PubCertKeyPath, kbxGet.EnvOr("KUBEX_AUTH_PUBLIC_KEY", kbxGet.ValOrType(os.ExpandEnv("$HOME/.gnyx/certs/be_rsa.pub"), ""))),
		Issuer:                kbxGet.ValOrType(InitArgs.Issuer, kbxGet.EnvOr("KUBEX_AUTH_ISSUER", "gnyx")),
		AuthProvidersConfig: AuthProvidersConfig{
			Google: AuthClientConfig{
				Web: *glgAuthConfig,
			},
		},
	}
	var srvConfig kbx.SrvConfig
	srvConfigPtr, _ := kbx.LoadConfigOrDefault[kbx.SrvConfig](InitArgs.ConfigFile, true)
	if srvConfigPtr == nil {
		srvConfig = kbx.NewSrvArgs()
	} else {
		srvConfig = *srvConfigPtr
	}

	return &Config{
		ServerConfig: &ServerConfig{
			SrvConfig: srvConfig,
			InitArgs:  InitArgs,
		},
		AuthConfig:           authCfg,
		Database:             ConfigFromEnv(),
		DataService:          dataServiceConfig,
		MailerConfigFilePath: kbxGet.EnvOr("MAILER_CONFIG_PATH", ""),
		TemplatesDir:         kbxGet.EnvOr("INVITE_TEMPLATES_DIR", kbxMod.DefaultTemplatesDir),
		Invite: &InviteConfig{
			BaseURL:     baseURL,
			SenderName:  kbxGet.EnvOr("INVITE_SENDER_NAME", "Equipe Kubex"),
			SenderEmail: kbxGet.EnvOr("INVITE_SENDER_EMAIL", "convites@kubex.world"),
			CompanyName: kbxGet.EnvOr("INVITE_COMPANY_NAME", "Kubex Ecosystem"),
			DefaultTTL:  defaultTTL,
		},
	}
}

// loadGoogleAuthConfig carrega ClientID/Secret/Redirect a partir de env ou do
// client_secret.json padrão do OAuth da Google.
func loadGoogleAuthConfig(initArgs *kbxMod.InitArgs) *AuthOAuthClientConfig {
	cfg, err := kbx.LoadConfigOrDefault[kbx.VendorAuthConfig](kbxGet.EnvOr("KUBEX_GOOGLE_CREDENTIALS_PATH", os.ExpandEnv(kbxMod.DefaultGoogleAuthClientPath)), true)
	if err != nil && cfg == nil {
		gl.Debugf("google oauth config load failed: %v", err)
		return nil
	}
	acCfg := cfg.AuthClientConfig.Web

	return &acCfg
}

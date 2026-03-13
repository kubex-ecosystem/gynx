package config

import (
	"time"

	kbxMod "github.com/kubex-ecosystem/gnyx/internal/module/kbx"
	"github.com/kubex-ecosystem/kbx"
	kbxGet "github.com/kubex-ecosystem/kbx/get"
)

// ServerConfig holds configuration for the gateway server
type ServerConfig struct {
	*kbxMod.InitArgs
	// Addr            string
	ProvidersConfig string
	LLMConfig       *kbx.LLMConfig
	// EnableCORS      bool
	DefaultTTL   time.Duration
	DBConfigFile string
}

func NewServerConfig() *ServerConfig {
	InitArgs := kbxMod.NewArgs("gnyx")
	kbxMod.InitArgsDefaults()

	cfg := &ServerConfig{
		InitArgs: InitArgs,
	}

	cfg.Files.ProvidersConfig = InitArgs.Files.ProvidersConfig

	cfg.Runtime.Bind = InitArgs.Runtime.Bind
	cfg.Runtime.Port = InitArgs.Runtime.Port
	cfg.Runtime.Host = InitArgs.Runtime.Host

	cfg.Basic.AppName = InitArgs.Basic.AppName
	cfg.Basic.Environment = kbxGet.EnvOr("KUBEX_ENV", kbxGet.EnvOr("ENV", "development"))
	cfg.Basic.CORSEnabled = InitArgs.Basic.CORSEnabled

	// cfg.SrvConfig.Runtime.TrustedProxies = InitArgs.TrustedProxies
	cfg.Runtime.AccessTokenTTL = InitArgs.Runtime.AccessTokenTTL
	cfg.Runtime.RefreshTokenTTL = InitArgs.Runtime.RefreshTokenTTL
	cfg.Runtime.PubCertKeyPath = InitArgs.Runtime.PubCertKeyPath
	cfg.Runtime.PubKeyPath = InitArgs.Runtime.PubKeyPath
	cfg.Runtime.PrivKeyPath = InitArgs.Runtime.PrivKeyPath
	cfg.Runtime.Issuer = InitArgs.Runtime.Issuer
	cfg.Runtime.AccessTokenTTL = InitArgs.Runtime.AccessTokenTTL
	cfg.Runtime.RefreshTokenTTL = InitArgs.Runtime.RefreshTokenTTL

	cfg.DefaultTTL = InitArgs.Auth.Invite.DefaultTTL
	// cfg.DBConfigFile = InitArgs.DBConfigFile
	// cfg.Mapper = types.NewMapperType(&cfg, cfg.ConfigFile)

	cfg.LLMConfig = &kbx.LLMConfig{
		Providers:   make(map[string]*kbx.LLMProviderConfig),
		Development: kbx.LLMDevelopmentConfig{},
	}

	return cfg
}

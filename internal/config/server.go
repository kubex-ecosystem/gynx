package config

import (
	"os"
	"path/filepath"
	"time"

	kbxMod "github.com/kubex-ecosystem/gnyx/internal/module/kbx"
	"github.com/kubex-ecosystem/kbx"
	kbxGet "github.com/kubex-ecosystem/kbx/get"
)

// ServerConfig holds configuration for the gateway server
type ServerConfig struct {
	kbx.SrvConfig
	InitArgs *kbxMod.InitArgs
	// Addr            string
	ProvidersConfig string
	// EnableCORS      bool
	DefaultTTL        time.Duration
	DataServiceConfig DataServiceConfig
}

func NewServerConfig() *ServerConfig {
	// ref := kbxMod.NewReference("github.com/kubex-ecosystem/gnyx_server").GetReference()
	baseURL := kbxMod.GetValueOrDefaultIf(kbxMod.GetEnvOrDefault("KUBEX_ENV", "development") == "production",
		"https://api.gnyx.app",
		"http://localhost:5000",
	)
	defaultTTL := kbxMod.GetEnvOrDefaultWithType("INVITE_EXPIRATION", 7*24*time.Hour)
	configPath := os.ExpandEnv(kbxMod.GetEnvOrDefault("KUBEX_BE_CONFIG_PATH", "/ALL/KUBEX/SHOWCASE/projects/gnyx/config/config.json"))
	dataServiceConfig := DataServiceConfig{
		ConfigPath: os.ExpandEnv(kbxMod.GetEnvOrDefault("KUBEX_DS_CONFIG_PATH", "/ALL/KUBEX/SHOWCASE/projects/domus/configs/config.json")),
		DBName:     kbxMod.GetEnvOrDefault("KUBEX_DS_DB_NAME", "postgres"),
	}
	pubKeyPath := os.ExpandEnv(kbxMod.GetEnvOrDefault("KUBEX_BE_PUBLIC_KEY_PATH", kbxMod.DefaultGNyxCertPath))
	privKeyPath := os.ExpandEnv(kbxMod.GetEnvOrDefault("KUBEX_BE_PRIVATE_KEY_PATH", kbxMod.DefaultGNyxKeyPath))
	InitArgs := kbxMod.NewInitArgs(
		os.ExpandEnv(configPath),
		filepath.Ext(configPath)[1:],
		os.ExpandEnv(dataServiceConfig.ConfigPath),
		filepath.Ext(dataServiceConfig.ConfigPath)[1:],
		os.ExpandEnv(kbxMod.GetEnvOrDefault("KUBEX_BE_ENV_PATH", "/ALL/KUBEX/SHOWCASE/projects/gnyx/.env")),
		os.ExpandEnv(kbxMod.GetEnvOrDefault("KUBEX_BE_LOG_FILE_PATH", "/ALL/KUBEX/SHOWCASE/projects/gnyx/gnyx.log")),
		kbxMod.GetEnvOrDefault("KUBEX_BE_PROCESS_NAME", "gnyx"),
		kbxMod.GetEnvOrDefaultWithType("KUBEX_BE_DEBUG_MODE", false),
		kbxMod.GetEnvOrDefaultWithType("KUBEX_BE_RELEASE_MODE", false),
		kbxMod.GetEnvOrDefaultWithType("KUBEX_BE_CONFIDENCIAL_MODE", false),
		kbxMod.GetEnvOrDefaultWithType("KUBEX_BE_PORT", "5000"),
		kbxMod.GetEnvOrDefaultWithType("KUBEX_BE_HOST", "localhost"),
		pubKeyPath,
		privKeyPath,
		kbxMod.GetEnvOrDefault("KUBEX_BE_PRIVATE_KEY_PASSWORD", ""),
	)
	cfg := &ServerConfig{
		SrvConfig: kbx.NewSrvArgs(),
	}
	cfg.InitArgs = InitArgs
	cfg.SrvConfig.Runtime.Bind = InitArgs.Bind
	cfg.SrvConfig.Runtime.Port = InitArgs.Port
	cfg.SrvConfig.Runtime.Host = InitArgs.Host
	cfg.SrvConfig.Basic.AppName = InitArgs.Name
	cfg.SrvConfig.Basic.Environment = kbxGet.EnvOr("KUBEX_ENV", kbxGet.EnvOr("ENV", "development"))
	// cfg.SrvConfig.Basic.BaseURL = baseURL
	cfg.ProvidersConfig = InitArgs.ProvidersConfig
	cfg.SrvConfig.Basic.CORSEnabled = kbxGet.EnvOrType("KUBEX_BE_ENABLE_CORS", true)
	cfg.SrvConfig.Runtime.AccessTokenTTL = kbxGet.EnvOrType("KUBEX_BE_ACCESS_TOKEN_TTL", 15*time.Minute)
	cfg.SrvConfig.Runtime.RefreshTokenTTL = kbxGet.EnvOrType("KUBEX_BE_REFRESH_TOKEN_TTL", 7*24*time.Hour)
	cfg.SrvConfig.Runtime.Host = baseURL

	cfg.DefaultTTL = defaultTTL
	cfg.DataServiceConfig = dataServiceConfig
	// cfg.Mapper = types.NewMapperType(&cfg, cfg.ConfigFile)
	return cfg
}

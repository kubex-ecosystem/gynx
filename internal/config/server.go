package config

import (
	"net"
	"net/url"
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
	LLMConfig       *kbx.LLMConfig
	// EnableCORS      bool
	DefaultTTL        time.Duration
	DataServiceConfig DataServiceConfig
}

func NewServerConfig() *ServerConfig {
	// ref := kbxMod.NewReference("github.com/kubex-ecosystem/gnyx_server").GetReference()
	scheme := os.ExpandEnv(kbxMod.GetEnvOrDefault("KUBEX_GNYX_SCHEME", "http"))
	host := os.ExpandEnv(kbxMod.GetEnvOrDefault("KUBEX_GNYX_HOST", kbxMod.DefaultServerHost))
	addr := net.JoinHostPort(host, kbxMod.GetEnvOrDefault("KUBEX_GNYX_PORT", "5000"))
	url := url.URL{Scheme: scheme, Host: addr}
	baseURL := kbxGet.ValueOrIf(kbxMod.GetEnvOrDefault("KUBEX_ENV", "development") == "production",
		"https://api.kubex.world",
		url.String(),
	)
	defaultTTL := kbxMod.GetEnvOrDefaultWithType("INVITE_EXPIRATION", 7*24*time.Hour)
	configPath := os.ExpandEnv(kbxMod.GetEnvOrDefault("KUBEX_GNYX_CONFIG_PATH", kbxMod.DefaultGNyxConfigPath))
	dataServiceConfig := DataServiceConfig{
		ConfigPath: os.ExpandEnv(kbxMod.GetEnvOrDefault("KUBEX_DOMUS_CONFIG_PATH", kbxMod.DefaultKubexDomusConfigPath)),
		DBName:     kbxMod.GetEnvOrDefault("KUBEX_DOMUS_DB_NAME", kbxMod.DefaultKubexDomusConfigPath),
	}
	pubKeyPath := os.ExpandEnv(kbxMod.GetEnvOrDefault("KUBEX_GNYX_PUBLIC_KEY_PATH", kbxMod.DefaultGNyxCertPath))
	privKeyPath := os.ExpandEnv(kbxMod.GetEnvOrDefault("KUBEX_GNYX_PRIVATE_KEY_PATH", kbxMod.DefaultGNyxKeyPath))
	InitArgs := kbxMod.NewInitArgs(
		os.ExpandEnv(configPath),
		filepath.Ext(configPath)[1:],
		os.ExpandEnv(dataServiceConfig.ConfigPath),
		filepath.Ext(dataServiceConfig.ConfigPath)[1:],
		os.ExpandEnv(kbxGet.EnvOr("KUBEX_GNYX_ENV_PATH", kbxMod.DefaultGNyxEnvPath)),
		os.ExpandEnv(kbxGet.EnvOr("KUBEX_GNYX_LOG_FILE_PATH", kbxMod.DefaultGNyxLogPath)),
		os.ExpandEnv(kbxGet.EnvOr("KUBEX_GNYX_PROCESS_NAME", kbxMod.DefaultServerHost)),
		kbxGet.EnvOrType("KUBEX_GNYX_DEBUG_MODE", false),
		kbxGet.EnvOrType("KUBEX_GNYX_RELEASE_MODE", false),
		kbxGet.EnvOrType("KUBEX_GNYX_CONFIDENTIAL_MODE", false),
		kbxGet.EnvOrType("KUBEX_GNYX_PORT", "5000"),
		os.ExpandEnv(kbxGet.EnvOr("KUBEX_GNYX_BIND", kbxMod.DefaultServerBind)),
		pubKeyPath,
		privKeyPath,
		kbxMod.GetEnvOrDefault("KUBEX_GNYX_PRIVATE_KEY_PASSWORD", ""),
		kbxMod.GetEnvOrDefault("KUBEX_GNYX_TEMPLATES_DIR", kbxMod.DefaultTemplatesDir),
		kbxGet.EnvOrType("KUBEX_GNYX_DISABLE_UI", false),
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
	cfg.SrvConfig.Basic.CORSEnabled = kbxGet.EnvOrType("KUBEX_GNYX_ENABLE_CORS", true)
	cfg.SrvConfig.Runtime.AccessTokenTTL = kbxGet.EnvOrType("KUBEX_GNYX_ACCESS_TOKEN_TTL", 15*time.Minute)
	cfg.SrvConfig.Runtime.RefreshTokenTTL = kbxGet.EnvOrType("KUBEX_GNYX_REFRESH_TOKEN_TTL", 7*24*time.Hour)
	cfg.SrvConfig.Runtime.Host = baseURL

	cfg.DefaultTTL = defaultTTL
	cfg.DataServiceConfig = dataServiceConfig
	// cfg.Mapper = types.NewMapperType(&cfg, cfg.ConfigFile)

	cfg.LLMConfig = &kbx.LLMConfig{
		Providers:   make(map[string]*kbx.LLMProviderConfig),
		Development: kbx.LLMDevelopmentConfig{},
	}

	return cfg
}

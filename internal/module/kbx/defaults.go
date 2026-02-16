// Package kbx has default configuration values
package kbx

// Default configuration constants
const (
	DefaultKubexConfigDir = "$HOME/.kubex/gnyx"

	DefaultGNyxCAPath   = "$HOME/.kubex/gnyx/ca-cert.pem"
	DefaultGNyxKeyPath  = "$HOME/.kubex/gnyx/gnyx.key" // Priv
	DefaultGNyxCertPath = "$HOME/.kubex/gnyx/gnyx.crt"

	DefaultGNyxConfigPath = "$HOME/.kubex/gnyx/config/config.json"
	DefaultGNyxEnvPath    = "$HOME/.kubex/gnyx/config/.env"
	DefaultGNyxLogPath    = "$HOME/.kubex/gnyx/logs/gnyx_process.log.txt"

	DefaultKubexDomusConfigPath = "$HOME/.kubex/domus/config/config.json"

	DefaultGoogleAuthClientPath = "$HOME/.kubex/gnyx/config/google_auth_client.json"

	DefaultVaultDir = "$HOME/.kubex/gnyx/secrets"

	DefaultVaultKey = "kubex_kubex-jwt_secret.secret"

	DefaultTemplatesDir = "templates"
)

const DefaultProvidersConfig = "$HOME/.kubex/gnyx/config/providers.yaml"

// Default General Rate Limiting Settings
const (
	DefaultRateLimitLimit  = 100
	DefaultRateLimitBurst  = 100
	DefaultRateLimitJitter = 0.1
	DefaultRequestWindow   = 1 * 60 * 1000 // 1 minute
)

// Default HTTP Client Settings
const (
	DefaultTLSHandshakeTimeout   = 10 * 1000 // 10 seconds
	DefaultExpectContinueTimeout = 1 * 1000  // 1 second
	DefaultResponseHeaderTimeout = 5 * 1000  // 5 seconds

	DefaultTimeout         = 30 * 1000 // 30 seconds
	DefaultKeepAlive       = 30 * 1000 // 30 seconds
	DefaultMaxConnsPerHost = 100
)

// Default Generic Retry and Connection Settings
const (
	DefaultMaxRetries = 3
	DefaultRetryDelay = 1 * 1000 // 1 second

	DefaultMaxIdleConns        = 100
	DefaultMaxIdleConnsPerHost = 100
	DefaultIdleConnTimeout     = 90 * 1000 // 90 seconds
)

// Default LLM Settings
const (
	DefaultLLMProvider    = "gemini"
	DefaultLLMModel       = "gemini-2.0-flash"
	DefaultLLMMaxTokens   = 1024
	DefaultLLMTemperature = 0.3
)

// Default Server Settings
const (
	DefaultServerPort = "5000"
	DefaultServerBind = "0.0.0.0"
	DefaultServerHost = "localhost"
)

// Default HTTP Basic Header Security Keys
const (
	HeaderRequestIDKey = "X-Request-ID"
	CookieSessionIDKey = "session_id"
)

// Default Authentication Types
const (
	AuthTypeNone   = "none"
	AuthTypeOIDC   = "oidc"
	AuthTypeBasic  = "basic"
	AuthTypeBearer = "bearer"
	AuthTypeAPIKey = "api_key" // pragma: allowlist secret
)

// Default Database Settings

type DBNameKey string

const (
	ContextDBNameKey      = DBNameKey("postgres")
	DefaultVolumesDir     = "$HOME/.kubex/domus/volumes"
	DefaultMongoVolume    = "$HOME/.kubex/domus/volumes/mongo"
	DefaultRedisVolume    = "$HOME/.kubex/domus/volumes/redis"
	DefaultPostgresVolume = "$HOME/.kubex/domus/volumes/postgresql"
	DefaultRabbitMQVolume = "$HOME/.kubex/domus/volumes/rabbitmq"
)

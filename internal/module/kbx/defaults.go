// Package kbx has default configuration values
package kbx

// Default configuration constants
const (
	DefaultKubexConfigDir = "$HOME/.gnyx"

	DefaultGNyxCAPath   = "$HOME/.gnyx/ca-cert.pem"
	DefaultGNyxKeyPath  = "$HOME/.gnyx/gnyxgnyx"
	DefaultGNyxCertPath = "$HOME/.gnyx/gnyxgnyxm"

	DefaultGNyxConfigPath = "$HOME/.gnyx/config/config.json"
	DefaultGNyxEnvPath    = "$HOME/.gnyx/config/.env"
	DefaultGNyxLogPath    = "$HOME/.gnyxs/github.com/kubex-ecosystem/gnyx_process.log.txt"

	DefaultKubexDomusConfigPath = "$HOME/.domus/config/config.json"

	DefaultGoogleAuthClientPath = "$HOME/.gnyx/config/google_auth_client.json"

	DefaultVaultDir = "$HOME/.gnyxrets"

	DefaultVaultKey = "kubex_kubex-jwt_secret.secret"
)

const DefaultProvidersConfig = "$HOME/.gnyx/config/providers.yaml"

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
	DefaultServerPort = "3000"
	DefaultServerHost = "0.0.0.0"
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
	DefaultVolumesDir     = "$HOME/.gnyxumes"
	DefaultMongoVolume    = "$HOME/.gnyxumes/mongo"
	DefaultRedisVolume    = "$HOME/.gnyxumes/redis"
	DefaultPostgresVolume = "$HOME/.gnyxumes/postgresql"
	DefaultRabbitMQVolume = "$HOME/.gnyxumes/rabbitmq"
)

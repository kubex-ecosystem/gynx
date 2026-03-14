package middlewares

import (
	"context"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	kbxMod "github.com/kubex-ecosystem/gnyx/internal/module/kbx"
	kbxGet "github.com/kubex-ecosystem/kbx/get"
	gl "github.com/kubex-ecosystem/logz"
)

var wasLoggedEntFeat bool

// ProductionConfig holds all production middleware configuration
type ProductionConfig struct {
	RateLimit struct {
		Enabled bool `yaml:"enabled"`
		Default struct {
			Capacity   int `yaml:"capacity"`    // requests per bucket
			RefillRate int `yaml:"refill_rate"` // tokens per second
		} `yaml:"default"`
		PerProvider map[string]struct {
			Capacity   int `yaml:"capacity"`
			RefillRate int `yaml:"refill_rate"`
		} `yaml:"per_provider"`
	} `yaml:"rate_limit"`

	CircuitBreaker struct {
		Enabled bool `yaml:"enabled"`
		Default struct {
			MaxFailures      int `yaml:"max_failures"`
			ResetTimeoutSec  int `yaml:"reset_timeout_sec"`
			SuccessThreshold int `yaml:"success_threshold"`
		} `yaml:"default"`
		PerProvider map[string]struct {
			MaxFailures      int `yaml:"max_failures"`
			ResetTimeoutSec  int `yaml:"reset_timeout_sec"`
			SuccessThreshold int `yaml:"success_threshold"`
		} `yaml:"per_provider"`
	} `yaml:"circuit_breaker"`

	HealthCheck struct {
		Enabled     bool `yaml:"enabled"`
		IntervalSec int  `yaml:"interval_sec"`
		TimeoutSec  int  `yaml:"timeout_sec"`
	} `yaml:"health_check"`

	Retry struct {
		Enabled     bool    `yaml:"enabled"`
		MaxRetries  int     `yaml:"max_retries"`
		BaseDelayMs int     `yaml:"base_delay_ms"`
		MaxDelayMs  int     `yaml:"max_delay_ms"`
		Multiplier  float64 `yaml:"multiplier"`
	} `yaml:"retry"`
}

// DefaultProductionConfig returns a sensible default configuration
func DefaultProductionConfig() ProductionConfig {
	config := ProductionConfig{}

	// Rate limiting defaults
	config.RateLimit.Enabled = true
	config.RateLimit.Default.Capacity = 100
	config.RateLimit.Default.RefillRate = 10

	// Circuit breaker defaults
	config.CircuitBreaker.Enabled = true
	config.CircuitBreaker.Default.MaxFailures = 5
	config.CircuitBreaker.Default.ResetTimeoutSec = 60
	config.CircuitBreaker.Default.SuccessThreshold = 3

	// Health check defaults
	config.HealthCheck.Enabled = true
	config.HealthCheck.IntervalSec = 30
	config.HealthCheck.TimeoutSec = 10

	// Retry defaults
	config.Retry.Enabled = true
	config.Retry.MaxRetries = 3
	config.Retry.BaseDelayMs = 100
	config.Retry.MaxDelayMs = 5000
	config.Retry.Multiplier = 2.0

	return config
}

// ProductionMiddleware wraps all production middleware functionality
type ProductionMiddleware struct {
	config         ProductionConfig
	rateLimiter    *RateLimiter
	circuitBreaker *CircuitBreakerManager
	healthMonitor  *HealthMonitor
	loggingLogz    *LoggingLogz
	retryConfig    RetryConfig
}

// NewProductionMiddleware creates a new production middleware manager
func NewProductionMiddleware(config ProductionConfig) *ProductionMiddleware {
	lgr := gl.GetLoggerZ("github.com/kubex-ecosystem/gnyx")

	pm := &ProductionMiddleware{
		config:      config,
		loggingLogz: NewLoggingLogz(lgr),
	}

	// Initialize rate limiter
	if config.RateLimit.Enabled {
		pm.rateLimiter = NewRateLimiter()
	}

	// Initialize circuit breaker
	if config.CircuitBreaker.Enabled {
		pm.circuitBreaker = NewCircuitBreakerManager()
	}

	// Initialize health monitor
	if config.HealthCheck.Enabled {
		interval := time.Duration(config.HealthCheck.IntervalSec) * time.Second
		pm.healthMonitor = NewHealthMonitor(interval)
	}

	// Setup retry config
	if config.Retry.Enabled {
		pm.retryConfig = RetryConfig{
			MaxRetries: config.Retry.MaxRetries,
			BaseDelay:  time.Duration(config.Retry.BaseDelayMs) * time.Millisecond,
			MaxDelay:   time.Duration(config.Retry.MaxDelayMs) * time.Millisecond,
			Multiplier: config.Retry.Multiplier,
		}
	}

	if !wasLoggedEntFeat {
		// gl.Println("[ProductionMiddleware] Initialized with enterprise features:")
		gl.Info("Starting server with Enterprise Features enabled:")
		if config.RateLimit.Enabled {
			gl.Infof("  Rate Limiting: %d capacity, %d/sec refill",
				config.RateLimit.Default.Capacity, config.RateLimit.Default.RefillRate)
		}
		if config.CircuitBreaker.Enabled {
			gl.Infof("  Circuit Breaker: %d max failures, %ds reset timeout",
				config.CircuitBreaker.Default.MaxFailures, config.CircuitBreaker.Default.ResetTimeoutSec)
		}
		if config.HealthCheck.Enabled {
			gl.Infof("  Health Checks: every %ds", config.HealthCheck.IntervalSec)
		}
		if config.Retry.Enabled {
			gl.Infof("  Retry Logic: %d max retries with exponential backoff", config.Retry.MaxRetries)
		}
		wasLoggedEntFeat = true
	}

	return pm
}

// RegisterProvider registers a provider with all middleware components
func (pm *ProductionMiddleware) RegisterProvider(provider string) {
	// Set up rate limiting
	if pm.rateLimiter != nil {
		capacity := pm.config.RateLimit.Default.Capacity
		refillRate := pm.config.RateLimit.Default.RefillRate

		// Check for provider-specific configuration
		if providerConfig, exists := pm.config.RateLimit.PerProvider[provider]; exists {
			capacity = providerConfig.Capacity
			refillRate = providerConfig.RefillRate
		}

		pm.rateLimiter.SetLimit(provider, capacity, refillRate)
	}

	// Set up circuit breaker
	if pm.circuitBreaker != nil {
		maxFailures := pm.config.CircuitBreaker.Default.MaxFailures
		resetTimeout := time.Duration(pm.config.CircuitBreaker.Default.ResetTimeoutSec) * time.Second
		successThreshold := pm.config.CircuitBreaker.Default.SuccessThreshold

		// Check for provider-specific configuration
		if providerConfig, exists := pm.config.CircuitBreaker.PerProvider[provider]; exists {
			maxFailures = providerConfig.MaxFailures
			resetTimeout = time.Duration(providerConfig.ResetTimeoutSec) * time.Second
			successThreshold = providerConfig.SuccessThreshold
		}

		pm.circuitBreaker.SetCircuitBreaker(provider, CircuitBreakerConfig{
			MaxFailures:      maxFailures,
			ResetTimeout:     resetTimeout,
			SuccessThreshold: successThreshold,
		})
	}

	// Register with health monitor
	if pm.healthMonitor != nil {
		pm.healthMonitor.RegisterProvider(provider)
	}
}

// WrapProvider wraps a provider call with all production middleware
func (pm *ProductionMiddleware) WrapProvider(provider string, operation func() error) error {
	startTime := time.Now()

	// 1. Check rate limit
	if pm.rateLimiter != nil {
		if !pm.rateLimiter.Allow(provider) {
			return gl.Errorf("rate limit exceeded for provider %s", provider)
		}
	}

	// 2. Check circuit breaker
	if pm.circuitBreaker != nil {
		if err := pm.circuitBreaker.Allow(provider); err != nil {
			return gl.Errorf("circuit breaker blocked request to %s: %v", provider, err)
		}
	}

	// 3. Execute with retry logic
	var err error
	if pm.config.Retry.Enabled {
		ctx := context.Background()
		err = RetryWithBackoff(ctx, pm.retryConfig, operation)
	} else {
		err = operation()
	}

	// 4. Record results
	responseTime := time.Since(startTime)
	success := err == nil

	// Record circuit breaker result
	if pm.circuitBreaker != nil {
		if success {
			pm.circuitBreaker.RecordSuccess(provider)
		} else {
			pm.circuitBreaker.RecordFailure(provider)
		}
	}

	// Record health check result
	if pm.healthMonitor != nil {
		errorMsg := ""
		if err != nil {
			errorMsg = err.Error()
		}
		pm.healthMonitor.RecordCheck(provider, success, responseTime, errorMsg)
	}

	return err
}

// GetStatus returns comprehensive status for all middleware components
func (pm *ProductionMiddleware) GetStatus() map[string]interface{} {
	status := make(map[string]interface{})

	// Rate limit status
	if pm.rateLimiter != nil {
		rateLimitStatus := make(map[string]interface{})
		// Note: You'd need to implement a way to get all provider names
		// For now, we'll just indicate that rate limiting is enabled
		rateLimitStatus["enabled"] = true
		status["rate_limit"] = rateLimitStatus
	}

	// Circuit breaker status
	if pm.circuitBreaker != nil {
		circuitBreakerStatus := make(map[string]interface{})
		circuitBreakerStatus["enabled"] = true
		status["circuit_breaker"] = circuitBreakerStatus
	}

	// Health check status
	if pm.healthMonitor != nil {
		healthStatus := pm.healthMonitor.GetAllHealth()
		status["health_checks"] = healthStatus
	}

	return status
}

// GetHealthMonitor returns the health monitor instance
func (pm *ProductionMiddleware) GetHealthMonitor() *HealthMonitor {
	return pm.healthMonitor
}

// Stop gracefully stops all middleware components
func (pm *ProductionMiddleware) Stop() {
	if pm.healthMonitor != nil {
		pm.healthMonitor.Stop()
	}
	gl.Log("info", "ProductionMiddleware Stopped all components")
}

func (pm *ProductionMiddleware) Logger(logger *gl.LoggerZ) gin.HandlerFunc {
	if pm.loggingLogz == nil {
		pm.loggingLogz = NewLoggingLogz(logger)
	}
	return pm.loggingLogz.Logger()
}

// ExecuteOperation executes the given operation with middleware checks
func (pm *ProductionMiddleware) ExecuteOperation(provider string, operation func() error) error {
	startTime := time.Now()

	// 1. Check rate limiter
	if pm.rateLimiter != nil {
		if !pm.rateLimiter.Allow(provider) {
			return gl.Errorf("rate limit exceeded for provider %s", provider)
		}
	}

	// 2. Check circuit breaker
	if pm.circuitBreaker != nil {
		if err := pm.circuitBreaker.Allow(provider); err != nil {
			return gl.Errorf("circuit breaker blocked request to %s: %v", provider, err)
		}
	}

	// 3. Execute with retry logic
	var err error
	if pm.config.Retry.Enabled {
		ctx := context.Background()
		err = RetryWithBackoff(ctx, pm.retryConfig, operation)
	} else {
		err = operation()
	}

	// 4. Record results
	responseTime := time.Since(startTime)
	success := err == nil

	// Record circuit breaker result
	if pm.circuitBreaker != nil {
		if success {
			pm.circuitBreaker.RecordSuccess(provider)
		} else {
			pm.circuitBreaker.RecordFailure(provider)
		}
	}

	// Record health check result
	if pm.healthMonitor != nil {
		errorMsg := ""
		if err != nil {
			errorMsg = err.Error()
		}
		pm.healthMonitor.RecordCheck(provider, success, responseTime, errorMsg)
	}

	return err
}

func (pm *ProductionMiddleware) HealthCheck() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Perform health check logic here
		c.JSON(http.StatusOK, gin.H{"status": "healthy"})
	}
}

// SecureServerInit sets up security-related server configurations
func (pm *ProductionMiddleware) SecureServerInit(r *gin.Engine, fullBindAddress string) error {
	trustedProxies, trustedProxiesErr := pm.GetTrustedProxies()
	if trustedProxiesErr != nil {
		return trustedProxiesErr
	}
	setTrustProxiesErr := r.SetTrustedProxies(trustedProxies)
	if setTrustProxiesErr != nil {
		return setTrustProxiesErr
	}

	r.Use(
		func(c *gin.Context) {
			// O método que checo os loopbacks e dou um "bypass" quando é produção...
			if !pm.ValidateExpectedHosts(fullBindAddress, c) {
				c.Abort()
				return
			} else {

				// Porque se for produção, ele cai aqui! kkkk E obriga a ter o header origin
				// Além de fazer toda a validação de headers que vem do front e configuração de cookies e cache
				c.Header("Access-Control-Allow-Origin", c.GetHeader("Origin"))
				c.Header("Access-Control-Allow-Credentials", "true")

				// Não se assuste com a lista de headers, é só pra garantir que tudo venha certinho e porque eu estava "explorando e aprendendo mais" sobre CORS
				c.Header("Access-Control-Allow-Headers", strings.Join([]string{
					"Accept",
					"Origin",
					"Accept-Language",
					"Accept-Encoding",
					"Authorization",
					"Referer",
					"Content-Type",
					"Content-Length",
					"Cache-Control",
					"Content-Security-Policy",
					"ETag",
					"Referrer-Policy",
					"Permissions-Policy",
					"Strict-Transport-Security",
					"Sec-Fetch-Dest",
					"Sec-Fetch-Mode",
					"Server",
					"Set-Cookie",
					"User-Agent",
					"X-CSRF-Token",
					"X-Requested-With",
					"X-External-API-Key",
					"X-Tenant-ID",
					"X-User-ID",
					"X-Request-ID",
					"X-Frame-Options",
					"X-XSS-Protection",
					"X-Content-Type-Options",
					"X-RateLimit-Limit",
					"X-RateLimit-Remaining",
					"X-RateLimit-Reset",
					"X-Circuit-Breaker-State",
					"X-Health-Status",
					"X-Retry-Attempt",
					"X-Retry-Delay",
					"X-Retry-Error",
					"X-Request-Duration",
					"X-Response-Time",
					"X-Server-Timestamp",
					"X-Request-ID",
					"X-Correlation-ID",
					"X-Trace-ID",
					"X-Span-ID",
				}, ", "))
				c.Header("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE, PATCH")

				// Handle OPTIONS preflight requests
				if c.Request.Method == "OPTIONS" {
					c.AbortWithStatus(http.StatusOK)
					return
				}

				// Set security headers
				c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains; preload")
				c.Header("Referrer-Policy", "strict-origin")
				c.Header("Permissions-Policy", "geolocation=(),midi=(),sync-xhr=(),microphone=(),camera=(),magnetometer=(),gyroscope=(),fullscreen=(self),payment=()")
				c.Header("Content-Security-Policy", "default-src 'self'; connect-src *; font-src *; script-src-elem * 'unsafe-inline'; img-src * data:; style-src * 'unsafe-inline';")

				c.Header("X-Frame-Options", "DENY")
				c.Header("X-XSS-Protection", "1; mode=block")
				c.Header("X-Content-Type-Options", "nosniff")

				c.Next()
			}
		},
	)

	return nil
}

func (pm *ProductionMiddleware) GetTrustedProxies() ([]string, error) {
	// trustedProxies := viper.GetStringSlice("trustedProxies")

	tpxStr := kbxGet.EnvOr("KUBEX_GNYX_TRUSTED_PROXIES", kbxGet.EnvOr("KUBEX_TRUSTED_PROXIES", kbxMod.DefaultGNyxLoopbackIP))
	trustedProxies := []string{}
	if len(tpxStr) > 0 {
		trustedProxies = strings.Split(tpxStr, ",")
	}

	if len(trustedProxies) == 0 {
		interfaces, err := net.Interfaces()
		if err != nil {
			return []string{}, err
		}

		for _, iface := range interfaces {
			if iface.Flags&net.FlagLoopback == 0 {
				addrs, addrsErr := iface.Addrs()
				if addrsErr != nil {
					return []string{}, gl.Errorf("error getting addresses for interface %s: %s", iface.Name, addrsErr)
					//continue // Ignora erro
				}

				for _, addr := range addrs {
					ipNet, ok := addr.(*net.IPNet)
					if ok {
						trustedProxies = append(trustedProxies, ipNet.IP.String())
					}
				}
			}
		}
	}

	gl.Noticef("Trusted Proxies: %v", trustedProxies)

	return trustedProxies, nil
}

func (pm *ProductionMiddleware) ValidateExpectedHosts(fullBindAddress string, c *gin.Context) bool {

	// Check if the environment is production
	envMode := kbxGet.EnvOr("BUILD_MODE", kbxGet.EnvOr("KUBEX_ENV", kbxGet.EnvOr("ENV", "development")))
	if envMode == "production" {
		return true
	}

	// Check if the host is the full bind address
	if c.Request.Host == fullBindAddress ||
		c.Request.URL.Host == fullBindAddress {
		return true
	}

	// Get the bind port
	_, bindPort, err := net.SplitHostPort(fullBindAddress)
	if err != nil {
		return false
	}

	// Create a list of trusted local addresses with the bind port
	trustedLocalList := []string{}
	for loopbackAddress := range strings.SplitSeq(kbxMod.DefaultGNyxLoopbackIP, ",") {
		trustedLocalList = append(trustedLocalList, loopbackAddress)
		trustedLocalList = append(trustedLocalList, loopbackAddress+":"+bindPort)
	}

	// Ensure localhost is in the trusted local list, with and without port
	trustedLocalList = append(trustedLocalList, "localhost")
	trustedLocalList = append(trustedLocalList, kbxGet.EnvOr("PUBLIC_DOMAIN", kbxGet.EnvOr("PUBLIC_URL", kbxGet.EnvOr("KUBEX_GNYX_PUBLIC_DOMAIN", "localhost"+bindPort))))
	trustedLocalList = append(trustedLocalList, "kubex.world")
	trustedLocalList = append(trustedLocalList, "lookatni.kubex.world")
	trustedLocalList = append(trustedLocalList, "docs.kubex.world")
	trustedLocalList = append(trustedLocalList, "gnyx.kubex.world")
	trustedLocalList = append(trustedLocalList, "api.kubex.world")
	trustedLocalList = append(trustedLocalList, "ws.kubex.world")
	trustedLocalList = append(trustedLocalList, "api.gnyx.kubex.world")

	// Check if the host is in the trusted local lists
	for _, trustedLocal := range trustedLocalList {
		if c.Request.Host == trustedLocal ||
			c.Request.URL.Host == trustedLocal {
			return true
		}
	}

	c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Unauthorized host: " + c.Request.Host})
	return false
}

// func max(a, b int) int {
// 	if a > b {
// 		return a
// 	}
// 	return b
// }

// func abs(a int) int {
// 	if a < 0 {
// 		return -a
// 	}
// 	return a
// }

// withCORS adds CORS headers to responses
func withCORS(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		allowLocalhost := origin != "" && (strings.HasPrefix(origin, "http://localhost") || strings.HasPrefix(origin, "http://127.0.0.1"))

		// Allow any localhost/127.0.0.1 origin in dev so the Ecosystem (5173/4173) can call the API with cookies/headers.
		if allowLocalhost {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Access-Control-Allow-Credentials", "true")
		} else {
			w.Header().Set("Access-Control-Allow-Origin", "*")
		}

		w.Header().Set("Access-Control-Allow-Headers", "content-type, authorization, x-external-api-key, x-tenant-id, x-user-id, accept, origin")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		if r.Method == http.MethodOptions {
			w.WriteHeader(204)
			return
		}
		h.ServeHTTP(w, r)
	})
}

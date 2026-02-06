package middleware

import (
	"sync"

	"github.com/gin-gonic/gin"

	"strings"

	gl "github.com/kubex-ecosystem/logz"
)

// healthCheckPaths are paths that should not be logged to reduce noise
var (
	swg              = sync.WaitGroup{}
	healthCheckPaths map[string]any
)

func init() {
	healthCheckPaths = make(map[string]any)
	healthCheckPaths = map[string]any{
		"prefix": []string{
			"/.well-known/acme-challenge/",
			"/assets/",
			"/static/",
		},
		"suffix": []string{
			"/status",
			"/ping",
			"/health",
			"/healthz",
			"/api/v1/health",
			"/favicon.ico",
			"/robots.txt",
			"/.well-known/ai-plugin.json",
			"/.well-known/openapi.yaml",
			"/.well-known/health-check.yaml",
			"/manifest.json",
		},
	}
}

type kbxContext struct {
	URI       string
	status    int
	clientIP  string
	userAgent string
	path      string
	method    string
	headers   string
}

type LoggingLogz struct {
	logger *gl.LoggerZ
}

func NewLoggingLogz(logger *gl.LoggerZ) *LoggingLogz {
	return &LoggingLogz{
		logger: logger,
	}
}

// shouldSkipLogging determines if a request path should skip logging
func shouldSkipLogging(path string) bool {
	// Check prefixes
	for whatCheck, patterns := range healthCheckPaths {
		for _, pattern := range patterns.([]string) {
			if whatCheck == "prefix" && strings.HasPrefix(path, pattern) {
				return true
			}
			if whatCheck == "suffix" && strings.HasSuffix(path, pattern) {
				return true
			}
		}
	}
	return false
}

func (l *LoggingLogz) Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		swg.Go(func() {
			func(c kbxContext) {
				// Skip logging for health check endpoints to reduce noise
				if !shouldSkipLogging(c.URI) {
					// Log only once with formatted message
					// Add detailed message to the log entry
					gl.Info(
						gl.
							NewLogzEntry(gl.Level("info")).
							WithMessage(
								gl.Sprintf("%s: %d - IP: %s, Path: %s",
									"HTTP Request",
									c.status,
									c.clientIP,
									c.path,
								),
							),
					)
				}
			}(kbxContext{
				URI:       c.Request.RequestURI,
				status:    c.Writer.Status(),
				clientIP:  c.ClientIP(),
				userAgent: c.Request.UserAgent(),
				path:      c.Request.URL.Path,
				method:    c.Request.Method,
				headers:   gl.Sprintf("%v", c.Request.Header),
			})
		})
		c.Next()
	}
}

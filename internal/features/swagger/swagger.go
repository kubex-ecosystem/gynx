package swagger

import (
	"net/http"
	"net/url"
	"strings"

	"github.com/gin-gonic/gin"
	docs "github.com/kubex-ecosystem/gnyx/docs"
	"github.com/kubex-ecosystem/gnyx/internal/config"
	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// Configure updates the generated Swagger metadata to reflect the active runtime.
func Configure(cfg *config.Config) {
	docs.SwaggerInfo.Title = "GNyx API"
	docs.SwaggerInfo.Version = "1.0.0"
	docs.SwaggerInfo.BasePath = "/"

	baseURL := ""
	if cfg != nil && cfg.Invite != nil {
		baseURL = strings.TrimSpace(cfg.Invite.BaseURL)
	}
	if baseURL == "" && cfg != nil && cfg.ServerConfig != nil {
		host := strings.TrimSpace(cfg.ServerConfig.Runtime.Host)
		if host != "" {
			baseURL = host
			if !strings.Contains(baseURL, "://") {
				baseURL = "http://" + baseURL
			}
		}
	}
	if baseURL == "" {
		docs.SwaggerInfo.Host = ""
		return
	}

	parsed, err := url.Parse(baseURL)
	if err != nil {
		docs.SwaggerInfo.Host = strings.TrimPrefix(strings.TrimPrefix(baseURL, "https://"), "http://")
		return
	}
	if parsed.Host != "" {
		docs.SwaggerInfo.Host = parsed.Host
	}
}

// Register mounts the Swagger UI in the active Gin runtime.
func Register(r gin.IRoutes, cfg *config.Config) {
	Configure(cfg)
	url := ginSwagger.URL("/swagger/doc.json")
	r.GET("/swagger", func(c *gin.Context) {
		c.Redirect(http.StatusTemporaryRedirect, "/swagger/index.html")
	})
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerfiles.Handler, url, ginSwagger.DefaultModelsExpandDepth(-1)))
}

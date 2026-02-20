// Package uiroutes defines the routes for the UI service.
package uiroutes

import (
	"io/fs"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/kubex-ecosystem/gnyx/internal/features/ui"
	"github.com/kubex-ecosystem/gnyx/internal/types"
	"github.com/kubex-ecosystem/logz"

	kbxGet "github.com/kubex-ecosystem/kbx/get"
	kbxIs "github.com/kubex-ecosystem/kbx/is"
)

// RegisterRoutes acopla a engine do Gin à feature de UI embarcada.
// Ele configura o servidor de arquivos estáticos e o Fallback para o SPA.
func RegisterRoutes(r *gin.RouterGroup, container types.IContainer) gin.IRoutes {
	uiSvc, ok := container.UIService().(*ui.UIService)
	if !ok || kbxIs.NilPtr(uiSvc) {
		logz.Warn("UIService is not of the expected type (*ui.UIService)")
		return nil
	}

	webFS := uiSvc.GetWebFS()
	if webFS == nil {
		logz.Warn("UIService returned a nil filesystem. Skipping UI route registration.")
		return nil
	}

	// Extrai a raiz da subpasta "web" (onde fica o index.html e /assets)
	strippedFS, err := fs.Sub(webFS, ".")
	if err != nil {
		logz.Errorf("Failed to load embedded web filesystem: %v", err)
		return nil
	}

	// Configura o servidor de arquivos estáticos para servir os arquivos do UI
	httpFS := http.FS(strippedFS)
	fileServer := http.FileServer(httpFS)

	// Wrapper para o NoRoute, para servir arquivos estáticos e fallback para SPA no group /app
	r.Any("/*path", func(c *gin.Context) {
		path := strings.TrimSpace(c.Request.URL.Path)

		if after, ok := strings.CutPrefix(path, "/app/"); ok {
			path = kbxGet.ValOrType(after, "/")
		}
		if strings.HasSuffix(path, "/") && len(path) > 1 {
			path += "index.html"
		} else if path == "/" {
			path = "index.html"
		}

		// Tenta servir o arquivo solicitado
		f, err := strippedFS.Open(path)
		if err == nil {
			stat, err := f.Stat()
			f.Close()
			if err == nil && !stat.IsDir() {
				fileServer.ServeHTTP(c.Writer, c.Request)
				return
			} else if err != nil {
				logz.Warnf("Error stating file %s: %v", path, err)
			}
		}

		// Se o arquivo não for encontrado, serve o index.html do SPA
		c.Request.URL.Path = "/"
		fileServer.ServeHTTP(c.Writer, c.Request)
	})

	logz.Info("UI Feature embedded routes registered successfully")
	return r
}

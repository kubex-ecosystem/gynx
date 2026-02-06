// Package web provides web interface for the kubexbe.
package web

import (
	gl "github.com/kubex-ecosystem/logz"
	"io"
	"io/fs"
	"net/http"
	"os"
	"path"
	"strings"
)

// Handler provides HTTP handlers for the web interface
type Handler struct {
	fsys fs.FS
}

// NewHandler creates a new web interface handler
func NewHandler() (*Handler, error) {
	// Strip the "embedded/guiweb" prefix from the embedded filesystem
	return &Handler{
		fsys: nil,
	}, nil
}

// ServeHTTP handles web interface requests
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Clean the path and remove leading slash
	cleanPath := path.Clean(r.URL.Path)
	if cleanPath == "/" {
		cleanPath = "/index.html"
	}
	cleanPath = strings.TrimPrefix(cleanPath, "/")

	if _, err := os.Stat(cleanPath); os.IsNotExist(err) {
		return
	}

	// Open the file from embedded filesystem
	file, err := h.fsys.Open(cleanPath)
	if err != nil {
		// If file not found, serve index.html for SPA routing
		file, err = h.fsys.Open("index.html")
		if err != nil {
			http.Error(w, "Not found", http.StatusNotFound)
			return
		}
	}
	defer file.Close()

	// Get file info for content type detection
	stat, err := file.Stat()
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Set content type based on file extension
	ext := strings.ToLower(path.Ext(cleanPath))
	switch ext {
	case ".html":
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
	case ".css":
		w.Header().Set("Content-Type", "text/css; charset=utf-8")
	case ".js":
		w.Header().Set("Content-Type", "application/javascript; charset=utf-8")
	case ".json":
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
	case ".png":
		w.Header().Set("Content-Type", "image/png")
	case ".ico":
		w.Header().Set("Content-Type", "image/x-icon")
	case ".svg":
		w.Header().Set("Content-Type", "image/svg+xml")
	default:
		// Let Go detect the content type
	}

	// Cache static assets for 1 hour
	if ext != ".html" {
		w.Header().Set("Cache-Control", "public, max-age=3600")
	}

	gl.Infof("Serving web file: %s", cleanPath)

	// Serve the file
	http.ServeContent(w, r, stat.Name(), stat.ModTime(), file.(io.ReadSeeker))
}

// RegisterRoutes registers web interface routes
func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	// Serve web interface on root path
	mux.Handle("/", h)

	// Also serve on /app/ for explicit access
	mux.Handle("/app/", http.StripPrefix("/app", h))
}

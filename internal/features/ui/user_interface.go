// Package ui encapsula o Frontend embarcado (SPA) como uma feature do GNyx.
package ui

import (
	"embed"
	"io/fs"

	"github.com/kubex-ecosystem/logz"
)

// O prefixo "all:" garante que TUDO dentro da pasta web
// (incluindo subpastas de assets e ícones) seja embutido de forma explícita e segura.
//
//go:embed all:web
var webFS embed.FS

// UIService representa o módulo de interface gráfica do sistema
type UIService struct {
	// Aqui podem entrar configurações de temas default, flag de enable/disable UI, etc.
}

func NewUIService() *UIService {
	return &UIService{}
}

func (s *UIService) GetWebFS() fs.FS {
	strippedFS, err := fs.Sub(webFS, "web")
	if err != nil {
		logz.Errorf("Failed to load embedded web filesystem: %v", err)
		return nil
	}
	return strippedFS
}

// Package templates fornece funcionalidades relacionadas a templates.
package templates

import (
	"embed"
	"io/fs"
	"path/filepath"
	"strings"

	kbxMod "github.com/kubex-ecosystem/gnyx/internal/module/kbx"
	kbxGet "github.com/kubex-ecosystem/kbx/get"
	gl "github.com/kubex-ecosystem/logz"
)

//go:embed all:email/*[.html,.json]
var TemplatesVarFS embed.FS

var (
	emailTemplateFS *EmailTemplateFSImpl
)

type EmailTemplateFSImpl struct {
	Path     string
	FS       fs.FS
	DEntries []fs.DirEntry
}

var EmailTemplateNames = []string{"deal_won", "lead_assigned", "user_invited"}

func GetEmailTemplateFS(path string) *EmailTemplateFSImpl {
	if path == "" {
		path = kbxGet.EnvOr("INVITE_TEMPLATES_DIR", kbxMod.DefaultTemplatesDir)
	}
	drs, err := TemplatesVarFS.ReadDir("email")
	if err != nil {
		gl.Error("Failed to read email templates directory: %v", err)
		return nil
	}
	return &EmailTemplateFSImpl{
		Path:     path,
		FS:       TemplatesVarFS,
		DEntries: drs,
	}
}

func (e *EmailTemplateFSImpl) ReadFile(name string) ([]byte, error) {
	if name == "" {
		return nil, fs.ErrNotExist
	}

	candidates := []string{name}
	if !strings.HasPrefix(name, "email/") {
		candidates = append(candidates, filepath.Join("email", name))
	}
	if filepath.Ext(name) == "" {
		candidates = append(candidates, filepath.Join("email", name+".html"))
	}

	for _, candidate := range candidates {
		if data, err := TemplatesVarFS.ReadFile(candidate); err == nil {
			return data, nil
		}
	}
	return nil, fs.ErrNotExist
}

// ListTemplates lista os nomes dos templates de email disponíveis.
func (e *EmailTemplateFSImpl) ListTemplates() []string {
	return EmailTemplateNames
}

// GetEmailTemplate retorna o conteúdo do template de email pelo nome.
func GetEmailTemplate(name string) ([]byte, error) {
	templateFS := GetEmailTemplateFS(kbxGet.EnvOr("INVITE_TEMPLATES_DIR", "templates"))
	return templateFS.ReadFile(name)
}

// ListEmailTemplates lista os nomes dos templates de email disponíveis.
func ListEmailTemplates() []string {
	return EmailTemplateNames
}

// func init() {
// 	if emailTemplateFS == nil {
// 		gl.Debug("Email templates loaded")
// 		emailTemplateFS = GetEmailTemplateFS(".")
// 		if emailTemplateFS == nil {
// 			gl.Error("Failed to initialize email template filesystem")
// 		}
// 		return
// 	}
// 	gl.Warnf("Email templates already loaded")
// }

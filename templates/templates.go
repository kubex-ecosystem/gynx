// Package templates fornece funcionalidades relacionadas a templates.
package templates

import (
	"embed"
	"io/fs"
	"path/filepath"

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
	for _, entry := range e.DEntries {
		if entry.Name() == name {
			return TemplatesVarFS.ReadFile(filepath.Join(e.Path, "email", name, "content.html"))
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
	return TemplatesVarFS.ReadFile(
		filepath.Join(
			kbxGet.EnvOr("INVITE_TEMPLATES_DIR", "templates"),
			"email",
			name,
			"content.html",
		),
	)
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

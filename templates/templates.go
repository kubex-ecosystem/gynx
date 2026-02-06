// Package templates fornece funcionalidades relacionadas a templates.
package templates

import (
	"embed"

	gl "github.com/kubex-ecosystem/logz"
)

//go:embed all:email/**/*
var EmailTemplates embed.FS

var (
	emailTemplateFS *EmailTemplateFS
)

type EmailTemplateFS struct {
	FS embed.FS
}

var EmailTemplateNames = []string{"deal_won", "lead_assigned", "user_invited"}

func GetEmailTemplateFS() *EmailTemplateFS {
	return &EmailTemplateFS{
		FS: EmailTemplates,
	}
}

func (e *EmailTemplateFS) ReadFile(name string) ([]byte, error) {
	return e.FS.ReadFile("email/" + name + "/content.html")
}

// ListTemplates lista os nomes dos templates de email disponíveis.
func (e *EmailTemplateFS) ListTemplates() []string {
	return EmailTemplateNames
}

// GetEmailTemplate retorna o conteúdo do template de email pelo nome.
func GetEmailTemplate(name string) ([]byte, error) {
	return EmailTemplates.ReadFile("email/" + name + ".html")
}

// ListEmailTemplates lista os nomes dos templates de email disponíveis.
func ListEmailTemplates() []string {
	return EmailTemplateNames
}

func init() {
	if emailTemplateFS == nil {
		gl.Debug("Email templates loaded")
		emailTemplateFS = GetEmailTemplateFS()
	}
}

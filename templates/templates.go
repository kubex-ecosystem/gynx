// Package templates fornece funcionalidades relacionadas a templates.
package templates

import (
	"embed"
	"path/filepath"

	"github.com/kubex-ecosystem/gnyx/templates/email"
	kbxGet "github.com/kubex-ecosystem/kbx/get"
	gl "github.com/kubex-ecosystem/logz"
)

//go:embed all:email/*[.html,.json]
var TemplatesVarFS embed.FS

var (
	emailTemplateFS *EmailTemplateFSImpl
)

type EmailTemplateFSImpl struct {
	FS embed.FS
}

var EmailTemplateNames = []string{"deal_won", "lead_assigned", "user_invited"}

func GetEmailTemplateFS() *EmailTemplateFSImpl {
	m := email.NewEmailHTMLRenderer()
	m.EmailTemplateFSImpl()

	return &EmailTemplateFSImpl{
		FS: TemplatesVarFS,
	}
}

func (e *EmailTemplateFSImpl) ReadFile(name string) ([]byte, error) {
	return e.FS.ReadFile(
		filepath.Join(
			kbxGet.EnvOr("INVITE_TEMPLATES_DIR", "templates"),
			"email",
			name,
			"content.html",
		),
	)
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

func init() {
	if emailTemplateFS == nil {
		gl.Debug("Email templates loaded")
		emailTemplateFS = GetEmailTemplateFS()
	}
}

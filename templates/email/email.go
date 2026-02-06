// Package email fornece tipos e funções relacionados ao envio de emails.
package email

import (
	"html"
	"strings"
	"text/template"

	"github.com/kubex-ecosystem/gnyx/internal/dsclient"
	"github.com/kubex-ecosystem/gnyx/templates"
)

// Type aliases para entidades e stores do DSClient
type (
	DSClient              = dsclient.DSClient
	User                  = dsclient.User
	CreateUserInput       = dsclient.CreateUserInput
	UpdateUserInput       = dsclient.UpdateUserInput
	UserFilters           = dsclient.UserFilters
	Invitation            = dsclient.Invitation
	InvitationType        = dsclient.InvitationType
	InvitationStatus      = dsclient.InvitationStatus
	CreateInvitationInput = dsclient.CreateInvitationInput
	UpdateInvitationInput = dsclient.UpdateInvitationInput
	InvitationFilters     = dsclient.InvitationFilters
	Company               = dsclient.Company
	CreateCompanyInput    = dsclient.CreateCompanyInput
	UpdateCompanyInput    = dsclient.UpdateCompanyInput
	CompanyFilters        = dsclient.CompanyFilters
)

// Invitation type constants
const (
	TypePartner  = dsclient.TypePartner
	TypeInternal = dsclient.TypeInternal
)

func init() {
	// Garantir que os pacotes sejam importados
	_ = templates.EmailTemplateFS{}
}

type EmailHTMLRenderer struct {
	*templates.EmailTemplateFS
}

func NewEmailHTMLRenderer() *EmailHTMLRenderer {
	return &EmailHTMLRenderer{
		EmailTemplateFS: templates.GetEmailTemplateFS(),
	}
}

func (r *EmailHTMLRenderer) RenderTemplate(templateType string, data any) (string, error) {
	tplFileContent, err := r.EmailTemplateFS.ReadFile(templateType)
	if err != nil {
		return "", err
	}
	tplFileContentStr := html.UnescapeString(string(tplFileContent))
	tpl, err := template.New(templateType).Parse(tplFileContentStr)
	if err != nil {
		return "", err
	}
	var renderedContent string
	buf := &strings.Builder{}
	if err := tpl.Execute(buf, data); err != nil {
		return "", err
	}
	renderedContent = buf.String()
	return renderedContent, nil
}

func (r *EmailHTMLRenderer) ListTemplates() []string {
	return r.EmailTemplateFS.ListTemplates()
}

func GetEmailHTMLRenderer() *EmailHTMLRenderer {
	return NewEmailHTMLRenderer()
}

// GetEmailTemplate retorna o conteúdo do template de email pelo nome.
func GetEmailTemplate(name string) ([]byte, error) {
	templateFS := GetEmailHTMLRenderer()
	return templateFS.EmailTemplateFS.ReadFile(name)
}

// ListEmailTemplates lista os nomes dos templates de email disponíveis.
func ListEmailTemplates() []string {
	templateFS := GetEmailHTMLRenderer()
	return templateFS.EmailTemplateFS.ListTemplates()
}

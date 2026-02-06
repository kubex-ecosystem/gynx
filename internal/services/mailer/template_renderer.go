package mailer

import (
	"bytes"
	"fmt"
	"html/template"
	"strings"

	gl "github.com/kubex-ecosystem/logz"
)

// TemplateData representa os dados para renderização de templates
type TemplateData struct {
	Type            string                 `json:"type" yaml:"type" xml:"type" toml:"type" mapstructure:"type"`                                                             // Tipo do email (ex: "Confirmação de Conta")
	Email           string                 `json:"email" yaml:"email" xml:"email" toml:"email" mapstructure:"email"`                                                        // Email do destinatário
	RecipientName   string                 `json:"recipient_name" yaml:"recipient_name" xml:"recipient_name" toml:"recipient_name" mapstructure:"recipient_name"`           // Nome do destinatário
	CompanyName     string                 `json:"company_name" yaml:"company_name" xml:"company_name" toml:"company_name" mapstructure:"company_name"`                     // Nome da empresa
	SiteURL         string                 `json:"site_url" yaml:"site_url" xml:"site_url" toml:"site_url" mapstructure:"site_url"`                                         // URL base do site
	ConfirmationURL string                 `json:"confirmation_url" yaml:"confirmation_url" xml:"confirmation_url" toml:"confirmation_url" mapstructure:"confirmation_url"` // URL de confirmação (se aplicável)
	RedirectTo      string                 `json:"redirect_to" yaml:"redirect_to" xml:"redirect_to" toml:"redirect_to" mapstructure:"redirect_to"`                          // URL de redirecionamento alternativo
	ActionURL       string                 `json:"action_url" yaml:"action_url" xml:"action_url" toml:"action_url" mapstructure:"action_url"`                               // URL de ação principal
	ActionLabel     string                 `json:"action_label" yaml:"action_label" xml:"action_label" toml:"action_label" mapstructure:"action_label"`                     // Label do botão de ação
	Data            map[string]interface{} `json:"data" yaml:"data" xml:"data" toml:"data" mapstructure:"data"`                                                             // Dados dinâmicos adicionais
}

// BaseEmailTemplate é o template base compatível com a maioria dos providers
const BaseEmailTemplate = `
<!DOCTYPE html>
<html xmlns="http://www.w3.org/1999/xhtml">
<head>
<meta name="viewport" content="width=device-width" />
<meta http-equiv="Content-Type" content="text/html; charset=UTF-8" />
<title>Kubex PRM - {{ .Type }}</title>
<style>
body {
  font-family: "Roboto", "Helvetica Neue", Helvetica, Arial, sans-serif;
  background: linear-gradient(to bottom, rgba(23,118,209,0.85), rgba(22,22,22,0.7)) no-repeat;
  -webkit-font-smoothing: antialiased !important;
  width: 100%!important;
  height: 100%!important;
  margin: 0;
  padding: 0;
}
.container {
  max-width: 600px;
  margin: 30px auto;
  background: #fff;
  border-radius: 12px;
  border: 1px solid #eee;
  padding: 20px;
}
h2 {
  text-align: center;
  color: #111;
  font-weight: 300;
  margin: 20px 0;
}
p {
  color: #333;
  line-height: 1.6;
  margin: 15px 0;
}
.btn {
  text-decoration: none;
  color: #FFF !important;
  padding: 12px 24px;
  border-radius: 25px;
  font-weight: bold;
  display: inline-block;
  margin: 10px;
}
.btn-primary {
  background-color: #157efb;
  border: 2px solid #157efb;
}
.btn-secondary {
  background-color: #999;
  border: 2px solid #999;
}
.btn:hover {
  opacity: 0.9;
}
.info-table {
  width: 100%;
  margin: 25px 0;
  border-collapse: collapse;
}
.info-table td {
  padding: 10px 5px;
  border-bottom: 1px solid #eee;
}
.info-table td:first-child {
  width: 30%;
  color: #555;
  font-weight: 500;
}
.footer {
  font-size: 12px;
  color: #666;
  text-align: center;
  margin-top: 30px;
  padding-top: 20px;
  border-top: 1px solid #eee;
}
.footer a {
  color: #157efb;
  text-decoration: none;
}
.logo {
  width: 120px;
  display: block;
  margin: 0 auto 20px;
}
</style>
</head>
<body>
  <div class="container">
    {{- if .Data.logo_url }}
    <img src="{{ .Data.logo_url }}" alt="{{ .CompanyName }}" class="logo" />
    {{- else }}
    <img src="{{ .SiteURL }}/static/logo.png" alt="GNyxM" class="logo" />
    {{- end }}

    <h2>{{ .Type }}</h2>

    {{- if .RecipientName }}
    <p>Olá <strong>{{ .RecipientName }}</strong>,</p>
    {{- else }}
    <p>Olá {{ .Email }},</p>
    {{- end }}

    {{ template "content" . }}

    {{- if .ActionURL }}
    <p style="text-align:center;margin-top:30px;">
      <a href="{{ .ActionURL }}" class="btn btn-primary">{{ .ActionLabel }}</a>
      {{- if .RedirectTo }}
        <a href="{{ .RedirectTo }}" class="btn btn-secondary">Cancelar</a>
      {{- end }}
    </p>
    {{- end }}

    {{- if .ConfirmationURL }}
    <p style="text-align:center;margin-top:30px;">
      <a href="{{ .ConfirmationURL }}" class="btn btn-primary">Confirmar</a>
      {{- if .RedirectTo }}
        <a href="{{ .RedirectTo }}" class="btn btn-secondary">Cancelar</a>
      {{- end }}
    </p>
    {{- end }}

    <div class="footer">
      Atenciosamente,<br>
      <strong>{{ if .CompanyName }}{{ .CompanyName }}{{ else }}Equipe GNyxM{{ end }}</strong><br>
      <a href="{{ .SiteURL }}">{{ .SiteURL }}</a>
      {{- if .Data.unsubscribe_url }}
       | <a href="{{ .Data.unsubscribe_url }}">Cancelar notificações</a>
      {{- end }}
    </div>
  </div>
</body>
</html>
`

// TemplateRenderer gerencia a renderização de templates de e-mail
type TemplateRenderer struct {
	templates map[string]*template.Template
}

// NewTemplateRenderer cria um novo renderizador de templates
func NewTemplateRenderer() *TemplateRenderer {
	return &TemplateRenderer{
		templates: make(map[string]*template.Template),
	}
}

// RegisterTemplate registra um novo template
func (tr *TemplateRenderer) RegisterTemplate(name string, contentTemplate string) error {
	// Combina o template base com o conteúdo específico
	fullTemplate := BaseEmailTemplate

	// Parse o template base
	tmpl, err := template.New("base").Parse(fullTemplate)
	if err != nil {
		return gl.Errorf("erro ao parsear template base: %v", err)
	}

	// Parse o conteúdo específico como sub-template "content"
	_, err = tmpl.New("content").Parse(contentTemplate)
	if err != nil {
		return gl.Errorf("erro ao parsear template de conteúdo: %v", err)
	}

	tr.templates[name] = tmpl
	return nil
}

// Render renderiza um template com os dados fornecidos
func (tr *TemplateRenderer) Render(templateName string, data TemplateData) (string, error) {
	tmpl, exists := tr.templates[templateName]
	if !exists {
		return "", gl.Errorf("template '%s' não encontrado", templateName)
	}

	// Garante que Data não seja nil
	if data.Data == nil {
		data.Data = make(map[string]interface{})
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", gl.Errorf("erro ao renderizar template: %v", err)
	}

	return buf.String(), nil
}

// RenderSimple renderiza um template simples (backward compatibility)
func RenderSimple(htmlTemplate string, data TemplateData) (string, error) {
	tmpl, err := template.New("email").Parse(htmlTemplate)
	if err != nil {
		return "", gl.Errorf("erro ao parsear template: %v", err)
	}

	if data.Data == nil {
		data.Data = make(map[string]interface{})
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", gl.Errorf("erro ao renderizar: %v", err)
	}

	return buf.String(), nil
}

// ReplaceVariables substitui variáveis simples no formato {{var}} (fallback simples)
func ReplaceVariables(template string, vars map[string]string) string {
	result := template
	for key, value := range vars {
		placeholder := fmt.Sprintf("{{%s}}", key)
		result = strings.ReplaceAll(result, placeholder, value)
	}
	return result
}

// ExtractVariables extrai todas as variáveis de um template (regex simples)
func ExtractVariables(template string) []string {
	var variables []string
	parts := strings.Split(template, "{{")

	for _, part := range parts {
		if endIdx := strings.Index(part, "}}"); endIdx > 0 {
			varName := strings.TrimSpace(part[:endIdx])
			// Remove qualquer ponto ou função (ex: .Data.field -> field)
			if dotIdx := strings.LastIndex(varName, "."); dotIdx >= 0 {
				varName = varName[dotIdx+1:]
			}
			if varName != "" && !contains(variables, varName) {
				variables = append(variables, varName)
			}
		}
	}

	return variables
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

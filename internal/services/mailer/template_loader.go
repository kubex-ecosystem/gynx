// Package mailer implementa o carregamento e renderização de templates de email.
package mailer

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"path/filepath"
	"slices"
	"strings"

	kbxMod "github.com/kubex-ecosystem/gnyx/internal/module/kbx"
	"github.com/kubex-ecosystem/gnyx/templates"
	kbxGet "github.com/kubex-ecosystem/kbx/get"
	gl "github.com/kubex-ecosystem/logz"
)

// TemplateMetadata representa os metadados de um template
type TemplateMetadata struct {
	TemplateKey string            `json:"template_key"`
	Subject     string            `json:"subject"`
	Description string            `json:"description"`
	Variables   map[string]string `json:"variables"`
}

// TemplateLoader carrega templates do filesystem
type TemplateLoader struct {
	embeddedFS   fs.FS
	templatesFS  *templates.EmailTemplateFSImpl `json:"-" yaml:"-" xml:"-" toml:"-" mapstructure:"-"`
	templatesDir string                         `json:"-" yaml:"-" xml:"-" toml:"-" mapstructure:"-"`
	renderer     *TemplateRenderer              `json:"-" yaml:"-" xml:"-" toml:"-" mapstructure:"-"`
	metadata     map[string]*TemplateMetadata   `json:"-" yaml:"-" xml:"-" toml:"-" mapstructure:"-"`
}

// NewTemplateLoader cria um novo loader de templates
func NewTemplateLoader(templatesDir string) *TemplateLoader {
	return &TemplateLoader{
		embeddedFS:   templates.TemplatesVarFS,
		templatesFS:  templates.GetEmailTemplateFS(templatesDir),
		templatesDir: templatesDir,
		renderer:     NewTemplateRenderer(),
		metadata:     make(map[string]*TemplateMetadata),
	}
}

// LoadAll carrega todos os templates do diretório
func (tl *TemplateLoader) LoadAll() error {
	// Lista todos os templates disponíveis
	templates := tl.templatesFS.ListTemplates()
	for _, template := range templates {
		// Carrega cada template
		if err := tl.loadTemplate(template, ""); err != nil {
			return gl.Errorf("erro ao carregar template '%s': %v", template, err)
		}
	}
	return nil
}

// loadTemplate carrega um template específico
func (tl *TemplateLoader) loadTemplate(name string, path string) error {
	if name == "" {
		return gl.Errorf("template name is empty")
	}
	templateName := name
	htmlPath := filepath.Join("email", templateName+".html")
	if path != "" {
		htmlPath = filepath.Join(path, templateName+".html")
	}

	file, err := tl.embeddedFS.Open(htmlPath)
	if err != nil {
		return gl.Errorf("erro ao ler conteúdo do template '%s': %v", templateName, err)
	}
	defer file.Close()
	contentBytes := make([]byte, 0)
	buffer := make([]byte, 1024)
	for {
		n, err := file.Read(buffer)
		if err != nil {
			break
		}
		contentBytes = append(contentBytes, buffer[:n]...)
	}

	// Registra o template no renderer
	gl.Debugf("Registrando template: %s", templateName)
	if err := tl.renderer.RegisterTemplate(templateName, string(contentBytes)); err != nil {
		return gl.Errorf("erro ao registrar template: %v", err)
	}

	// Lê os metadados do FS embed, se existir
	metaPath := filepath.Join("email", templateName+".json")
	if path != "" {
		metaPath = filepath.Join(path, templateName+".json")
	}
	metaBytes, err := fs.ReadFile(tl.embeddedFS, metaPath)
	if err == nil {
		var metadata TemplateMetadata
		if err := json.Unmarshal(metaBytes, &metadata); err != nil {
			return gl.Errorf("erro ao parsear metadata do template '%s': %v", templateName, err)
		}
		if strings.TrimSpace(metadata.TemplateKey) == "" {
			metadata.TemplateKey = templateName
		}
		tl.metadata[templateName] = &metadata
	} else {
		// Garante a listagem mesmo sem metadata explícita
		if _, ok := tl.metadata[templateName]; !ok {
			tl.metadata[templateName] = &TemplateMetadata{TemplateKey: templateName}
		}
	}

	return nil
}

// Render renderiza um template pelo nome
func (tl *TemplateLoader) Render(templateName string, data TemplateData) (string, string, error) {
	// Obtém o subject do metadata
	subject := "Notificação - Kubex Ecosystem"
	if meta, exists := tl.metadata[templateName]; exists {
		subject = meta.Subject
		// Substitui variáveis no subject
		subject = replaceSubjectVars(subject, data)
	}

	// Renderiza o HTML
	html, err := tl.renderer.Render(templateName, data)
	if err != nil {
		return "", "", err
	}

	return subject, html, nil
}

// GetMetadata retorna os metadados de um template
func (tl *TemplateLoader) GetMetadata(templateName string) (*TemplateMetadata, error) {
	meta, exists := tl.metadata[templateName]
	if !exists {
		return nil, gl.Errorf("metadados do template '%s' não encontrados", templateName)
	}
	return meta, nil
}

// ListTemplates retorna a lista de templates disponíveis
func (tl *TemplateLoader) ListTemplates() []string {
	var templates []string
	for name := range tl.metadata {
		templates = append(templates, name)
	}
	slices.Sort(templates)
	return templates
}

// replaceSubjectVars substitui variáveis no subject
func replaceSubjectVars(subject string, data TemplateData) string {
	result := subject

	// Substitui variáveis de nível superior
	result = strings.ReplaceAll(result, "{{Email}}", data.Email)
	result = strings.ReplaceAll(result, "{{RecipientName}}", data.RecipientName)
	result = strings.ReplaceAll(result, "{{CompanyName}}", data.CompanyName)

	// Substitui variáveis do mapa Data
	if data.Data != nil {
		for key, value := range data.Data {
			placeholder := fmt.Sprintf("{{%s}}", key)
			result = strings.ReplaceAll(result, placeholder, fmt.Sprintf("%v", value))
		}
	}

	return result
}

// GetDefaultTemplateLoader retorna um loader com templates do diretório padrão
func GetDefaultTemplateLoader() (*TemplateLoader, error) {
	loader := NewTemplateLoader(kbxGet.EnvOr("INVITE_TEMPLATES_DIR", kbxMod.DefaultTemplatesDir))
	if err := loader.LoadAll(); err != nil {
		return nil, err
	}
	return loader, nil
}

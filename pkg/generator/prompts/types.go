package prompts

import (
	"bytes"
	"fmt"
	"text/template"
)

// Template represents a prompt template with variables
type Template struct {
	name     string
	template *template.Template
}

// Execute executes the template with the given data
func (t *Template) Execute(data interface{}) (string, error) {
	var buf bytes.Buffer
	if err := t.template.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("failed to execute template %s: %w", t.name, err)
	}
	return buf.String(), nil
}

// Name returns the template name
func (t *Template) Name() string {
	return t.name
}

// TemplateManager handles all prompt templates
type TemplateManager struct {
	templates map[string]*Template
}

// NewTemplateManager creates a new template manager
func NewTemplateManager() *TemplateManager {
	return &TemplateManager{
		templates: make(map[string]*Template),
	}
}

// RegisterTemplate registers a new template
func (tm *TemplateManager) RegisterTemplate(name, content string) error {
	tmpl, err := template.New(name).Parse(content)
	if err != nil {
		return fmt.Errorf("failed to parse template %s: %w", name, err)
	}

	tm.templates[name] = &Template{
		name:     name,
		template: tmpl,
	}

	return nil
}

// Execute executes a template with the given data
func (tm *TemplateManager) Execute(name string, data interface{}) (string, error) {

	tmpl, exists := tm.templates[name]
	if !exists {
		return "", fmt.Errorf("template %s not found", name)
	}

	var buf bytes.Buffer
	if err := tmpl.template.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("failed to execute template %s: %w", name, err)
	}

	return buf.String(), nil
}

// GetTemplate returns a template by name
func (tm *TemplateManager) GetTemplate(name string) (*Template, bool) {
	tmpl, exists := tm.templates[name]
	return tmpl, exists
}

// ListTemplates returns all registered template names
func (tm *TemplateManager) ListTemplates() []string {
	names := make([]string, 0, len(tm.templates))
	for name := range tm.templates {
		names = append(names, name)
	}
	return names
}

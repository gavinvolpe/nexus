package impl

import (
	"bytes"
	"fmt"
	"text/template"

	"github.com/gavinvolpe/nexus/pkg/types"
)

// Prompt implements types.IPrompt
type Prompt struct {
	Base
	Content    string            `json:"content"`
	Variables  map[string]string `json:"variables"`
	Purpose    string            `json:"purpose"`
	Target     string            `json:"target"`
	When       string            `json:"when"`
	IsTemplate bool              `json:"is_template"`
}

func NewPrompt(content, purpose, target, when string) *Prompt {
	return &Prompt{
		Base:      NewBase(),
		Content:   content,
		Variables: make(map[string]string),
		Purpose:   purpose,
		Target:    target,
		When:      when,
	}
}

func (p *Prompt) GetContent() string              { return p.Content }
func (p *Prompt) GetVariables() map[string]string { return p.Variables }
func (p *Prompt) GetPurpose() string              { return p.Purpose }
func (p *Prompt) GetTarget() string               { return p.Target }
func (p *Prompt) GetWhen() string                 { return p.When }

func (p *Prompt) Validate() error {
	if p.Content == "" {
		return fmt.Errorf("prompt content cannot be empty")
	}
	if p.IsTemplate {
		_, err := template.New("validate").Parse(p.Content)
		if err != nil {
			return fmt.Errorf("invalid template: %w", err)
		}
	}
	return nil
}

// PromptTemplate implements types.IPromptTemplate
type PromptTemplate struct {
	Base
	Template     string `json:"template"`
	TemplateType string `json:"template_type"`
}

func NewPromptTemplate(templateContent, templateType string) *PromptTemplate {
	return &PromptTemplate{
		Base:         NewBase(),
		Template:     templateContent,
		TemplateType: templateType,
	}
}

func (pt *PromptTemplate) Render(vars map[string]string) (string, error) {
	tmpl, err := template.New("prompt").Parse(pt.Template)
	if err != nil {
		return "", fmt.Errorf("template parse error: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, vars); err != nil {
		return "", fmt.Errorf("template execution error: %w", err)
	}

	return buf.String(), nil
}

func (pt *PromptTemplate) Parse(content string) error {
	_, err := template.New("validate").Parse(content)
	if err != nil {
		return fmt.Errorf("invalid template: %w", err)
	}
	pt.Template = content
	return nil
}

func (pt *PromptTemplate) GetTemplateType() string { return pt.TemplateType }

// PromptSelector implements types.IPromptSelector
type PromptSelector struct {
	Base
	Scenarios map[string]types.IPrompt `json:"scenarios"`
}

func NewPromptSelector() *PromptSelector {
	return &PromptSelector{
		Base:      NewBase(),
		Scenarios: make(map[string]types.IPrompt),
	}
}

func (ps *PromptSelector) SelectPrompt(scenario string, context map[string]interface{}) (types.IPrompt, error) {
	prompt, exists := ps.Scenarios[scenario]
	if !exists {
		return nil, fmt.Errorf("no prompt found for scenario: %s", scenario)
	}
	return prompt, nil
}

func (ps *PromptSelector) AddScenario(scenario string, prompt types.IPrompt) error {
	if _, exists := ps.Scenarios[scenario]; exists {
		return fmt.Errorf("scenario already exists: %s", scenario)
	}
	ps.Scenarios[scenario] = prompt
	return nil
}

func (ps *PromptSelector) RemoveScenario(scenario string) error {
	if _, exists := ps.Scenarios[scenario]; !exists {
		return fmt.Errorf("scenario not found: %s", scenario)
	}
	delete(ps.Scenarios, scenario)
	return nil
}

// Copyright (c) 2025 Gavin Volpe
//
// Permission is hereby granted, free of charge, to any person obtaining a copy of
// this software and associated documentation files (the "Software"), to deal in
// the Software without restriction, including without limitation the rights to
// use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of
// the Software, and to permit persons to whom the Software is furnished to do so,
// subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS
// FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR
// COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER
// IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN
// CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.

package models

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// ModelProvider represents different AI model providers
type ModelProvider string

const (
	OpenAI    ModelProvider = "openai"
	Azure     ModelProvider = "azure"
	Anthropic ModelProvider = "anthropic"
	Ollama    ModelProvider = "ollama"
	Groq      ModelProvider = "groq"
	Custom    ModelProvider = "custom"
)

// ModelRole represents the role in a conversation
type ModelRole string

const (
	RoleSystem    ModelRole = "system"
	RoleUser      ModelRole = "user"
	RoleAssistant ModelRole = "assistant"
	RoleFunction  ModelRole = "function"
)

// Message represents a single message in the conversation
type Message struct {
	Role         ModelRole     `json:"role"`
	Content      string        `json:"content"`
	Name         string        `json:"name,omitempty"`
	FunctionCall *FunctionCall `json:"function_call,omitempty"`
	ToolCalls    []ToolCall    `json:"tool_calls,omitempty"`
	ToolCallID   string        `json:"tool_call_id,omitempty"`
}

// FunctionCall represents a function call in the conversation
type FunctionCall struct {
	Name      string          `json:"name"`
	Arguments json.RawMessage `json:"arguments"`
}

// ToolCall represents a tool call in the conversation
type ToolCall struct {
	ID       string       `json:"id"`
	Type     string       `json:"type"`
	Function FunctionCall `json:"function"`
}

// ModelConfig represents the configuration for an AI model
type ModelConfig struct {
	Provider         ModelProvider `json:"provider"`
	ModelID          string        `json:"model_id"`
	BaseEndpoint     string        `json:"base_endpoint"`
	APIKey           string        `json:"api_key,omitempty"` // Optional for Ollama
	OrgID            string        `json:"org_id,omitempty"`
	Temperature      float32       `json:"temperature"`
	MaxTokens        int           `json:"max_tokens"`
	TopP             float32       `json:"top_p"`
	FrequencyPenalty float32       `json:"frequency_penalty"`
	PresencePenalty  float32       `json:"presence_penalty"`
	Stop             []string      `json:"stop,omitempty"`
	HTTPClient       *http.Client  `json:"-"`
	Headers          http.Header   `json:"-"`
	Timeout          time.Duration `json:"-"`
	RetryConfig      *RetryConfig  `json:"-"`
	// Ollama specific options
	Format  string         `json:"format,omitempty"`  // For Ollama: json or text
	Options map[string]any `json:"options,omitempty"` // Provider-specific options
}

// RetryConfig represents retry configuration
type RetryConfig struct {
	MaxRetries  int           `json:"max_retries"`
	InitialWait time.Duration `json:"initial_wait"`
	MaxWait     time.Duration `json:"max_wait"`
	Multiplier  float64       `json:"multiplier"`
}

// ModelResponse represents the response from an AI model
type ModelResponse struct {
	ID                string   `json:"id"`
	Created           int64    `json:"created"`
	Model             string   `json:"model"`
	SystemFingerprint string   `json:"system_fingerprint,omitempty"`
	Choices           []Choice `json:"choices"`
	Usage             Usage    `json:"usage"`
}

// Choice represents a single choice in the model response
type Choice struct {
	Index        int     `json:"index"`
	Message      Message `json:"message"`
	FinishReason string  `json:"finish_reason"`
}

// Usage represents token usage information
type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// IModel defines the interface for interacting with AI models
type IModel interface {
	// Core methods
	Complete(ctx context.Context, messages []Message) (*ModelResponse, error)
	Stream(ctx context.Context, messages []Message) (<-chan ModelResponse, error)

	// Configuration
	GetConfig() ModelConfig
	UpdateConfig(config ModelConfig) error

	// Token management
	CountTokens(messages []Message) (int, error)
	ValidateTokenCount(messages []Message) error

	// Function/Tool calling
	RegisterFunction(name string, parameters interface{}) error
	RegisterTool(name string, parameters interface{}) error

	// Error handling and retry logic
	WithRetry(config RetryConfig) IModel
	WithTimeout(timeout time.Duration) IModel
}

// BaseModel provides a base implementation of IModel
type BaseModel struct {
	config ModelConfig
	client *http.Client
}

// NewBaseModel creates a new BaseModel with the given configuration
func NewBaseModel(config ModelConfig) (*BaseModel, error) {
	if err := validateConfig(config); err != nil {
		return nil, err
	}

	client := config.HTTPClient
	if client == nil {
		client = &http.Client{
			Timeout: config.Timeout,
		}
	}

	return &BaseModel{
		config: config,
		client: client,
	}, nil
}

// validateConfig validates the model configuration
func validateConfig(config ModelConfig) error {
	if config.ModelID == "" {
		return fmt.Errorf("model ID is required")
	}
	if config.BaseEndpoint == "" {
		switch config.Provider {
		case Ollama:
			config.BaseEndpoint = "http://localhost:11434"
		case Groq:
			config.BaseEndpoint = "https://api.groq.com/v1"
		case OpenAI:
			config.BaseEndpoint = "https://api.openai.com/v1"
		default:
			return fmt.Errorf("base endpoint is required for provider %s", config.Provider)
		}
	}
	if config.APIKey == "" && config.Provider != Ollama {
		return fmt.Errorf("API key is required for provider %s", config.Provider)
	}
	return nil
}

// GetConfig returns the current model configuration
func (m *BaseModel) GetConfig() ModelConfig {
	return m.config
}

// UpdateConfig updates the model configuration
func (m *BaseModel) UpdateConfig(config ModelConfig) error {
	if err := validateConfig(config); err != nil {
		return err
	}
	m.config = config
	return nil
}

// WithRetry sets retry configuration for the model
func (m *BaseModel) WithRetry(config RetryConfig) IModel {
	m.config.RetryConfig = &config
	return m
}

// WithTimeout sets timeout for the model
func (m *BaseModel) WithTimeout(timeout time.Duration) IModel {
	m.config.Timeout = timeout
	if m.client != nil {
		m.client.Timeout = timeout
	}
	return m
}

// Complete implements IModel
func (m *BaseModel) Complete(ctx context.Context, messages []Message) (*ModelResponse, error) {
	return nil, fmt.Errorf("Complete method must be implemented by specific model provider")
}

// Stream implements IModel
func (m *BaseModel) Stream(ctx context.Context, messages []Message) (<-chan ModelResponse, error) {
	return nil, fmt.Errorf("Stream method must be implemented by specific model provider")
}

// CountTokens implements IModel
func (m *BaseModel) CountTokens(messages []Message) (int, error) {
	return 0, fmt.Errorf("CountTokens method must be implemented by specific model provider")
}

// ValidateTokenCount implements IModel
func (m *BaseModel) ValidateTokenCount(messages []Message) error {
	count, err := m.CountTokens(messages)
	if err != nil {
		return fmt.Errorf("failed to count tokens: %w", err)
	}
	if count > m.config.MaxTokens {
		return fmt.Errorf("token count %d exceeds maximum allowed %d", count, m.config.MaxTokens)
	}
	return nil
}

// RegisterFunction implements IModel
func (m *BaseModel) RegisterFunction(name string, parameters interface{}) error {
	return fmt.Errorf("RegisterFunction method must be implemented by specific model provider")
}

// RegisterTool implements IModel
func (m *BaseModel) RegisterTool(name string, parameters interface{}) error {
	return fmt.Errorf("RegisterTool method must be implemented by specific model provider")
}

// DefaultConfig returns a default model configuration
func DefaultConfig() ModelConfig {
	return ModelConfig{
		Provider:         OpenAI,
		Temperature:      0.7,
		MaxTokens:        2000,
		TopP:             1.0,
		FrequencyPenalty: 0.0,
		PresencePenalty:  0.0,
		Timeout:          30 * time.Second,
		Headers:          make(http.Header),
		RetryConfig: &RetryConfig{
			MaxRetries:  3,
			InitialWait: 1 * time.Second,
			MaxWait:     10 * time.Second,
			Multiplier:  2.0,
		},
	}
}

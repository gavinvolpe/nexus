package models

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// GroqModel implements IModel for Groq
type GroqModel struct {
	*BaseModel
	*MCPModelMixin
	client *http.Client
	apiKey string
}

// GroqRequest represents the request format for Groq API
type GroqRequest struct {
	Model       string    `json:"model"`
	Messages    []Message `json:"messages"`
	Stream      bool      `json:"stream"`
	Temperature float32   `json:"temperature,omitempty"`
	MaxTokens   int       `json:"max_tokens,omitempty"`
	TopP        float32   `json:"top_p,omitempty"`
}

// GroqResponse represents the response format from Groq API
type GroqResponse struct {
	ID                string   `json:"id"`
	Object            string   `json:"object"`
	Created           int64    `json:"created"`
	Model             string   `json:"model"`
	SystemFingerprint string   `json:"system_fingerprint"`
	Choices           []Choice `json:"choices"`
	Usage             Usage    `json:"usage"`
}

// NewGroqModel creates a new Groq model instance
func NewGroqModel(config ModelConfig) (*GroqModel, error) {
	if config.Provider != Groq {
		config.Provider = Groq
	}

	base, err := NewBaseModel(config)
	if err != nil {
		return nil, err
	}

	return &GroqModel{
		BaseModel:     base,
		MCPModelMixin: NewMCPModelMixin(),
		client:        &http.Client{},
		apiKey:        config.APIKey,
	}, nil
}

// Complete implements IModel
func (m *GroqModel) Complete(ctx context.Context, messages []Message) (*ModelResponse, error) {
	url := fmt.Sprintf("%s/chat/completions", m.config.BaseEndpoint)

	req := GroqRequest{
		Model:       m.config.ModelID,
		Messages:    messages,
		Stream:      false,
		Temperature: m.config.Temperature,
		MaxTokens:   m.config.MaxTokens,
		TopP:        m.config.TopP,
	}

	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", fmt.Sprintf("Bearer %s", m.apiKey))

	resp, err := m.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var groqResp GroqResponse
	if err := json.NewDecoder(resp.Body).Decode(&groqResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &ModelResponse{
		ID:                groqResp.ID,
		Created:          groqResp.Created,
		Model:            groqResp.Model,
		SystemFingerprint: groqResp.SystemFingerprint,
		Choices:          groqResp.Choices,
		Usage:            groqResp.Usage,
	}, nil
}

// Stream implements IModel
func (m *GroqModel) Stream(ctx context.Context, messages []Message) (<-chan ModelResponse, error) {
	url := fmt.Sprintf("%s/chat/completions", m.config.BaseEndpoint)
	responseChan := make(chan ModelResponse)

	req := GroqRequest{
		Model:       m.config.ModelID,
		Messages:    messages,
		Stream:      true,
		Temperature: m.config.Temperature,
		MaxTokens:   m.config.MaxTokens,
		TopP:        m.config.TopP,
	}

	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", fmt.Sprintf("Bearer %s", m.apiKey))

	go func() {
		defer close(responseChan)

		resp, err := m.client.Do(httpReq)
		if err != nil {
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return
		}

		reader := bufio.NewReader(resp.Body)
		for {
			line, err := reader.ReadBytes('\n')
			if err != nil {
				if err != io.EOF {
					return
				}
				break
			}

			if !bytes.HasPrefix(line, []byte("data: ")) {
				continue
			}

			data := bytes.TrimPrefix(line, []byte("data: "))
			if len(data) == 0 {
				continue
			}

			var groqResp GroqResponse
			if err := json.Unmarshal(data, &groqResp); err != nil {
				continue
			}

			response := ModelResponse{
				ID:                groqResp.ID,
				Created:          groqResp.Created,
				Model:            groqResp.Model,
				SystemFingerprint: groqResp.SystemFingerprint,
				Choices:          groqResp.Choices,
				Usage:            groqResp.Usage,
			}

			select {
			case <-ctx.Done():
				return
			case responseChan <- response:
			}

			if len(groqResp.Choices) > 0 && groqResp.Choices[0].FinishReason != "" {
				return
			}
		}
	}()

	return responseChan, nil
}

// CountTokens implements IModel
func (m *GroqModel) CountTokens(messages []Message) (int, error) {
	// Groq doesn't provide a token counting endpoint
	// This is a rough estimate based on character count
	total := 0
	for _, msg := range messages {
		total += len(msg.Content) / 4 // rough estimate: 4 chars per token
	}
	return total, nil
}

// RegisterFunction implements IModel
func (m *GroqModel) RegisterFunction(name string, parameters interface{}) error {
	// Groq supports OpenAI-compatible function calling
	if m.config.Options == nil {
		m.config.Options = make(map[string]any)
	}

	functions, _ := m.config.Options["functions"].([]map[string]any)
	functions = append(functions, map[string]any{
		"name":       name,
		"parameters": parameters,
	})
	m.config.Options["functions"] = functions

	return nil
}

// RegisterTool implements IModel
func (m *GroqModel) RegisterTool(name string, parameters interface{}) error {
	// Groq supports OpenAI-compatible tool calling
	if m.config.Options == nil {
		m.config.Options = make(map[string]any)
	}

	tools, _ := m.config.Options["tools"].([]map[string]any)
	tools = append(tools, map[string]any{
		"type": "function",
		"function": map[string]any{
			"name":       name,
			"parameters": parameters,
		},
	})
	m.config.Options["tools"] = tools

	return nil
}

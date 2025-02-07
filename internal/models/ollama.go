package models

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// OllamaModel implements IModel for Ollama
type OllamaModel struct {
	*BaseModel
}

// OllamaRequest represents the request format for Ollama API
type OllamaRequest struct {
	Model    string         `json:"model"`
	Messages []Message      `json:"messages"`
	Stream   bool           `json:"stream"`
	Options  map[string]any `json:"options,omitempty"`
	Format   string         `json:"format,omitempty"`
	Tools    []Tool         `json:"tools,omitempty"`
}

// Tool represents a tool that can be used by the model
type Tool struct {
	Type     string          `json:"type"`
	Function ToolFunction    `json:"function"`
}

// ToolFunction represents a function that can be called by the model
type ToolFunction struct {
	Name        string          `json:"name"`
	Description string          `json:"description"`
	Parameters  interface{}     `json:"parameters"`
}

// OllamaResponse represents the response format from Ollama API
type OllamaResponse struct {
	Model     string          `json:"model"`
	CreatedAt string          `json:"created_at"`
	Message   Message         `json:"message"`
	Done      bool           `json:"done"`
	Context   []int          `json:"context,omitempty"`
	TotalDuration int64      `json:"total_duration,omitempty"`
	LoadDuration  int64      `json:"load_duration,omitempty"`
	PromptEvalCount  int     `json:"prompt_eval_count,omitempty"`
	EvalCount       int     `json:"eval_count,omitempty"`
	EvalDuration    int64   `json:"eval_duration,omitempty"`
}

// NewOllamaModel creates a new Ollama model instance
func NewOllamaModel(config ModelConfig) (*OllamaModel, error) {
	if config.Provider != Ollama {
		config.Provider = Ollama
	}
	
	base, err := NewBaseModel(config)
	if err != nil {
		return nil, err
	}
	
	return &OllamaModel{
		BaseModel: base,
	}, nil
}

// Complete implements IModel
func (m *OllamaModel) Complete(ctx context.Context, messages []Message) (*ModelResponse, error) {
	url := fmt.Sprintf("%s/api/chat", m.config.BaseEndpoint)
	
	tools, _ := m.config.Options["tools"].([]Tool)
	
	req := OllamaRequest{
		Model:    m.config.ModelID,
		Messages: messages,
		Stream:   false,
		Options:  m.config.Options,
		Format:   m.config.Format,
		Tools:    tools,
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
	
	resp, err := m.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("request failed with status %d: %s", resp.StatusCode, string(body))
	}
	
	var ollamaResp OllamaResponse
	if err := json.NewDecoder(resp.Body).Decode(&ollamaResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}
	
	return &ModelResponse{
		ID:      fmt.Sprintf("ollama-%s-%s", m.config.ModelID, ollamaResp.CreatedAt),
		Created: 0, // Ollama uses different timestamp format
		Model:   m.config.ModelID,
		Choices: []Choice{
			{
				Index:        0,
				Message:      ollamaResp.Message,
				FinishReason: "stop",
			},
		},
		Usage: Usage{
			PromptTokens:     ollamaResp.PromptEvalCount,
			CompletionTokens: ollamaResp.EvalCount,
			TotalTokens:      ollamaResp.PromptEvalCount + ollamaResp.EvalCount,
		},
	}, nil
}

// Stream implements IModel
func (m *OllamaModel) Stream(ctx context.Context, messages []Message) (<-chan ModelResponse, error) {
	url := fmt.Sprintf("%s/api/chat", m.config.BaseEndpoint)
	responseChan := make(chan ModelResponse)
	
	tools, _ := m.config.Options["tools"].([]Tool)
	
	req := OllamaRequest{
		Model:    m.config.ModelID,
		Messages: messages,
		Stream:   true,
		Options:  m.config.Options,
		Format:   m.config.Format,
		Tools:    tools,
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
		
		decoder := json.NewDecoder(resp.Body)
		for {
			var ollamaResp OllamaResponse
			if err := decoder.Decode(&ollamaResp); err != nil {
				if err != io.EOF {
					return
				}
				break
			}
			
			response := ModelResponse{
				ID:      fmt.Sprintf("ollama-%s-%s", m.config.ModelID, ollamaResp.CreatedAt),
				Created: 0,
				Model:   m.config.ModelID,
				Choices: []Choice{
					{
						Index:        0,
						Message:      ollamaResp.Message,
						FinishReason: "stop",
					},
				},
				Usage: Usage{
					PromptTokens:     ollamaResp.PromptEvalCount,
					CompletionTokens: ollamaResp.EvalCount,
					TotalTokens:      ollamaResp.PromptEvalCount + ollamaResp.EvalCount,
				},
			}
			
			select {
			case <-ctx.Done():
				return
			case responseChan <- response:
			}
			
			if ollamaResp.Done {
				break
			}
		}
	}()
	
	return responseChan, nil
}

// CountTokens implements IModel
func (m *OllamaModel) CountTokens(messages []Message) (int, error) {
	// Ollama doesn't provide a direct token counting endpoint
	// This is a rough estimate based on character count
	total := 0
	for _, msg := range messages {
		total += len(msg.Content) / 4 // rough estimate: 4 chars per token
	}
	return total, nil
}

// RegisterFunction implements IModel
func (m *OllamaModel) RegisterFunction(name string, parameters interface{}) error {
	if m.config.Options == nil {
		m.config.Options = make(map[string]any)
	}
	
	tools, _ := m.config.Options["tools"].([]Tool)
	if tools == nil {
		tools = make([]Tool, 0)
	}
	
	tools = append(tools, Tool{
		Type: "function",
		Function: ToolFunction{
			Name:       name,
			Parameters: parameters,
		},
	})
	
	m.config.Options["tools"] = tools
	return nil
}

// RegisterTool implements IModel
func (m *OllamaModel) RegisterTool(name string, parameters interface{}) error {
	return m.RegisterFunction(name, parameters) // Ollama uses the same format for tools and functions
}

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

package mcp

import (
	"encoding/json"
	"time"

	"github.com/gorilla/websocket"
)

// MCPMethod represents the method type for MCP requests
type MCPMethod string

const (
	Initialize     MCPMethod = "initialize"
	Initialized    MCPMethod = "initialized"
	ToolsList      MCPMethod = "tools/list"
	ToolsCall      MCPMethod = "tools/call"
	ResourcesList  MCPMethod = "resources/list"
	ResourcesRead  MCPMethod = "resources/read"
	ResourcesWrite MCPMethod = "resources/write"
	PromptsRender  MCPMethod = "prompts/render"
	PromptsList    MCPMethod = "prompts/list"
	Notification   MCPMethod = "$/notification"
	CancelRequest  MCPMethod = "$/cancelRequest"
)

// MCPMessage represents the base message structure for MCP
type MCPMessage struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      interface{}     `json:"id,omitempty"`
	Method  MCPMethod       `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
	Result  json.RawMessage `json:"result,omitempty"`
	Error   *MCPError       `json:"error,omitempty"`
	Conn    *websocket.Conn `json:"-"`
}

// MCPError represents an error in MCP
type MCPError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

// InitializeParams represents the parameters for initialize request
type InitializeParams struct {
	RootURI      string             `json:"rootUri"`
	Capabilities ClientCapabilities `json:"capabilities"`
}

// ClientCapabilities represents the capabilities of an MCP client
type ClientCapabilities struct {
	Tools     ToolsClientCapabilities     `json:"tools,omitempty"`
	Resources ResourcesClientCapabilities `json:"resources,omitempty"`
	Prompts   PromptsClientCapabilities   `json:"prompts,omitempty"`
}

// ToolsClientCapabilities represents tool-related capabilities
type ToolsClientCapabilities struct {
	Call bool `json:"call"`
	List bool `json:"list"`
}

// ResourcesClientCapabilities represents resource-related capabilities
type ResourcesClientCapabilities struct {
	Read  bool `json:"read"`
	Write bool `json:"write"`
	List  bool `json:"list"`
}

// PromptsClientCapabilities represents prompt-related capabilities
type PromptsClientCapabilities struct {
	Render bool `json:"render"`
	List   bool `json:"list"`
}

// ServerCapabilities represents the capabilities of an MCP server
type ServerCapabilities struct {
	Tools     ToolsServerCapabilities     `json:"tools,omitempty"`
	Resources ResourcesServerCapabilities `json:"resources,omitempty"`
	Prompts   PromptsServerCapabilities   `json:"prompts,omitempty"`
}

// ToolsServerCapabilities represents server tool capabilities
type ToolsServerCapabilities struct {
	Supported bool     `json:"supported"`
	Types     []string `json:"types,omitempty"`
}

// ResourcesServerCapabilities represents server resource capabilities
type ResourcesServerCapabilities struct {
	Supported bool     `json:"supported"`
	Types     []string `json:"types,omitempty"`
}

// PromptsServerCapabilities represents server prompt capabilities
type PromptsServerCapabilities struct {
	Supported bool     `json:"supported"`
	Types     []string `json:"types,omitempty"`
}

// Tool represents an MCP tool
type Tool struct {
	Name        string      `json:"name"`
	Description string      `json:"description"`
	Parameters  interface{} `json:"parameters"`
	Returns     interface{} `json:"returns,omitempty"`
}

// ToolCallParams represents parameters for a tool call
type ToolCallParams struct {
	Name      string          `json:"name"`
	Arguments json.RawMessage `json:"arguments"`
	Timeout   time.Duration   `json:"timeout,omitempty"`
}

// Resource represents an MCP resource
type Resource struct {
	URI         string      `json:"uri"`
	Type        string      `json:"type"`
	Name        string      `json:"name"`
	Description string      `json:"description"`
	Metadata    interface{} `json:"metadata,omitempty"`
}

// Prompt represents an MCP prompt
type Prompt struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Template    string                 `json:"template"`
	Variables   map[string]interface{} `json:"variables,omitempty"`
}

// NotificationParams represents parameters for notifications
type NotificationParams struct {
	Type    string `json:"type"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

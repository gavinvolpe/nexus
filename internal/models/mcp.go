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

	"github.com/gavinvolpe/nexus/internal/mcp"
)

// IMCPModel extends IModel with MCP capabilities
type IMCPModel interface {
	IModel

	// MCP server capabilities
	StartMCPServer(addr string) error
	StopMCPServer() error
	RegisterMCPTool(tool mcp.Tool) error
	RegisterMCPResource(resource mcp.Resource) error
	RegisterMCPPrompt(prompt mcp.Prompt) error

	// MCP client capabilities
	ConnectToMCP(url string) error
	DisconnectFromMCP() error
	CallMCPTool(ctx context.Context, name string, args interface{}) (json.RawMessage, error)
	ListMCPTools(ctx context.Context) ([]mcp.Tool, error)
}

// MCPModelMixin provides MCP capabilities to a model
type MCPModelMixin struct {
	server *mcp.Server
	client *mcp.Client
}

// NewMCPModelMixin creates a new MCPModelMixin
func NewMCPModelMixin() *MCPModelMixin {
	return &MCPModelMixin{}
}

// StartMCPServer starts the MCP server
func (m *MCPModelMixin) StartMCPServer(addr string) error {
	if m.server != nil {
		return fmt.Errorf("MCP server already running")
	}

	m.server = mcp.NewServer()
	// TODO: Start HTTP server with WebSocket upgrade
	return nil
}

// StopMCPServer stops the MCP server
func (m *MCPModelMixin) StopMCPServer() error {
	if m.server == nil {
		return fmt.Errorf("MCP server not running")
	}

	// TODO: Stop HTTP server
	m.server = nil
	return nil
}

// RegisterMCPTool registers a tool with the MCP server
func (m *MCPModelMixin) RegisterMCPTool(tool mcp.Tool) error {
	if m.server == nil {
		return fmt.Errorf("MCP server not running")
	}
	return m.server.RegisterTool(tool)
}

// RegisterMCPResource registers a resource with the MCP server
func (m *MCPModelMixin) RegisterMCPResource(resource mcp.Resource) error {
	if m.server == nil {
		return fmt.Errorf("MCP server not running")
	}
	return m.server.RegisterResource(resource)
}

// RegisterMCPPrompt registers a prompt with the MCP server
func (m *MCPModelMixin) RegisterMCPPrompt(prompt mcp.Prompt) error {
	if m.server == nil {
		return fmt.Errorf("MCP server not running")
	}
	return m.server.RegisterPrompt(prompt)
}

// ConnectToMCP connects to an MCP server
func (m *MCPModelMixin) ConnectToMCP(url string) error {
	if m.client != nil {
		return fmt.Errorf("already connected to MCP server")
	}

	capabilities := mcp.ClientCapabilities{
		Tools: mcp.ToolsClientCapabilities{
			Call: true,
			List: true,
		},
		Resources: mcp.ResourcesClientCapabilities{
			Read:  true,
			Write: true,
			List:  true,
		},
		Prompts: mcp.PromptsClientCapabilities{
			Render: true,
			List:   true,
		},
	}

	client, err := mcp.NewClient(url, capabilities)
	if err != nil {
		return fmt.Errorf("error connecting to MCP server: %w", err)
	}

	m.client = client
	return nil
}

// DisconnectFromMCP disconnects from the MCP server
func (m *MCPModelMixin) DisconnectFromMCP() error {
	if m.client == nil {
		return fmt.Errorf("not connected to MCP server")
	}

	if err := m.client.Close(); err != nil {
		return fmt.Errorf("error closing MCP connection: %w", err)
	}

	m.client = nil
	return nil
}

// CallMCPTool calls a tool on the connected MCP server
func (m *MCPModelMixin) CallMCPTool(ctx context.Context, name string, args interface{}) (json.RawMessage, error) {
	if m.client == nil {
		return nil, fmt.Errorf("not connected to MCP server")
	}

	return m.client.CallTool(ctx, name, args)
}

// ListMCPTools lists tools available on the connected MCP server
func (m *MCPModelMixin) ListMCPTools(ctx context.Context) ([]mcp.Tool, error) {
	if m.client == nil {
		return nil, fmt.Errorf("not connected to MCP server")
	}

	return m.client.ListTools(ctx)
}

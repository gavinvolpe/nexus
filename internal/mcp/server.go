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
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/bytedance/sonic"
	"github.com/gorilla/websocket"
)

// Server represents an MCP server
type Server struct {
	capabilities ServerCapabilities
	tools        map[string]Tool
	resources    map[string]Resource
	prompts      map[string]Prompt
	handlers     map[MCPMethod]HandlerFunc
	clients      map[*websocket.Conn]*ClientState
	mu           sync.RWMutex
}

// ClientState represents the state of a connected client
type ClientState struct {
	Capabilities ClientCapabilities
	RootURI      string
	Initialized  bool
}

// HandlerFunc represents a function that handles MCP messages
type HandlerFunc func(ctx context.Context, msg *MCPMessage) (*MCPMessage, error)

// NewServer creates a new MCP server
func NewServer() *Server {
	s := &Server{
		tools:     make(map[string]Tool),
		resources: make(map[string]Resource),
		prompts:   make(map[string]Prompt),
		handlers:  make(map[MCPMethod]HandlerFunc),
		clients:   make(map[*websocket.Conn]*ClientState),
	}

	// Register default handlers
	s.handlers[Initialize] = s.handleInitialize
	s.handlers[Initialized] = s.handleInitialized
	s.handlers[ToolsList] = s.handleToolsList
	s.handlers[ToolsCall] = s.handleToolsCall
	s.handlers[ResourcesList] = s.handleResourcesList
	s.handlers[ResourcesRead] = s.handleResourcesRead
	s.handlers[ResourcesWrite] = s.handleResourcesWrite
	s.handlers[PromptsList] = s.handlePromptsList
	s.handlers[PromptsRender] = s.handlePromptsRender

	return s
}

// RegisterTool registers a new tool with the server
func (s *Server) RegisterTool(tool Tool) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.tools[tool.Name]; exists {
		return fmt.Errorf("tool %s already registered", tool.Name)
	}

	s.tools[tool.Name] = tool
	return nil
}

// RegisterResource registers a new resource with the server
func (s *Server) RegisterResource(resource Resource) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.resources[resource.URI]; exists {
		return fmt.Errorf("resource %s already registered", resource.URI)
	}

	s.resources[resource.URI] = resource
	return nil
}

// RegisterPrompt registers a new prompt with the server
func (s *Server) RegisterPrompt(prompt Prompt) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.prompts[prompt.Name]; exists {
		return fmt.Errorf("prompt %s already registered", prompt.Name)
	}

	s.prompts[prompt.Name] = prompt
	return nil
}

// HandleConnection handles a new WebSocket connection
func (s *Server) HandleConnection(conn *websocket.Conn) {
	defer conn.Close()

	s.mu.Lock()
	s.clients[conn] = &ClientState{}
	s.mu.Unlock()

	defer func() {
		s.mu.Lock()
		delete(s.clients, conn)
		s.mu.Unlock()
	}()

	for {
		var msg MCPMessage
		if err := conn.ReadJSON(&msg); err != nil {
			log.Printf("error reading message: %v", err)
			return
		}

		// Set the connection for the message
		msg.Conn = conn

		handler, ok := s.handlers[msg.Method]
		if !ok {
			s.sendError(conn, msg.ID, -32601, "method not found")
			continue
		}

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		response, err := handler(ctx, &msg)
		cancel()

		if err != nil {
			s.sendError(conn, msg.ID, -32000, err.Error())
			continue
		}

		if response != nil {
			if err := conn.WriteJSON(response); err != nil {
				log.Printf("error writing response: %v", err)
				return
			}
		}
	}
}

func (s *Server) sendError(conn *websocket.Conn, id interface{}, code int, message string) {
	response := MCPMessage{
		JSONRPC: "2.0",
		ID:      id,
		Error: &MCPError{
			Code:    code,
			Message: message,
		},
	}

	if err := conn.WriteJSON(response); err != nil {
		log.Printf("error sending error response: %v", err)
	}
}

// Handler implementations
func (s *Server) handleInitialize(ctx context.Context, msg *MCPMessage) (*MCPMessage, error) {
	var params InitializeParams
	if err := sonic.Unmarshal(msg.Params, &params); err != nil {
		return nil, fmt.Errorf("invalid initialize params: %w", err)
	}

	s.mu.Lock()
	if client, ok := s.clients[msg.Conn]; ok {
		client.Capabilities = params.Capabilities
		client.RootURI = params.RootURI
	}
	s.mu.Unlock()

	response := MCPMessage{
		JSONRPC: "2.0",
		ID:      msg.ID,
		Result: json.RawMessage(`{
			"capabilities": {
				"tools": {
					"supported": true,
					"types": ["function"]
				},
				"resources": {
					"supported": true,
					"types": ["file", "memory"]
				},
				"prompts": {
					"supported": true,
					"types": ["text", "chat"]
				}
			}
		}`),
	}

	return &response, nil
}

func (s *Server) handleInitialized(ctx context.Context, msg *MCPMessage) (*MCPMessage, error) {
	s.mu.Lock()
	if client, ok := s.clients[msg.Conn]; ok {
		client.Initialized = true
	}
	s.mu.Unlock()

	return nil, nil
}

func (s *Server) handleToolsList(ctx context.Context, msg *MCPMessage) (*MCPMessage, error) {
	s.mu.RLock()
	tools := make([]Tool, 0, len(s.tools))
	for _, tool := range s.tools {
		tools = append(tools, tool)
	}
	s.mu.RUnlock()

	result, err := sonic.Marshal(map[string]interface{}{
		"tools": tools,
	})
	if err != nil {
		return nil, fmt.Errorf("error marshaling tools: %w", err)
	}

	return &MCPMessage{
		JSONRPC: "2.0",
		ID:      msg.ID,
		Result:  result,
	}, nil
}

func (s *Server) handleToolsCall(ctx context.Context, msg *MCPMessage) (*MCPMessage, error) {
	var params ToolCallParams
	if err := sonic.Unmarshal(msg.Params, &params); err != nil {
		return nil, fmt.Errorf("invalid tool call params: %w", err)
	}

	s.mu.RLock()
	_, exists := s.tools[params.Name]
	s.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("tool %s not found", params.Name)
	}

	// TODO: Implement actual tool execution
	result := json.RawMessage(`{"status": "success"}`)

	return &MCPMessage{
		JSONRPC: "2.0",
		ID:      msg.ID,
		Result:  result,
	}, nil
}

func (s *Server) handleResourcesList(ctx context.Context, msg *MCPMessage) (*MCPMessage, error) {
	s.mu.RLock()
	resources := make([]Resource, 0, len(s.resources))
	for _, resource := range s.resources {
		resources = append(resources, resource)
	}
	s.mu.RUnlock()

	result, err := sonic.Marshal(map[string]interface{}{
		"resources": resources,
	})
	if err != nil {
		return nil, fmt.Errorf("error marshaling resources: %w", err)
	}

	return &MCPMessage{
		JSONRPC: "2.0",
		ID:      msg.ID,
		Result:  result,
	}, nil
}

func (s *Server) handleResourcesRead(ctx context.Context, msg *MCPMessage) (*MCPMessage, error) {
	var params struct {
		URI string `json:"uri"`
	}
	if err := sonic.Unmarshal(msg.Params, &params); err != nil {
		return nil, fmt.Errorf("invalid resource read params: %w", err)
	}

	s.mu.RLock()
	resource, ok := s.resources[params.URI]
	s.mu.RUnlock()

	if !ok {
		return nil, fmt.Errorf("resource %s not found", params.URI)
	}

	result, err := sonic.Marshal(map[string]interface{}{
		"resource": resource,
	})
	if err != nil {
		return nil, fmt.Errorf("error marshaling resource: %w", err)
	}

	return &MCPMessage{
		JSONRPC: "2.0",
		ID:      msg.ID,
		Result:  result,
	}, nil
}

func (s *Server) handleResourcesWrite(ctx context.Context, msg *MCPMessage) (*MCPMessage, error) {
	var params struct {
		URI     string      `json:"uri"`
		Content interface{} `json:"content"`
	}
	if err := sonic.Unmarshal(msg.Params, &params); err != nil {
		return nil, fmt.Errorf("invalid resource write params: %w", err)
	}

	s.mu.Lock()
	if resource, ok := s.resources[params.URI]; ok {
		resource.Metadata = params.Content
		s.resources[params.URI] = resource
	} else {
		return nil, fmt.Errorf("resource %s not found", params.URI)
	}
	s.mu.Unlock()

	return &MCPMessage{
		JSONRPC: "2.0",
		ID:      msg.ID,
		Result:  json.RawMessage(`{"status":"success"}`),
	}, nil
}

func (s *Server) handlePromptsList(ctx context.Context, msg *MCPMessage) (*MCPMessage, error) {
	s.mu.RLock()
	prompts := make([]Prompt, 0, len(s.prompts))
	for _, prompt := range s.prompts {
		prompts = append(prompts, prompt)
	}
	s.mu.RUnlock()

	result, err := sonic.Marshal(map[string]interface{}{
		"prompts": prompts,
	})
	if err != nil {
		return nil, fmt.Errorf("error marshaling prompts: %w", err)
	}

	return &MCPMessage{
		JSONRPC: "2.0",
		ID:      msg.ID,
		Result:  result,
	}, nil
}

func (s *Server) handlePromptsRender(ctx context.Context, msg *MCPMessage) (*MCPMessage, error) {
	var params struct {
		Name      string                 `json:"name"`
		Variables map[string]interface{} `json:"variables"`
	}
	if err := sonic.Unmarshal(msg.Params, &params); err != nil {
		return nil, fmt.Errorf("invalid prompt render params: %w", err)
	}

	s.mu.RLock()
	prompt, ok := s.prompts[params.Name]
	s.mu.RUnlock()

	if !ok {
		return nil, fmt.Errorf("prompt %s not found", params.Name)
	}

	// Simple template rendering - in a real implementation, you'd want to use a proper template engine
	rendered := prompt.Template
	for key, value := range params.Variables {
		rendered = strings.ReplaceAll(rendered, "{{"+key+"}}", fmt.Sprintf("%v", value))
	}

	result, err := sonic.Marshal(map[string]interface{}{
		"rendered": rendered,
	})
	if err != nil {
		return nil, fmt.Errorf("error marshaling rendered prompt: %w", err)
	}

	return &MCPMessage{
		JSONRPC: "2.0",
		ID:      msg.ID,
		Result:  result,
	}, nil
}

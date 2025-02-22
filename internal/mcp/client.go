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
	"sync"
	"sync/atomic"

	"github.com/bytedance/sonic"
	"github.com/gorilla/websocket"
)

// Client represents an MCP client
type Client struct {
	conn         *websocket.Conn
	nextID       atomic.Int64
	capabilities ClientCapabilities
	handlers     map[MCPMethod]HandlerFunc
	responses    map[interface{}]chan *MCPMessage
	mu           sync.RWMutex
}

// NewClient creates a new MCP client
func NewClient(url string, capabilities ClientCapabilities) (*Client, error) {
	conn, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		return nil, fmt.Errorf("error connecting to server: %w", err)
	}

	client := &Client{
		conn:         conn,
		capabilities: capabilities,
		handlers:     make(map[MCPMethod]HandlerFunc),
		responses:    make(map[interface{}]chan *MCPMessage),
	}

	// Start message handler
	go client.handleMessages()

	return client, nil
}

// Initialize initializes the connection with the server
func (c *Client) Initialize(ctx context.Context, rootURI string) (*ServerCapabilities, error) {
	params := InitializeParams{
		RootURI:      rootURI,
		Capabilities: c.capabilities,
	}

	paramsBytes, err := sonic.Marshal(params)
	if err != nil {
		return nil, fmt.Errorf("error marshaling params: %w", err)
	}

	response, err := c.sendRequest(ctx, Initialize, paramsBytes)
	if err != nil {
		return nil, fmt.Errorf("initialize request failed: %w", err)
	}

	var result struct {
		Capabilities ServerCapabilities `json:"capabilities"`
	}
	if err := sonic.Unmarshal(response.Result, &result); err != nil {
		return nil, fmt.Errorf("error unmarshaling server capabilities: %w", err)
	}

	// Send initialized notification
	if err := c.sendNotification(Initialized, nil); err != nil {
		return nil, fmt.Errorf("error sending initialized notification: %w", err)
	}

	return &result.Capabilities, nil
}

// ListTools retrieves the list of available tools from the server
func (c *Client) ListTools(ctx context.Context) ([]Tool, error) {
	response, err := c.sendRequest(ctx, ToolsList, nil)
	if err != nil {
		return nil, fmt.Errorf("tools/list request failed: %w", err)
	}

	var result struct {
		Tools []Tool `json:"tools"`
	}
	if err := sonic.Unmarshal(response.Result, &result); err != nil {
		return nil, fmt.Errorf("error unmarshaling tools list: %w", err)
	}

	return result.Tools, nil
}

// CallTool calls a tool on the server
func (c *Client) CallTool(ctx context.Context, name string, arguments interface{}) (json.RawMessage, error) {
	argsBytes, err := sonic.Marshal(arguments)
	if err != nil {
		return nil, fmt.Errorf("error marshaling arguments: %w", err)
	}

	params := ToolCallParams{
		Name:      name,
		Arguments: argsBytes,
	}

	paramsBytes, err := sonic.Marshal(params)
	if err != nil {
		return nil, fmt.Errorf("error marshaling params: %w", err)
	}

	response, err := c.sendRequest(ctx, ToolsCall, paramsBytes)
	if err != nil {
		return nil, fmt.Errorf("tools/call request failed: %w", err)
	}

	return response.Result, nil
}

// Close closes the client connection
func (c *Client) Close() error {
	return c.conn.Close()
}

func (c *Client) handleMessages() {
	for {
		var msg MCPMessage
		if err := c.conn.ReadJSON(&msg); err != nil {
			// Handle connection closed
			return
		}

		if msg.Method == Notification {
			if handler, ok := c.handlers[Notification]; ok {
				go func() {
					if _, err := handler(context.Background(), &msg); err != nil {
						// Handle notification handler error
					}
				}()
			}
			continue
		}

		c.mu.RLock()
		ch, ok := c.responses[msg.ID]
		c.mu.RUnlock()

		if ok {
			ch <- &msg
		}
	}
}

func (c *Client) sendRequest(ctx context.Context, method MCPMethod, params json.RawMessage) (*MCPMessage, error) {
	id := c.nextID.Add(1)
	ch := make(chan *MCPMessage, 1)

	c.mu.Lock()
	c.responses[id] = ch
	c.mu.Unlock()

	defer func() {
		c.mu.Lock()
		delete(c.responses, id)
		c.mu.Unlock()
	}()

	msg := MCPMessage{
		JSONRPC: "2.0",
		ID:      id,
		Method:  method,
		Params:  params,
	}

	if err := c.conn.WriteJSON(msg); err != nil {
		return nil, fmt.Errorf("error sending request: %w", err)
	}

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case response := <-ch:
		if response.Error != nil {
			return nil, fmt.Errorf("server error: %s", response.Error.Message)
		}
		return response, nil
	}
}

func (c *Client) sendNotification(method MCPMethod, params json.RawMessage) error {
	msg := MCPMessage{
		JSONRPC: "2.0",
		Method:  method,
		Params:  params,
	}

	return c.conn.WriteJSON(msg)
}

// RegisterNotificationHandler registers a handler for notifications
func (c *Client) RegisterNotificationHandler(handler HandlerFunc) {
	c.handlers[Notification] = handler
}

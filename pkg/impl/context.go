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

package impl

import (
	"sync"
)

// Context implements types.IContext
type Context struct {
	Base
	mu    sync.RWMutex
	store map[string]interface{}
}

func NewContext() *Context {
	return &Context{
		Base:  NewBase(),
		store: make(map[string]interface{}),
	}
}

func (c *Context) Set(key string, value interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.store[key] = value
}

func (c *Context) Get(key string) interface{} {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.store[key]
}

func (c *Context) GetAll() map[string]interface{} {
	c.mu.RLock()
	defer c.mu.RUnlock()
	result := make(map[string]interface{}, len(c.store))
	for k, v := range c.store {
		result[k] = v
	}
	return result
}

func (c *Context) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.store = make(map[string]interface{})
}

// ActionResult implements types.IActionResult
type ActionResult struct {
	Base
	Output   interface{}            `json:"output"`
	Err      error                  `json:"error,omitempty"`
	Metadata map[string]interface{} `json:"metadata"`
	Success  bool                   `json:"success"`
}

func NewActionResult(output interface{}, err error, metadata map[string]interface{}) *ActionResult {
	return &ActionResult{
		Base:     NewBase(),
		Output:   output,
		Err:      err,
		Metadata: metadata,
		Success:  err == nil,
	}
}

func (ar *ActionResult) GetOutput() interface{}              { return ar.Output }
func (ar *ActionResult) GetError() error                     { return ar.Err }
func (ar *ActionResult) GetMetadata() map[string]interface{} { return ar.Metadata }
func (ar *ActionResult) IsSuccess() bool                     { return ar.Success }

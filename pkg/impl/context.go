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

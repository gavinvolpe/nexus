package types

// Config represents the configuration for a model
type Config struct {
	Provider string
	ModelID  string
	APIKey   string
	Options  map[string]interface{}
}

// Tool represents a tool that can be registered with a model
type Tool struct {
	Name        string
	Description string
	Parameters  map[string]interface{}
}

// Model represents an AI model interface
type Model interface {
	RegisterTool(tool *Tool) error
	StartMCPServer(addr string) error
}

package main

import (
	"log"
	"os"

	"github.com/gavinvolpe/nexus/pkg/impl"
	"github.com/gavinvolpe/nexus/pkg/types"
)

func main() {
	// Create configuration
	config := &types.Config{
		Provider: "groq",
		ModelID:  "mixtral-8x7b-32768",
		APIKey:   os.Getenv("NEXUS_API_KEY"),
	}

	// Initialize model
	model, err := impl.NewModel(config)
	if err != nil {
		log.Fatal(err)
	}

	// Register tools
	model.RegisterTool(&types.Tool{
		Name:        "calculator",
		Description: "Performs basic arithmetic operations",
		Parameters: map[string]interface{}{
			"operation": map[string]string{
				"type":        "string",
				"description": "The operation to perform (add, subtract, multiply, divide)",
			},
			"numbers": map[string]interface{}{
				"type":        "array",
				"items":       map[string]string{"type": "number"},
				"description": "The numbers to operate on",
			},
		},
	})

	// Start MCP server
	log.Println("Starting MCP server on :8080...")
	if err := model.StartMCPServer(":8080"); err != nil {
		log.Fatal(err)
	}
}

# Nexus: Universal Access Pattern Framework

Nexus is a powerful Go framework that provides universal access patterns for connecting and managing diverse resources, from AI models to tools and prompts. It implements the Model Context Protocol (MCP) to enable seamless interaction between different AI models and tools.

## 🌟 Key Features

- **Universal Access Patterns**: Create standardized ways to access any type of resource
- **MCP Integration**: Full support for Anthropic's Model Context Protocol
- **AI Model Support**: Built-in support for various AI models (Groq, Ollama)
- **Tool Management**: Register and manage tools that can be used by AI models
- **Resource Management**: Handle different types of resources with a unified interface
- **Prompt System**: Flexible prompt management with variable substitution

## 🚀 Quick Start

```go
import "github.com/gavinvolpe/nexus"

// Create a new Groq model with MCP support
config := &models.Config{
    Provider: models.Groq,
    ModelID: "mixtral-8x7b-32768",
    APIKey: "your-groq-api-key",
}
model, _ := models.NewGroqModel(config)

// Start MCP server
model.StartMCPServer(":8080")

// Register a tool
model.RegisterMCPTool(mcp.Tool{
    Name: "summarize",
    Description: "Summarizes text",
    Parameters: map[string]any{
        "text": map[string]any{
            "type": "string",
            "description": "Text to summarize",
        },
    },
})
```

## 📦 Installation

```bash
go get github.com/gavinvolpe/nexus
```

## 📚 Documentation

- [Components](COMPONENTS.md): Detailed component documentation
- [Notes](NOTES.md): Development notes and change log
- [API Reference](https://pkg.go.dev/github.com/gavinvolpe/nexus)

## 🔧 Project Structure

```
nexus/
├── internal/
│   └── mcp/           # Model Context Protocol implementation
├── pkg/
│   ├── impl/          # Core implementations
│   └── types/         # Interface definitions
└── prompts/           # Prompt management system
```

## 🤝 Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🙏 Acknowledgments

- Anthropic for the Model Context Protocol specification
- The Go community for their excellent packages and tools
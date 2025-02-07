# Configuration Guide

## Environment Variables

Nexus can be configured using the following environment variables:

- `NEXUS_API_KEY`: Your API key for the model provider
- `NEXUS_MODEL_PROVIDER`: The model provider to use (e.g., "groq", "ollama")
- `NEXUS_MODEL_ID`: The specific model ID to use
- `NEXUS_MCP_PORT`: Port for the MCP server (default: 8080)
- `NEXUS_LOG_LEVEL`: Logging level (default: "info")

## Configuration File

Alternatively, you can use a configuration file (`config.yaml`) in your project root:

```yaml
api:
  key: "your-api-key"
  model:
    provider: "groq"
    id: "mixtral-8x7b-32768"
  mcp:
    port: 8080
    host: "localhost"
logging:
  level: "info"
  format: "json"
```

## Priority Order

Configuration values are loaded in the following order (highest priority first):
1. Environment variables
2. Configuration file
3. Default values

## Examples

### Basic Configuration

```go
import "github.com/gavinvolpe/nexus/pkg/types"

config := &types.Config{
    Provider: "groq",
    ModelID:  "mixtral-8x7b-32768",
    APIKey:   os.Getenv("NEXUS_API_KEY"),
}
```

### Advanced Configuration

```go
import (
    "github.com/gavinvolpe/nexus/pkg/types"
    "github.com/gavinvolpe/nexus/pkg/impl"
)

config := &types.Config{
    Provider: "groq",
    ModelID:  "mixtral-8x7b-32768",
    APIKey:   os.Getenv("NEXUS_API_KEY"),
    Options: map[string]interface{}{
        "temperature": 0.7,
        "maxTokens":   2048,
    },
}

model, err := impl.NewModel(config)
if err != nil {
    log.Fatal(err)
}
```

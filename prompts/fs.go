package prompts

import (
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"strings"
	"sync"

	"github.com/gavinvolpe/nexus/pkg/impl"
)

//go:embed data/*
var promptdata embed.FS

//go:embed templates/*
var prompttemplates embed.FS

var (
	DataDir, _      = fs.Sub(promptdata, "data")
	TemplatesDir, _ = fs.Sub(prompttemplates, "templates")
)

// PromptFS represents a filesystem for managing prompts
type PromptFS struct {
	sync.RWMutex
	cache    map[string]*impl.Prompt
	metadata map[string]PromptMetadata
}

// PromptMetadata stores additional information about prompts
type PromptMetadata struct {
	Topics     []string    `json:"topics"`
	Target     string      `json:"target"`
	Priority   int         `json:"priority"`
	Tags       []string    `json:"tags"`
	Categories []string    `json:"categories"`
	Version    string      `json:"version"`
	Stats      PromptStats `json:"stats"`
}

// PromptStats tracks usage statistics for prompts
type PromptStats struct {
	UsageCount     int     `json:"usage_count"`
	SuccessRate    float64 `json:"success_rate"`
	AverageLatency int64   `json:"average_latency_ms"`
	LastUsed       int64   `json:"last_used_timestamp"`
}

// NewPromptFS creates a new prompt filesystem
func NewPromptFS() *PromptFS {
	return &PromptFS{
		cache:    make(map[string]*impl.Prompt),
		metadata: make(map[string]PromptMetadata),
	}
}

// LoadPrompts loads all prompts from the filesystem
func (pfs *PromptFS) LoadPrompts() error {
	return fs.WalkDir(DataDir, ".", func(filePath string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || !strings.HasSuffix(filePath, ".json") {
			return nil
		}

		prompt, metadata, err := pfs.loadPromptFile(filePath)
		if err != nil {
			return fmt.Errorf("error loading prompt %s: %w", filePath, err)
		}

		pfs.Lock()
		pfs.cache[filePath] = prompt
		pfs.metadata[filePath] = metadata
		pfs.Unlock()

		return nil
	})
}

// loadPromptFile loads a single prompt file and its metadata
func (pfs *PromptFS) loadPromptFile(filePath string) (*impl.Prompt, PromptMetadata, error) {
	file, err := DataDir.Open(filePath)
	if err != nil {
		return nil, PromptMetadata{}, err
	}
	defer file.Close()

	var data struct {
		Prompt   *impl.Prompt   `json:"prompt"`
		Metadata PromptMetadata `json:"metadata"`
	}

	if err := json.NewDecoder(file).Decode(&data); err != nil {
		return nil, PromptMetadata{}, err
	}

	return data.Prompt, data.Metadata, nil
}

// GetPromptsByTopic retrieves prompts by topic
func (pfs *PromptFS) GetPromptsByTopic(topic string) []*impl.Prompt {
	pfs.RLock()
	defer pfs.RUnlock()

	var prompts []*impl.Prompt
	for filePath, prompt := range pfs.cache {
		metadata := pfs.metadata[filePath]
		for _, t := range metadata.Topics {
			if t == topic {
				prompts = append(prompts, prompt)
				break
			}
		}
	}
	return prompts
}

// GetPromptsByTarget retrieves prompts by target agent
func (pfs *PromptFS) GetPromptsByTarget(target string) []*impl.Prompt {
	pfs.RLock()
	defer pfs.RUnlock()

	var prompts []*impl.Prompt
	for filePath, prompt := range pfs.cache {
		if pfs.metadata[filePath].Target == target {
			prompts = append(prompts, prompt)
		}
	}
	return prompts
}

// GetPromptsByTags retrieves prompts that match all given tags
func (pfs *PromptFS) GetPromptsByTags(tags []string) []*impl.Prompt {
	pfs.RLock()
	defer pfs.RUnlock()

	var prompts []*impl.Prompt
	for filePath, prompt := range pfs.cache {
		metadata := pfs.metadata[filePath]
		if containsAll(metadata.Tags, tags) {
			prompts = append(prompts, prompt)
		}
	}
	return prompts
}

// UpdatePromptStats updates usage statistics for a prompt
func (pfs *PromptFS) UpdatePromptStats(filePath string, success bool, latencyMs int64) {
	pfs.Lock()
	defer pfs.Unlock()

	if metadata, exists := pfs.metadata[filePath]; exists {
		metadata.Stats.UsageCount++
		metadata.Stats.LastUsed = latencyMs

		// Update success rate
		oldSuccesses := metadata.Stats.SuccessRate * float64(metadata.Stats.UsageCount-1)
		if success {
			oldSuccesses++
		}
		metadata.Stats.SuccessRate = oldSuccesses / float64(metadata.Stats.UsageCount)

		// Update average latency
		oldTotal := metadata.Stats.AverageLatency * int64(metadata.Stats.UsageCount-1)
		metadata.Stats.AverageLatency = (oldTotal + latencyMs) / int64(metadata.Stats.UsageCount)

		pfs.metadata[filePath] = metadata
	}
}

// Helper function to check if slice contains all elements
func containsAll(slice []string, elements []string) bool {
	for _, e := range elements {
		found := false
		for _, s := range slice {
			if s == e {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
}

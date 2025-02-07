package prompts

import (
	"fmt"
	"io/fs"
	"strings"
	"sync"

	"github.com/gavinvolpe/nexus/pkg/impl"
)

// Global prompt filesystem instance
var (
	globalFS     *PromptFS
	globalFSOnce sync.Once
)

// PromptQuery represents a query for finding prompts
type PromptQuery struct {
	Topics     []string // Topics to match
	Target     string   // Target agent
	Tags       []string // Tags to match
	Categories []string // Categories to match
	MinSuccess float64  // Minimum success rate
	Priority   int      // Minimum priority
	Limit      int      // Maximum number of results
}

// PromptResult represents a prompt with its metadata and score
type PromptResult struct {
	Prompt   *impl.Prompt
	Metadata PromptMetadata
	Score    float64
	FilePath string
}

// Initialize the global filesystem
func initGlobalFS() {
	globalFSOnce.Do(func() {
		globalFS = NewPromptFS()
		if err := globalFS.LoadPrompts(); err != nil {
			panic(fmt.Sprintf("failed to load prompts: %v", err))
		}
	})
}

// GetPromptFS returns the global prompt filesystem
func GetPromptFS() *PromptFS {
	initGlobalFS()
	return globalFS
}

// FastFindPrompts quickly finds prompts matching the query
func FastFindPrompts(query PromptQuery) []PromptResult {
	fs := GetPromptFS()
	fs.RLock()
	defer fs.RUnlock()

	results := make([]PromptResult, 0)
	for filePath, prompt := range fs.cache {
		metadata := fs.metadata[filePath]

		// Quick disqualification checks
		if query.Target != "" && metadata.Target != query.Target {
			continue
		}
		if query.MinSuccess > 0 && metadata.Stats.SuccessRate < query.MinSuccess {
			continue
		}
		if query.Priority > 0 && metadata.Priority < query.Priority {
			continue
		}

		// Calculate match score
		score := calculateMatchScore(query, metadata)
		if score > 0 {
			results = append(results, PromptResult{
				Prompt:   prompt,
				Metadata: metadata,
				Score:    score,
				FilePath: filePath,
			})
		}

		// Apply limit if specified
		if query.Limit > 0 && len(results) >= query.Limit {
			break
		}
	}

	// Sort results by score (highest first)
	sortPromptResults(results)
	return results
}

// Helper function to calculate match score
func calculateMatchScore(query PromptQuery, metadata PromptMetadata) float64 {
	score := 0.0

	// Topic matches
	topicMatches := 0
	for _, qt := range query.Topics {
		for _, mt := range metadata.Topics {
			if strings.EqualFold(qt, mt) {
				topicMatches++
			}
		}
	}
	if len(query.Topics) > 0 && topicMatches == 0 {
		return 0 // No topic matches, disqualify
	}
	score += float64(topicMatches) * 2.0

	// Tag matches
	tagMatches := 0
	for _, qt := range query.Tags {
		for _, mt := range metadata.Tags {
			if strings.EqualFold(qt, mt) {
				tagMatches++
			}
		}
	}
	if len(query.Tags) > 0 && tagMatches == 0 {
		return 0 // No tag matches, disqualify
	}
	score += float64(tagMatches)

	// Category matches
	catMatches := 0
	for _, qc := range query.Categories {
		for _, mc := range metadata.Categories {
			if strings.EqualFold(qc, mc) {
				catMatches++
			}
		}
	}
	if len(query.Categories) > 0 && catMatches == 0 {
		return 0 // No category matches, disqualify
	}
	score += float64(catMatches)

	// Success rate bonus
	score += metadata.Stats.SuccessRate

	// Priority bonus
	score += float64(metadata.Priority) * 0.1

	return score
}

// Helper function to sort prompt results
func sortPromptResults(results []PromptResult) {
	// Implementation of quick sort for PromptResult slice
	if len(results) < 2 {
		return
	}

	left, right := 0, len(results)-1
	pivot := results[len(results)/2].Score

	for left <= right {
		for results[left].Score > pivot {
			left++
		}
		for results[right].Score < pivot {
			right--
		}
		if left <= right {
			results[left], results[right] = results[right], results[left]
			left++
			right--
		}
	}

	if right > 0 {
		sortPromptResults(results[:right+1])
	}
	if left < len(results)-1 {
		sortPromptResults(results[left:])
	}
}

// Convenience functions for common use cases

// QuickFindPrompt finds a single best matching prompt
func QuickFindPrompt(target string, topics []string, tags []string) (*impl.Prompt, error) {
	results := FastFindPrompts(PromptQuery{
		Target:     target,
		Topics:     topics,
		Tags:       tags,
		MinSuccess: 0.0,
		Limit:      1,
	})

	if len(results) == 0 {
		return nil, fmt.Errorf("no matching prompt found")
	}
	return results[0].Prompt, nil
}

// FindSimilarPrompts finds prompts similar to a given one
func FindSimilarPrompts(prompt *impl.Prompt, limit int) []PromptResult {
	fs := GetPromptFS()
	fs.RLock()
	metadata := PromptMetadata{}
	for fp, p := range fs.cache {
		if p == prompt {
			metadata = fs.metadata[fp]
			break
		}
	}
	fs.RUnlock()

	return FastFindPrompts(PromptQuery{
		Target:     metadata.Target,
		Topics:     metadata.Topics,
		Tags:       metadata.Tags,
		Categories: metadata.Categories,
		Limit:      limit,
	})
}

func DataFS() fs.FS {
	return DataDir
}

func TemplatesFS() fs.FS {
	return TemplatesDir
}

func openPromptTemplate(name string) (fs.File, error) {
	return TemplatesDir.Open(name)
}

func openDataFile(name string) (fs.File, error) {
	return DataDir.Open(name)
}

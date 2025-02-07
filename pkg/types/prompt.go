package types

// Prompt represents a single prompt for an AI to generate a perfect response.
type Prompt struct {
	Id      string         `json:"id"`
	Title   string         `json:"title"`
	Prompt  string         `json:"prompt"`
	Vars    map[string]any `json:"vars"`
	Purpose string         `json:"purpose"`
	Target  string         `json:"target"`
	When    string         `json:"when"`
}

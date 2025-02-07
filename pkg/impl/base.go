package impl

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// Base struct for common fields across implementations
type Base struct {
	ID        string                 `json:"id"`
	CreatedAt time.Time              `json:"created_at"`
	UpdatedAt time.Time              `json:"updated_at"`
	Version   string                 `json:"version"`
	Metadata  map[string]interface{} `json:"metadata"`
}

func NewBase() Base {
	now := time.Now()
	return Base{
		ID:        uuid.New().String(),
		CreatedAt: now,
		UpdatedAt: now,
		Version:   "1.0.0",
		Metadata:  make(map[string]interface{}),
	}
}

// Serialization helpers
func (b *Base) MarshalJSON() ([]byte, error) {
	type Alias Base
	return json.Marshal(&struct {
		*Alias
	}{
		Alias: (*Alias)(b),
	})
}

func (b *Base) UnmarshalJSON(data []byte) error {
	type Alias Base
	aux := &struct {
		*Alias
	}{
		Alias: (*Alias)(b),
	}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	return nil
}

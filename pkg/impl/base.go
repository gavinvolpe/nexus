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
	"time"

	"github.com/bytedance/sonic"
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
	return sonic.Marshal(&struct {
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
	if err := sonic.Unmarshal(data, &aux); err != nil {
		return err
	}
	return nil
}

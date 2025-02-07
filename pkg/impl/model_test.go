package impl

import (
	"testing"

	"github.com/gavinvolpe/nexus/pkg/types"
)

func TestNewModel(t *testing.T) {
	tests := []struct {
		name    string
		config  *types.Config
		wantErr bool
	}{
		{
			name: "valid configuration",
			config: &types.Config{
				Provider: "groq",
				ModelID:  "mixtral-8x7b-32768",
				APIKey:   "test-key",
			},
			wantErr: false,
		},
		{
			name: "missing model ID",
			config: &types.Config{
				Provider: "groq",
				APIKey:   "test-key",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewModel(tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewModel() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

package horde

import (
	"os"
	"testing"
)

func TestImageGenerate(t *testing.T) {
	tests := []struct {
		name    string
		prompt  string
		steps   int
		width   int
		height  int
		wantErr bool
	}{
		{
			name:    "Default values",
			prompt:  "Test image",
			steps:   0,
			width:   0,
			height:  0,
			wantErr: false,
		},
		{
			name:    "Custom dimensions",
			prompt:  "Test image",
			steps:   40,
			width:   1024,
			height:  768,
			wantErr: false,
		},
		{
			name:    "Empty prompt",
			prompt:  "",
			steps:   30,
			width:   512,
			height:  512,
			wantErr: true,
		},
	}

	apiKey := os.Getenv("HORDE_API_KEY")
	if apiKey == "" {
		t.Errorf("Set a stable horde API key")
	}

	client := NewClient(apiKey)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			imageData, err := client.ImageGenerate(tt.prompt, tt.steps, tt.width, tt.height, defaultModel)
			if tt.wantErr {
				if err == nil {
					t.Error("ImageGenerate() expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Errorf("ImageGenerate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && len(imageData) == 0 {
				t.Error("ImageGenerate() returned empty image data")
			}
		})
	}
}

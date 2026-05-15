package config

import (
	"os"
	"testing"
)

func TestConfigValidation(t *testing.T) {
	tests := []struct {
		name      string
		config    Config
		expectErr bool
	}{
		{
			name: "valid config",
			config: Config{
				Floors:   2,
				Monsters: 2,
				OpenAt:   "14:05:00",
				Duration: 2,
			},
			expectErr: false,
		},
		{
			name: "invalid floors - zero",
			config: Config{
				Floors:   0,
				Monsters: 2,
				OpenAt:   "14:05:00",
				Duration: 2,
			},
			expectErr: true,
		},
		{
			name: "invalid floors - negative",
			config: Config{
				Floors:   -1,
				Monsters: 2,
				OpenAt:   "14:05:00",
				Duration: 2,
			},
			expectErr: true,
		},
		{
			name: "invalid monsters - zero",
			config: Config{
				Floors:   2,
				Monsters: 0,
				OpenAt:   "14:05:00",
				Duration: 2,
			},
			expectErr: true,
		},
		{
			name: "invalid duration - zero",
			config: Config{
				Floors:   2,
				Monsters: 2,
				OpenAt:   "14:05:00",
				Duration: 0,
			},
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if tt.expectErr && err == nil {
				t.Errorf("expected error but got none")
			}
			if !tt.expectErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestConfigLoadFromFile(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "config*.json")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())

	validConfig := `{
		"Floors": 2,
		"Monsters": 2,
		"OpenAt": "14:05:00",
		"Duration": 2
	}`

	if _, err := tmpFile.Write([]byte(validConfig)); err != nil {
		t.Fatal(err)
	}
	tmpFile.Close()

	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	os.Args = []string{"cmd", "--config", tmpFile.Name()}

	cfg := MustLoad()

	if cfg.Floors != 2 {
		t.Errorf("Floors: got %d, want 2", cfg.Floors)
	}
	if cfg.Monsters != 2 {
		t.Errorf("Monsters: got %d, want 2", cfg.Monsters)
	}
	if cfg.OpenAt != "14:05:00" {
		t.Errorf("OpenAt: got %s, want 14:05:00", cfg.OpenAt)
	}
	if cfg.Duration != 2 {
		t.Errorf("Duration: got %d, want 2", cfg.Duration)
	}
}

func TestConfigLoadFromFileInvalid(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "config*.json")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())

	invalidConfig := `{
		"Floors": "not a number",
		"Monsters": 2,
		"OpenAt": "14:05:00",
		"Duration": 2
	}`

	if _, err := tmpFile.Write([]byte(invalidConfig)); err != nil {
		t.Fatal(err)
	}
	tmpFile.Close()

	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	os.Args = []string{"cmd", "--config", tmpFile.Name()}

	defer func() {
		if r := recover(); r == nil {
			t.Errorf("expected panic for invalid JSON")
		}
	}()

	MustLoad()
}

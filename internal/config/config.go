package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// Config holds user configuration.
type Config struct {
	Theme string `json:"theme,omitempty"`
}

// Load reads ~/.config/markcli/config.json.
// Returns Config{} (zero value) if the file doesn't exist or is invalid JSON.
func Load() Config {
	home, err := os.UserHomeDir()
	if err != nil {
		return Config{}
	}
	return loadFrom(filepath.Join(home, ".config", "markcli", "config.json"))
}

func loadFrom(path string) Config {
	data, err := os.ReadFile(path)
	if err != nil {
		return Config{}
	}
	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return Config{}
	}
	return cfg
}

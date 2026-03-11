package config

import (
	"os"
	"path/filepath"
	"testing"
)

func writeConfigFile(t *testing.T, dir, content string) {
	t.Helper()
	cfgDir := filepath.Join(dir, ".config", "markcli")
	if err := os.MkdirAll(cfgDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(cfgDir, "config.json"), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func withTempHome(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	t.Setenv("HOME", dir)
	return dir
}

func TestLoad_FileNotExist(t *testing.T) {
	withTempHome(t)
	cfg := Load()
	if cfg.Theme != "" {
		t.Errorf("expected empty theme, got %q", cfg.Theme)
	}
}

func TestLoad_ValidJSON(t *testing.T) {
	home := withTempHome(t)
	writeConfigFile(t, home, `{"theme":"tokyonight-moon"}`)
	cfg := Load()
	if cfg.Theme != "tokyonight-moon" {
		t.Errorf("expected tokyonight-moon, got %q", cfg.Theme)
	}
}

func TestLoad_InvalidJSON(t *testing.T) {
	home := withTempHome(t)
	writeConfigFile(t, home, `{not valid json`)
	cfg := Load()
	if cfg.Theme != "" {
		t.Errorf("expected empty theme on bad JSON, got %q", cfg.Theme)
	}
}

func TestLoad_EmptyObject(t *testing.T) {
	home := withTempHome(t)
	writeConfigFile(t, home, `{}`)
	cfg := Load()
	if cfg.Theme != "" {
		t.Errorf("expected empty theme for empty object, got %q", cfg.Theme)
	}
}

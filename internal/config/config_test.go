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
	_ = cfg // Config is now empty struct, just verify no panic
}

func TestLoad_ValidJSON(t *testing.T) {
	home := withTempHome(t)
	writeConfigFile(t, home, `{}`)
	cfg := Load()
	_ = cfg
}

func TestLoad_InvalidJSON(t *testing.T) {
	home := withTempHome(t)
	writeConfigFile(t, home, `{not valid json`)
	cfg := Load()
	_ = cfg // should not panic on invalid JSON
}

func TestLoad_EmptyObject(t *testing.T) {
	home := withTempHome(t)
	writeConfigFile(t, home, `{}`)
	cfg := Load()
	_ = cfg
}

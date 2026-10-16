package server

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadConfig_Defaults(t *testing.T) {
	os.Unsetenv("RAVEN_PORT")
	os.Unsetenv("PORT")

	cfg, err := LoadConfig(filepath.Join(t.TempDir(), "nonexistent.json"))
	if err != nil {
		t.Fatalf("unexpected error for missing config: %v", err)
	}
	if cfg.Port != 9471 {
		t.Errorf("expected default port 9471, got %d", cfg.Port)
	}
	if cfg.DataDir != "data" || cfg.WebDir != "web" {
		t.Errorf("unexpected defaults: data=%s web=%s", cfg.DataDir, cfg.WebDir)
	}
}

func TestLoadConfig_FromFile(t *testing.T) {
	os.Unsetenv("RAVEN_PORT")
	os.Unsetenv("PORT")

	tmpDir := t.TempDir()
	cfgPath := filepath.Join(tmpDir, "config.json")
	content := `{"port": 8080, "data_dir": "custom_data", "web_dir": "custom_web"}`
	if err := os.WriteFile(cfgPath, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write test config: %v", err)
	}

	cfg, err := LoadConfig(cfgPath)
	if err != nil {
		t.Fatalf("failed to load valid config: %v", err)
	}
	if cfg.Port != 8080 {
		t.Errorf("expected port 8080 from file, got %d", cfg.Port)
	}
	if cfg.DataDir != "custom_data" || cfg.WebDir != "custom_web" {
		t.Errorf("unexpected directories: data=%s web=%s", cfg.DataDir, cfg.WebDir)
	}
}

func TestLoadConfig_EnvRavenPort(t *testing.T) {
	t.Setenv("RAVEN_PORT", "9999")
	os.Unsetenv("PORT")

	cfg, err := LoadConfig(filepath.Join(t.TempDir(), "nonexistent.json"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Port != 9999 {
		t.Errorf("expected port 9999 from RAVEN_PORT env, got %d", cfg.Port)
	}
}

func TestLoadConfig_EnvPort(t *testing.T) {
	os.Unsetenv("RAVEN_PORT")
	t.Setenv("PORT", "7777")

	cfg, err := LoadConfig(filepath.Join(t.TempDir(), "nonexistent.json"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Port != 7777 {
		t.Errorf("expected port 7777 from PORT env, got %d", cfg.Port)
	}
}

func TestLoadConfig_EnvPrecedence(t *testing.T) {
	t.Setenv("RAVEN_PORT", "8888")
	t.Setenv("PORT", "7777")

	cfg, err := LoadConfig(filepath.Join(t.TempDir(), "nonexistent.json"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Port != 8888 {
		t.Errorf("expected RAVEN_PORT 8888 to take precedence, got %d", cfg.Port)
	}
}

func TestLoadConfig_InvalidJSON(t *testing.T) {
	tmpDir := t.TempDir()
	cfgPath := filepath.Join(tmpDir, "invalid.json")
	if err := os.WriteFile(cfgPath, []byte("{invalid json"), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	_, err := LoadConfig(cfgPath)
	if err == nil {
		t.Errorf("expected error for invalid JSON, got nil")
	}
}

package internal

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSetFabVersion_NewFile(t *testing.T) {
	root := t.TempDir()
	configPath := filepath.Join(root, "fab", "project", "config.yaml")

	if err := setFabVersion(configPath, "0.42.0"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("config not created: %v", err)
	}

	content := string(data)
	if !strings.Contains(content, "fab_version") || !strings.Contains(content, "0.42.0") {
		t.Errorf("expected fab_version: 0.42.0 in config, got:\n%s", content)
	}
}

func TestSetFabVersion_ExistingFile(t *testing.T) {
	root := t.TempDir()
	configDir := filepath.Join(root, "fab", "project")
	if err := os.MkdirAll(configDir, 0o755); err != nil {
		t.Fatal(err)
	}

	configPath := filepath.Join(configDir, "config.yaml")
	initial := []byte("project:\n    name: myproject\nsource_paths:\n    - src/\n")
	if err := os.WriteFile(configPath, initial, 0o644); err != nil {
		t.Fatal(err)
	}

	if err := setFabVersion(configPath, "0.42.0"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatal(err)
	}

	content := string(data)
	if !strings.Contains(content, "fab_version") || !strings.Contains(content, "0.42.0") {
		t.Errorf("expected fab_version in config, got:\n%s", content)
	}
	if !strings.Contains(content, "myproject") {
		t.Errorf("expected existing project name preserved, got:\n%s", content)
	}
}

func TestParseConfigFile_Valid(t *testing.T) {
	root := t.TempDir()
	configPath := filepath.Join(root, "config.yaml")
	if err := os.WriteFile(configPath, []byte("fab_version: \"1.0.0\"\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	cfg, err := parseConfigFile(configPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.FabVersion != "1.0.0" {
		t.Errorf("expected 1.0.0, got %s", cfg.FabVersion)
	}
}

func TestParseConfigFile_Missing(t *testing.T) {
	_, err := parseConfigFile("/nonexistent/config.yaml")
	if err == nil {
		t.Error("expected error for missing file")
	}
}

func TestParseConfigFile_NoVersion(t *testing.T) {
	root := t.TempDir()
	configPath := filepath.Join(root, "config.yaml")
	if err := os.WriteFile(configPath, []byte("project:\n  name: test\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	cfg, err := parseConfigFile(configPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.FabVersion != "" {
		t.Errorf("expected empty fab_version, got %s", cfg.FabVersion)
	}
}


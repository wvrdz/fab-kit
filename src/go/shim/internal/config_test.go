package internal

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDiscoverConfig_Found(t *testing.T) {
	// Create a temp directory tree: root/fab/project/config.yaml
	root := t.TempDir()
	configDir := filepath.Join(root, "fab", "project")
	if err := os.MkdirAll(configDir, 0o755); err != nil {
		t.Fatal(err)
	}
	configContent := []byte("project:\n  name: test\nfab_version: \"1.2.3\"\n")
	if err := os.WriteFile(filepath.Join(configDir, "config.yaml"), configContent, 0o644); err != nil {
		t.Fatal(err)
	}

	// Discover from a subdirectory.
	subdir := filepath.Join(root, "src", "deep")
	if err := os.MkdirAll(subdir, 0o755); err != nil {
		t.Fatal(err)
	}

	cfg, err := DiscoverConfig(subdir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.FabVersion != "1.2.3" {
		t.Errorf("expected fab_version 1.2.3, got %s", cfg.FabVersion)
	}
	if cfg.RepoRoot != root {
		t.Errorf("expected repo root %s, got %s", root, cfg.RepoRoot)
	}
}

func TestDiscoverConfig_FoundAtRoot(t *testing.T) {
	root := t.TempDir()
	configDir := filepath.Join(root, "fab", "project")
	if err := os.MkdirAll(configDir, 0o755); err != nil {
		t.Fatal(err)
	}
	configContent := []byte("fab_version: \"0.5.0\"\n")
	if err := os.WriteFile(filepath.Join(configDir, "config.yaml"), configContent, 0o644); err != nil {
		t.Fatal(err)
	}

	cfg, err := DiscoverConfig(root)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.FabVersion != "0.5.0" {
		t.Errorf("expected fab_version 0.5.0, got %s", cfg.FabVersion)
	}
}

func TestDiscoverConfig_NotFound(t *testing.T) {
	root := t.TempDir()
	_, err := DiscoverConfig(root)
	if err != ErrNoConfig {
		t.Errorf("expected ErrNoConfig, got %v", err)
	}
}

func TestDiscoverConfig_NoFabVersion(t *testing.T) {
	root := t.TempDir()
	configDir := filepath.Join(root, "fab", "project")
	if err := os.MkdirAll(configDir, 0o755); err != nil {
		t.Fatal(err)
	}
	configContent := []byte("project:\n  name: test\n")
	if err := os.WriteFile(filepath.Join(configDir, "config.yaml"), configContent, 0o644); err != nil {
		t.Fatal(err)
	}

	cfg, err := DiscoverConfig(root)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.FabVersion != "" {
		t.Errorf("expected empty fab_version, got %s", cfg.FabVersion)
	}
}

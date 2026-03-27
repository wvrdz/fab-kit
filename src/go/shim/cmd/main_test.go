package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestRun_Version(t *testing.T) {
	// --version should work without any config.
	if err := run([]string{"--version"}); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestRun_Help(t *testing.T) {
	if err := run([]string{"--help"}); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestRun_HelpNoArgs(t *testing.T) {
	if err := run(nil); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestRun_NoConfig(t *testing.T) {
	// Run from a temp dir with no config — should error for non-init commands.
	tmpDir := t.TempDir()
	origDir, _ := os.Getwd()
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(origDir)

	err := run([]string{"status"})
	if err == nil {
		t.Error("expected error for missing config")
	}
}

func TestRun_NoFabVersion(t *testing.T) {
	// Config exists but no fab_version.
	tmpDir := t.TempDir()
	configDir := filepath.Join(tmpDir, "fab", "project")
	if err := os.MkdirAll(configDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(configDir, "config.yaml"), []byte("project:\n  name: test\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	origDir, _ := os.Getwd()
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(origDir)

	err := run([]string{"status"})
	if err == nil {
		t.Error("expected error for missing fab_version")
	}
}

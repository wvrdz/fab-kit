package internal

import (
	"os"
	"path/filepath"
	"testing"
)

func TestIsCached_Hit(t *testing.T) {
	// Override home for test isolation.
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	version := "1.0.0"
	runtimeDir := filepath.Join(tmpHome, cacheDir, versionsSubdir, version, "fab", ".kit", "bin")
	if err := os.MkdirAll(runtimeDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(runtimeDir, "fab"), []byte("#!/bin/sh\n"), 0o755); err != nil {
		t.Fatal(err)
	}

	cached, err := IsCached(version)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !cached {
		t.Error("expected cached=true, got false")
	}
}

func TestIsCached_Miss(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	cached, err := IsCached("9.9.9")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cached {
		t.Error("expected cached=false, got true")
	}
}

func TestVersionDir(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	vdir, err := VersionDir("2.0.0")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := filepath.Join(tmpHome, cacheDir, versionsSubdir, "2.0.0")
	if vdir != expected {
		t.Errorf("expected %s, got %s", expected, vdir)
	}
}

func TestCacheDir(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	dir, err := CacheDir()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := filepath.Join(tmpHome, cacheDir, versionsSubdir)
	if dir != expected {
		t.Errorf("expected %s, got %s", expected, dir)
	}
}

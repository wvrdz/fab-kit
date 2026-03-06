package runtime

import (
	"os"
	"path/filepath"
	"testing"

	"gopkg.in/yaml.v3"
)

func setupTestEnv(t *testing.T) (string, string) {
	t.Helper()
	dir := t.TempDir()
	fabRoot := filepath.Join(dir, "fab")
	os.MkdirAll(fabRoot, 0o755)
	// Create a changes dir with a dummy change for resolve context
	changesDir := filepath.Join(fabRoot, "changes", "260306-test-change")
	os.MkdirAll(changesDir, 0o755)
	return fabRoot, dir
}

func readRuntime(t *testing.T, dir string) map[string]*changeEntry {
	t.Helper()
	path := filepath.Join(dir, ".fab-runtime.yaml")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read runtime file: %v", err)
	}
	var entries map[string]*changeEntry
	if err := yaml.Unmarshal(data, &entries); err != nil {
		t.Fatalf("failed to parse runtime file: %v", err)
	}
	return entries
}

func TestSetIdle_CreatesFile(t *testing.T) {
	fabRoot, dir := setupTestEnv(t)

	err := SetIdle(fabRoot, "260306-test-change")
	if err != nil {
		t.Fatalf("SetIdle failed: %v", err)
	}

	entries := readRuntime(t, dir)
	entry, ok := entries["260306-test-change"]
	if !ok {
		t.Fatal("expected change entry to exist")
	}
	if entry.Agent == nil {
		t.Fatal("expected agent block to exist")
	}
	if entry.Agent.IdleSince <= 0 {
		t.Errorf("expected positive timestamp, got %d", entry.Agent.IdleSince)
	}
}

func TestSetIdle_UpdatesExisting(t *testing.T) {
	fabRoot, dir := setupTestEnv(t)

	// Set idle twice
	SetIdle(fabRoot, "260306-test-change")
	SetIdle(fabRoot, "260306-test-change")

	entries := readRuntime(t, dir)
	if len(entries) != 1 {
		t.Errorf("expected 1 entry, got %d", len(entries))
	}
}

func TestSetIdle_PreservesOtherEntries(t *testing.T) {
	fabRoot, dir := setupTestEnv(t)

	SetIdle(fabRoot, "260306-test-change")
	SetIdle(fabRoot, "260306-other-change")

	entries := readRuntime(t, dir)
	if len(entries) != 2 {
		t.Errorf("expected 2 entries, got %d", len(entries))
	}
	if _, ok := entries["260306-test-change"]; !ok {
		t.Error("expected first change entry")
	}
	if _, ok := entries["260306-other-change"]; !ok {
		t.Error("expected second change entry")
	}
}

func TestClearIdle_RemovesEntry(t *testing.T) {
	fabRoot, dir := setupTestEnv(t)

	SetIdle(fabRoot, "260306-test-change")
	err := ClearIdle(fabRoot, "260306-test-change")
	if err != nil {
		t.Fatalf("ClearIdle failed: %v", err)
	}

	entries := readRuntime(t, dir)
	if len(entries) != 0 {
		t.Errorf("expected 0 entries after clear, got %d", len(entries))
	}
}

func TestClearIdle_PreservesOtherEntries(t *testing.T) {
	fabRoot, dir := setupTestEnv(t)

	SetIdle(fabRoot, "260306-test-change")
	SetIdle(fabRoot, "260306-other-change")
	ClearIdle(fabRoot, "260306-test-change")

	entries := readRuntime(t, dir)
	if len(entries) != 1 {
		t.Errorf("expected 1 entry after clear, got %d", len(entries))
	}
	if _, ok := entries["260306-other-change"]; !ok {
		t.Error("expected other change entry to remain")
	}
}

func TestClearIdle_NoEntry(t *testing.T) {
	fabRoot, _ := setupTestEnv(t)

	// Set one change, clear a different one
	SetIdle(fabRoot, "260306-test-change")
	err := ClearIdle(fabRoot, "260306-nonexistent")
	if err != nil {
		t.Fatalf("ClearIdle should be no-op for missing entry: %v", err)
	}
}

func TestClearIdle_MissingFile(t *testing.T) {
	fabRoot, _ := setupTestEnv(t)

	err := ClearIdle(fabRoot, "260306-test-change")
	if err != nil {
		t.Fatalf("ClearIdle should be no-op for missing file: %v", err)
	}
}

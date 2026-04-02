package internal

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSyncHooks_CreateNew(t *testing.T) {
	hooksDir := t.TempDir()
	settingsDir := t.TempDir()
	settingsPath := filepath.Join(settingsDir, "settings.local.json")

	// Create hook scripts
	os.WriteFile(filepath.Join(hooksDir, "on-session-start.sh"), []byte("#!/bin/bash\n"), 0644)
	os.WriteFile(filepath.Join(hooksDir, "on-stop.sh"), []byte("#!/bin/bash\n"), 0644)

	msg, err := syncHooks(hooksDir, settingsPath)
	if err != nil {
		t.Fatalf("syncHooks failed: %v", err)
	}

	if !strings.Contains(msg, "Created:") {
		t.Errorf("expected Created message, got: %s", msg)
	}

	// Verify settings file has hooks
	data, _ := os.ReadFile(settingsPath)
	var settings map[string]json.RawMessage
	json.Unmarshal(data, &settings)

	var hooks map[string][]hookEntry
	json.Unmarshal(settings["hooks"], &hooks)

	if len(hooks["SessionStart"]) != 1 {
		t.Errorf("expected 1 SessionStart hook, got %d", len(hooks["SessionStart"]))
	}
	if len(hooks["Stop"]) != 1 {
		t.Errorf("expected 1 Stop hook, got %d", len(hooks["Stop"]))
	}
}

func TestSyncHooks_Idempotent(t *testing.T) {
	hooksDir := t.TempDir()
	settingsDir := t.TempDir()
	settingsPath := filepath.Join(settingsDir, "settings.local.json")

	os.WriteFile(filepath.Join(hooksDir, "on-session-start.sh"), []byte("#!/bin/bash\n"), 0644)

	// Run twice
	syncHooks(hooksDir, settingsPath)
	msg, err := syncHooks(hooksDir, settingsPath)
	if err != nil {
		t.Fatalf("second syncHooks failed: %v", err)
	}

	if !strings.Contains(msg, "OK") {
		t.Errorf("expected OK on second run, got: %s", msg)
	}

	// Verify no duplicates
	data, _ := os.ReadFile(settingsPath)
	var settings map[string]json.RawMessage
	json.Unmarshal(data, &settings)

	var hooks map[string][]hookEntry
	json.Unmarshal(settings["hooks"], &hooks)

	if len(hooks["SessionStart"]) != 1 {
		t.Errorf("expected 1 SessionStart hook (no duplicates), got %d", len(hooks["SessionStart"]))
	}
}

func TestSyncHooks_PathMigration(t *testing.T) {
	hooksDir := t.TempDir()
	settingsDir := t.TempDir()
	settingsPath := filepath.Join(settingsDir, "settings.local.json")

	os.WriteFile(filepath.Join(hooksDir, "on-session-start.sh"), []byte("#!/bin/bash\n"), 0644)

	// Create settings with old-format path
	oldSettings := map[string]interface{}{
		"hooks": map[string]interface{}{
			"SessionStart": []interface{}{
				map[string]interface{}{
					"matcher": "",
					"hooks": []interface{}{
						map[string]interface{}{
							"type":    "command",
							"command": "bash fab/.kit/hooks/on-session-start.sh",
						},
					},
				},
			},
		},
	}
	data, _ := json.MarshalIndent(oldSettings, "", "  ")
	os.WriteFile(settingsPath, data, 0644)

	msg, err := syncHooks(hooksDir, settingsPath)
	if err != nil {
		t.Fatalf("syncHooks failed: %v", err)
	}

	if !strings.Contains(msg, "migrated") {
		t.Errorf("expected migration message, got: %s", msg)
	}

	// Verify path was migrated
	data, _ = os.ReadFile(settingsPath)
	content := string(data)
	if strings.Contains(content, `"bash fab/.kit/hooks/`) {
		t.Error("old-format path should have been migrated")
	}
	if !strings.Contains(content, `$CLAUDE_PROJECT_DIR`) {
		t.Error("expected $CLAUDE_PROJECT_DIR in migrated path")
	}
}

func TestSyncHooks_MissingHooksDir(t *testing.T) {
	settingsDir := t.TempDir()
	settingsPath := filepath.Join(settingsDir, "settings.local.json")

	// Non-existent hooks dir — should still work (no hooks to register)
	msg, err := syncHooks("/nonexistent/hooks", settingsPath)
	if err != nil {
		t.Fatalf("syncHooks should handle missing dir gracefully, got: %v", err)
	}

	if !strings.Contains(msg, "OK") {
		t.Errorf("expected OK with no hooks to register, got: %s", msg)
	}
}

func TestSyncHooks_ArtifactWriteDoubleMapping(t *testing.T) {
	hooksDir := t.TempDir()
	settingsDir := t.TempDir()
	settingsPath := filepath.Join(settingsDir, "settings.local.json")

	// Create the artifact-write hook script
	os.WriteFile(filepath.Join(hooksDir, "on-artifact-write.sh"), []byte("#!/bin/bash\n"), 0644)

	msg, err := syncHooks(hooksDir, settingsPath)
	if err != nil {
		t.Fatalf("syncHooks failed: %v", err)
	}

	if !strings.Contains(msg, "Created:") {
		t.Errorf("expected Created message, got: %s", msg)
	}

	// Verify both Write and Edit matchers exist
	data, _ := os.ReadFile(settingsPath)
	var settings map[string]json.RawMessage
	json.Unmarshal(data, &settings)

	var hooks map[string][]hookEntry
	json.Unmarshal(settings["hooks"], &hooks)

	if len(hooks["PostToolUse"]) != 2 {
		t.Errorf("expected 2 PostToolUse hooks (Write + Edit), got %d", len(hooks["PostToolUse"]))
	}
}

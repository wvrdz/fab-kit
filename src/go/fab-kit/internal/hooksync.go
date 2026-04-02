package internal

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// hookMapping defines a mapping from a hook script to a Claude Code event.
type hookMapping struct {
	Script  string
	Event   string
	Matcher string
}

// defaultHookMappings maps hook scripts to Claude Code events.
var defaultHookMappings = []hookMapping{
	{Script: "on-session-start.sh", Event: "SessionStart", Matcher: ""},
	{Script: "on-stop.sh", Event: "Stop", Matcher: ""},
	{Script: "on-user-prompt.sh", Event: "UserPromptSubmit", Matcher: ""},
	{Script: "on-artifact-write.sh", Event: "PostToolUse", Matcher: "Write"},
	{Script: "on-artifact-write.sh", Event: "PostToolUse", Matcher: "Edit"},
}

type hookEntry struct {
	Matcher string     `json:"matcher"`
	Hooks   []hookSpec `json:"hooks"`
}

type hookSpec struct {
	Type    string `json:"type"`
	Command string `json:"command"`
}

// syncHooks discovers hook scripts in hooksDir, maps them to Claude Code events,
// and merges entries into settingsPath. Idempotent.
func syncHooks(hooksDir, settingsPath string) (string, error) {
	// Discover existing hook scripts
	existingScripts := make(map[string]bool)
	entries, err := os.ReadDir(hooksDir)
	if err != nil && !os.IsNotExist(err) {
		return "", fmt.Errorf("reading hooks dir: %w", err)
	}
	for _, e := range entries {
		if !e.IsDir() {
			existingScripts[e.Name()] = true
		}
	}

	// Build desired hook entries from mappings (only for scripts that exist)
	type desiredEntry struct {
		event   string
		matcher string
		command string
	}
	var desired []desiredEntry
	for _, m := range defaultHookMappings {
		if !existingScripts[m.Script] {
			continue
		}
		cmd := `bash "$CLAUDE_PROJECT_DIR"/fab/.kit/hooks/` + m.Script
		desired = append(desired, desiredEntry{event: m.Event, matcher: m.Matcher, command: cmd})
	}

	// Ensure settings directory exists
	settingsDir := filepath.Dir(settingsPath)
	if err := os.MkdirAll(settingsDir, 0o755); err != nil {
		return "", fmt.Errorf("creating settings dir: %w", err)
	}

	// Load or initialize settings
	settings := make(map[string]json.RawMessage)
	if data, err := os.ReadFile(settingsPath); err == nil {
		trimmed := bytes.TrimSpace(data)
		if len(trimmed) > 0 {
			if err := json.Unmarshal(trimmed, &settings); err != nil {
				return "", fmt.Errorf("parsing settings: %w", err)
			}
		}
	}

	// Parse existing hooks section
	existingHooks := make(map[string][]hookEntry)
	if raw, ok := settings["hooks"]; ok {
		if err := json.Unmarshal(raw, &existingHooks); err != nil {
			existingHooks = make(map[string][]hookEntry)
		}
	}

	// Migrate old-format commands (relative path) to new format ($CLAUDE_PROJECT_DIR)
	migrated := 0
	for event, eventEntries := range existingHooks {
		for i, entry := range eventEntries {
			for j, h := range entry.Hooks {
				if strings.HasPrefix(h.Command, "bash fab/.kit/hooks/") {
					existingHooks[event][i].Hooks[j].Command = strings.Replace(
						h.Command,
						"bash fab/.kit/hooks/",
						`bash "$CLAUDE_PROJECT_DIR"/fab/.kit/hooks/`,
						1,
					)
					migrated++
				}
			}
		}
	}

	// Count existing entries for change detection
	existingCount := 0
	for _, entries := range existingHooks {
		existingCount += len(entries)
	}

	// Merge desired entries (deduplicate by matcher + command pair)
	added := 0
	for _, d := range desired {
		eventEntries := existingHooks[d.event]
		if !hookHasDuplicate(eventEntries, d.matcher, d.command) {
			eventEntries = append(eventEntries, hookEntry{
				Matcher: d.matcher,
				Hooks: []hookSpec{
					{Type: "command", Command: d.command},
				},
			})
			existingHooks[d.event] = eventEntries
			added++
		}
	}

	// Serialize hooks back into settings
	hooksJSON, err := json.Marshal(existingHooks)
	if err != nil {
		return "", fmt.Errorf("marshaling hooks: %w", err)
	}
	settings["hooks"] = hooksJSON

	// Write settings file
	data, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		return "", fmt.Errorf("marshaling settings: %w", err)
	}
	data = append(data, '\n')

	if err := os.WriteFile(settingsPath, data, 0o644); err != nil {
		return "", fmt.Errorf("writing settings: %w", err)
	}

	// Determine result message
	newCount := 0
	for _, entries := range existingHooks {
		newCount += len(entries)
	}

	if added == 0 && migrated == 0 {
		return ".claude/settings.local.json hooks: OK", nil
	}

	if existingCount == 0 {
		return fmt.Sprintf("Created: .claude/settings.local.json hooks (%d hook entries)", newCount), nil
	}

	var parts []string
	if added > 0 {
		parts = append(parts, fmt.Sprintf("added %d hook entries", added))
	}
	if migrated > 0 {
		parts = append(parts, fmt.Sprintf("migrated %d to absolute paths", migrated))
	}
	return fmt.Sprintf("Updated: .claude/settings.local.json hooks (%s)", strings.Join(parts, ", ")), nil
}

// hookHasDuplicate checks if an entry with the same matcher and command already exists.
func hookHasDuplicate(entries []hookEntry, matcher, command string) bool {
	for _, e := range entries {
		if e.Matcher == matcher {
			for _, h := range e.Hooks {
				if h.Command == command {
					return true
				}
			}
		}
	}
	return false
}

package runtime

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"gopkg.in/yaml.v3"
)

// agentBlock holds the agent runtime state for a change.
type agentBlock struct {
	IdleSince int64 `yaml:"idle_since"`
}

// changeEntry holds all runtime state for a change.
type changeEntry struct {
	Agent *agentBlock `yaml:"agent,omitempty"`
}

// runtimeFilePath returns the absolute path to .fab-runtime.yaml.
func runtimeFilePath(fabRoot string) string {
	repoRoot := filepath.Dir(fabRoot)
	return filepath.Join(repoRoot, ".fab-runtime.yaml")
}

// loadFile reads and parses .fab-runtime.yaml. Returns empty map if file doesn't exist.
func loadFile(path string) (map[string]*changeEntry, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return make(map[string]*changeEntry), nil
		}
		return nil, fmt.Errorf("reading runtime file: %w", err)
	}

	var entries map[string]*changeEntry
	if err := yaml.Unmarshal(data, &entries); err != nil {
		return nil, fmt.Errorf("parsing runtime file: %w", err)
	}
	if entries == nil {
		entries = make(map[string]*changeEntry)
	}
	return entries, nil
}

// saveFile writes the runtime entries to .fab-runtime.yaml.
func saveFile(path string, entries map[string]*changeEntry) error {
	data, err := yaml.Marshal(entries)
	if err != nil {
		return fmt.Errorf("marshaling runtime file: %w", err)
	}
	return os.WriteFile(path, data, 0o644)
}

// SetIdle writes agent.idle_since timestamp for the given change folder.
func SetIdle(fabRoot, folder string) error {
	path := runtimeFilePath(fabRoot)
	entries, err := loadFile(path)
	if err != nil {
		return err
	}

	entry := entries[folder]
	if entry == nil {
		entry = &changeEntry{}
		entries[folder] = entry
	}
	entry.Agent = &agentBlock{
		IdleSince: time.Now().Unix(),
	}

	return saveFile(path, entries)
}

// ClearIdle removes the agent block for the given change folder.
func ClearIdle(fabRoot, folder string) error {
	path := runtimeFilePath(fabRoot)

	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil // no-op if file doesn't exist
	}

	entries, err := loadFile(path)
	if err != nil {
		return err
	}

	entry := entries[folder]
	if entry == nil {
		return nil // no-op if change not in file
	}

	// Remove the entire entry — currently Agent is the only field
	delete(entries, folder)

	return saveFile(path, entries)
}

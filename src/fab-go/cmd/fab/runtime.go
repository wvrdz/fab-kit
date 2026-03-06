package main

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"
	"github.com/wvrdz/fab-kit/src/fab-go/internal/resolve"
	"gopkg.in/yaml.v3"
)

func runtimeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "runtime",
		Short: "Manage runtime state (.fab-runtime.yaml)",
	}

	cmd.AddCommand(
		runtimeSetIdleCmd(),
		runtimeClearIdleCmd(),
	)

	return cmd
}

// runtimeFilePath returns the absolute path to .fab-runtime.yaml at the repo root.
func runtimeFilePath(fabRoot string) string {
	repoRoot := filepath.Dir(fabRoot)
	return filepath.Join(repoRoot, ".fab-runtime.yaml")
}

// loadRuntimeFile reads and parses .fab-runtime.yaml, returning an empty map if the file doesn't exist.
func loadRuntimeFile(path string) (map[string]interface{}, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return make(map[string]interface{}), nil
		}
		return nil, err
	}
	var m map[string]interface{}
	if err := yaml.Unmarshal(data, &m); err != nil {
		return nil, fmt.Errorf("parsing %s: %w", path, err)
	}
	if m == nil {
		m = make(map[string]interface{})
	}
	return m, nil
}

// saveRuntimeFile marshals the map and writes it to the given path.
func saveRuntimeFile(path string, m map[string]interface{}) error {
	data, err := yaml.Marshal(m)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}

func runtimeSetIdleCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "set-idle <change>",
		Short: "Record agent idle timestamp for a change",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			fabRoot, err := resolve.FabRoot()
			if err != nil {
				return err
			}

			folder, err := resolve.ToFolder(fabRoot, args[0])
			if err != nil {
				return err
			}

			rtPath := runtimeFilePath(fabRoot)
			m, err := loadRuntimeFile(rtPath)
			if err != nil {
				return err
			}

			// Ensure the folder entry exists as a map
			folderEntry, ok := m[folder].(map[string]interface{})
			if !ok {
				folderEntry = make(map[string]interface{})
			}

			// Set agent.idle_since
			folderEntry["agent"] = map[string]interface{}{
				"idle_since": time.Now().Unix(),
			}
			m[folder] = folderEntry

			return saveRuntimeFile(rtPath, m)
		},
	}
}

func runtimeClearIdleCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "clear-idle <change>",
		Short: "Clear agent idle state for a change",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			fabRoot, err := resolve.FabRoot()
			if err != nil {
				return err
			}

			folder, err := resolve.ToFolder(fabRoot, args[0])
			if err != nil {
				return err
			}

			rtPath := runtimeFilePath(fabRoot)
			if _, err := os.Stat(rtPath); os.IsNotExist(err) {
				return nil // no-op if file doesn't exist
			}

			m, err := loadRuntimeFile(rtPath)
			if err != nil {
				return err
			}

			// Delete the agent block for this folder
			if folderEntry, ok := m[folder].(map[string]interface{}); ok {
				delete(folderEntry, "agent")
				if len(folderEntry) == 0 {
					delete(m, folder)
				} else {
					m[folder] = folderEntry
				}
			}

			return saveRuntimeFile(rtPath, m)
		},
	}
}

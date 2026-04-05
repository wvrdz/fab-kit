package main

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

// operatorRepoRootOverride is used in tests to redirect .fab-operator.yaml I/O
// to a temp directory instead of the real git repo root.
var operatorRepoRootOverride string

func operatorTickStartCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "tick-start",
		Short: "Increment tick_count and record last_tick_at in .fab-operator.yaml",
		RunE:  runOperatorTickStart,
	}
}

func runOperatorTickStart(cmd *cobra.Command, args []string) error {
	var repoRoot string
	if operatorRepoRootOverride != "" {
		repoRoot = operatorRepoRootOverride
	} else {
		var err error
		repoRoot, err = gitRepoRoot()
		if err != nil {
			return fmt.Errorf("cannot determine repo root: %w", err)
		}
	}

	yamlPath := filepath.Join(repoRoot, ".fab-operator.yaml")

	// Read existing file, or start with empty map if missing
	data := make(map[string]interface{})
	raw, err := os.ReadFile(yamlPath)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("cannot read %s: %w", yamlPath, err)
	}
	if err == nil && len(raw) > 0 {
		if parseErr := yaml.Unmarshal(raw, &data); parseErr != nil {
			return fmt.Errorf("cannot parse %s: %w", yamlPath, parseErr)
		}
	}

	// Increment tick_count
	tickCount := 0
	if v, ok := data["tick_count"]; ok {
		switch n := v.(type) {
		case int:
			tickCount = n
		case float64:
			tickCount = int(n)
		}
	}
	tickCount++

	data["tick_count"] = tickCount
	data["last_tick_at"] = time.Now().UTC().Format(time.RFC3339)

	// Write back
	out, err := yaml.Marshal(data)
	if err != nil {
		fmt.Fprintf(cmd.ErrOrStderr(), "Error: cannot marshal YAML: %v\n", err)
		os.Exit(1)
	}
	if err := os.WriteFile(yamlPath, out, 0644); err != nil {
		fmt.Fprintf(cmd.ErrOrStderr(), "Error: cannot write %s: %v\n", yamlPath, err)
		os.Exit(1)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "tick: %d\nnow: %s\n", tickCount, time.Now().Format("15:04"))
	return nil
}

package internal

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

const configRelPath = "fab/project/config.yaml"

// Config holds the parsed fab project configuration.
type Config struct {
	FabVersion string `yaml:"fab_version"`
	// RepoRoot is the directory containing fab/project/config.yaml.
	RepoRoot string `yaml:"-"`
}

// ErrNoConfig indicates no fab/project/config.yaml was found in any ancestor.
var ErrNoConfig = fmt.Errorf("not in a fab-managed repo")

// DiscoverConfig walks up from startDir to find fab/project/config.yaml.
// Returns the parsed config or ErrNoConfig if not found.
func DiscoverConfig(startDir string) (*Config, error) {
	dir, err := filepath.Abs(startDir)
	if err != nil {
		return nil, fmt.Errorf("resolving start directory: %w", err)
	}

	for {
		configPath := filepath.Join(dir, configRelPath)
		if _, err := os.Stat(configPath); err == nil {
			return parseConfig(configPath, dir)
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			return nil, ErrNoConfig
		}
		dir = parent
	}
}

func parseConfig(path, repoRoot string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading config: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing config: %w", err)
	}

	cfg.RepoRoot = repoRoot
	return &cfg, nil
}

package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	wt "github.com/sahil87/fab-kit/src/go/wt/internal/worktree"
)

func initCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "init",
		Short: "Run worktree init script",
		Long: `Run the worktree init script for the current repository.
If the init script doesn't exist, exits with guidance.`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := wt.ValidateGitRepo(); err != nil {
				wt.ExitWithError(wt.ExitGitError,
					"Not a git repository",
					"This command requires a git repository",
					"Navigate to a git repository and try again")
			}

			return runInitScript()
		},
	}

	return cmd
}

func runInitScript() error {
	// Resolve main repo root using git-common-dir
	out, err := exec.Command("git", "rev-parse", "--git-common-dir").Output()
	if err != nil {
		wt.ExitWithError(wt.ExitGitError, "Cannot determine git common dir", err.Error(), "")
	}

	gitCommonDir := strings.TrimSpace(string(out))
	absPath, err := filepath.Abs(gitCommonDir)
	if err != nil {
		wt.ExitWithError(wt.ExitGeneralError, "Cannot resolve path", err.Error(), "")
	}
	resolved, err := filepath.EvalSymlinks(absPath)
	if err != nil {
		resolved = absPath
	}

	repoRoot := strings.TrimSuffix(resolved, string(filepath.Separator)+".git")

	// Get current toplevel (worktree or main repo)
	topOut, err := exec.Command("git", "rev-parse", "--show-toplevel").Output()
	if err != nil {
		wt.ExitWithError(wt.ExitGitError, "Cannot determine toplevel", err.Error(), "")
	}
	currentRoot := strings.TrimSpace(string(topOut))

	initScriptRel := wt.InitScriptPath()

	fmt.Println("Running worktree init...")
	fmt.Println()

	// Determine if it's a command (contains spaces) or a file path
	var cmd *exec.Cmd
	if strings.Contains(initScriptRel, " ") {
		// Command invocation: check if the command is available
		parts := strings.Fields(initScriptRel)
		if _, err := exec.LookPath(parts[0]); err != nil {
			fmt.Printf("Warning: %q not found on PATH, skipping init\n", parts[0])
			fmt.Println("Install fab-kit or set WORKTREE_INIT_SCRIPT to a custom script.")
			return nil
		}
		cmd = exec.Command(parts[0], parts[1:]...)
	} else {
		// File path: resolve relative to repo root
		initScript := filepath.Join(repoRoot, initScriptRel)
		if _, err := os.Stat(initScript); os.IsNotExist(err) {
			fmt.Printf("No init script found at: %s\n", initScript)
			fmt.Println()
			fmt.Println("To add an init script:")
			fmt.Printf("  mkdir -p %s\n", filepath.Dir(initScriptRel))
			fmt.Printf("  touch %s\n", initScriptRel)
			return nil
		}
		cmd = exec.Command("bash", initScript)
	}

	cmd.Dir = currentRoot
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("init script failed: %w", err)
	}

	fmt.Println()
	fmt.Println("Worktree init complete.")
	return nil
}

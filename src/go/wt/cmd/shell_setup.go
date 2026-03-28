package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	wt "github.com/wvrdz/fab-kit/src/go/wt/internal/worktree"
)

const shellWrapper = `wt() {
  local line last="" rc
  while IFS= read -r line; do
    printf '%s\n' "$line"
    last=$line
  done < <(WT_WRAPPER=1 command wt "$@")
  rc=$?
  if [[ "$last" == cd\ * ]]; then
    eval "$last"
  fi
  return $rc
}
`

func shellSetupCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "shell-setup",
		Short: "Output shell wrapper function for eval in your shell profile",
		RunE: func(cmd *cobra.Command, args []string) error {
			shell := filepath.Base(os.Getenv("SHELL"))
			switch shell {
			case "bash", "zsh":
				_, err := os.Stdout.WriteString(shellWrapper)
				return err
			default:
				fmt.Fprintf(os.Stderr, "wt shell-setup: unsupported shell %q (supported: bash, zsh)\n", shell)
				os.Exit(wt.ExitGeneralError)
				return nil
			}
		},
	}
}

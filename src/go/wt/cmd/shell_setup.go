package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	wt "github.com/wvrdz/fab-kit/src/go/wt/internal/worktree"
)

const shellWrapper = `wt() {
  local line last="" rc=""
  while IFS= read -r line; do
    if [[ "$line" == __WT_RC__* ]]; then
      rc=${line#__WT_RC__}
      continue
    fi
    printf '%s\n' "$line"
    last=$line
  done < <(
    WT_WRAPPER=1 command wt "$@"
    printf '__WT_RC__%s\n' "$?"
  )
  if [[ -z "$rc" ]]; then
    rc=0
  fi
  if [[ "$last" == cd\ * ]]; then
    eval "$last"
  fi
  return "$rc"
}
`

func shellSetupCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "shell-setup",
		Short: "Output shell wrapper function for eval in your shell profile",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			shellEnv := os.Getenv("SHELL")
			if shellEnv == "" {
				fmt.Fprintf(os.Stderr, "wt shell-setup: SHELL environment variable is not set (supported: bash, zsh)\n")
				os.Exit(wt.ExitGeneralError)
				return nil
			}
			shell := filepath.Base(shellEnv)
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

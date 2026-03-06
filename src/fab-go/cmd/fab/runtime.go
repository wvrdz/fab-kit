package main

import (
	"github.com/spf13/cobra"
	"github.com/wvrdz/fab-kit/src/fab-go/internal/resolve"
	"github.com/wvrdz/fab-kit/src/fab-go/internal/runtime"
)

func runtimeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "runtime",
		Short: "Manage ephemeral agent runtime state (.fab-runtime.yaml)",
	}

	cmd.AddCommand(
		runtimeSetIdleCmd(),
		runtimeClearIdleCmd(),
	)

	return cmd
}

func runtimeSetIdleCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "set-idle <change>",
		Short: "Write agent.idle_since timestamp",
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
			return runtime.SetIdle(fabRoot, folder)
		},
	}
}

func runtimeClearIdleCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "clear-idle <change>",
		Short: "Delete agent block from runtime state",
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
			return runtime.ClearIdle(fabRoot, folder)
		},
	}
}

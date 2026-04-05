package main

import (
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
)

func operatorTimeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "time",
		Short: "Print current time and optionally next-tick time",
		RunE:  runOperatorTime,
	}
	cmd.Flags().String("interval", "", "Duration until next tick (e.g. 3m). If given, outputs next: HH:MM")
	return cmd
}

func runOperatorTime(cmd *cobra.Command, args []string) error {
	now := time.Now()
	fmt.Fprintf(cmd.OutOrStdout(), "now: %s\n", now.Format("15:04"))

	interval, _ := cmd.Flags().GetString("interval")
	if interval == "" {
		return nil
	}

	d, err := time.ParseDuration(interval)
	if err != nil {
		fmt.Fprintf(cmd.ErrOrStderr(), "Error: invalid --interval %q: %v\n", interval, err)
		os.Exit(1)
	}

	next := now.Add(d)
	fmt.Fprintf(cmd.OutOrStdout(), "next: %s\n", next.Format("15:04"))
	return nil
}

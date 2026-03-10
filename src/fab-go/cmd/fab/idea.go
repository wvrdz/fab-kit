package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/wvrdz/fab-kit/src/fab-go/internal/idea"
)

func ideaCmd() *cobra.Command {
	var fileFlag string

	cmd := &cobra.Command{
		Use:   "idea",
		Short: "Backlog idea management (CRUD for fab/backlog.md)",
	}

	cmd.PersistentFlags().StringVar(&fileFlag, "file", "", "Override backlog file path (relative to git root)")

	cmd.AddCommand(
		ideaAddCmd(&fileFlag),
		ideaListCmd(&fileFlag),
		ideaShowCmd(&fileFlag),
		ideaDoneCmd(&fileFlag),
		ideaReopenCmd(&fileFlag),
		ideaEditCmd(&fileFlag),
		ideaRmCmd(&fileFlag),
	)

	return cmd
}

func resolveIdeaFile(fileFlag *string) (string, error) {
	repoRoot, err := idea.GitRepoRoot()
	if err != nil {
		return "", err
	}
	return idea.ResolveFilePath(repoRoot, *fileFlag), nil
}

func ideaAddCmd(fileFlag *string) *cobra.Command {
	var customID, customDate string

	cmd := &cobra.Command{
		Use:   "add <text>",
		Short: "Add a new idea to the backlog",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			path, err := resolveIdeaFile(fileFlag)
			if err != nil {
				return err
			}
			text := args[0]
			i, err := idea.Add(path, text, customID, customDate)
			if err != nil {
				return err
			}
			fmt.Printf("Added: [%s] %s: %s\n", i.ID, i.Date, i.Text)
			return nil
		},
	}

	cmd.Flags().StringVar(&customID, "id", "", "Custom 4-char ID")
	cmd.Flags().StringVar(&customDate, "date", "", "Custom date (YYYY-MM-DD)")

	return cmd
}

func ideaListCmd(fileFlag *string) *cobra.Command {
	var all, done, jsonOut, reverse bool
	var sortField string

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List ideas from the backlog",
		RunE: func(cmd *cobra.Command, args []string) error {
			path, err := resolveIdeaFile(fileFlag)
			if err != nil {
				return err
			}

			// Check if file exists
			if _, statErr := os.Stat(path); os.IsNotExist(statErr) {
				if jsonOut {
					fmt.Println("[]")
				} else {
					fmt.Println("No ideas file yet. Add one with: fab idea add \"your idea\"")
				}
				return nil
			}

			filter := idea.FilterOpen
			if all {
				filter = idea.FilterAll
			} else if done {
				filter = idea.FilterDone
			}

			ideas, err := idea.List(path, filter, sortField, reverse)
			if err != nil {
				return err
			}

			if len(ideas) == 0 {
				if jsonOut {
					fmt.Println("[]")
				} else {
					fmt.Println("No ideas found.")
				}
				return nil
			}

			if jsonOut {
				enc := json.NewEncoder(os.Stdout)
				enc.SetIndent("", "  ")
				return enc.Encode(ideas)
			}

			for _, i := range ideas {
				fmt.Println(idea.FormatLine(i))
			}
			return nil
		},
	}

	cmd.Flags().BoolVarP(&all, "all", "a", false, "Show all ideas (open + done)")
	cmd.Flags().BoolVar(&done, "done", false, "Show only done ideas")
	cmd.Flags().BoolVar(&jsonOut, "json", false, "Output as JSON")
	cmd.Flags().StringVar(&sortField, "sort", "date", "Sort by field (id or date)")
	cmd.Flags().BoolVar(&reverse, "reverse", false, "Reverse sort order")

	return cmd
}

func ideaShowCmd(fileFlag *string) *cobra.Command {
	var jsonOut bool

	cmd := &cobra.Command{
		Use:   "show <query>",
		Short: "Show a single idea",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			path, err := resolveIdeaFile(fileFlag)
			if err != nil {
				return err
			}

			i, err := idea.Show(path, args[0])
			if err != nil {
				return err
			}

			if jsonOut {
				enc := json.NewEncoder(os.Stdout)
				return enc.Encode(i)
			}

			fmt.Println(idea.FormatLine(i))
			return nil
		},
	}

	cmd.Flags().BoolVar(&jsonOut, "json", false, "Output as JSON")

	return cmd
}

func ideaDoneCmd(fileFlag *string) *cobra.Command {
	return &cobra.Command{
		Use:   "done <query>",
		Short: "Mark an idea as done",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			path, err := resolveIdeaFile(fileFlag)
			if err != nil {
				return err
			}

			i, err := idea.Done(path, args[0])
			if err != nil {
				return err
			}

			fmt.Printf("Done: %s\n", idea.FormatLine(i))
			return nil
		},
	}
}

func ideaReopenCmd(fileFlag *string) *cobra.Command {
	return &cobra.Command{
		Use:   "reopen <query>",
		Short: "Reopen a completed idea",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			path, err := resolveIdeaFile(fileFlag)
			if err != nil {
				return err
			}

			i, err := idea.Reopen(path, args[0])
			if err != nil {
				return err
			}

			fmt.Printf("Reopened: %s\n", idea.FormatLine(i))
			return nil
		},
	}
}

func ideaEditCmd(fileFlag *string) *cobra.Command {
	var newID, newDate string

	cmd := &cobra.Command{
		Use:   "edit <query> <new-text>",
		Short: "Modify an idea's text",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			path, err := resolveIdeaFile(fileFlag)
			if err != nil {
				return err
			}

			i, err := idea.Edit(path, args[0], args[1], newID, newDate)
			if err != nil {
				return err
			}

			fmt.Printf("Updated: %s\n", idea.FormatLine(i))
			return nil
		},
	}

	cmd.Flags().StringVar(&newID, "id", "", "Change the idea's ID")
	cmd.Flags().StringVar(&newDate, "date", "", "Change the idea's date")

	return cmd
}

func ideaRmCmd(fileFlag *string) *cobra.Command {
	var force bool

	cmd := &cobra.Command{
		Use:   "rm <query>",
		Short: "Delete an idea from the backlog",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			path, err := resolveIdeaFile(fileFlag)
			if err != nil {
				return err
			}

			i, err := idea.Rm(path, args[0], force)
			if err != nil {
				return err
			}

			fmt.Printf("Removed: %s\n", idea.FormatLine(i))
			return nil
		},
	}

	cmd.Flags().BoolVar(&force, "force", false, "Confirm deletion")

	return cmd
}

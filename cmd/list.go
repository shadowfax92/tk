package cmd

import (
	"fmt"
	"os"

	"github.com/nickhudkins/tk/model"
	"github.com/nickhudkins/tk/render"
	"github.com/spf13/cobra"
)

var (
	listInbox  bool
	listAll    bool
	listStale  bool
	listStatus string
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List tasks (plain output, use `tk pick` for interactive)",
	Aliases: []string{"ls", "l"},
	RunE: func(cmd *cobra.Command, args []string) error {
		filter := func(t *model.Task) bool {
			if listAll {
				return true
			}
			if listInbox {
				return t.Status == model.StatusInbox
			}
			if listStatus != "" {
				return t.Status == listStatus
			}
			if listStale {
				return t.IsActive() && t.DaysSinceUpdate() > cfg.StaleWarnDays
			}
			// Default: show all active (inbox, todo, next, now)
			return t.IsActive()
		}

		tasks, err := st.List(filter)
		if err != nil {
			return err
		}

		if len(tasks) == 0 {
			fmt.Println("No tasks found.")
			return nil
		}

		if jsonOutput {
			return render.TaskJSON(tasks)
		}

		// Stale warning
		if !listStale && !listAll && listStatus == "" {
			staleCount := 0
			for _, t := range tasks {
				if t.DaysSinceUpdate() > cfg.StaleWarnDays {
					staleCount++
				}
			}
			if staleCount > 0 {
				fmt.Fprintf(os.Stderr, "⚠ %d stale tasks. Run `tk review` to clean up.\n\n", staleCount)
			}
		}

		render.TaskList(tasks, cfg.StaleWarnDays, cfg.StaleCritDays)
		return nil
	},
}

func init() {
	listCmd.Flags().BoolVar(&listInbox, "inbox", false, "Show only inbox items")
	listCmd.Flags().BoolVar(&listAll, "all", false, "Show all including done/archived")
	listCmd.Flags().BoolVar(&listStale, "stale", false, "Show stale tasks")
	listCmd.Flags().StringVar(&listStatus, "status", "", "Filter by status (inbox|todo|next|now|done|archived)")
	rootCmd.AddCommand(listCmd)
}

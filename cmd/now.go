package cmd

import (
	"fmt"
	"strconv"

	"github.com/nickhudkins/tk/model"
	"github.com/nickhudkins/tk/render"
	"github.com/spf13/cobra"
)

var nowCmd = &cobra.Command{
	Use:         "now [id]",
	Short:       "Show or set tasks with status 'now'",
	Aliases:     []string{"today", "td"},
	Annotations: map[string]string{"group": "Status:"},
	Args:        cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 1 {
			id, err := strconv.Atoi(args[0])
			if err != nil {
				return fmt.Errorf("invalid task ID: %s", args[0])
			}
			t, err := st.Get(id)
			if err != nil {
				return fmt.Errorf("task #%d not found", id)
			}
			prev := t.Status
			t.Status = model.StatusNow
			if err := st.Save(t); err != nil {
				return err
			}
			fmt.Printf("#%d: %s → now (%s)\n", t.ID, prev, t.Title)
			return nil
		}

		tasks, err := st.List(func(t *model.Task) bool {
			return t.Status == model.StatusNow
		})
		if err != nil {
			return err
		}
		sortTasksByPriority(tasks)

		if len(tasks) == 0 {
			fmt.Println("No tasks for now. Run `tk plan` to pick from 'next'.")
			return nil
		}

		if jsonOutput {
			return render.TaskJSON(tasks)
		}

		fmt.Println("🔥 Now:")
		for i, t := range tasks {
			fmt.Printf("  %d. %s\n", i+1, render.TaskLine(t, cfg.StaleWarnDays, cfg.StaleCritDays))
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(nowCmd)
}

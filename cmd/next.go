package cmd

import (
	"fmt"
	"strconv"

	"github.com/nickhudkins/tk/model"
	"github.com/nickhudkins/tk/render"
	"github.com/spf13/cobra"
)

var nextCmd = &cobra.Command{
	Use:         "next [id]",
	Short:       "Show or set tasks with status 'next'",
	Aliases:     []string{"n"},
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
			if prev != model.StatusNext {
				if err := checkWIPLimit(model.StatusNext); err != nil {
					return err
				}
			}
			t.Status = model.StatusNext
			if err := st.Save(t); err != nil {
				return err
			}
			fmt.Printf("#%d: %s → next (%s)\n", t.ID, prev, t.Title)
			return nil
		}

		tasks, err := st.List(func(t *model.Task) bool {
			return t.Status == model.StatusNext
		})
		if err != nil {
			return err
		}
		sortTasksByPriority(tasks)

		if len(tasks) == 0 {
			fmt.Println("No tasks in next. Promote tasks or add with `--next`.")
			return nil
		}

		if jsonOutput {
			return render.TaskJSON(tasks)
		}

		fmt.Println("🔜 Next:")
		for i, t := range tasks {
			fmt.Printf("  %d. %s\n", i+1, render.TaskLine(t, cfg.StaleWarnDays, cfg.StaleCritDays))
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(nextCmd)
}

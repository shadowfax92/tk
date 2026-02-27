package cmd

import (
	"fmt"
	"strconv"

	"github.com/nickhudkins/tk/model"
	"github.com/nickhudkins/tk/render"
	"github.com/spf13/cobra"
)

var backlogCmd = &cobra.Command{
	Use:         "backlog [id]",
	Short:       "Show or park tasks in backlog",
	Aliases:     []string{"bl"},
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
			t.Status = model.StatusBacklog
			if err := st.Save(t); err != nil {
				return err
			}
			fmt.Printf("#%d: %s → backlog (%s)\n", t.ID, prev, t.Title)
			return nil
		}

		tasks, err := st.List(func(t *model.Task) bool {
			return t.Status == model.StatusBacklog
		})
		if err != nil {
			return err
		}
		sortTasksByPriority(tasks)

		if len(tasks) == 0 {
			fmt.Println("No backlog tasks.")
			return nil
		}

		if jsonOutput {
			return render.TaskJSON(tasks)
		}

		render.TaskListWithDue(tasks, cfg.StaleWarnDays, cfg.StaleCritDays, cfg.DueSoonDays)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(backlogCmd)
}

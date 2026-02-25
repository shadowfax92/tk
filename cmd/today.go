package cmd

import (
	"fmt"

	"github.com/nickhudkins/tk/model"
	"github.com/nickhudkins/tk/render"
	"github.com/spf13/cobra"
)

var todayCmd = &cobra.Command{
	Use:     "today",
	Short:   "Show tasks with status 'now' (today's focus)",
	Aliases: []string{"td"},
	RunE: func(cmd *cobra.Command, args []string) error {
		tasks, err := st.List(func(t *model.Task) bool {
			return t.Status == model.StatusNow
		})
		if err != nil {
			return err
		}
		sortTasksByPriority(tasks)

		if len(tasks) == 0 {
			fmt.Println("No tasks for today. Run `tk plan` to pick from 'next'.")
			return nil
		}

		if jsonOutput {
			return render.TaskJSON(tasks)
		}

		fmt.Println("Now:")
		for i, t := range tasks {
			fmt.Printf("  %d. %s\n", i+1, render.TaskLine(t, cfg.StaleWarnDays, cfg.StaleCritDays))
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(todayCmd)
}

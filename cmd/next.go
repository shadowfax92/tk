package cmd

import (
	"fmt"

	"github.com/nickhudkins/tk/model"
	"github.com/nickhudkins/tk/render"
	"github.com/spf13/cobra"
)

var nextCmd = &cobra.Command{
	Use:     "next",
	Short:   "Show tasks with status 'next'",
	Aliases: []string{"n"},
	RunE: func(cmd *cobra.Command, args []string) error {
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

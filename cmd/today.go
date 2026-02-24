package cmd

import (
	"fmt"
	"time"

	"github.com/nickhudkins/tk/model"
	"github.com/nickhudkins/tk/render"
	"github.com/spf13/cobra"
)

var todayCmd = &cobra.Command{
	Use:   "today",
	Short: "Show today's focus tasks",
	Aliases: []string{"td"},
	RunE: func(cmd *cobra.Command, args []string) error {
		today := time.Now().Format("2006-01-02")
		tasks, err := st.List(func(t *model.Task) bool {
			return t.FocusDate == today && t.Status != model.StatusDone && t.Status != model.StatusArchived
		})
		if err != nil {
			return err
		}

		if len(tasks) == 0 {
			fmt.Println("No tasks planned for today. Run `tk plan` to set focus.")
			return nil
		}

		if jsonOutput {
			return render.TaskJSON(tasks)
		}

		fmt.Println("Today's focus:")
		for i, t := range tasks {
			fmt.Printf("  %d. %s\n", i+1, render.TaskLine(t, cfg.StaleWarnDays, cfg.StaleCritDays))
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(todayCmd)
}

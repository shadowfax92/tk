package cmd

import (
	"fmt"

	"github.com/nickhudkins/tk/model"
	"github.com/nickhudkins/tk/render"
	"github.com/spf13/cobra"
)

var (
	actionsNow  bool
	actionsNext bool
)

var actionsCmd = &cobra.Command{
	Use:     "actions",
	Short:       "Show next action per task",
	Aliases:     []string{"v"},
	Annotations: map[string]string{"group": "Views:"},
	RunE: func(cmd *cobra.Command, args []string) error {
		status := model.StatusNow
		if actionsNext {
			status = model.StatusNext
		}
		if actionsNow {
			status = model.StatusNow
		}

		tasks, err := st.List(func(t *model.Task) bool {
			return t.Status == status
		})
		if err != nil {
			return err
		}
		sortTasksByPriority(tasks)

		if len(tasks) == 0 {
			fmt.Printf("No tasks in %s.\n", status)
			return nil
		}

		render.NextActions(tasks)
		return nil
	},
}

func init() {
	actionsCmd.Flags().BoolVar(&actionsNow, "now", false, "Show actions for now tasks")
	actionsCmd.Flags().BoolVar(&actionsNext, "next", false, "Show actions for next tasks")
	actionsCmd.MarkFlagsMutuallyExclusive("now", "next")
	rootCmd.AddCommand(actionsCmd)
}

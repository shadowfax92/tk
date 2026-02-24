package cmd

import (
	"github.com/nickhudkins/tk/model"
	"github.com/nickhudkins/tk/render"
	"github.com/spf13/cobra"
)

var nextCmd = &cobra.Command{
	Use:   "next",
	Short: "Show next actionable item per active task",
	Aliases: []string{"n"},
	RunE: func(cmd *cobra.Command, args []string) error {
		tasks, err := st.List(func(t *model.Task) bool {
			return t.IsActive()
		})
		if err != nil {
			return err
		}

		render.NextActions(tasks)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(nextCmd)
}

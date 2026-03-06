package cmd

import (
	"github.com/nickhudkins/tk/model"
	"github.com/spf13/cobra"
)

var pickAll bool
var pickProject string

var pickCmd = &cobra.Command{
	Use:   "pick",
	Short:       "Interactive task picker (fzf)",
	Annotations: map[string]string{"group": "Interactive:"},
	Aliases: []string{"i"},
	RunE: func(cmd *cobra.Command, args []string) error {
		return fzfPick(func(t *model.Task) bool {
			if pickProject != "" && t.Project != pickProject {
				return false
			}
			if pickAll {
				return t.Status != model.StatusArchived
			}
			return t.IsActive()
		})
	},
}

func init() {
	pickCmd.Flags().BoolVar(&pickAll, "all", false, "Include done tasks")
	pickCmd.Flags().StringVarP(&pickProject, "project", "P", "", "Filter by project")
	rootCmd.AddCommand(pickCmd)
}

package cmd

import (
	"strings"

	"github.com/nickhudkins/tk/model"
	"github.com/nickhudkins/tk/render"
	"github.com/spf13/cobra"
)

var searchCmd = &cobra.Command{
	Use:   "search [query]",
	Short: "Search tasks (interactive fzf with actions when no query)",
	Aliases: []string{"s"},
	RunE: func(cmd *cobra.Command, args []string) error {
		// No query: interactive fzf picker with actions
		if len(args) == 0 {
			return fzfPick(func(t *model.Task) bool {
				return t.Status != model.StatusArchived
			})
		}

		// With a query: plain filtered output
		query := strings.ToLower(strings.Join(args, " "))
		tasks, err := st.List(func(t *model.Task) bool {
			return t.Status != model.StatusArchived
		})
		if err != nil {
			return err
		}

		var filtered []*model.Task
		for _, t := range tasks {
			if strings.Contains(strings.ToLower(t.Title), query) ||
				strings.Contains(strings.ToLower(t.Body), query) ||
				containsTag(t.Tags, query) {
				filtered = append(filtered, t)
			}
		}
		if jsonOutput {
			return render.TaskJSON(filtered)
		}
		render.TaskList(filtered, cfg.StaleWarnDays, cfg.StaleCritDays)
		return nil
	},
}

func containsTag(tags []string, query string) bool {
	for _, tag := range tags {
		if strings.Contains(strings.ToLower(tag), query) {
			return true
		}
	}
	return false
}

func init() {
	rootCmd.AddCommand(searchCmd)
}

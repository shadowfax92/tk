package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/nickhudkins/tk/model"
	"github.com/nickhudkins/tk/render"
	"github.com/spf13/cobra"
)

var searchCmd = &cobra.Command{
	Use:   "search [query]",
	Short: "Search tasks by title and body (fzf)",
	Aliases: []string{"s"},
	RunE: func(cmd *cobra.Command, args []string) error {
		tasks, err := st.List(func(t *model.Task) bool {
			return t.Status != model.StatusArchived
		})
		if err != nil {
			return err
		}

		if len(tasks) == 0 {
			fmt.Println("No tasks found.")
			return nil
		}

		// If query provided, filter in-process
		if len(args) > 0 {
			query := strings.ToLower(strings.Join(args, " "))
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
		}

		// Interactive fzf search
		if !hasFzf() {
			return fmt.Errorf("fzf required for interactive search")
		}

		var lines []string
		for _, t := range tasks {
			line := render.TaskLine(t, cfg.StaleWarnDays, cfg.StaleCritDays)
			lines = append(lines, line)
		}

		fzf := exec.Command("fzf", "--ansi", "--no-sort", "--header", "Search tasks")
		fzf.Stdin = strings.NewReader(strings.Join(lines, "\n"))
		fzf.Stdout = os.Stdout
		fzf.Stderr = os.Stderr
		return fzf.Run()
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

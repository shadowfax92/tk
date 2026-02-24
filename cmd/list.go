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

var (
	listInbox bool
	listAll   bool
	listStale bool
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List tasks",
	Aliases: []string{"ls", "l"},
	RunE: func(cmd *cobra.Command, args []string) error {
		filter := func(t *model.Task) bool {
			if listAll {
				return true
			}
			if listInbox {
				return t.Status == model.StatusInbox
			}
			if listStale {
				return (t.Status == model.StatusInbox || t.Status == model.StatusActive) &&
					t.DaysSinceUpdate() > cfg.StaleWarnDays
			}
			return t.Status == model.StatusInbox || t.Status == model.StatusActive
		}

		tasks, err := st.List(filter)
		if err != nil {
			return err
		}

		if len(tasks) == 0 {
			fmt.Println("No tasks found.")
			return nil
		}

		if jsonOutput {
			return render.TaskJSON(tasks)
		}

		// Show stale warning
		if !listStale && !listAll {
			staleCount := 0
			for _, t := range tasks {
				if t.DaysSinceUpdate() > cfg.StaleWarnDays {
					staleCount++
				}
			}
			if staleCount > 0 {
				fmt.Fprintf(os.Stderr, "⚠ %d stale tasks. Run `tk list --stale` to review.\n\n", staleCount)
			}
		}

		// If interactive TTY and fzf available, use fzf with preview
		if isInteractive() && hasFzf() {
			return fzfList(tasks)
		}

		render.TaskList(tasks, cfg.StaleWarnDays, cfg.StaleCritDays)
		return nil
	},
}

func isInteractive() bool {
	fi, err := os.Stdout.Stat()
	if err != nil {
		return false
	}
	return fi.Mode()&os.ModeCharDevice != 0
}

func hasFzf() bool {
	_, err := exec.LookPath("fzf")
	return err == nil
}

func fzfList(tasks []*model.Task) error {
	var lines []string
	for _, t := range tasks {
		line := render.TaskLine(t, cfg.StaleWarnDays, cfg.StaleCritDays)
		lines = append(lines, line)
	}

	input := strings.Join(lines, "\n")
	fzf := exec.Command("fzf", "--ansi", "--no-sort",
		"--preview", fmt.Sprintf("cat %s/{1}.md 2>/dev/null | head -40", st.TaskFilePath(0)[:len(st.TaskFilePath(0))-6]),
	)
	fzf.Stdin = strings.NewReader(input)
	fzf.Stdout = os.Stdout
	fzf.Stderr = os.Stderr
	return fzf.Run()
}

func init() {
	listCmd.Flags().BoolVar(&listInbox, "inbox", false, "Show only inbox items")
	listCmd.Flags().BoolVar(&listAll, "all", false, "Show all tasks including done/archived")
	listCmd.Flags().BoolVar(&listStale, "stale", false, "Show stale tasks")
	rootCmd.AddCommand(listCmd)
}

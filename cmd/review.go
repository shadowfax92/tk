package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/nickhudkins/tk/model"
	"github.com/nickhudkins/tk/render"
	"github.com/spf13/cobra"
)

var reviewCmd = &cobra.Command{
	Use:   "review",
	Short: "Review and clean up stale tasks (fzf multi-select to archive/delete)",
	RunE: func(cmd *cobra.Command, args []string) error {
		tasks, err := st.List(func(t *model.Task) bool {
			return (t.Status == model.StatusInbox || t.Status == model.StatusActive) &&
				t.DaysSinceUpdate() > cfg.StaleWarnDays
		})
		if err != nil {
			return err
		}

		if len(tasks) == 0 {
			fmt.Println("No stale tasks. You're clean!")
			return nil
		}

		fmt.Printf("Found %d stale tasks:\n\n", len(tasks))

		if !hasFzf() {
			render.TaskList(tasks, cfg.StaleWarnDays, cfg.StaleCritDays)
			fmt.Println("\nInstall fzf for interactive cleanup: brew install fzf")
			return nil
		}

		// fzf multi-select stale tasks to archive
		var lines []string
		for _, t := range tasks {
			line := fmt.Sprintf("%d\t%s", t.ID, render.TaskLine(t, cfg.StaleWarnDays, cfg.StaleCritDays))
			lines = append(lines, line)
		}

		fzf := exec.Command("fzf", "--ansi", "--multi",
			"--header", "Select stale tasks to ARCHIVE (TAB to multi-select, ENTER to confirm, ESC to skip)",
			"--with-nth", "2..",
		)
		fzf.Stdin = strings.NewReader(strings.Join(lines, "\n"))
		fzf.Stderr = os.Stderr

		out, err := fzf.Output()
		if err != nil {
			fmt.Println("Skipped review.")
			return nil
		}

		selected := strings.TrimSpace(string(out))
		if selected == "" {
			return nil
		}

		count := 0
		for _, line := range strings.Split(selected, "\n") {
			parts := strings.SplitN(line, "\t", 2)
			if len(parts) == 0 {
				continue
			}
			id, err := strconv.Atoi(strings.TrimSpace(parts[0]))
			if err != nil {
				continue
			}

			t, err := st.Get(id)
			if err != nil {
				continue
			}
			t.Status = model.StatusArchived
			if err := st.Save(t); err != nil {
				fmt.Fprintf(os.Stderr, "Failed to archive #%d: %v\n", id, err)
				continue
			}
			count++
		}

		fmt.Printf("Archived %d stale tasks.\n", count)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(reviewCmd)
}

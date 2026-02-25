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

var reviewCmd = &cobra.Command{
	Use:   "review",
	Short:       "Review and clean up stale tasks",
	Annotations: map[string]string{"group": "Organize:"},
	RunE: func(cmd *cobra.Command, args []string) error {
		tasks, err := st.List(func(t *model.Task) bool {
			return t.IsActive() && t.DaysSinceUpdate() > cfg.StaleWarnDays
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

		var lines []string
		for _, t := range tasks {
			line := fmt.Sprintf("%d\t%s", t.ID, render.TaskLine(t, cfg.StaleWarnDays, cfg.StaleCritDays))
			lines = append(lines, line)
		}

		previewCmd := fmt.Sprintf(
			`cat "$(printf '%%s/%%03d.md' '%s' {1})" 2>/dev/null | tail -n +2`,
			strings.ReplaceAll(st.Root, "'", "'\\''"),
		)

		fzf := exec.Command("fzf", "--ansi", "--multi",
			"--header", "Select stale tasks to ARCHIVE (TAB=multi-select, ENTER=confirm, ESC=skip)",
			"--with-nth", "2..",
			"--delimiter", "\t",
			"--preview", previewCmd,
			"--preview-window", "right:50%:wrap",
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

		ids := extractIDs(strings.Split(selected, "\n"))
		batchSetStatus(ids, model.StatusArchived, "Archived")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(reviewCmd)
}

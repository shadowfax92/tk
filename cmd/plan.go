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

var planCmd = &cobra.Command{
	Use:   "plan",
	Short:       "Move tasks to 'now' (fzf multi-select)",
	Annotations: map[string]string{"group": "Interactive:"},
	RunE: func(cmd *cobra.Command, args []string) error {
		tasks, err := st.List(func(t *model.Task) bool {
			return t.Status == model.StatusTodo || t.Status == model.StatusNext
		})
		if err != nil {
			return err
		}

		if len(tasks) == 0 {
			fmt.Println("No todo/next tasks to plan from. Add tasks first.")
			return nil
		}

		if !hasFzf() {
			return fmt.Errorf("fzf required for `tk plan`. Install: brew install fzf")
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
			"--header", "Select tasks for NOW (TAB=multi-select, ENTER=confirm)",
			"--with-nth", "2..",
			"--delimiter", "\t",
			"--preview", previewCmd,
			"--preview-window", "right:50%:wrap",
		)
		fzf.Stdin = strings.NewReader(strings.Join(lines, "\n"))
		fzf.Stderr = os.Stderr

		out, err := fzf.Output()
		if err != nil {
			return nil
		}

		selected := strings.TrimSpace(string(out))
		if selected == "" {
			return nil
		}

		ids := extractIDs(strings.Split(selected, "\n"))
		count := 0
		for _, id := range ids {
			t, err := st.Get(id)
			if err != nil {
				continue
			}
			t.Status = model.StatusNow
			if err := st.Save(t); err != nil {
				fmt.Fprintf(os.Stderr, "Failed to update #%d: %v\n", id, err)
				continue
			}
			fmt.Printf("#%d → now: %s\n", t.ID, t.Title)
			count++
		}

		fmt.Printf("\n%d tasks moved to now.\n", count)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(planCmd)
}

package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/nickhudkins/tk/model"
	"github.com/nickhudkins/tk/render"
	"github.com/spf13/cobra"
)

var planCmd = &cobra.Command{
	Use:   "plan",
	Short: "Pick tasks for tomorrow's focus (fzf multi-select)",
	RunE: func(cmd *cobra.Command, args []string) error {
		tasks, err := st.List(func(t *model.Task) bool {
			return t.Status == model.StatusActive || t.Status == model.StatusInbox
		})
		if err != nil {
			return err
		}

		if len(tasks) == 0 {
			fmt.Println("No active tasks to plan from.")
			return nil
		}

		if !hasFzf() {
			return fmt.Errorf("fzf is required for `tk plan`. Install it: brew install fzf")
		}

		// Build fzf input: "ID | rendered line"
		var lines []string
		for _, t := range tasks {
			line := fmt.Sprintf("%d\t%s", t.ID, render.TaskLine(t, cfg.StaleWarnDays, cfg.StaleCritDays))
			lines = append(lines, line)
		}

		input := strings.Join(lines, "\n")

		tomorrow := time.Now().AddDate(0, 0, 1).Format("2006-01-02")

		fzf := exec.Command("fzf", "--ansi", "--multi",
			"--header", fmt.Sprintf("Select tasks for %s (TAB to multi-select, ENTER to confirm)", tomorrow),
			"--with-nth", "2..",
		)
		fzf.Stdin = strings.NewReader(input)
		fzf.Stderr = os.Stderr

		out, err := fzf.Output()
		if err != nil {
			return nil // user cancelled
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

			t.FocusDate = tomorrow
			if t.Status == model.StatusInbox {
				t.Status = model.StatusActive
			}
			if err := st.Save(t); err != nil {
				fmt.Fprintf(os.Stderr, "Failed to update #%d: %v\n", id, err)
				continue
			}
			count++
		}

		fmt.Printf("Planned %d tasks for %s\n", count, tomorrow)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(planCmd)
}

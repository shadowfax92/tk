package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"sort"
	"strings"

	"github.com/fatih/color"
	"github.com/nickhudkins/tk/model"
	"github.com/nickhudkins/tk/render"
	"github.com/spf13/cobra"
)

var reviewDimColor = color.New(color.Faint)

var reviewCmd = &cobra.Command{
	Use:         "review",
	Short:       "Review stale and backlog tasks",
	Annotations: map[string]string{"group": "Organize:"},
	RunE: func(cmd *cobra.Command, args []string) error {
		tasks, err := st.List(func(t *model.Task) bool {
			if t.Status == model.StatusBacklog {
				return true
			}
			return t.IsActive() && t.DaysSinceUpdate() > cfg.StaleWarnDays
		})
		if err != nil {
			return err
		}

		if len(tasks) == 0 {
			fmt.Println("Nothing to review. You're clean!")
			return nil
		}

		sort.SliceStable(tasks, func(i, j int) bool {
			return tasks[i].Created.Before(tasks[j].Created)
		})

		fmt.Printf("Found %d tasks to review:\n\n", len(tasks))

		if !hasFzf() {
			render.TaskList(tasks, cfg.StaleWarnDays, cfg.StaleCritDays)
			fmt.Println("\nInstall fzf for interactive review: brew install fzf")
			return nil
		}

		var lines []string
		for _, t := range tasks {
			age := fmt.Sprintf("%-10s", humanizeDaysAgo(t.AgeDays()))
			line := fmt.Sprintf("%d\t%s %s",
				t.ID,
				reviewDimColor.Sprint(age),
				render.TaskLine(t, cfg.StaleWarnDays, cfg.StaleCritDays),
			)
			lines = append(lines, line)
		}

		previewCmd := fmt.Sprintf(
			`cat "$(printf '%%s/%%03d.md' '%s' {1})" 2>/dev/null | tail -n +2`,
			strings.ReplaceAll(st.Root, "'", "'\\''"),
		)

		fzf := exec.Command("fzf", "--ansi", "--multi",
			"--header", "enter:set-status  ^o:archive  ^x:delete  tab:multi  esc:quit",
			"--header-first",
			"--with-nth", "2..",
			"--delimiter", "\t",
			"--expect", "ctrl-o,ctrl-x",
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

		output := strings.TrimRight(string(out), "\n")
		outputLines := strings.Split(output, "\n")
		if len(outputLines) < 2 {
			return nil
		}

		action := outputLines[0]
		ids := extractIDs(outputLines[1:])
		if len(ids) == 0 {
			return nil
		}

		switch action {
		case "ctrl-o":
			batchSetStatus(ids, model.StatusArchived, "Archived")
		case "ctrl-x":
			batchDelete(ids)
		default:
			batchStatusPick(ids)
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(reviewCmd)
}

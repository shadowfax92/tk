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

var editDue string
var editStatus string
var editProj string
var editProjSet bool

var validStatuses = map[string]bool{
	model.StatusInbox: true, model.StatusTodo: true, model.StatusNext: true,
	model.StatusNow: true, model.StatusDone: true, model.StatusArchived: true,
	model.StatusBacklog: true,
}

var editCmd = &cobra.Command{
	Use:         "edit <id>",
	Short:       "Open a task in your editor",
	Aliases:     []string{"e"},
	Annotations: map[string]string{"group": "Tasks:"},
	Args:        cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return fzfEdit()
		}

		id, err := strconv.Atoi(args[0])
		if err != nil {
			return fmt.Errorf("invalid task ID: %s", args[0])
		}

		if editStatus != "" && !validStatuses[editStatus] {
			names := make([]string, 0, len(validStatuses))
			for s := range validStatuses {
				names = append(names, s)
			}
			return fmt.Errorf("invalid status %q (valid: %s)", editStatus, strings.Join(names, ", "))
		}

		editProjSet = cmd.Flags().Changed("project")

		if editDue != "" || editStatus != "" || editProjSet {
			t, err := st.Get(id)
			if err != nil {
				return fmt.Errorf("task #%d not found", id)
			}
			if editDue != "" {
				due, err := parseDue(editDue)
				if err != nil {
					return err
				}
				t.Due = due
				fmt.Printf("#%d: due → %s\n", t.ID, due)
			}
			if editStatus != "" {
				prev := t.Status
				t.Status = editStatus
				fmt.Printf("#%d: %s → %s (%s)\n", t.ID, prev, editStatus, t.Title)
			}
			if editProjSet {
				if editProj != "" {
					if _, err := st.GetProject(editProj); err != nil {
						return fmt.Errorf("project %q not found", editProj)
					}
				}
				prev := t.Project
				t.Project = editProj
				if editProj == "" {
					fmt.Printf("#%d: removed from project %q\n", t.ID, prev)
				} else {
					fmt.Printf("#%d: project → %s\n", t.ID, editProj)
				}
			}
			return st.Save(t)
		}

		path := st.TaskFilePath(id)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			return fmt.Errorf("task #%d not found", id)
		}

		editor := cfg.Editor
		c := exec.Command(editor, path)
		c.Stdin = os.Stdin
		c.Stdout = os.Stdout
		c.Stderr = os.Stderr
		return c.Run()
	},
}

func fzfEdit() error {
	if !hasFzf() {
		return fmt.Errorf("fzf is required for edit without ID. Install: brew install fzf")
	}

	tasks, err := st.List(func(t *model.Task) bool {
		return t.Status != model.StatusDone && t.Status != model.StatusArchived
	})
	if err != nil {
		return err
	}
	if len(tasks) == 0 {
		fmt.Println("No tasks.")
		return nil
	}

	var lines []string
	for _, t := range tasks {
		lines = append(lines, fmt.Sprintf("%d\t%s", t.ID, render.TaskLine(t, cfg.StaleWarnDays, cfg.StaleCritDays)))
	}

	previewCmd := fmt.Sprintf(
		`cat "$(printf '%%s/%%03d.md' '%s' {1})" 2>/dev/null | tail -n +2`,
		strings.ReplaceAll(st.Root, "'", "'\\''"),
	)

	fzf := exec.Command("fzf",
		"--ansi",
		"--no-multi",
		"--with-nth", "2..",
		"--delimiter", "\t",
		"--header", "Select task to edit",
		"--preview", previewCmd,
		"--preview-window", "right:50%:wrap",
	)
	fzf.Stdin = strings.NewReader(strings.Join(lines, "\n"))
	fzf.Stderr = os.Stderr

	out, err := fzf.Output()
	if err != nil {
		return nil
	}

	ids := extractIDs(strings.Split(strings.TrimSpace(string(out)), "\n"))
	if len(ids) == 0 {
		return nil
	}

	return editTask(ids[0])
}

func init() {
	editCmd.Flags().StringVar(&editDue, "due", "", "Set due date (number of days or YYYY-MM-DD)")
	editCmd.Flags().StringVarP(&editStatus, "status", "s", "", "Set status (inbox, todo, next, now, done, archived)")
	editCmd.Flags().StringVarP(&editProj, "project", "P", "", "Set project (empty string to remove)")
	rootCmd.AddCommand(editCmd)
}

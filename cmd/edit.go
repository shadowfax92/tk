package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/nickhudkins/tk/model"
	"github.com/spf13/cobra"
)

var editDue string
var editStatus string

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
	Args:        cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
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

		if editDue != "" || editStatus != "" {
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

func init() {
	editCmd.Flags().StringVar(&editDue, "due", "", "Set due date (number of days or YYYY-MM-DD)")
	editCmd.Flags().StringVarP(&editStatus, "status", "s", "", "Set status (inbox, todo, next, now, done, archived)")
	rootCmd.AddCommand(editCmd)
}

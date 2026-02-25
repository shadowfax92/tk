package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"

	"github.com/spf13/cobra"
)

var editDue string

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

		if editDue != "" {
			t, err := st.Get(id)
			if err != nil {
				return fmt.Errorf("task #%d not found", id)
			}
			due, err := parseDue(editDue)
			if err != nil {
				return err
			}
			t.Due = due
			if err := st.Save(t); err != nil {
				return err
			}
			fmt.Printf("Set #%d due → %s\n", t.ID, due)
			return nil
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
	rootCmd.AddCommand(editCmd)
}

package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"

	"github.com/spf13/cobra"
)

var editCmd = &cobra.Command{
	Use:         "edit <id>",
	Short:       "Open a task in your editor",
	Aliases:     []string{"e"},
	Annotations: map[string]string{"group": "Tasks:"},
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id, err := strconv.Atoi(args[0])
		if err != nil {
			return fmt.Errorf("invalid task ID: %s", args[0])
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
	rootCmd.AddCommand(editCmd)
}

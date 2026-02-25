package cmd

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"

	"github.com/spf13/cobra"
)

var copyCmd = &cobra.Command{
	Use:     "copy <id>",
	Short:       "Copy task file path to clipboard",
	Aliases:     []string{"c"},
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

		absPath, err := filepath.Abs(path)
		if err != nil {
			return err
		}

		if err := copyToClipboard(absPath); err != nil {
			return err
		}

		fmt.Println(absPath)
		return nil
	},
}

func copyToClipboard(text string) error {
	type tool struct {
		name string
		args []string
	}

	tools := []tool{
		{name: "pbcopy"},
		{name: "wl-copy"},
		{name: "xclip", args: []string{"-selection", "clipboard"}},
	}

	var lastErr error
	for _, t := range tools {
		if _, err := exec.LookPath(t.name); err != nil {
			continue
		}

		c := exec.Command(t.name, t.args...)
		c.Stdin = bytes.NewBufferString(text)
		if err := c.Run(); err != nil {
			lastErr = err
			continue
		}
		return nil
	}

	if lastErr != nil {
		return fmt.Errorf("failed to copy to clipboard: %w", lastErr)
	}
	return fmt.Errorf("no clipboard tool found (tried: pbcopy, wl-copy, xclip)")
}

func init() {
	rootCmd.AddCommand(copyCmd)
}

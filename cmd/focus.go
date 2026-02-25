package cmd

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/nickhudkins/tk/render"
	"github.com/spf13/cobra"
)

var focusShort bool

var focusCmd = &cobra.Command{
	Use:     "focus",
	Short:   "Show or edit focus items",
	Aliases: []string{"f"},
	RunE: func(cmd *cobra.Command, args []string) error {
		st.EnsureFocus()

		content, err := st.ReadFocus()
		if err != nil {
			return err
		}

		if content == "" {
			fmt.Println("No focus items. Run `tk focus edit` to add some.")
			return nil
		}

		max := cfg.FocusItems
		if focusShort && max > 3 {
			max = 3
		}
		render.FocusItems(content, max)
		return nil
	},
}

var focusEditCmd = &cobra.Command{
	Use:   "edit",
	Short: "Edit focus file in your editor",
	RunE: func(cmd *cobra.Command, args []string) error {
		st.EnsureFocus()
		path := st.FocusFilePath()
		c := exec.Command(cfg.Editor, path)
		c.Stdin = os.Stdin
		c.Stdout = os.Stdout
		c.Stderr = os.Stderr
		return c.Run()
	},
}

func init() {
	focusCmd.Flags().BoolVar(&focusShort, "short", false, "Show fewer items (for shell prompt)")
	focusCmd.AddCommand(focusEditCmd)
	rootCmd.AddCommand(focusCmd)
}

package cmd

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/fatih/color"
	"github.com/nickhudkins/tk/render"
	"github.com/spf13/cobra"
)

var goalCmd = &cobra.Command{
	Use:     "goals",
	Short:   "Show or edit goals",
	Aliases: []string{"g"},
	RunE: func(cmd *cobra.Command, args []string) error {
		goals, err := st.ReadGoals()
		if err != nil {
			return err
		}

		if len(goals) == 0 {
			fmt.Println("No goals. Run `tk goals edit` to add some.")
			return nil
		}

		color.New(color.FgGreen, color.Bold).Println("🏁 Goals")
		render.Goals(goals)
		return nil
	},
}

var goalEditCmd = &cobra.Command{
	Use:   "edit",
	Short: "Edit goals file in your editor",
	RunE: func(cmd *cobra.Command, args []string) error {
		st.EnsureGoals()
		path := st.GoalsFilePath()
		c := exec.Command(cfg.Editor, path)
		c.Stdin = os.Stdin
		c.Stdout = os.Stdout
		c.Stderr = os.Stderr
		return c.Run()
	},
}

func init() {
	goalCmd.AddCommand(goalEditCmd)
	rootCmd.AddCommand(goalCmd)
}

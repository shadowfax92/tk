package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

var addDesc string

var addCmd = &cobra.Command{
	Use:   "add [title...]",
	Short: "Add a task to inbox",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		title := strings.Join(args, " ")
		t, err := st.Add(title, addDesc)
		if err != nil {
			return err
		}
		fmt.Printf("Added #%d: %s [inbox]\n", t.ID, t.Title)
		return nil
	},
}

func init() {
	addCmd.Flags().StringVarP(&addDesc, "desc", "d", "", "Task description/body")
	rootCmd.AddCommand(addCmd)
}

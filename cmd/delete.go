package cmd

import (
	"fmt"
	"strconv"

	"github.com/spf13/cobra"
)

var deleteCmd = &cobra.Command{
	Use:   "delete <id>",
	Short:       "Delete a task",
	Annotations: map[string]string{"group": "Tasks:"},
	Aliases: []string{"rm"},
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id, err := strconv.Atoi(args[0])
		if err != nil {
			return fmt.Errorf("invalid task ID: %s", args[0])
		}
		t, err := st.Get(id)
		if err != nil {
			return fmt.Errorf("task #%d not found", id)
		}
		if err := st.Delete(id); err != nil {
			return err
		}
		fmt.Printf("Deleted #%d: %s\n", t.ID, t.Title)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(deleteCmd)
}

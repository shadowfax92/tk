package cmd

import (
	"fmt"
	"strconv"

	"github.com/nickhudkins/tk/model"
	"github.com/spf13/cobra"
)

var promoteCmd = &cobra.Command{
	Use:   "promote <id>",
	Short: "Move a task from inbox to active",
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
		if t.Status != model.StatusInbox {
			return fmt.Errorf("task #%d is already %s", id, t.Status)
		}
		t.Status = model.StatusActive
		if err := st.Save(t); err != nil {
			return err
		}
		fmt.Printf("Promoted #%d: %s → active\n", t.ID, t.Title)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(promoteCmd)
}

package cmd

import (
	"fmt"
	"strconv"

	"github.com/nickhudkins/tk/model"
	"github.com/spf13/cobra"
)

var archiveCmd = &cobra.Command{
	Use:   "archive <id>",
	Short: "Archive a task",
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
		t.Status = model.StatusArchived
		if err := st.Save(t); err != nil {
			return err
		}
		fmt.Printf("Archived #%d: %s\n", t.ID, t.Title)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(archiveCmd)
}

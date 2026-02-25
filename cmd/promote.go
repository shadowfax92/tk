package cmd

import (
	"fmt"
	"strconv"

	"github.com/nickhudkins/tk/model"
	"github.com/spf13/cobra"
)

var promoteCmd = &cobra.Command{
	Use:     "promote <id>",
	Short:   "Advance a task to the next status (inbox→todo→next→now→done)",
	Aliases: []string{"adv", "p"},
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id, err := strconv.Atoi(args[0])
		if err != nil {
			return fmt.Errorf("invalid task ID: %s", args[0])
		}
		t, err := st.Get(id)
		if err != nil {
			return fmt.Errorf("task #%d not found", id)
		}
		next := model.Advance(t.Status)
		if next == "" {
			return fmt.Errorf("task #%d is already %s (terminal)", id, t.Status)
		}
		prev := t.Status
		t.Status = next
		if err := st.Save(t); err != nil {
			return err
		}
		fmt.Printf("#%d: %s → %s (%s)\n", t.ID, prev, next, t.Title)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(promoteCmd)
}

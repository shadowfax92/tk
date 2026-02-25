package cmd

import (
	"fmt"
	"strconv"

	"github.com/nickhudkins/tk/model"
	"github.com/spf13/cobra"
)

var demoteCmd = &cobra.Command{
	Use:     "demote <id>",
	Short:   "Move a task back to the previous status (done→now→next→todo→inbox)",
	Aliases: []string{"dem", "b"},
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
		prev := model.Demote(t.Status)
		if prev == "" {
			return fmt.Errorf("task #%d is already %s (can't demote further)", id, t.Status)
		}
		old := t.Status
		t.Status = prev
		if err := st.Save(t); err != nil {
			return err
		}
		fmt.Printf("#%d: %s → %s (%s)\n", t.ID, old, prev, t.Title)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(demoteCmd)
}

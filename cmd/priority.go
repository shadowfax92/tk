package cmd

import (
	"fmt"
	"strconv"

	"github.com/spf13/cobra"
)

var priorityCmd = &cobra.Command{
	Use:     "priority <id> <p0|p1|p2>",
	Short:   "Set task priority",
	Aliases: []string{"prio"},
	Args:    cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		id, err := strconv.Atoi(args[0])
		if err != nil {
			return fmt.Errorf("invalid task ID: %s", args[0])
		}

		prio := args[1]
		if prio != "p0" && prio != "p1" && prio != "p2" && prio != "" {
			return fmt.Errorf("priority must be p0, p1, or p2")
		}

		t, err := st.Get(id)
		if err != nil {
			return fmt.Errorf("task #%d not found", id)
		}
		t.Priority = prio
		if err := st.Save(t); err != nil {
			return err
		}
		fmt.Printf("Set #%d priority → %s\n", t.ID, prio)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(priorityCmd)
}

package cmd

import (
	"fmt"
	"strconv"

	"github.com/spf13/cobra"
)

var priorityCmd = &cobra.Command{
	Use:     "priority <id> <p0|p1|p2>",
	Short:   "Set task priority",
	Aliases: []string{"prio", "piority", "priorty"},
	Args:    cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		return setTaskPriority(args[0], args[1])
	},
}

var p0Cmd = &cobra.Command{
	Use:   "p0 <id>",
	Short: "Set task priority to p0",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return setTaskPriority(args[0], "p0")
	},
}

var p1Cmd = &cobra.Command{
	Use:   "p1 <id>",
	Short: "Set task priority to p1",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return setTaskPriority(args[0], "p1")
	},
}

var p2Cmd = &cobra.Command{
	Use:   "p2 <id>",
	Short: "Set task priority to p2",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return setTaskPriority(args[0], "p2")
	},
}

func setTaskPriority(idArg, prio string) error {
	id, err := strconv.Atoi(idArg)
	if err != nil {
		return fmt.Errorf("invalid task ID: %s", idArg)
	}

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
}

func init() {
	rootCmd.AddCommand(priorityCmd)
	rootCmd.AddCommand(p0Cmd)
	rootCmd.AddCommand(p1Cmd)
	rootCmd.AddCommand(p2Cmd)
}

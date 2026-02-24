package cmd

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
)

var tagCmd = &cobra.Command{
	Use:   "tag <id> <tag>",
	Short: "Add a tag to a task",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		id, err := strconv.Atoi(args[0])
		if err != nil {
			return fmt.Errorf("invalid task ID: %s", args[0])
		}

		tag := strings.TrimPrefix(args[1], "#")

		t, err := st.Get(id)
		if err != nil {
			return fmt.Errorf("task #%d not found", id)
		}

		for _, existing := range t.Tags {
			if existing == tag {
				fmt.Printf("#%d already has tag #%s\n", id, tag)
				return nil
			}
		}

		t.Tags = append(t.Tags, tag)
		if err := st.Save(t); err != nil {
			return err
		}
		fmt.Printf("Tagged #%d with #%s\n", t.ID, tag)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(tagCmd)
}

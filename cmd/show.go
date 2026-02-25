package cmd

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var showCmd = &cobra.Command{
	Use:         "show <id>",
	Short:       "View a task's details",
	Annotations: map[string]string{"group": "Tasks:"},
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

		bold := color.New(color.Bold)
		dim := color.New(color.Faint)

		bold.Printf("#%d %s\n", t.ID, t.Title)
		fmt.Printf("Status: %s", t.Status)
		if t.Priority != "" {
			fmt.Printf("  Priority: %s", t.Priority)
		}
		fmt.Println()

		if len(t.Tags) > 0 {
			fmt.Printf("Tags: #%s\n", strings.Join(t.Tags, " #"))
		}

		dim.Printf("Created: %s  Updated: %s\n", t.Created.Format("2006-01-02"), t.Updated.Format("2006-01-02"))

		total, done := t.SubTaskStats()
		if total > 0 {
			fmt.Printf("Progress: %d/%d sub-tasks\n", done, total)
		}

		if strings.TrimSpace(t.Body) != "" {
			fmt.Println()
			fmt.Println(t.Body)
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(showCmd)
}

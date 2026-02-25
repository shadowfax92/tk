package cmd

import (
	"fmt"
	"strings"

	"github.com/nickhudkins/tk/model"
	"github.com/spf13/cobra"
)

var exportCmd = &cobra.Command{
	Use:   "export",
	Short:       "Export all tasks as markdown",
	Annotations: map[string]string{"group": "Views:"},
	RunE: func(cmd *cobra.Command, args []string) error {
		tasks, err := st.List(nil)
		if err != nil {
			return err
		}

		grouped := map[string][]*model.Task{}
		for _, t := range tasks {
			grouped[t.Status] = append(grouped[t.Status], t)
		}

		fmt.Println("# Tasks")
		fmt.Println()

		sections := []struct{ status, label string }{
			{model.StatusNow, "Now"},
			{model.StatusNext, "Next"},
			{model.StatusTodo, "Todo"},
			{model.StatusInbox, "Inbox"},
			{model.StatusDone, "Done"},
			{model.StatusArchived, "Archived"},
		}
		for _, s := range sections {
			if len(grouped[s.status]) > 0 {
				fmt.Printf("## %s\n", s.label)
				printSection(grouped[s.status])
			}
		}

		return nil
	},
}

func printSection(tasks []*model.Task) {
	for _, t := range tasks {
		prio := ""
		if t.Priority != "" {
			prio = fmt.Sprintf(" [%s]", t.Priority)
		}
		tags := ""
		if len(t.Tags) > 0 {
			tags = " #" + strings.Join(t.Tags, " #")
		}
		age := ""
		if t.DaysSinceUpdate() > 28 {
			age = fmt.Sprintf(" (%dd stale)", t.DaysSinceUpdate())
		}
		fmt.Printf("- #%d %s%s%s%s\n", t.ID, t.Title, prio, tags, age)

		total, completed := t.SubTaskStats()
		if total > 0 {
			fmt.Printf("  Progress: %d/%d\n", completed, total)
		}
	}
	fmt.Println()
}

func init() {
	rootCmd.AddCommand(exportCmd)
}

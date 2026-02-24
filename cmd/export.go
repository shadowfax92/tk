package cmd

import (
	"fmt"
	"strings"

	"github.com/nickhudkins/tk/model"
	"github.com/spf13/cobra"
)

var exportCmd = &cobra.Command{
	Use:   "export",
	Short: "Export all tasks as a single markdown overview",
	RunE: func(cmd *cobra.Command, args []string) error {
		tasks, err := st.List(nil)
		if err != nil {
			return err
		}

		var inbox, active, done, archived []*model.Task
		for _, t := range tasks {
			switch t.Status {
			case model.StatusInbox:
				inbox = append(inbox, t)
			case model.StatusActive:
				active = append(active, t)
			case model.StatusDone:
				done = append(done, t)
			case model.StatusArchived:
				archived = append(archived, t)
			}
		}

		fmt.Println("# Tasks")
		fmt.Println()

		if len(active) > 0 {
			fmt.Println("## Active")
			printSection(active)
		}
		if len(inbox) > 0 {
			fmt.Println("## Inbox")
			printSection(inbox)
		}
		if len(done) > 0 {
			fmt.Println("## Done")
			printSection(done)
		}
		if len(archived) > 0 {
			fmt.Println("## Archived")
			printSection(archived)
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

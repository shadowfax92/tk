package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
)

var tagsCmd = &cobra.Command{
	Use:     "tags",
	Short:   "Manage tags",
	Aliases: []string{"tag"},
	Args:    cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		return listTags()
	},
}

var tagsAddCmd = &cobra.Command{
	Use:   "add <id> <tags>",
	Short: "Add tag(s) to a task",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		return addTagsToTask(args[0], args[1])
	},
}

var tagsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all tags with counts",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		return listTags()
	},
}

var tagsDeleteCmd = &cobra.Command{
	Use:   "delete <tags>",
	Short: "Delete tag(s) from all tasks",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return deleteTags(args[0])
	},
}

func addTagsToTask(idArg, tagsArg string) error {
	id, err := strconv.Atoi(idArg)
	if err != nil {
		return fmt.Errorf("invalid task ID: %s", idArg)
	}

	parsed := normalizeTags([]string{tagsArg})
	if len(parsed) == 0 {
		return fmt.Errorf("no valid tags provided")
	}

	t, err := st.Get(id)
	if err != nil {
		return fmt.Errorf("task #%d not found", id)
	}

	seen := map[string]bool{}
	for _, existing := range t.Tags {
		seen[existing] = true
	}

	var added []string
	for _, tag := range parsed {
		if seen[tag] {
			continue
		}
		seen[tag] = true
		t.Tags = append(t.Tags, tag)
		added = append(added, tag)
	}

	if len(added) == 0 {
		fmt.Printf("#%d already has all requested tags\n", id)
		return nil
	}

	if err := st.Save(t); err != nil {
		return err
	}

	for i := range added {
		added[i] = "#" + added[i]
	}
	fmt.Printf("Tagged #%d with %s\n", t.ID, strings.Join(added, ", "))
	return nil
}

func deleteTags(tagsArg string) error {
	targets := normalizeTags([]string{tagsArg})
	if len(targets) == 0 {
		return fmt.Errorf("no valid tags provided")
	}

	targetSet := map[string]bool{}
	for _, t := range targets {
		targetSet[t] = true
	}

	tasks, err := st.List(nil)
	if err != nil {
		return err
	}

	updatedTasks := 0
	removedCount := 0

	for _, task := range tasks {
		if len(task.Tags) == 0 {
			continue
		}

		filtered := make([]string, 0, len(task.Tags))
		removedHere := 0
		for _, tag := range task.Tags {
			if targetSet[tag] {
				removedHere++
				continue
			}
			filtered = append(filtered, tag)
		}

		if removedHere == 0 {
			continue
		}

		task.Tags = filtered
		if err := st.Save(task); err != nil {
			return err
		}

		updatedTasks++
		removedCount += removedHere
	}

	if removedCount == 0 {
		fmt.Printf("No tasks had %s\n", formatTagList(targets))
		return nil
	}

	fmt.Printf("Removed %s from %d task(s) (%d total removal(s))\n", formatTagList(targets), updatedTasks, removedCount)
	return nil
}

func formatTagList(tags []string) string {
	out := make([]string, 0, len(tags))
	for _, tag := range tags {
		out = append(out, "#"+tag)
	}
	return strings.Join(out, ", ")
}

func listTags() error {
	tasks, err := st.List(nil)
	if err != nil {
		return err
	}

	counts := map[string]int{}
	for _, t := range tasks {
		for _, tag := range t.Tags {
			if tag == "" {
				continue
			}
			counts[tag]++
		}
	}

	if len(counts) == 0 {
		fmt.Println("No tags found.")
		return nil
	}

	type tagCount struct {
		Tag   string `json:"tag"`
		Count int    `json:"count"`
	}

	var out []tagCount
	for tag, count := range counts {
		out = append(out, tagCount{Tag: tag, Count: count})
	}

	sort.Slice(out, func(i, j int) bool {
		if out[i].Count == out[j].Count {
			return out[i].Tag < out[j].Tag
		}
		return out[i].Count > out[j].Count
	})

	if jsonOutput {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(out)
	}

	for _, item := range out {
		fmt.Printf("#%s (%d)\n", item.Tag, item.Count)
	}

	return nil
}

func init() {
	tagsCmd.AddCommand(tagsAddCmd)
	tagsCmd.AddCommand(tagsDeleteCmd)
	tagsCmd.AddCommand(tagsListCmd)
	rootCmd.AddCommand(tagsCmd)
}

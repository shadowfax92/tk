package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/nickhudkins/tk/model"
	"github.com/spf13/cobra"
)

var boardCmd = &cobra.Command{
	Use:         "board",
	Short:       "Edit all tasks in a board view",
	Aliases:     []string{"b"},
	Annotations: map[string]string{"group": "Interactive:"},
	RunE: func(cmd *cobra.Command, args []string) error {
		return runBoard()
	},
}

const boardDeleteBucket = "_delete"

var boardSectionOrder = []string{model.StatusNow, model.StatusNext, model.StatusTodo, model.StatusInbox, model.StatusBacklog, model.StatusDone, boardDeleteBucket}
var boardSectionLabels = map[string]string{
	model.StatusNow:     "now",
	model.StatusNext:    "next",
	model.StatusTodo:    "todo",
	model.StatusInbox:   "inbox",
	model.StatusBacklog: "backlog",
	model.StatusDone:    "done",
	boardDeleteBucket:   "delete",
}

func runBoard() error {
	tasks, err := st.List(func(t *model.Task) bool {
		return t.Status != model.StatusDone && t.Status != model.StatusArchived
	})
	if err != nil {
		return err
	}

	original := generateBoardMarkdown(tasks)

	tmpDir := os.TempDir()
	tmpFile := filepath.Join(tmpDir, "tk-board.md")
	if err := os.WriteFile(tmpFile, []byte(original), 0644); err != nil {
		return err
	}
	defer os.Remove(tmpFile)

	c := exec.Command(cfg.Editor, tmpFile)
	c.Stdin = os.Stdin
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	if err := c.Run(); err != nil {
		return fmt.Errorf("editor exited with error: %w", err)
	}

	edited, err := os.ReadFile(tmpFile)
	if err != nil {
		return err
	}

	if string(edited) == original {
		fmt.Println("No changes.")
		return nil
	}

	return applyBoardEdits(tasks, string(edited))
}

func generateBoardMarkdown(tasks []*model.Task) string {
	var b strings.Builder
	b.WriteString("# tk board\n")

	grouped := map[string][]*model.Task{}
	for _, t := range tasks {
		grouped[t.Status] = append(grouped[t.Status], t)
	}

	for _, status := range boardSectionOrder {
		label := boardSectionLabels[status]
		b.WriteString(fmt.Sprintf("\n### %s\n", label))

		if status == model.StatusDone {
			continue
		}

		bucket := grouped[status]
		sortTasksByPriority(bucket)
		for _, t := range bucket {
			b.WriteString(formatBoardLine(t))
		}
	}

	return b.String()
}

func formatBoardLine(t *model.Task) string {
	var parts []string
	parts = append(parts, fmt.Sprintf("[%d]", t.ID))
	parts = append(parts, t.Title)
	if t.Priority != "" {
		parts = append(parts, t.Priority)
	}
	for _, tag := range t.Tags {
		parts = append(parts, "#"+tag)
	}
	if t.Due != "" {
		parts = append(parts, "due:"+t.Due)
	}
	if t.Project != "" {
		parts = append(parts, fmt.Sprintf("‹%s›", t.Project))
	}
	return "- " + strings.Join(parts, " ") + "\n"
}

func applyBoardEdits(originalTasks []*model.Task, edited string) error {
	sections := parseBoardSections(edited)

	originalIDs := map[int]bool{}
	for _, t := range originalTasks {
		originalIDs[t.ID] = true
	}

	seenIDs := map[int]bool{}
	created := 0
	moved := 0
	updated := 0
	deleted := 0

	for status, lines := range sections {
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if line == "" || strings.HasPrefix(line, "#") {
				continue
			}

			if m := editIDPattern.FindStringSubmatch(line); m != nil {
				id, _ := strconv.Atoi(m[1])
				seenIDs[id] = true

				if status == boardDeleteBucket {
					t, err := st.Get(id)
					if err != nil {
						fmt.Fprintf(os.Stderr, "skip [%d]: not found\n", id)
						continue
					}
					if err := st.Delete(id); err != nil {
						fmt.Fprintf(os.Stderr, "error deleting #%d: %v\n", id, err)
						continue
					}
					fmt.Printf("Deleted #%d: %s\n", id, t.Title)
					deleted++
					continue
				}

				rest := strings.TrimSpace(m[2])

				t, err := st.Get(id)
				if err != nil {
					fmt.Fprintf(os.Stderr, "skip [%d]: not found\n", id)
					continue
				}

				changed := false
				if t.Status != status {
					prev := t.Status
					t.Status = status
					fmt.Printf("#%d: %s → %s (%s)\n", t.ID, prev, status, t.Title)
					changed = true
					moved++
				}

				newTitle, newPrio, newTags, newDue := parseBoardRest(rest)
				if newTitle != "" && newTitle != t.Title {
					t.Title = newTitle
					changed = true
					updated++
				}
				if newPrio != t.Priority {
					t.Priority = newPrio
					changed = true
				}
				if newDue != "" && newDue != t.Due {
					t.Due = newDue
					changed = true
				}
				if !tagsEqual(newTags, t.Tags) && len(newTags) > 0 {
					t.Tags = newTags
					changed = true
				}

				if changed {
					if err := st.Save(t); err != nil {
						fmt.Fprintf(os.Stderr, "error saving #%d: %v\n", id, err)
					}
				}
			} else if m := editNewPattern.FindStringSubmatch(line); m != nil {
				raw := strings.TrimSpace(m[1])
				_, title, tags, prio := parseBulkLine(raw)
				taskStatus := status

				t, err := st.AddFull(title, "", taskStatus, tags, "")
				if err != nil {
					fmt.Fprintf(os.Stderr, "error creating task: %v\n", err)
					continue
				}
				if prio != "" {
					t.Priority = prio
					_ = st.Save(t)
				}
				fmt.Printf("Added #%d: %s [%s]\n", t.ID, t.Title, t.Status)
				created++
			}
		}
	}

	// Removed lines are silently ignored — use tk delete for deletion.

	if created == 0 && moved == 0 && updated == 0 && deleted == 0 {
		fmt.Println("No changes detected.")
	} else {
		fmt.Printf("\nSummary: %d created, %d moved, %d updated, %d deleted\n", created, moved, updated, deleted)
	}

	return nil
}

func parseBoardSections(content string) map[string][]string {
	sections := map[string][]string{}
	currentStatus := ""

	for _, line := range strings.Split(content, "\n") {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "### ") {
			label := strings.TrimSpace(strings.TrimPrefix(trimmed, "### "))
			label = strings.ToLower(label)
			for status, l := range boardSectionLabels {
				if l == label {
					currentStatus = status
					break
				}
			}
			continue
		}
		if strings.HasPrefix(trimmed, "# ") {
			continue
		}
		if currentStatus != "" {
			sections[currentStatus] = append(sections[currentStatus], line)
		}
	}

	return sections
}

func parseBoardRest(rest string) (title, prio string, tags []string, due string) {
	// Strip project annotation ‹slug› (read-only, not editable here)
	var cleaned []string
	for _, word := range strings.Fields(rest) {
		if strings.HasPrefix(word, "‹") && strings.HasSuffix(word, "›") {
			continue
		}
		cleaned = append(cleaned, word)
	}
	return parseEditRest(strings.Join(cleaned, " "))
}

func init() {
	rootCmd.AddCommand(boardCmd)
}

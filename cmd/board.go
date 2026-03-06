package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/nickhudkins/tk/model"
	"github.com/spf13/cobra"
)

var boardProjects bool
var boardTags bool

var boardCmd = &cobra.Command{
	Use:         "board",
	Short:       "Edit all tasks in a board view",
	Aliases:     []string{"b"},
	Annotations: map[string]string{"group": "Interactive:"},
	RunE: func(cmd *cobra.Command, args []string) error {
		switch {
		case boardProjects:
			return runBoardByProject()
		case boardTags:
			return runBoardByTag()
		default:
			return runBoard()
		}
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

// openBoardEditor writes content to a temp file, opens editor, returns edited content.
func openBoardEditor(name, content string) (string, error) {
	tmpDir := os.TempDir()
	tmpFile := filepath.Join(tmpDir, fmt.Sprintf("tk-board-%s.md", name))
	if err := os.WriteFile(tmpFile, []byte(content), 0644); err != nil {
		return "", err
	}
	defer os.Remove(tmpFile)

	c := exec.Command(cfg.Editor, tmpFile)
	c.Stdin = os.Stdin
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	if err := c.Run(); err != nil {
		return "", fmt.Errorf("editor exited with error: %w", err)
	}

	edited, err := os.ReadFile(tmpFile)
	if err != nil {
		return "", err
	}
	return string(edited), nil
}

// --- Status board (default) ---

func runBoard() error {
	tasks, err := st.List(func(t *model.Task) bool {
		return t.Status != model.StatusDone && t.Status != model.StatusArchived
	})
	if err != nil {
		return err
	}

	original := generateBoardMarkdown(tasks)
	edited, err := openBoardEditor("status", original)
	if err != nil {
		return err
	}
	if edited == original {
		fmt.Println("No changes.")
		return nil
	}
	return applyBoardEdits(tasks, edited)
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

// formatBoardLineWithStatus shows [status] inline — used when grouping by project or tag.
func formatBoardLineWithStatus(t *model.Task, showProject, showTags bool) string {
	var parts []string
	parts = append(parts, fmt.Sprintf("[%d]", t.ID))
	parts = append(parts, t.Title)
	parts = append(parts, fmt.Sprintf("[%s]", t.Status))
	if t.Priority != "" {
		parts = append(parts, t.Priority)
	}
	if showTags {
		for _, tag := range t.Tags {
			parts = append(parts, "#"+tag)
		}
	}
	if t.Due != "" {
		parts = append(parts, "due:"+t.Due)
	}
	if showProject && t.Project != "" {
		parts = append(parts, fmt.Sprintf("‹%s›", t.Project))
	}
	return "- " + strings.Join(parts, " ") + "\n"
}

func applyBoardEdits(_ []*model.Task, edited string) error {
	sections := parseBoardSections(edited)

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

				t, err := st.AddFull(title, "", status, tags, "")
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

	printBoardSummary(created, moved, updated, deleted)
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
	var cleaned []string
	for _, word := range strings.Fields(rest) {
		if strings.HasPrefix(word, "‹") && strings.HasSuffix(word, "›") {
			continue
		}
		cleaned = append(cleaned, word)
	}
	return parseEditRest(strings.Join(cleaned, " "))
}

// --- Project board ---

const boardNoProject = "(no project)"

func runBoardByProject() error {
	tasks, err := st.List(func(t *model.Task) bool {
		return t.Status != model.StatusDone && t.Status != model.StatusArchived
	})
	if err != nil {
		return err
	}

	projects, err := st.ReadProjects()
	if err != nil {
		return err
	}

	original := generateBoardByProject(tasks, projects)
	edited, err := openBoardEditor("projects", original)
	if err != nil {
		return err
	}
	if edited == original {
		fmt.Println("No changes.")
		return nil
	}
	return applyBoardProjectEdits(tasks, projects, edited)
}

func generateBoardByProject(tasks []*model.Task, projects []model.Project) string {
	var b strings.Builder
	b.WriteString("# tk board (projects)\n")

	grouped := map[string][]*model.Task{}
	for _, t := range tasks {
		key := t.Project
		if key == "" {
			key = boardNoProject
		}
		grouped[key] = append(grouped[key], t)
	}

	// Active projects in order
	for _, p := range projects {
		if !p.IsActive() {
			continue
		}
		b.WriteString(fmt.Sprintf("\n### %s\n", p.Slug))
		bucket := grouped[p.Slug]
		sortTasksByPriority(bucket)
		for _, t := range bucket {
			b.WriteString(formatBoardLineWithStatus(t, false, true))
		}
	}

	// No-project section
	b.WriteString(fmt.Sprintf("\n### %s\n", boardNoProject))
	bucket := grouped[boardNoProject]
	sortTasksByPriority(bucket)
	for _, t := range bucket {
		b.WriteString(formatBoardLineWithStatus(t, false, true))
	}

	b.WriteString("\n### delete\n")
	return b.String()
}

func applyBoardProjectEdits(_ []*model.Task, _ []model.Project, edited string) error {
	sections := parseGenericSections(edited)

	seenIDs := map[int]bool{}
	created := 0
	moved := 0
	deleted := 0

	for section, lines := range sections {
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if line == "" || strings.HasPrefix(line, "#") {
				continue
			}

			if m := editIDPattern.FindStringSubmatch(line); m != nil {
				id, _ := strconv.Atoi(m[1])
				seenIDs[id] = true

				if section == "delete" {
					t, err := st.Get(id)
					if err != nil {
						continue
					}
					if err := st.Delete(id); err != nil {
						continue
					}
					fmt.Printf("Deleted #%d: %s\n", id, t.Title)
					deleted++
					continue
				}

				t, err := st.Get(id)
				if err != nil {
					fmt.Fprintf(os.Stderr, "skip [%d]: not found\n", id)
					continue
				}

				newProject := section
				if newProject == boardNoProject {
					newProject = ""
				}

				if t.Project != newProject {
					prev := t.Project
					if prev == "" {
						prev = boardNoProject
					}
					t.Project = newProject
					if err := st.Save(t); err != nil {
						fmt.Fprintf(os.Stderr, "error saving #%d: %v\n", id, err)
						continue
					}
					dest := newProject
					if dest == "" {
						dest = boardNoProject
					}
					fmt.Printf("#%d: %s → %s (%s)\n", t.ID, prev, dest, t.Title)
					moved++
				}
			} else if m := editNewPattern.FindStringSubmatch(line); m != nil {
				raw := strings.TrimSpace(m[1])
				_, title, tags, prio := parseBulkLine(raw)

				project := section
				if project == boardNoProject {
					project = ""
				}

				t, err := st.AddFullWithProject(title, "", model.StatusInbox, tags, "", project)
				if err != nil {
					fmt.Fprintf(os.Stderr, "error creating task: %v\n", err)
					continue
				}
				if prio != "" {
					t.Priority = prio
					_ = st.Save(t)
				}
				fmt.Printf("Added #%d: %s [%s]", t.ID, t.Title, t.Status)
				if project != "" {
					fmt.Printf(" (%s)", project)
				}
				fmt.Println()
				created++
			}
		}
	}

	printBoardSummary(created, moved, 0, deleted)
	return nil
}

// --- Tag board ---

const boardUntagged = "(untagged)"

func runBoardByTag() error {
	tasks, err := st.List(func(t *model.Task) bool {
		return t.Status != model.StatusDone && t.Status != model.StatusArchived
	})
	if err != nil {
		return err
	}

	original := generateBoardByTag(tasks)
	edited, err := openBoardEditor("tags", original)
	if err != nil {
		return err
	}
	if edited == original {
		fmt.Println("No changes.")
		return nil
	}
	return applyBoardTagEdits(tasks, edited)
}

func generateBoardByTag(tasks []*model.Task) string {
	var b strings.Builder
	b.WriteString("# tk board (tags)\n")

	// Collect all tags in sorted order and track which tasks go where
	tagOrder := []string{}
	tagSet := map[string]bool{}
	grouped := map[string][]*model.Task{}

	for _, t := range tasks {
		if len(t.Tags) == 0 {
			grouped[boardUntagged] = append(grouped[boardUntagged], t)
			continue
		}
		// Task goes under its first tag
		primaryTag := t.Tags[0]
		if !tagSet[primaryTag] {
			tagSet[primaryTag] = true
			tagOrder = append(tagOrder, primaryTag)
		}
		grouped[primaryTag] = append(grouped[primaryTag], t)
	}
	sort.Strings(tagOrder)

	for _, tag := range tagOrder {
		b.WriteString(fmt.Sprintf("\n### %s\n", tag))
		bucket := grouped[tag]
		sortTasksByPriority(bucket)
		for _, t := range bucket {
			b.WriteString(formatBoardLineWithStatus(t, true, false))
		}
	}

	// Untagged section
	b.WriteString(fmt.Sprintf("\n### %s\n", boardUntagged))
	bucket := grouped[boardUntagged]
	sortTasksByPriority(bucket)
	for _, t := range bucket {
		b.WriteString(formatBoardLineWithStatus(t, true, false))
	}

	b.WriteString("\n### delete\n")
	return b.String()
}

func applyBoardTagEdits(_ []*model.Task, edited string) error {
	sections := parseGenericSections(edited)

	seenIDs := map[int]bool{}
	created := 0
	moved := 0
	deleted := 0

	for section, lines := range sections {
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if line == "" || strings.HasPrefix(line, "#") {
				continue
			}

			if m := editIDPattern.FindStringSubmatch(line); m != nil {
				id, _ := strconv.Atoi(m[1])
				seenIDs[id] = true

				if section == "delete" {
					t, err := st.Get(id)
					if err != nil {
						continue
					}
					if err := st.Delete(id); err != nil {
						continue
					}
					fmt.Printf("Deleted #%d: %s\n", id, t.Title)
					deleted++
					continue
				}

				t, err := st.Get(id)
				if err != nil {
					fmt.Fprintf(os.Stderr, "skip [%d]: not found\n", id)
					continue
				}

				changed := false

				if section == boardUntagged {
					if len(t.Tags) > 0 {
						t.Tags = nil
						changed = true
						fmt.Printf("#%d: tags cleared (%s)\n", t.ID, t.Title)
						moved++
					}
				} else {
					oldPrimary := ""
					if len(t.Tags) > 0 {
						oldPrimary = t.Tags[0]
					}
					if oldPrimary != section {
						// Replace primary tag, keep the rest
						newTags := []string{section}
						for _, tag := range t.Tags {
							if tag != oldPrimary && tag != section {
								newTags = append(newTags, tag)
							}
						}
						t.Tags = newTags
						changed = true
						if oldPrimary == "" {
							fmt.Printf("#%d: → #%s (%s)\n", t.ID, section, t.Title)
						} else {
							fmt.Printf("#%d: #%s → #%s (%s)\n", t.ID, oldPrimary, section, t.Title)
						}
						moved++
					}
				}

				if changed {
					if err := st.Save(t); err != nil {
						fmt.Fprintf(os.Stderr, "error saving #%d: %v\n", id, err)
					}
				}
			} else if m := editNewPattern.FindStringSubmatch(line); m != nil {
				raw := strings.TrimSpace(m[1])
				_, title, tags, prio := parseBulkLine(raw)

				if section != boardUntagged {
					hasSection := false
					for _, tag := range tags {
						if tag == section {
							hasSection = true
							break
						}
					}
					if !hasSection {
						tags = append([]string{section}, tags...)
					}
				}

				t, err := st.AddFull(title, "", model.StatusInbox, tags, "")
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

	printBoardSummary(created, moved, 0, deleted)
	return nil
}

// --- Shared helpers ---

// parseGenericSections parses ### headings into section name → lines map.
func parseGenericSections(content string) map[string][]string {
	sections := map[string][]string{}
	current := ""

	for _, line := range strings.Split(content, "\n") {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "### ") {
			current = strings.TrimSpace(strings.TrimPrefix(trimmed, "### "))
			continue
		}
		if strings.HasPrefix(trimmed, "# ") {
			continue
		}
		if current != "" {
			sections[current] = append(sections[current], line)
		}
	}

	return sections
}

func printBoardSummary(created, moved, updated, deleted int) {
	if created == 0 && moved == 0 && updated == 0 && deleted == 0 {
		fmt.Println("No changes detected.")
		return
	}
	var parts []string
	if created > 0 {
		parts = append(parts, fmt.Sprintf("%d created", created))
	}
	if moved > 0 {
		parts = append(parts, fmt.Sprintf("%d moved", moved))
	}
	if updated > 0 {
		parts = append(parts, fmt.Sprintf("%d updated", updated))
	}
	if deleted > 0 {
		parts = append(parts, fmt.Sprintf("%d deleted", deleted))
	}
	fmt.Printf("\nSummary: %s\n", strings.Join(parts, ", "))
}

func init() {
	boardCmd.Flags().BoolVar(&boardProjects, "projects", false, "Group by project")
	boardCmd.Flags().BoolVar(&boardTags, "tags", false, "Group by tag")
	boardCmd.MarkFlagsMutuallyExclusive("projects", "tags")
	rootCmd.AddCommand(boardCmd)
}

package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/fatih/color"
	"github.com/nickhudkins/tk/model"
	"github.com/nickhudkins/tk/render"
	"github.com/spf13/cobra"
)

var projectAll bool

var projectCmd = &cobra.Command{
	Use:         "project [slug]",
	Short:       "Manage projects",
	Aliases:     []string{"proj", "pr"},
	Annotations: map[string]string{"group": "Organize:"},
	Args:        cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 1 {
			return showProject(args[0])
		}
		return listProjects()
	},
}

var projectAddNext bool

var projectAddCmd = &cobra.Command{
	Use:   "add <slug> [title...]",
	Short: "Create a new project",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		slug := args[0]
		title := slug
		if len(args) > 1 {
			title = strings.Join(args[1:], " ")
		}
		status := model.ProjectStatusTodo
		if projectAddNext {
			status = model.ProjectStatusNext
		}
		p, err := st.AddProject(slug, title, status)
		if err != nil {
			return err
		}
		fmt.Printf("Created project: %s (%s) [%s]\n", p.Slug, p.Title, p.Status)
		return nil
	},
}

var projectNextCmd = &cobra.Command{
	Use:   "next <slug>",
	Short: "Set project status to next",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := st.UpdateProjectStatus(args[0], model.ProjectStatusNext); err != nil {
			return err
		}
		fmt.Printf("Project %s → next\n", args[0])
		return nil
	},
}

var projectDoneCmd = &cobra.Command{
	Use:   "done <slug>",
	Short: "Set project status to done",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := st.UpdateProjectStatus(args[0], model.ProjectStatusDone); err != nil {
			return err
		}
		fmt.Printf("Project %s → done\n", args[0])
		return nil
	},
}

var projectArchiveCmd = &cobra.Command{
	Use:   "archive <slug>",
	Short: "Archive a project",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := st.UpdateProjectStatus(args[0], model.ProjectStatusArchived); err != nil {
			return err
		}
		fmt.Printf("Project %s → archived\n", args[0])
		return nil
	},
}

var projectEditCmd = &cobra.Command{
	Use:   "edit <slug>",
	Short: "Batch edit project tasks in your editor",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return editProject(args[0])
	},
}

func listProjects() error {
	projects, err := st.ReadProjects()
	if err != nil {
		return err
	}
	if len(projects) == 0 {
		fmt.Println("No projects. Create one with `tk project add <slug> <title>`.")
		return nil
	}

	allTasks, err := st.List(nil)
	if err != nil {
		return err
	}

	type projectInfo struct {
		Slug   string `json:"slug"`
		Title  string `json:"title"`
		Status string `json:"status"`
		Now    int    `json:"now"`
		Next   int    `json:"next"`
		Todo   int    `json:"todo"`
		Done   int    `json:"done"`
	}

	var infos []projectInfo
	for _, p := range projects {
		if !projectAll && (p.Status == model.ProjectStatusDone || p.Status == model.ProjectStatusArchived) {
			continue
		}
		info := projectInfo{Slug: p.Slug, Title: p.Title, Status: p.Status}
		for _, t := range allTasks {
			if t.Project != p.Slug {
				continue
			}
			switch t.Status {
			case model.StatusNow:
				info.Now++
			case model.StatusNext:
				info.Next++
			case model.StatusTodo, model.StatusInbox:
				info.Todo++
			case model.StatusDone:
				info.Done++
			}
		}
		infos = append(infos, info)
	}

	if jsonOutput {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(infos)
	}

	if len(infos) == 0 {
		fmt.Println("No active projects.")
		return nil
	}

	bold := color.New(color.Bold)
	dim := color.New(color.Faint)
	bold.Println("📋 Projects")
	for _, info := range infos {
		statusC := render.ProjectStatusColor(info.Status)
		counts := dim.Sprintf("(%d now, %d next, %d todo)", info.Now, info.Next, info.Todo)
		if info.Done > 0 {
			counts = dim.Sprintf("(%d now, %d next, %d todo, %d done)", info.Now, info.Next, info.Todo, info.Done)
		}
		fmt.Printf("  %-20s %-24s %s  %s\n",
			bold.Sprint(info.Slug),
			info.Title,
			statusC.Sprintf("[%s]", info.Status),
			counts,
		)
	}
	return nil
}

func showProject(slug string) error {
	p, err := st.GetProject(slug)
	if err != nil {
		return err
	}

	tasks, err := st.List(func(t *model.Task) bool {
		return t.Project == slug
	})
	if err != nil {
		return err
	}

	if jsonOutput {
		return render.TaskJSON(tasks)
	}

	bold := color.New(color.Bold)
	dim := color.New(color.Faint)
	statusC := render.ProjectStatusColor(p.Status)

	bold.Printf("📋 %s %s\n\n", p.Title, statusC.Sprintf("[%s]", p.Status))

	buckets := []struct {
		label  string
		status []string
		emoji  string
	}{
		{"Now", []string{model.StatusNow}, "🔥"},
		{"Next", []string{model.StatusNext}, "⏭ "},
		{"Todo", []string{model.StatusTodo, model.StatusInbox}, "📥"},
		{"Done", []string{model.StatusDone}, "✅"},
	}

	for _, b := range buckets {
		var bucket []*model.Task
		for _, t := range tasks {
			for _, s := range b.status {
				if t.Status == s {
					bucket = append(bucket, t)
					break
				}
			}
		}
		if len(bucket) == 0 {
			continue
		}
		sortTasksByPriority(bucket)
		bold.Printf("%s %s:\n", b.emoji, b.label)
		if b.label == "Done" {
			dim.Printf("  %d tasks\n", len(bucket))
		} else {
			for i, t := range bucket {
				fmt.Printf("  %d. %s\n", i+1, render.TaskLine(t, cfg.StaleWarnDays, cfg.StaleCritDays))
			}
		}
		fmt.Println()
	}

	return nil
}

// Batch editor: generate markdown, open in editor, diff and apply changes.

var editSectionOrder = []string{model.StatusNow, model.StatusNext, model.StatusTodo, model.StatusInbox, model.StatusDone}
var editSectionLabels = map[string]string{
	model.StatusNow:   "now",
	model.StatusNext:  "next",
	model.StatusTodo:  "todo",
	model.StatusInbox: "inbox",
	model.StatusDone:  "done",
}

func editProject(slug string) error {
	p, err := st.GetProject(slug)
	if err != nil {
		return err
	}

	tasks, err := st.List(func(t *model.Task) bool {
		return t.Project == slug
	})
	if err != nil {
		return err
	}

	original := generateProjectMarkdown(p, tasks)

	tmpDir := os.TempDir()
	tmpFile := filepath.Join(tmpDir, fmt.Sprintf("tk-project-%s.md", slug))
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

	return applyProjectEdits(slug, tasks, string(edited))
}

func generateProjectMarkdown(p *model.Project, tasks []*model.Task) string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("# %s\n", p.Title))

	grouped := map[string][]*model.Task{}
	for _, t := range tasks {
		grouped[t.Status] = append(grouped[t.Status], t)
	}

	for _, status := range editSectionOrder {
		label := editSectionLabels[status]
		b.WriteString(fmt.Sprintf("\n### %s\n", label))
		bucket := grouped[status]
		sortTasksByPriority(bucket)

		if status == model.StatusDone && len(bucket) > 5 {
			for _, t := range bucket[:5] {
				b.WriteString(formatEditLine(t))
			}
			b.WriteString(fmt.Sprintf("# ... and %d more done\n", len(bucket)-5))
		} else {
			for _, t := range bucket {
				b.WriteString(formatEditLine(t))
			}
		}
	}

	return b.String()
}

func formatEditLine(t *model.Task) string {
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
	return "- " + strings.Join(parts, " ") + "\n"
}

var editIDPattern = regexp.MustCompile(`^\s*-\s+\[(\d+)\]\s*(.*)$`)
var editNewPattern = regexp.MustCompile(`^\s*-\s+(.+)$`)

type parsedLine struct {
	id    int
	title string
	raw   string
}

func applyProjectEdits(slug string, originalTasks []*model.Task, edited string) error {
	sections := parseEditSections(edited)

	originalIDs := map[int]bool{}
	for _, t := range originalTasks {
		originalIDs[t.ID] = true
	}

	seenIDs := map[int]bool{}
	created := 0
	moved := 0
	updated := 0

	for status, lines := range sections {
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if line == "" || strings.HasPrefix(line, "#") {
				continue
			}

			if m := editIDPattern.FindStringSubmatch(line); m != nil {
				id, _ := strconv.Atoi(m[1])
				seenIDs[id] = true
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

				newTitle, newPrio, newTags, newDue := parseEditRest(rest)
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
				bulkStatus, title, tags, prio := parseBulkLine(raw)
				// Section status takes precedence over inline status prefix
				_ = bulkStatus
				taskStatus := status

				t, err := st.AddFullWithProject(title, "", taskStatus, tags, "", slug)
				if err != nil {
					fmt.Fprintf(os.Stderr, "error creating task: %v\n", err)
					continue
				}
				if prio != "" {
					t.Priority = prio
					_ = st.Save(t)
				}
				fmt.Printf("Added #%d: %s [%s] (%s)\n", t.ID, t.Title, t.Status, slug)
				created++
			}
		}
	}

	// Tasks in original but not in edited → unassign from project
	removed := 0
	for _, t := range originalTasks {
		if !seenIDs[t.ID] && t.Status != model.StatusDone {
			t.Project = ""
			if err := st.Save(t); err != nil {
				fmt.Fprintf(os.Stderr, "error unassigning #%d: %v\n", t.ID, err)
				continue
			}
			fmt.Printf("Unassigned #%d from %s (%s)\n", t.ID, slug, t.Title)
			removed++
		}
	}

	if created == 0 && moved == 0 && updated == 0 && removed == 0 {
		fmt.Println("No changes detected.")
	} else {
		fmt.Printf("\nSummary: %d created, %d moved, %d updated, %d unassigned\n", created, moved, updated, removed)
	}

	return nil
}

func parseEditSections(content string) map[string][]string {
	sections := map[string][]string{}
	currentStatus := ""

	for _, line := range strings.Split(content, "\n") {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "### ") {
			label := strings.TrimSpace(strings.TrimPrefix(trimmed, "### "))
			label = strings.ToLower(label)
			for status, l := range editSectionLabels {
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

func parseEditRest(rest string) (title, prio string, tags []string, due string) {
	// Extract due:YYYY-MM-DD
	dueRe := regexp.MustCompile(`\bdue:(\S+)`)
	if m := dueRe.FindStringSubmatch(rest); m != nil {
		due = m[1]
		rest = dueRe.ReplaceAllString(rest, "")
	}

	// Extract priority
	if m := prioPattern.FindString(rest); m != "" {
		prio = m
		rest = prioPattern.ReplaceAllString(rest, "")
	}

	// Extract tags
	tagMatches := tagPattern.FindAllStringSubmatch(rest, -1)
	seen := map[string]bool{}
	for _, m := range tagMatches {
		tag := m[1]
		if tag == "" || seen[tag] || tag == "p0" || tag == "p1" || tag == "p2" {
			continue
		}
		seen[tag] = true
		tags = append(tags, tag)
	}
	rest = tagPattern.ReplaceAllString(rest, "")

	title = strings.Join(strings.Fields(rest), " ")
	return
}

func tagsEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	sa := make([]string, len(a))
	copy(sa, a)
	sb := make([]string, len(b))
	copy(sb, b)
	sort.Strings(sa)
	sort.Strings(sb)
	for i := range sa {
		if sa[i] != sb[i] {
			return false
		}
	}
	return true
}

func init() {
	projectCmd.Flags().BoolVar(&projectAll, "all", false, "Include done/archived projects")
	projectAddCmd.Flags().BoolVar(&projectAddNext, "next", false, "Create with status next")
	projectCmd.AddCommand(projectAddCmd)
	projectCmd.AddCommand(projectNextCmd)
	projectCmd.AddCommand(projectDoneCmd)
	projectCmd.AddCommand(projectArchiveCmd)
	projectCmd.AddCommand(projectEditCmd)
	rootCmd.AddCommand(projectCmd)
}

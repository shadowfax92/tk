package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"sort"
	"strconv"
	"strings"

	"github.com/fatih/color"
	"github.com/nickhudkins/tk/model"
	"github.com/nickhudkins/tk/render"
)

func hasFzf() bool {
	_, err := exec.LookPath("fzf")
	return err == nil
}

func isInteractive() bool {
	fi, err := os.Stdout.Stat()
	if err != nil {
		return false
	}
	return fi.Mode()&os.ModeCharDevice != 0
}

var statusFilters = []string{"", model.StatusInbox, model.StatusTodo, model.StatusNext, model.StatusNow, model.StatusBacklog}
var statusFilterLabels = []string{"all", "inbox", "todo", "next", "now", "backlog"}
var pickerFilterLabelColor = color.New(color.FgYellow, color.Bold)

// fzfPick runs a looping interactive fzf picker.
// After each action, reloads tasks and re-enters fzf. ESC to exit.
func fzfPick(filterFn func(*model.Task) bool) error {
	if !hasFzf() {
		return fmt.Errorf("fzf is required. Install: brew install fzf")
	}

	filterIdx := 0
	currentTag := ""

	// Pre-check: exit early if no tasks match the base filter at all.
	allTasks, err := st.List(func(t *model.Task) bool {
		return filterFn == nil || filterFn(t)
	})
	if err != nil {
		return err
	}
	if len(allTasks) == 0 {
		fmt.Println("No tasks.")
		return nil
	}

	for {
		statusScopedFilter := func(t *model.Task) bool {
			if filterFn != nil && !filterFn(t) {
				return false
			}
			if statusFilters[filterIdx] != "" {
				return t.Status == statusFilters[filterIdx]
			}
			return true
		}

		statusScopedTasks, err := st.List(statusScopedFilter)
		if err != nil {
			return err
		}

		tagFilters := collectTagFilters(statusScopedTasks)
		if !containsTagFilter(tagFilters, currentTag) {
			currentTag = ""
		}

		effectiveFilter := func(t *model.Task) bool {
			if !statusScopedFilter(t) {
				return false
			}
			if currentTag != "" {
				return hasExactTag(t, currentTag)
			}
			return true
		}

		tasks, err := st.List(effectiveFilter)
		if err != nil {
			return err
		}

		var lines []string
		for _, t := range tasks {
			line := fmt.Sprintf("%d\t%s", t.ID, render.TaskLine(t, cfg.StaleWarnDays, cfg.StaleCritDays))
			lines = append(lines, line)
		}

		previewCmd := fmt.Sprintf(
			`cat "$(printf '%%s/%%03d.md' '%s' {1})" 2>/dev/null | tail -n +2`,
			strings.ReplaceAll(st.Root, "'", "'\\''"),
		)

		statusLabel := statusFilterLabels[filterIdx]
		tagLabel := "all"
		if currentTag != "" {
			tagLabel = "#" + currentTag
		}
		filterLabel := pickerFilterLabelColor.Sprintf("[status:%s tag:%s]", statusLabel, tagLabel)
		header := fmt.Sprintf("%s  enter:edit  ^e:set-status  ^d:done  ^o:archive  ^b:backlog\n^x:delete  ^r:priority  ^t:add-tag  ^p:project  ^f:filter  ^g:tag  tab:multi  esc:quit", filterLabel)

		fzf := exec.Command("fzf",
			"--ansi",
			"--multi",
			"--no-sort",
			"--with-nth", "2..",
			"--delimiter", "\t",
			"--header", header,
			"--header-first",
			"--expect", "ctrl-d,ctrl-e,ctrl-o,ctrl-x,ctrl-r,ctrl-t,ctrl-f,ctrl-g,ctrl-b,ctrl-p",
			"--preview", previewCmd,
			"--preview-window", "right:50%:wrap",
		)
		fzf.Stdin = strings.NewReader(strings.Join(lines, "\n"))
		fzf.Stderr = os.Stderr

		out, err := fzf.Output()
		output := strings.TrimRight(string(out), "\n")

		if err != nil {
			// fzf failed — but check if an expect key was pressed (e.g. on empty list)
			if output != "" {
				action := strings.Split(output, "\n")[0]
				switch action {
				case "ctrl-f":
					filterIdx = (filterIdx + 1) % len(statusFilters)
					continue
				case "ctrl-g":
					currentTag = nextTagFilter(tagFilters, currentTag)
					continue
				}
			}
			return nil
		}

		outputLines := strings.Split(output, "\n")
		if len(outputLines) < 2 {
			continue
		}

		action := outputLines[0]

		if action == "ctrl-f" {
			filterIdx = (filterIdx + 1) % len(statusFilters)
			continue
		}
		if action == "ctrl-g" {
			currentTag = nextTagFilter(tagFilters, currentTag)
			continue
		}

		selected := outputLines[1:]
		ids := extractIDs(selected)
		if len(ids) == 0 {
			continue
		}

		switch action {
		case "": // enter = edit (single) or show selection (multi)
			if len(ids) == 1 {
				editTask(ids[0])
			} else {
				for _, id := range ids {
					t, _ := st.Get(id)
					if t != nil {
						fmt.Printf("Selected #%d: %s\n", t.ID, t.Title)
					}
				}
			}
		case "ctrl-e":
			batchStatusPick(ids)
		case "ctrl-d":
			batchSetStatus(ids, model.StatusDone, "Done")
		case "ctrl-o":
			batchSetStatus(ids, model.StatusArchived, "Archived")
		case "ctrl-b":
			batchSetStatus(ids, model.StatusBacklog, "Backlog")
		case "ctrl-x":
			batchDelete(ids)
		case "ctrl-r":
			batchPriority(ids)
		case "ctrl-t":
			batchTag(ids)
		case "ctrl-p":
			batchProject(ids)
		}

		fmt.Println() // blank line before re-entering fzf
	}
}

func collectTagFilters(tasks []*model.Task) []string {
	if len(tasks) == 0 {
		return []string{""}
	}

	seen := map[string]bool{}
	var tags []string
	for _, t := range tasks {
		for _, tag := range t.Tags {
			if tag == "" || seen[tag] {
				continue
			}
			seen[tag] = true
			tags = append(tags, tag)
		}
	}
	sort.Strings(tags)

	filters := make([]string, 0, len(tags)+1)
	filters = append(filters, "")
	filters = append(filters, tags...)
	return filters
}

func containsTagFilter(filters []string, target string) bool {
	for _, f := range filters {
		if f == target {
			return true
		}
	}
	return false
}

func nextTagFilter(filters []string, current string) string {
	if len(filters) == 0 {
		return ""
	}
	for i, f := range filters {
		if f == current {
			return filters[(i+1)%len(filters)]
		}
	}
	return filters[0]
}

func hasExactTag(t *model.Task, tag string) bool {
	for _, existing := range t.Tags {
		if existing == tag {
			return true
		}
	}
	return false
}

func extractIDs(lines []string) []int {
	var ids []int
	for _, line := range lines {
		parts := strings.SplitN(strings.TrimSpace(line), "\t", 2)
		if len(parts) == 0 {
			continue
		}
		id, err := strconv.Atoi(strings.TrimSpace(parts[0]))
		if err != nil {
			continue
		}
		ids = append(ids, id)
	}
	return ids
}

func editTask(id int) error {
	path := st.TaskFilePath(id)
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return fmt.Errorf("task #%d not found", id)
	}
	c := exec.Command(cfg.Editor, path)
	c.Stdin = os.Stdin
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	return c.Run()
}

func batchStatusPick(ids []int) {
	if !hasFzf() {
		return
	}

	statuses := strings.Join(model.StatusOrder, "\n") + "\n" + model.StatusBacklog + "\n" + model.StatusArchived

	fzf := exec.Command("fzf",
		"--header", "Set status to:",
		"--no-multi",
	)
	fzf.Stdin = strings.NewReader(statuses)
	fzf.Stderr = os.Stderr

	out, err := fzf.Output()
	if err != nil {
		return
	}

	status := strings.TrimSpace(string(out))
	if status == "" {
		return
	}

	for _, id := range ids {
		t, err := st.Get(id)
		if err != nil {
			fmt.Fprintf(os.Stderr, "skip #%d: not found\n", id)
			continue
		}
		prev := t.Status
		t.Status = status
		if err := st.Save(t); err != nil {
			fmt.Fprintf(os.Stderr, "skip #%d: %v\n", id, err)
			continue
		}
		fmt.Printf("#%d: %s → %s (%s)\n", t.ID, prev, status, t.Title)
	}
}

func batchSetStatus(ids []int, status, label string) {
	for _, id := range ids {
		t, err := st.Get(id)
		if err != nil {
			fmt.Fprintf(os.Stderr, "skip #%d: not found\n", id)
			continue
		}
		t.Status = status
		if err := st.Save(t); err != nil {
			fmt.Fprintf(os.Stderr, "skip #%d: %v\n", id, err)
			continue
		}
		fmt.Printf("%s #%d: %s\n", label, t.ID, t.Title)
	}
}

func batchDelete(ids []int) {
	for _, id := range ids {
		t, err := st.Get(id)
		if err != nil {
			fmt.Fprintf(os.Stderr, "skip #%d: not found\n", id)
			continue
		}
		if err := st.Delete(id); err != nil {
			fmt.Fprintf(os.Stderr, "skip #%d: %v\n", id, err)
			continue
		}
		fmt.Printf("Deleted #%d: %s\n", t.ID, t.Title)
	}
}

func batchTag(ids []int) {
	if !hasFzf() {
		return
	}

	// Collect existing tags across all tasks
	allTasks, err := st.List(nil)
	if err != nil {
		return
	}
	seen := map[string]bool{}
	var tags []string
	for _, t := range allTasks {
		for _, tag := range t.Tags {
			if !seen[tag] {
				seen[tag] = true
				tags = append(tags, tag)
			}
		}
	}

	fzf := exec.Command("fzf",
		"--header", "Select tag or type new one",
		"--no-multi",
		"--print-query",
	)
	fzf.Stdin = strings.NewReader(strings.Join(tags, "\n"))
	fzf.Stderr = os.Stderr

	// --print-query makes fzf exit 1 when no match is selected,
	// but it still writes the query to stdout. We need to capture that.
	out, err := fzf.Output()
	if err != nil && len(out) == 0 {
		return
	}

	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	// --print-query: first line is typed query, second (if present) is selected match
	tag := ""
	if len(lines) >= 2 && lines[1] != "" {
		tag = lines[1]
	} else if len(lines) >= 1 {
		tag = lines[0]
	}
	tag = strings.TrimSpace(strings.TrimPrefix(tag, "#"))
	if tag == "" {
		return
	}

	for _, id := range ids {
		t, err := st.Get(id)
		if err != nil {
			continue
		}
		already := false
		for _, existing := range t.Tags {
			if existing == tag {
				already = true
				break
			}
		}
		if already {
			fmt.Printf("#%d already has #%s\n", t.ID, tag)
			continue
		}
		t.Tags = append(t.Tags, tag)
		if err := st.Save(t); err != nil {
			continue
		}
		fmt.Printf("Tagged #%d with #%s\n", t.ID, tag)
	}
}

func batchProject(ids []int) {
	if !hasFzf() {
		return
	}

	projects, err := st.ReadProjects()
	if err != nil || len(projects) == 0 {
		fmt.Println("No projects. Create one with `tk project add <slug> <title>`.")
		return
	}

	var lines []string
	lines = append(lines, "(none)")
	for _, p := range projects {
		if p.IsActive() {
			lines = append(lines, fmt.Sprintf("%s\t%s [%s]", p.Slug, p.Title, p.Status))
		}
	}

	fzf := exec.Command("fzf",
		"--header", "Assign to project:",
		"--no-multi",
		"--with-nth", "1..",
		"--delimiter", "\t",
	)
	fzf.Stdin = strings.NewReader(strings.Join(lines, "\n"))
	fzf.Stderr = os.Stderr

	out, err := fzf.Output()
	if err != nil {
		return
	}

	selected := strings.TrimSpace(string(out))
	slug := strings.SplitN(selected, "\t", 2)[0]
	if slug == "(none)" {
		slug = ""
	}

	for _, id := range ids {
		t, err := st.Get(id)
		if err != nil {
			continue
		}
		t.Project = slug
		if err := st.Save(t); err != nil {
			continue
		}
		if slug == "" {
			fmt.Printf("#%d: removed from project (%s)\n", t.ID, t.Title)
		} else {
			fmt.Printf("#%d: → project %s (%s)\n", t.ID, slug, t.Title)
		}
	}
}

func batchPriority(ids []int) {
	if !hasFzf() {
		return
	}

	fzf := exec.Command("fzf",
		"--header", "Select priority",
		"--no-multi",
	)
	fzf.Stdin = strings.NewReader("p0\np1\np2")
	fzf.Stderr = os.Stderr

	out, err := fzf.Output()
	if err != nil {
		return
	}

	prio := strings.TrimSpace(string(out))
	for _, id := range ids {
		t, err := st.Get(id)
		if err != nil {
			continue
		}
		t.Priority = prio
		if err := st.Save(t); err != nil {
			continue
		}
		fmt.Printf("Set #%d → %s\n", t.ID, prio)
	}
}

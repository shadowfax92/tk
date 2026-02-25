package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"

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

var statusFilters = []string{"", model.StatusTodo, model.StatusNext, model.StatusNow}
var statusFilterLabels = []string{"all", "todo", "next", "now"}

// fzfPick runs a looping interactive fzf picker.
// After each action, reloads tasks and re-enters fzf. ESC to exit.
func fzfPick(filterFn func(*model.Task) bool) error {
	if !hasFzf() {
		return fmt.Errorf("fzf is required. Install: brew install fzf")
	}

	filterIdx := 0

	for {
		effectiveFilter := func(t *model.Task) bool {
			if filterFn != nil && !filterFn(t) {
				return false
			}
			if statusFilters[filterIdx] != "" {
				return t.Status == statusFilters[filterIdx]
			}
			return true
		}

		tasks, err := st.List(effectiveFilter)
		if err != nil {
			return err
		}

		if len(tasks) == 0 {
			if statusFilters[filterIdx] != "" {
				fmt.Printf("No %s tasks.\n", statusFilterLabels[filterIdx])
			} else {
				fmt.Println("No tasks.")
			}
			return nil
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

		filterLabel := fmt.Sprintf("[%s]", statusFilterLabels[filterIdx])
		header := fmt.Sprintf("%s  enter:edit  ^p:advance  ^b:demote  ^d:done  ^a:archive\n^x:delete  ^r:priority  ^t:tag  ^f:filter  tab:multi  esc:quit", filterLabel)

		fzf := exec.Command("fzf",
			"--ansi",
			"--multi",
			"--no-sort",
			"--with-nth", "2..",
			"--delimiter", "\t",
			"--header", header,
			"--header-first",
			"--expect", "ctrl-d,ctrl-p,ctrl-b,ctrl-a,ctrl-x,ctrl-r,ctrl-t,ctrl-f",
			"--preview", previewCmd,
			"--preview-window", "right:50%:wrap",
		)
		fzf.Stdin = strings.NewReader(strings.Join(lines, "\n"))
		fzf.Stderr = os.Stderr

		out, err := fzf.Output()
		if err != nil {
			// ESC or ctrl-c — exit the loop
			return nil
		}

		output := strings.TrimRight(string(out), "\n")
		outputLines := strings.Split(output, "\n")
		if len(outputLines) < 2 {
			continue
		}

		action := outputLines[0]

		if action == "ctrl-f" {
			filterIdx = (filterIdx + 1) % len(statusFilters)
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
		case "ctrl-p":
			batchAdvance(ids)
		case "ctrl-b":
			batchDemote(ids)
		case "ctrl-d":
			batchSetStatus(ids, model.StatusDone, "Done")
		case "ctrl-a":
			batchSetStatus(ids, model.StatusArchived, "Archived")
		case "ctrl-x":
			batchDelete(ids)
		case "ctrl-r":
			batchPriority(ids)
		case "ctrl-t":
			batchTag(ids)
		}

		fmt.Println() // blank line before re-entering fzf
	}
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

func batchDemote(ids []int) {
	for _, id := range ids {
		t, err := st.Get(id)
		if err != nil {
			fmt.Fprintf(os.Stderr, "skip #%d: not found\n", id)
			continue
		}
		prev := model.Demote(t.Status)
		if prev == "" {
			fmt.Fprintf(os.Stderr, "skip #%d: already %s (can't demote)\n", id, t.Status)
			continue
		}
		old := t.Status
		t.Status = prev
		if err := st.Save(t); err != nil {
			fmt.Fprintf(os.Stderr, "skip #%d: %v\n", id, err)
			continue
		}
		fmt.Printf("#%d: %s → %s (%s)\n", t.ID, old, prev, t.Title)
	}
}

func batchAdvance(ids []int) {
	for _, id := range ids {
		t, err := st.Get(id)
		if err != nil {
			fmt.Fprintf(os.Stderr, "skip #%d: not found\n", id)
			continue
		}
		next := model.Advance(t.Status)
		if next == "" {
			fmt.Fprintf(os.Stderr, "skip #%d: already %s (terminal)\n", id, t.Status)
			continue
		}
		prev := t.Status
		t.Status = next
		if err := st.Save(t); err != nil {
			fmt.Fprintf(os.Stderr, "skip #%d: %v\n", id, err)
			continue
		}
		fmt.Printf("#%d: %s → %s (%s)\n", t.ID, prev, next, t.Title)
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

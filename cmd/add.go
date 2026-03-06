package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/nickhudkins/tk/model"
	"github.com/spf13/cobra"
)

var addDesc string
var addNow bool
var addNext bool
var addBacklog bool
var addTags []string
var addDue string
var addBulk bool
var addProject string

var addCmd = &cobra.Command{
	Use:         "add [title...]",
	Short:       "Add a task (defaults to inbox)",
	Aliases:     []string{"a"},
	Annotations: map[string]string{"group": "Tasks:"},
	Args: cobra.ArbitraryArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		if addBulk {
			return bulkAdd()
		}

		if len(args) == 0 {
			return fmt.Errorf("requires at least 1 arg(s), only received 0")
		}

		status := model.StatusInbox
		if addNow {
			status = model.StatusNow
		}
		if addNext {
			status = model.StatusNext
		}
		if addBacklog {
			status = model.StatusBacklog
		}
		tags := normalizeTags(addTags)

		due, err := parseDue(addDue)
		if err != nil {
			return err
		}

		title := strings.Join(args, " ")
		if addProject != "" {
			if _, err := st.GetProject(addProject); err != nil {
				return fmt.Errorf("project %q not found", addProject)
			}
		}
		t, err := st.AddFullWithProject(title, addDesc, status, tags, due, addProject)
		if err != nil {
			return err
		}
		msg := fmt.Sprintf("Added #%d: %s [%s]", t.ID, t.Title, t.Status)
		if t.HasDue() {
			msg += fmt.Sprintf(" (due %s)", t.Due)
		}
		if t.Project != "" {
			msg += fmt.Sprintf(" (%s)", t.Project)
		}
		fmt.Println(msg)
		return nil
	},
}

func parseDue(raw string) (string, error) {
	if raw == "" {
		return "", nil
	}
	if days, err := strconv.Atoi(raw); err == nil {
		return time.Now().AddDate(0, 0, days).Format("2006-01-02"), nil
	}
	if _, err := time.Parse("2006-01-02", raw); err == nil {
		return raw, nil
	}
	return "", fmt.Errorf("invalid --due value %q (use number of days or YYYY-MM-DD)", raw)
}

var bulkTemplate = `# tk bulk add — one task per line, default status: now
#
# Format:  task title #tag1 #tag2 p0
# Override status with prefix:  next: task title
# Statuses: now (default), next, todo, inbox, backlog
# Priorities: p0, p1, p2
#
# Examples:
#   fix auth bug #api p0
#   next: write blog post #content
#   ship endpoint #api
#
`

var statusPrefixes = map[string]string{
	"now:":     model.StatusNow,
	"next:":    model.StatusNext,
	"todo:":    model.StatusTodo,
	"inbox:":   model.StatusInbox,
	"backlog:": model.StatusBacklog,
}

var tagPattern = regexp.MustCompile(`#(\w[\w-]*)`)
var prioPattern = regexp.MustCompile(`\b(p[012])\b`)

func bulkAdd() error {
	tmpDir := os.TempDir()
	tmpFile := filepath.Join(tmpDir, "tk-bulk-add.md")

	if err := os.WriteFile(tmpFile, []byte(bulkTemplate), 0644); err != nil {
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

	data, err := os.ReadFile(tmpFile)
	if err != nil {
		return err
	}

	lines := strings.Split(string(data), "\n")
	added := 0
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		status, title, tags, prio := parseBulkLine(line)

		t, err := st.AddFull(title, "", status, tags, "")
		if err != nil {
			fmt.Fprintf(os.Stderr, "skip: %s (%v)\n", line, err)
			continue
		}
		if prio != "" {
			t.Priority = prio
			if err := st.Save(t); err != nil {
				fmt.Fprintf(os.Stderr, "warning: added #%d but failed to set priority: %v\n", t.ID, err)
			}
		}
		fmt.Printf("Added #%d: %s [%s]", t.ID, t.Title, t.Status)
		if prio != "" {
			fmt.Printf(" %s", prio)
		}
		if len(tags) > 0 {
			fmt.Printf(" #%s", strings.Join(tags, " #"))
		}
		fmt.Println()
		added++
	}

	if added == 0 {
		fmt.Println("No tasks added.")
	} else {
		fmt.Printf("\n%d tasks added.\n", added)
	}
	return nil
}

func parseBulkLine(line string) (status, title string, tags []string, prio string) {
	status = model.StatusNow

	// Check for status prefix
	lower := strings.ToLower(line)
	for prefix, s := range statusPrefixes {
		if strings.HasPrefix(lower, prefix) {
			status = s
			line = strings.TrimSpace(line[len(prefix):])
			break
		}
	}

	// Extract tags
	tagMatches := tagPattern.FindAllStringSubmatch(line, -1)
	seen := map[string]bool{}
	for _, m := range tagMatches {
		tag := m[1]
		if tag == "" || seen[tag] {
			continue
		}
		// Skip if it looks like a priority (p0, p1, p2)
		if tag == "p0" || tag == "p1" || tag == "p2" {
			continue
		}
		seen[tag] = true
		tags = append(tags, tag)
	}

	// Extract priority
	if m := prioPattern.FindString(line); m != "" {
		prio = m
	}

	// Clean title: remove tags and priority
	title = tagPattern.ReplaceAllString(line, "")
	title = prioPattern.ReplaceAllString(title, "")
	title = strings.Join(strings.Fields(title), " ")

	return
}

func init() {
	addCmd.Flags().StringVarP(&addDesc, "desc", "d", "", "Task description/body")
	addCmd.Flags().BoolVar(&addNow, "now", false, "Add task directly to now")
	addCmd.Flags().BoolVar(&addNext, "next", false, "Add task directly to next")
	addCmd.Flags().StringSliceVar(&addTags, "tags", nil, "Add tag(s), e.g. --tags \"#cli,#x\" or --tags cli --tags x")
	addCmd.Flags().StringVar(&addDue, "due", "", "Due date (number of days or YYYY-MM-DD)")
	addCmd.Flags().BoolVar(&addBacklog, "backlog", false, "Add task directly to backlog")
	addCmd.Flags().BoolVar(&addBulk, "bulk", false, "Bulk add tasks via editor")
	addCmd.Flags().StringVarP(&addProject, "project", "P", "", "Assign to project")
	addCmd.MarkFlagsMutuallyExclusive("now", "next", "backlog")
	rootCmd.AddCommand(addCmd)
}

func normalizeTags(inputs []string) []string {
	seen := map[string]bool{}
	var out []string

	for _, input := range inputs {
		parts := strings.FieldsFunc(input, func(r rune) bool {
			return r == ',' || r == ' ' || r == '\t' || r == '\n'
		})
		for _, p := range parts {
			tag := strings.TrimSpace(strings.TrimPrefix(p, "#"))
			if tag == "" || seen[tag] {
				continue
			}
			seen[tag] = true
			out = append(out, tag)
		}
	}

	return out
}

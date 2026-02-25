package model

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

type Task struct {
	ID       int       `yaml:"id"`
	Title    string    `yaml:"title"`
	Status   string    `yaml:"status"`
	Priority string    `yaml:"priority,omitempty"`
	Tags     []string  `yaml:"tags,omitempty"`
	Due      string    `yaml:"due,omitempty"`
	Created  time.Time `yaml:"created"`
	Updated  time.Time `yaml:"updated"`

	Body string `yaml:"-"`
}

func (t *Task) HasDue() bool {
	return t.Due != ""
}

func (t *Task) DueTime() (time.Time, error) {
	return time.Parse("2006-01-02", t.Due)
}

func (t *Task) DaysUntilDue() int {
	dl, err := t.DueTime()
	if err != nil {
		return 0
	}
	now := time.Now().Truncate(24 * time.Hour)
	dl = dl.Truncate(24 * time.Hour)
	return int(dl.Sub(now).Hours() / 24)
}

const (
	StatusInbox    = "inbox"
	StatusTodo     = "todo"
	StatusNext     = "next"
	StatusNow      = "now"
	StatusDone     = "done"
	StatusArchived = "archived"
)

// StatusOrder defines the natural progression.
var StatusOrder = []string{StatusInbox, StatusTodo, StatusNext, StatusNow, StatusDone}

// Advance returns the next status in the natural flow, or "" if already terminal.
func Advance(current string) string {
	for i, s := range StatusOrder {
		if s == current && i+1 < len(StatusOrder) {
			return StatusOrder[i+1]
		}
	}
	return ""
}

// Demote returns the previous status in the natural flow, or "" if already at inbox.
func Demote(current string) string {
	for i, s := range StatusOrder {
		if s == current && i > 0 {
			return StatusOrder[i-1]
		}
	}
	return ""
}

// IsActive returns true if the task is in a working state (not done/archived).
func (t *Task) IsActive() bool {
	return t.Status != StatusDone && t.Status != StatusArchived
}

func ParseFile(path string) (*Task, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return Parse(data)
}

func Parse(data []byte) (*Task, error) {
	content := string(data)

	if !strings.HasPrefix(content, "---\n") {
		return nil, fmt.Errorf("missing frontmatter")
	}

	end := strings.Index(content[4:], "\n---")
	if end == -1 {
		return nil, fmt.Errorf("unclosed frontmatter")
	}

	fm := content[4 : 4+end]
	body := ""
	rest := content[4+end+4:]
	if len(rest) > 0 {
		body = strings.TrimPrefix(rest, "\n")
	}

	var t Task
	if err := yaml.Unmarshal([]byte(fm), &t); err != nil {
		return nil, fmt.Errorf("parse frontmatter: %w", err)
	}
	t.Body = body

	// Migrate old "active" status to "todo"
	if t.Status == "active" {
		t.Status = StatusTodo
	}

	return &t, nil
}

func (t *Task) Serialize() ([]byte, error) {
	fm, err := yaml.Marshal(t)
	if err != nil {
		return nil, err
	}

	var buf strings.Builder
	buf.WriteString("---\n")
	buf.Write(fm)
	buf.WriteString("---\n")
	if t.Body != "" {
		buf.WriteString("\n")
		buf.WriteString(t.Body)
		if !strings.HasSuffix(t.Body, "\n") {
			buf.WriteString("\n")
		}
	}

	return []byte(buf.String()), nil
}

func (t *Task) SaveTo(path string) error {
	data, err := t.Serialize()
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

func (t *Task) AgeDays() int {
	return int(time.Since(t.Created).Hours() / 24)
}

func (t *Task) DaysSinceUpdate() int {
	return int(time.Since(t.Updated).Hours() / 24)
}

func (t *Task) NextAction() string {
	scanner := bufio.NewScanner(strings.NewReader(t.Body))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "- [ ] ") {
			return strings.TrimPrefix(line, "- [ ] ")
		}
	}
	return t.Title
}

func (t *Task) SubTaskStats() (int, int) {
	total, done := 0, 0
	scanner := bufio.NewScanner(strings.NewReader(t.Body))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "- [ ] ") {
			total++
		} else if strings.HasPrefix(line, "- [x] ") || strings.HasPrefix(line, "- [X] ") {
			total++
			done++
		}
	}
	return total, done
}

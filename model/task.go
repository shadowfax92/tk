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
	ID        int       `yaml:"id"`
	Title     string    `yaml:"title"`
	Status    string    `yaml:"status"`
	Priority  string    `yaml:"priority,omitempty"`
	Tags      []string  `yaml:"tags,omitempty"`
	Created   time.Time `yaml:"created"`
	Updated   time.Time `yaml:"updated"`
	FocusDate string    `yaml:"focus_date,omitempty"`

	// Body is the markdown content below frontmatter
	Body string `yaml:"-"`
}

const (
	StatusInbox    = "inbox"
	StatusActive   = "active"
	StatusDone     = "done"
	StatusArchived = "archived"
)

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
	rest := content[4+end+4:] // skip closing ---\n
	if len(rest) > 0 {
		body = strings.TrimPrefix(rest, "\n")
	}

	var t Task
	if err := yaml.Unmarshal([]byte(fm), &t); err != nil {
		return nil, fmt.Errorf("parse frontmatter: %w", err)
	}
	t.Body = body

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

// NextAction returns the first unchecked sub-task from the body, or the title if none.
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

// SubTaskStats returns (total, completed) counts of sub-tasks.
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

package render

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/fatih/color"
	"github.com/nickhudkins/tk/model"
)

var (
	p0Color    = color.New(color.FgRed, color.Bold)
	p1Color    = color.New(color.FgYellow)
	p2Color    = color.New(color.FgBlue)
	inboxColor = color.New(color.FgWhite, color.Faint)
	doneColor  = color.New(color.FgGreen, color.CrossedOut)
	staleWarn  = color.New(color.FgYellow, color.Faint)
	staleCrit  = color.New(color.FgRed, color.Faint)
	dimColor   = color.New(color.Faint)
	boldColor  = color.New(color.Bold)
)

func TaskLine(t *model.Task, staleWarnDays, staleCritDays int) string {
	id := fmt.Sprintf("#%-3d", t.ID)
	status := fmt.Sprintf("[%s]", t.Status)
	title := t.Title

	var tags string
	if len(t.Tags) > 0 {
		tags = " " + dimColor.Sprint("#"+strings.Join(t.Tags, " #"))
	}

	var prio string
	switch t.Priority {
	case "p0":
		prio = p0Color.Sprint(" p0")
	case "p1":
		prio = p1Color.Sprint(" p1")
	case "p2":
		prio = p2Color.Sprint(" p2")
	}

	var stale string
	age := t.DaysSinceUpdate()
	if age > staleCritDays {
		stale = staleCrit.Sprintf(" (%dd stale)", age)
	} else if age > staleWarnDays {
		stale = staleWarn.Sprintf(" (%dd stale)", age)
	}

	// Subtask progress
	total, done := t.SubTaskStats()
	var progress string
	if total > 0 {
		progress = dimColor.Sprintf(" [%d/%d]", done, total)
	}

	var line string
	switch t.Status {
	case model.StatusInbox:
		line = fmt.Sprintf("%s %s %s%s%s%s%s", dimColor.Sprint(id), inboxColor.Sprint(status), title, prio, tags, progress, stale)
	case model.StatusDone:
		line = fmt.Sprintf("%s %s %s", dimColor.Sprint(id), doneColor.Sprint(status), doneColor.Sprint(title))
	default:
		line = fmt.Sprintf("%s %s %s%s%s%s%s", boldColor.Sprint(id), status, boldColor.Sprint(title), prio, tags, progress, stale)
	}

	return line
}

func TaskList(tasks []*model.Task, staleWarnDays, staleCritDays int) {
	for _, t := range tasks {
		fmt.Println(TaskLine(t, staleWarnDays, staleCritDays))
	}
}

func TaskJSON(tasks []*model.Task) error {
	type jsonTask struct {
		ID        int      `json:"id"`
		Title     string   `json:"title"`
		Status    string   `json:"status"`
		Priority  string   `json:"priority,omitempty"`
		Tags      []string `json:"tags,omitempty"`
		Created   string   `json:"created"`
		Updated   string   `json:"updated"`
		FocusDate string   `json:"focus_date,omitempty"`
		AgeDays   int      `json:"age_days"`
	}

	var out []jsonTask
	for _, t := range tasks {
		out = append(out, jsonTask{
			ID:        t.ID,
			Title:     t.Title,
			Status:    t.Status,
			Priority:  t.Priority,
			Tags:      t.Tags,
			Created:   t.Created.Format("2006-01-02"),
			Updated:   t.Updated.Format("2006-01-02"),
			FocusDate: t.FocusDate,
			AgeDays:   t.AgeDays(),
		})
	}

	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(out)
}

func FocusItems(content string, max int) {
	lines := strings.Split(strings.TrimSpace(content), "\n")
	count := 0
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if count >= max {
			break
		}
		fmt.Println(line)
		count++
	}
}

func NextActions(tasks []*model.Task) {
	for _, t := range tasks {
		if t.Status != model.StatusActive {
			continue
		}
		action := t.NextAction()
		id := fmt.Sprintf("#%-3d", t.ID)
		fmt.Printf("%s %s → %s\n", boldColor.Sprint(id), t.Title, color.CyanString(action))
	}
}

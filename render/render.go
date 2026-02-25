package render

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/nickhudkins/tk/model"
)

var (
	p0Color    = color.New(color.FgRed, color.Bold)
	p1Color    = color.New(color.FgYellow)
	p2Color    = color.New(color.FgBlue)
	doneColor  = color.New(color.FgGreen, color.CrossedOut)
	staleWarn  = color.New(color.FgYellow, color.Faint)
	staleCrit  = color.New(color.FgRed, color.Faint)
	dimColor   = color.New(color.Faint)
	boldColor  = color.New(color.Bold)
	cyanColor  = color.New(color.FgCyan)
	nowColor   = color.New(color.FgGreen, color.Bold)
	focusColor = color.New(color.FgHiBlue)
)

func statusColor(status string) *color.Color {
	switch status {
	case model.StatusInbox:
		return dimColor
	case model.StatusTodo:
		return color.New(color.FgWhite)
	case model.StatusNext:
		return cyanColor
	case model.StatusNow:
		return nowColor
	case model.StatusDone:
		return doneColor
	case model.StatusArchived:
		return dimColor
	}
	return color.New(color.Reset)
}

func TaskLine(t *model.Task, staleWarnDays, staleCritDays int) string {
	id := fmt.Sprintf("#%-3d", t.ID)
	status := fmt.Sprintf("[%s]", t.Status)

	var tags string
	if len(t.Tags) > 0 {
		tags = " " + color.New(color.FgCyan, color.Faint).Sprint("#"+strings.Join(t.Tags, " #"))
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
	if t.IsActive() {
		age := t.DaysSinceUpdate()
		if age > staleCritDays {
			stale = staleCrit.Sprintf(" (%dd stale)", age)
		} else if age > staleWarnDays {
			stale = staleWarn.Sprintf(" (%dd stale)", age)
		}
	}

	total, done := t.SubTaskStats()
	var progress string
	if total > 0 {
		progress = dimColor.Sprintf(" [%d/%d]", done, total)
	}

	sc := statusColor(t.Status)

	switch t.Status {
	case model.StatusDone:
		return fmt.Sprintf("%s %s %s", dimColor.Sprint(id), doneColor.Sprint(status), doneColor.Sprint(t.Title))
	case model.StatusArchived:
		return fmt.Sprintf("%s %s %s", dimColor.Sprint(id), dimColor.Sprint(status), dimColor.Sprint(t.Title))
	case model.StatusNow:
		return fmt.Sprintf("%s %s %s%s%s%s%s", nowColor.Sprint(id), nowColor.Sprint(status), boldColor.Sprint(t.Title), prio, tags, progress, stale)
	default:
		return fmt.Sprintf("%s %s %s%s%s%s%s", sc.Sprint(id), sc.Sprint(status), t.Title, prio, tags, progress, stale)
	}
}

func TaskList(tasks []*model.Task, staleWarnDays, staleCritDays int) {
	for _, t := range tasks {
		fmt.Println(TaskLine(t, staleWarnDays, staleCritDays))
	}
}

func TaskJSON(tasks []*model.Task) error {
	type jsonTask struct {
		ID       int      `json:"id"`
		Title    string   `json:"title"`
		Status   string   `json:"status"`
		Priority string   `json:"priority,omitempty"`
		Tags     []string `json:"tags,omitempty"`
		Created  string   `json:"created"`
		Updated  string   `json:"updated"`
		AgeDays  int      `json:"age_days"`
	}

	var out []jsonTask
	for _, t := range tasks {
		out = append(out, jsonTask{
			ID:       t.ID,
			Title:    t.Title,
			Status:   t.Status,
			Priority: t.Priority,
			Tags:     t.Tags,
			Created:  t.Created.Format("2006-01-02"),
			Updated:  t.Updated.Format("2006-01-02"),
			AgeDays:  t.AgeDays(),
		})
	}

	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(out)
}

func FocusItems(content string, max int) {
	lines := PickFocusItems(content, max)
	for _, line := range lines {
		focusColor.Println(line)
	}
}

func PickFocusItems(content string, max int) []string {
	lines := strings.Split(strings.TrimSpace(content), "\n")

	var filtered []string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		filtered = append(filtered, line)
	}

	if len(filtered) == 0 {
		return nil
	}

	if max <= 0 || max > len(filtered) {
		max = len(filtered)
	}

	if len(filtered) > 1 {
		r := rand.New(rand.NewSource(time.Now().UnixNano()))
		r.Shuffle(len(filtered), func(i, j int) {
			filtered[i], filtered[j] = filtered[j], filtered[i]
		})
	}

	return filtered[:max]
}

func NextActions(tasks []*model.Task) {
	for _, t := range tasks {
		action := t.NextAction()
		id := fmt.Sprintf("#%-3d", t.ID)
		fmt.Printf("%s %s → %s\n", boldColor.Sprint(id), t.Title, cyanColor.Sprint(action))
	}
}

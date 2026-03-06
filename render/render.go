package render

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"sort"
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
	focusColor = color.New(color.FgHiMagenta)
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
	case model.StatusBacklog:
		return dimColor
	}
	return color.New(color.Reset)
}

var (
	dueWarn = color.New(color.FgYellow)
	dueUrgent = color.New(color.FgRed, color.Bold)
)

func dueColor(daysLeft, dueSoonDays int) *color.Color {
	switch {
	case daysLeft <= 1:
		return dueUrgent
	case daysLeft <= dueSoonDays:
		return dueWarn
	default:
		return dimColor
	}
}

func DueCountdown(daysLeft int, c *color.Color) string {
	switch {
	case daysLeft < 0:
		return c.Sprintf("OVERDUE %dd", -daysLeft)
	case daysLeft == 0:
		return c.Sprint("due today")
	case daysLeft == 1:
		return c.Sprint("due tomorrow")
	default:
		return c.Sprintf("due in %dd", daysLeft)
	}
}

func TaskLine(t *model.Task, staleWarnDays, staleCritDays int) string {
	return TaskLineWithDue(t, staleWarnDays, staleCritDays, 3)
}

func TaskLineWithDue(t *model.Task, staleWarnDays, staleCritDays, dueSoonDays int) string {
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

	var due string
	if t.HasDue() && t.IsActive() {
		days := t.DaysUntilDue()
		c := dueColor(days, dueSoonDays)
		due = "  " + DueCountdown(days, c)
	}

	sc := statusColor(t.Status)

	proj := ProjectAnnotation(t.Project)

	switch t.Status {
	case model.StatusDone:
		return fmt.Sprintf("%s %s %s", dimColor.Sprint(id), doneColor.Sprint(status), doneColor.Sprint(t.Title))
	case model.StatusArchived:
		return fmt.Sprintf("%s %s %s", dimColor.Sprint(id), dimColor.Sprint(status), dimColor.Sprint(t.Title))
	case model.StatusNow:
		return fmt.Sprintf("%s %s %s%s%s%s%s%s%s", nowColor.Sprint(id), nowColor.Sprint(status), boldColor.Sprint(t.Title), prio, tags, progress, stale, due, proj)
	default:
		return fmt.Sprintf("%s %s %s%s%s%s%s%s%s", sc.Sprint(id), sc.Sprint(status), t.Title, prio, tags, progress, stale, due, proj)
	}
}

func TaskList(tasks []*model.Task, staleWarnDays, staleCritDays int) {
	for _, t := range tasks {
		fmt.Println(TaskLine(t, staleWarnDays, staleCritDays))
	}
}

func TaskListWithDue(tasks []*model.Task, staleWarnDays, staleCritDays, dueSoonDays int) {
	for _, t := range tasks {
		fmt.Println(TaskLineWithDue(t, staleWarnDays, staleCritDays, dueSoonDays))
	}
}

func TaskJSON(tasks []*model.Task) error {
	type jsonTask struct {
		ID       int      `json:"id"`
		Title    string   `json:"title"`
		Status   string   `json:"status"`
		Priority string   `json:"priority,omitempty"`
		Tags     []string `json:"tags,omitempty"`
		Project  string   `json:"project,omitempty"`
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
			Project:  t.Project,
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

var (
	goalGreen  = color.New(color.FgGreen)
	goalYellow = color.New(color.FgYellow)
	goalRed    = color.New(color.FgRed, color.Bold)
)

func goalColor(daysLeft int) *color.Color {
	switch {
	case daysLeft > 14:
		return goalGreen
	case daysLeft >= 4:
		return goalYellow
	default:
		return goalRed
	}
}

func progressBar(current, target int, c *color.Color) string {
	const width = 12
	ratio := float64(current) / float64(target)
	if ratio > 1 {
		ratio = 1
	}
	filled := int(ratio * width)
	empty := width - filled
	return c.Sprint(strings.Repeat("█", filled)) + dimColor.Sprint(strings.Repeat("░", empty))
}

func GoalLine(g *model.Goal) string {
	days := g.DaysLeft()
	c := goalColor(days)

	var countdown string
	switch {
	case days < 0:
		countdown = c.Sprintf("OVERDUE %dd", -days)
	case days == 0:
		countdown = c.Sprint("due today")
	case days == 1:
		countdown = c.Sprint("1 day to go")
	default:
		countdown = c.Sprintf("%d days to go", days)
	}

	if !g.HasMetric() {
		return fmt.Sprintf("  %-24s %s", g.Title, countdown)
	}

	metricStr := fmt.Sprintf("%d/%d %s", g.Current, g.Target, g.Metric)
	bar := progressBar(g.Current, g.Target, c)
	return fmt.Sprintf("  %-24s %-18s %s  %s", g.Title, metricStr, bar, countdown)
}

func Goals(goals []model.Goal) {
	sort.SliceStable(goals, func(i, j int) bool {
		di := goals[i].DaysLeft()
		dj := goals[j].DaysLeft()
		return di < dj
	})
	for _, g := range goals {
		fmt.Println(GoalLine(&g))
	}
}

func NextActions(tasks []*model.Task) {
	for _, t := range tasks {
		action := t.NextAction()
		id := fmt.Sprintf("#%-3d", t.ID)
		fmt.Printf("%s %s → %s\n", boldColor.Sprint(id), t.Title, cyanColor.Sprint(action))
	}
}

var projectColor = color.New(color.FgMagenta, color.Faint)

func ProjectStatusColor(status string) *color.Color {
	switch status {
	case model.ProjectStatusNext:
		return cyanColor
	case model.ProjectStatusDone:
		return doneColor
	case model.ProjectStatusArchived:
		return dimColor
	default:
		return color.New(color.FgWhite)
	}
}

func ProjectAnnotation(project string) string {
	if project == "" {
		return ""
	}
	return "  " + projectColor.Sprintf("‹%s›", project)
}

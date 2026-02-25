package cmd

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/nickhudkins/tk/model"
	"github.com/nickhudkins/tk/render"
)

func dashboard() error {
	if cfg.Demo {
		fmt.Println("tk demo mode")
		return nil
	}

	bold := color.New(color.Bold)
	dim := color.New(color.Faint)
	focusHeadingColor := color.New(color.FgHiBlue, color.Bold)
	focusItemColor := color.New(color.FgHiBlue)

	// Focus
	content, _ := st.ReadFocus()
	if content != "" {
		lines := render.PickFocusItems(content, cfg.FocusItems)
		if len(lines) > 0 {
			focusHeadingColor.Println("🎯 Focus")
			for _, line := range lines {
				fmt.Printf("  %s\n", focusItemColor.Sprint(line))
			}
			fmt.Println()
		}
	}

	// Goals
	goalHeadingColor := color.New(color.FgGreen, color.Bold)
	goals, _ := st.ReadGoals()
	if len(goals) > 0 {
		goalHeadingColor.Println("🏁 Goals")
		render.Goals(goals)
		fmt.Println()
	}

	// Now tasks
	nowTasks, _ := st.List(func(t *model.Task) bool {
		return t.Status == model.StatusNow
	})
	sortTasksByPriority(nowTasks)

	bold.Printf("🔥 Now (%s)\n", time.Now().Format("Mon Jan 2"))
	if len(nowTasks) == 0 {
		dim.Println("  No tasks for now. Run `tk plan` to pick from next.")
	} else {
		for i, t := range nowTasks {
			fmt.Printf("  %d. %s\n", i+1, render.TaskLine(t, cfg.StaleWarnDays, cfg.StaleCritDays))
		}
	}
	fmt.Println()

	// Counts
	all, _ := st.List(nil)
	counts := map[string]int{}
	stale := 0
	for _, t := range all {
		counts[t.Status]++
		if t.IsActive() && t.DaysSinceUpdate() > cfg.StaleWarnDays {
			stale++
		}
	}

	dim.Printf("Inbox: %d  Todo: %d  Next: %d  Now: %d  Done: %d",
		counts[model.StatusInbox], counts[model.StatusTodo],
		counts[model.StatusNext], counts[model.StatusNow],
		counts[model.StatusDone])
	if stale > 0 {
		color.New(color.FgYellow, color.Faint).Printf("  Stale: %d", stale)
	}
	fmt.Println()

	return nil
}

func sortTasksByPriority(tasks []*model.Task) {
	sort.SliceStable(tasks, func(i, j int) bool {
		a, b := tasks[i], tasks[j]
		ar, br := dashboardPriorityRank(a.Priority), dashboardPriorityRank(b.Priority)
		if ar == br {
			return a.ID < b.ID
		}
		return ar < br
	})
}

func dashboardPriorityRank(priority string) int {
	switch strings.ToLower(priority) {
	case "p0":
		return 0
	case "p1":
		return 1
	case "p2":
		return 2
	case "":
		return 3
	default:
		return 99
	}
}

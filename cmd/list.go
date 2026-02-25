package cmd

import (
	"fmt"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/nickhudkins/tk/model"
	"github.com/nickhudkins/tk/render"
	"github.com/spf13/cobra"
)

var (
	listInbox       bool
	listAll         bool
	listStale       bool
	listDue         bool
	listStatus      string
	listSort        string
	listDesc        bool
	listShowUpdated bool
)

var listCmd = &cobra.Command{
	Use:     "list",
	Short:   "List tasks (plain output, use `tk pick` for interactive)",
	Aliases: []string{"ls", "l"},
	RunE: func(cmd *cobra.Command, args []string) error {
		filter := func(t *model.Task) bool {
			if listAll {
				return true
			}
			if listInbox {
				return t.Status == model.StatusInbox
			}
			if listStatus != "" {
				return t.Status == listStatus
			}
			if listStale {
				return t.IsActive() && t.DaysSinceUpdate() > cfg.StaleWarnDays
			}
			if listDue {
				return t.IsActive() && t.HasDue()
			}
			// Default: show all active (inbox, todo, next, now)
			return t.IsActive()
		}

		tasks, err := st.List(filter)
		if err != nil {
			return err
		}

		if listDue {
			sort.SliceStable(tasks, func(i, j int) bool {
				return tasks[i].DaysUntilDue() < tasks[j].DaysUntilDue()
			})
		} else if err := sortTasks(tasks, listSort, listDesc); err != nil {
			return err
		}

		if len(tasks) == 0 {
			fmt.Println("No tasks found.")
			return nil
		}

		if jsonOutput {
			return render.TaskJSON(tasks)
		}

		// Stale warning
		if !listStale && !listAll && listStatus == "" {
			staleCount := 0
			for _, t := range tasks {
				if t.DaysSinceUpdate() > cfg.StaleWarnDays {
					staleCount++
				}
			}
			if staleCount > 0 {
				fmt.Fprintf(os.Stderr, "⚠ %d stale tasks. Run `tk review` to clean up.\n\n", staleCount)
			}
		}

		if listShowUpdated {
			renderTaskListWithUpdated(tasks)
			return nil
		}

		render.TaskListWithDue(tasks, cfg.StaleWarnDays, cfg.StaleCritDays, cfg.DueSoonDays)
		return nil
	},
}

func sortTasks(tasks []*model.Task, field string, desc bool) error {
	field = strings.ToLower(strings.TrimSpace(field))
	if field == "" {
		field = "id"
	}

	switch field {
	case "id", "created", "updated", "title", "status", "priority":
	default:
		return fmt.Errorf("invalid sort field %q (use: id|created|updated|title|status|priority)", field)
	}

	compare := func(a, b *model.Task) int {
		switch field {
		case "id":
			return cmpInt(a.ID, b.ID)
		case "created":
			// Newest first by default for date-based sorting.
			return cmpTimeDesc(a.Created, b.Created)
		case "updated":
			// Newest first by default for date-based sorting.
			return cmpTimeDesc(a.Updated, b.Updated)
		case "title":
			at, bt := strings.ToLower(a.Title), strings.ToLower(b.Title)
			if at == bt {
				return cmpInt(a.ID, b.ID)
			}
			return cmpString(at, bt)
		case "status":
			ar, br := statusRank(a.Status), statusRank(b.Status)
			if ar == br {
				return cmpInt(a.ID, b.ID)
			}
			return cmpInt(ar, br)
		case "priority":
			ar, br := priorityRank(a.Priority), priorityRank(b.Priority)
			if ar == br {
				return cmpInt(a.ID, b.ID)
			}
			return cmpInt(ar, br)
		default:
			return cmpInt(a.ID, b.ID)
		}
	}

	sort.SliceStable(tasks, func(i, j int) bool {
		cmp := compare(tasks[i], tasks[j])
		if desc {
			cmp = -cmp
		}
		return cmp < 0
	})
	return nil
}

func statusRank(status string) int {
	switch strings.ToLower(status) {
	case model.StatusNow:
		return 0
	case model.StatusNext:
		return 1
	case model.StatusTodo:
		return 2
	case model.StatusInbox:
		return 3
	case model.StatusDone:
		return 4
	case model.StatusArchived:
		return 5
	default:
		return 99
	}
}

func priorityRank(priority string) int {
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

func cmpInt(a, b int) int {
	switch {
	case a < b:
		return -1
	case a > b:
		return 1
	default:
		return 0
	}
}

func cmpString(a, b string) int {
	switch {
	case a < b:
		return -1
	case a > b:
		return 1
	default:
		return 0
	}
}

func cmpTimeDesc(a, b time.Time) int {
	switch {
	case a.After(b):
		return -1
	case a.Before(b):
		return 1
	default:
		return 0
	}
}

func renderTaskListWithUpdated(tasks []*model.Task) {
	for _, t := range tasks {
		line := render.TaskLine(t, cfg.StaleWarnDays, cfg.StaleCritDays)
		days := t.DaysSinceUpdate()
		age := humanizeDaysAgo(days)
		meta := fmt.Sprintf("updated %s", age)
		fmt.Printf("%s  (%s)\n", line, updatedAgeColor(days).Sprint(meta))
	}
}

func updatedAgeColor(days int) *color.Color {
	switch {
	case days < 7:
		return color.New(color.FgYellow, color.Faint)
	case days < 30:
		return color.New(color.FgBlue, color.Faint)
	case days < 365:
		return color.New(color.FgMagenta, color.Faint)
	default:
		return color.New(color.FgRed, color.Faint)
	}
}

func humanizeDaysAgo(days int) string {
	switch {
	case days <= 0:
		return "today"
	case days == 1:
		return "1 day ago"
	case days < 7:
		return fmt.Sprintf("%d days ago", days)
	case days < 30:
		weeks := days / 7
		if weeks <= 1 {
			return "1 week ago"
		}
		return fmt.Sprintf("%d weeks ago", weeks)
	case days < 365:
		months := days / 30
		if months <= 1 {
			return "1 month ago"
		}
		return fmt.Sprintf("%d months ago", months)
	default:
		years := days / 365
		if years <= 1 {
			return "1 year ago"
		}
		return fmt.Sprintf("%d years ago", years)
	}
}

func init() {
	listCmd.Flags().BoolVar(&listInbox, "inbox", false, "Show only inbox items")
	listCmd.Flags().BoolVar(&listAll, "all", false, "Show all including done/archived")
	listCmd.Flags().BoolVar(&listStale, "stale", false, "Show stale tasks")
	listCmd.Flags().BoolVar(&listDue, "due", false, "Show tasks with due dates, sorted nearest first")
	listCmd.Flags().StringVar(&listStatus, "status", "", "Filter by status (inbox|todo|next|now|done|archived)")
	listCmd.Flags().StringVar(&listSort, "sort", "id", "Sort by (id|created|updated|title|status|priority)")
	listCmd.Flags().BoolVar(&listDesc, "desc", false, "Reverse sort order for selected field")
	listCmd.Flags().BoolVar(&listShowUpdated, "show-updated", false, "Show updated date and relative age")
	rootCmd.AddCommand(listCmd)
}

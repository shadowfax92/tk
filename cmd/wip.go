package cmd

import (
	"fmt"

	"github.com/fatih/color"
	"github.com/nickhudkins/tk/model"
)

func countByStatus(status string) (int, error) {
	tasks, err := st.List(func(t *model.Task) bool {
		return t.Status == status
	})
	if err != nil {
		return 0, err
	}
	return len(tasks), nil
}

func checkWIPLimit(targetStatus string) error {
	var limit int
	switch targetStatus {
	case model.StatusNow:
		limit = cfg.MaxNow
	case model.StatusNext:
		limit = cfg.MaxNext
	default:
		return nil
	}
	if limit <= 0 {
		return nil
	}
	count, err := countByStatus(targetStatus)
	if err != nil {
		return err
	}
	if count >= limit {
		return fmt.Errorf("%s is full (%d/%d) — finish or demote a task first", targetStatus, count, limit)
	}
	return nil
}

func wipRemaining(targetStatus string) int {
	var limit int
	switch targetStatus {
	case model.StatusNow:
		limit = cfg.MaxNow
	case model.StatusNext:
		limit = cfg.MaxNext
	default:
		return -1
	}
	if limit <= 0 {
		return -1
	}
	count, _ := countByStatus(targetStatus)
	return max(limit-count, 0)
}

func autoDemoteStale() {
	warn := color.New(color.FgYellow)
	demoted := false

	if cfg.NowStaleDays > 0 {
		nowTasks, _ := st.List(func(t *model.Task) bool {
			return t.Status == model.StatusNow
		})
		for _, t := range nowTasks {
			if t.DaysSinceUpdate() >= cfg.NowStaleDays {
				days := t.DaysSinceUpdate()
				t.Status = model.StatusNext
				if err := st.Save(t); err == nil {
					warn.Printf("  ↓ #%d now → next (untouched %dd): %s\n", t.ID, days, t.Title)
					demoted = true
				}
			}
		}
	}

	if cfg.NextStaleDays > 0 {
		nextTasks, _ := st.List(func(t *model.Task) bool {
			return t.Status == model.StatusNext
		})
		for _, t := range nextTasks {
			if t.DaysSinceUpdate() >= cfg.NextStaleDays {
				days := t.DaysSinceUpdate()
				t.Status = model.StatusTodo
				if err := st.Save(t); err == nil {
					warn.Printf("  ↓ #%d next → todo (untouched %dd): %s\n", t.ID, days, t.Title)
					demoted = true
				}
			}
		}
	}

	if demoted {
		fmt.Println()
	}
}

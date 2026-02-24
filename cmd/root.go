package cmd

import (
	"fmt"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/nickhudkins/tk/config"
	"github.com/nickhudkins/tk/model"
	"github.com/nickhudkins/tk/render"
	"github.com/nickhudkins/tk/store"
	"github.com/spf13/cobra"
)

var (
	jsonOutput bool
	cfg        *config.Config
	st         *store.Store
)

var rootCmd = &cobra.Command{
	Use:   "tk",
	Short: "Terminal task manager backed by markdown",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		var err error
		cfg, err = config.Load()
		if err != nil {
			return err
		}
		st, err = store.New(cfg)
		return err
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		return dashboard()
	},
}

func dashboard() error {
	bold := color.New(color.Bold)
	dim := color.New(color.Faint)

	// Focus
	content, _ := st.ReadFocus()
	if content != "" {
		bold.Println("Focus")
		lines := strings.Split(strings.TrimSpace(content), "\n")
		count := 0
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if line == "" || strings.HasPrefix(line, "#") {
				continue
			}
			if count >= 3 {
				break
			}
			fmt.Printf("  %s\n", line)
			count++
		}
		fmt.Println()
	}

	// Now tasks
	nowTasks, _ := st.List(func(t *model.Task) bool {
		return t.Status == model.StatusNow
	})

	bold.Printf("Now (%s)\n", time.Now().Format("Mon Jan 2"))
	if len(nowTasks) == 0 {
		dim.Println("  No tasks for now. Run `tk plan` to pick from next.")
	} else {
		for i, t := range nowTasks {
			fmt.Printf("  %d. %s\n", i+1, render.TaskLine(t, cfg.StaleWarnDays, cfg.StaleCritDays))
		}
	}
	fmt.Println()

	// Next tasks (this week)
	nextTasks, _ := st.List(func(t *model.Task) bool {
		return t.Status == model.StatusNext
	})
	if len(nextTasks) > 0 {
		bold.Println("Next")
		for _, t := range nextTasks {
			fmt.Printf("  %s\n", render.TaskLine(t, cfg.StaleWarnDays, cfg.StaleCritDays))
		}
		fmt.Println()
	}

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

func init() {
	rootCmd.PersistentFlags().BoolVar(&jsonOutput, "json", false, "Output as JSON")
}

func Execute() error {
	return rootCmd.Execute()
}

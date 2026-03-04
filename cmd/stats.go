package cmd

import (
	"fmt"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/nickhudkins/tk/model"
	"github.com/spf13/cobra"
)

var statsDays int

var statsCmd = &cobra.Command{
	Use:         "stats",
	Short:       "Show task statistics and trends",
	Annotations: map[string]string{"group": "Views:"},
	RunE: func(cmd *cobra.Command, args []string) error {
		all, err := st.List(nil)
		if err != nil {
			return err
		}
		if len(all) == 0 {
			fmt.Println("No tasks.")
			return nil
		}

		bold := color.New(color.Bold)
		dim := color.New(color.Faint)

		// --- Status breakdown ---
		counts := map[string]int{}
		stale, overdue := 0, 0
		var totalAge int
		activeCount := 0
		for _, t := range all {
			counts[t.Status]++
			if t.IsActive() {
				activeCount++
				totalAge += t.AgeDays()
				if t.DaysSinceUpdate() > cfg.StaleWarnDays {
					stale++
				}
				if t.HasDue() && t.DaysUntilDue() < 0 {
					overdue++
				}
			}
		}

		bold.Println("📊 Status")
		statuses := []struct {
			key   string
			label string
			c     *color.Color
		}{
			{model.StatusNow, "now", color.New(color.FgGreen, color.Bold)},
			{model.StatusNext, "next", color.New(color.FgCyan)},
			{model.StatusTodo, "todo", color.New(color.FgWhite)},
			{model.StatusInbox, "inbox", dim},
			{model.StatusBacklog, "backlog", dim},
			{model.StatusDone, "done", color.New(color.FgGreen)},
		}

		maxCount := 0
		for _, s := range statuses {
			if counts[s.key] > maxCount {
				maxCount = counts[s.key]
			}
		}

		barWidth := 20
		for _, s := range statuses {
			n := counts[s.key]
			if n == 0 {
				continue
			}
			filled := 0
			if maxCount > 0 {
				filled = n * barWidth / maxCount
			}
			if filled == 0 && n > 0 {
				filled = 1
			}
			bar := s.c.Sprint(strings.Repeat("█", filled)) + dim.Sprint(strings.Repeat("░", barWidth-filled))
			fmt.Printf("  %-8s %s %d\n", s.label, bar, n)
		}
		fmt.Println()

		// --- Completion trend ---
		now := time.Now()
		today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

		doneCounts := make([]int, statsDays)
		createdCounts := make([]int, statsDays)
		dayLabels := make([]string, statsDays)

		for i := 0; i < statsDays; i++ {
			day := today.AddDate(0, 0, -(statsDays - 1 - i))
			dayLabels[i] = day.Format("Mon")
			if i == statsDays-1 {
				dayLabels[i] = "Today"
			}
		}

		for _, t := range all {
			for i := 0; i < statsDays; i++ {
				day := today.AddDate(0, 0, -(statsDays - 1 - i))
				nextDay := day.AddDate(0, 0, 1)

				if t.Status == model.StatusDone && !t.Updated.Before(day) && t.Updated.Before(nextDay) {
					doneCounts[i]++
				}
				if !t.Created.Before(day) && t.Created.Before(nextDay) {
					createdCounts[i]++
				}
			}
		}

		totalDone := 0
		for _, c := range doneCounts {
			totalDone += c
		}
		totalCreated := 0
		for _, c := range createdCounts {
			totalCreated += c
		}

		bold.Printf("✅ Completed (%d last %d days)\n", totalDone, statsDays)
		renderVerticalBars(doneCounts, dayLabels, color.New(color.FgGreen))
		fmt.Println()

		bold.Printf("📥 Created (%d last %d days)\n", totalCreated, statsDays)
		renderVerticalBars(createdCounts, dayLabels, color.New(color.FgCyan))
		fmt.Println()

		// --- Summary ---
		avgAge := 0
		if activeCount > 0 {
			avgAge = totalAge / activeCount
		}

		bold.Print("📋 Summary  ")
		fmt.Printf("%d active", activeCount)
		if overdue > 0 {
			color.New(color.FgRed).Printf(" · %d overdue", overdue)
		}
		if stale > 0 {
			color.New(color.FgYellow).Printf(" · %d stale", stale)
		}
		dim.Printf(" · avg age %dd", avgAge)
		fmt.Println()

		return nil
	},
}

func renderVerticalBars(values []int, labels []string, c *color.Color) {
	maxVal := 0
	for _, v := range values {
		if v > maxVal {
			maxVal = v
		}
	}

	if maxVal == 0 {
		fmt.Println("  (none)")
		return
	}

	chartHeight := 8
	if maxVal < chartHeight {
		chartHeight = maxVal
	}

	colWidth := 5
	for _, l := range labels {
		if len(l)+1 > colWidth {
			colWidth = len(l) + 1
		}
	}

	dim := color.New(color.Faint)

	// Render bars top to bottom
	for row := chartHeight; row >= 1; row-- {
		threshold := float64(row) / float64(chartHeight) * float64(maxVal)
		fmt.Print("  ")
		for _, v := range values {
			if float64(v) >= threshold && v > 0 {
				bar := c.Sprint("██")
				pad := colWidth - 2
				fmt.Printf("%s%s", bar, strings.Repeat(" ", pad))
			} else {
				fmt.Print(strings.Repeat(" ", colWidth))
			}
		}
		fmt.Println()
	}

	// Labels
	fmt.Print("  ")
	for _, l := range labels {
		fmt.Printf("%-*s", colWidth, l)
	}
	fmt.Println()

	// Values
	fmt.Print("  ")
	for _, v := range values {
		dim.Printf("%-*d", colWidth, v)
	}
	fmt.Println()
}

func init() {
	statsCmd.Flags().IntVar(&statsDays, "days", 7, "Number of days to show in trends")
	rootCmd.AddCommand(statsCmd)
}

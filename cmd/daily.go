package cmd

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/nickhudkins/tk/model"
	"github.com/spf13/cobra"
)

var dailyCmd = &cobra.Command{
	Use:         "daily [date]",
	Short:       "Open today's daily note",
	Aliases:     []string{"dn"},
	Annotations: map[string]string{"group": "Organize:"},
	Args:        cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		day := time.Now()
		if len(args) == 1 {
			parsed, err := parseDay(args[0])
			if err != nil {
				return err
			}
			day = parsed
		}

		if err := st.EnsureDailyDir(); err != nil {
			return err
		}

		path := st.DailyPath(day)
		isNew := false
		if _, err := os.Stat(path); os.IsNotExist(err) {
			isNew = true
			template, err := generateDailyTemplate(day)
			if err != nil {
				return err
			}
			if err := os.WriteFile(path, []byte(template), 0644); err != nil {
				return err
			}
		}

		c := exec.Command(cfg.Editor, path)
		c.Stdin = os.Stdin
		c.Stdout = os.Stdout
		c.Stderr = os.Stderr
		if err := c.Run(); err != nil {
			return err
		}

		// If new and user left it empty/unchanged, clean up
		if isNew {
			data, _ := os.ReadFile(path)
			if isTemplateUnchanged(string(data)) {
				os.Remove(path)
				fmt.Println("Empty daily note removed.")
			}
		}

		return nil
	},
}

var dailyWeekCmd = &cobra.Command{
	Use:   "week",
	Short: "Show last 7 days of daily notes",
	RunE: func(cmd *cobra.Command, args []string) error {
		dim := color.New(color.Faint)
		bold := color.New(color.Bold)
		today := time.Now()
		found := 0

		for i := 6; i >= 0; i-- {
			day := today.AddDate(0, 0, -i)
			content, err := st.ReadDaily(day)
			if err != nil || content == "" {
				continue
			}
			found++
			label := day.Format("Mon Jan 2")
			if i == 0 {
				label += " (today)"
			}
			bold.Printf("── %s ──\n", label)
			// Print non-empty lines, skip comment-only lines at top
			scanner := bufio.NewScanner(strings.NewReader(content))
			for scanner.Scan() {
				line := scanner.Text()
				if strings.HasPrefix(strings.TrimSpace(line), "#") && strings.HasPrefix(line, "# ") {
					bold.Println(line)
				} else {
					fmt.Println(line)
				}
			}
			dim.Println()
		}

		if found == 0 {
			fmt.Println("No daily notes in the last 7 days.")
		}
		return nil
	},
}

func parseDay(s string) (time.Time, error) {
	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	switch strings.ToLower(s) {
	case "today":
		return today, nil
	case "yesterday", "yday":
		return today.AddDate(0, 0, -1), nil
	}

	// Try YYYY-MM-DD
	if t, err := time.Parse("2006-01-02", s); err == nil {
		return t, nil
	}

	// Try weekday name (last occurrence)
	weekdays := map[string]time.Weekday{
		"mon": time.Monday, "monday": time.Monday,
		"tue": time.Tuesday, "tuesday": time.Tuesday,
		"wed": time.Wednesday, "wednesday": time.Wednesday,
		"thu": time.Thursday, "thursday": time.Thursday,
		"fri": time.Friday, "friday": time.Friday,
		"sat": time.Saturday, "saturday": time.Saturday,
		"sun": time.Sunday, "sunday": time.Sunday,
	}
	if wd, ok := weekdays[strings.ToLower(s)]; ok {
		d := today
		for i := 0; i < 7; i++ {
			d = d.AddDate(0, 0, -1)
			if d.Weekday() == wd {
				return d, nil
			}
		}
	}

	return time.Time{}, fmt.Errorf("invalid date %q (use: today, yesterday, YYYY-MM-DD, or weekday name)", s)
}

func generateDailyTemplate(day time.Time) (string, error) {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("# %s\n\n", day.Format("Mon Jan 2, 2006")))

	// Now tasks
	nowTasks, err := st.List(func(t *model.Task) bool {
		return t.Status == model.StatusNow
	})
	if err == nil && len(nowTasks) > 0 {
		sortTasksByPriority(nowTasks)
		b.WriteString("## Now\n")
		for _, t := range nowTasks {
			b.WriteString(fmt.Sprintf("- #%d %s", t.ID, t.Title))
			if t.Priority != "" {
				b.WriteString(" " + t.Priority)
			}
			if t.Project != "" {
				b.WriteString(fmt.Sprintf(" <%s>", t.Project))
			}
			b.WriteString("\n")
		}
		b.WriteString("\n")
	}

	// Overdue tasks
	overdue, err := st.List(func(t *model.Task) bool {
		return t.IsActive() && t.HasDue() && t.DaysUntilDue() < 0
	})
	if err == nil && len(overdue) > 0 {
		b.WriteString("## Overdue\n")
		for _, t := range overdue {
			b.WriteString(fmt.Sprintf("- #%d %s (overdue %dd)\n", t.ID, t.Title, -t.DaysUntilDue()))
		}
		b.WriteString("\n")
	}

	b.WriteString("## Notes\n\n")

	return b.String(), nil
}

func isTemplateUnchanged(content string) bool {
	for _, line := range strings.Split(content, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, "- #") {
			continue
		}
		return false
	}
	return true
}

func dailyDashboard() {
	dim := color.New(color.Faint)
	headingColor := color.New(color.FgHiBlue, color.Bold)

	if !st.DailyExists(time.Now()) {
		headingColor.Print("📝 Daily  ")
		dim.Println("no note yet — `tk daily` to start")
		fmt.Println()
		return
	}

	content, _ := st.ReadDaily(time.Now())
	headingColor.Println("📝 Daily")

	scanner := bufio.NewScanner(strings.NewReader(content))
	shown := 0
	for scanner.Scan() {
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "# ") {
			continue
		}
		if strings.HasPrefix(trimmed, "## ") {
			color.New(color.Bold).Printf("  %s\n", trimmed)
		} else {
			fmt.Printf("  %s\n", trimmed)
		}
		shown++
		if shown >= 5 {
			dim.Println("  ...")
			break
		}
	}

	fmt.Println()
}

func init() {
	dailyCmd.AddCommand(dailyWeekCmd)
	rootCmd.AddCommand(dailyCmd)
}

package cmd

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/fatih/color"
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
	case "tomorrow", "tmrw":
		return today.AddDate(0, 0, 1), nil
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
	b.WriteString(fmt.Sprintf("# %s, %s\n\n", day.Format("January 2"), day.Format("Monday")))
	b.WriteString("## Tasks\n\n")
	b.WriteString("## Notes\n\n")
	b.WriteString("## Exercise\n> Do graham weaver to set the goals\n")

	return b.String(), nil
}

func isTemplateUnchanged(content string) bool {
	for _, line := range strings.Split(content, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, ">") {
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

	// Collect checkbox tasks
	var tasks []string
	total, done := 0, 0
	scanner := bufio.NewScanner(strings.NewReader(content))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "- [x] ") || strings.HasPrefix(line, "- [X] ") {
			total++
			done++
			tasks = append(tasks, line)
		} else if strings.HasPrefix(line, "- [ ] ") {
			total++
			tasks = append(tasks, line)
		}
	}

	if total > 0 {
		headingColor.Printf("📝 Daily  ")
		dim.Printf("[%d/%d]\n", done, total)
		for _, t := range tasks {
			fmt.Printf("  %s\n", t)
		}
	} else {
		headingColor.Printf("📝 Daily  ")
		dim.Println("no tasks yet")
	}

	fmt.Println()
}

func init() {
	dailyCmd.AddCommand(dailyWeekCmd)
	rootCmd.AddCommand(dailyCmd)
}

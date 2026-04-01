package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var weeklyCmd = &cobra.Command{
	Use:         "weekly [week]",
	Short:       "Open this week's note",
	Aliases:     []string{"wk"},
	Annotations: map[string]string{"group": "Organize:"},
	Args:        cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		day := time.Now()
		if len(args) == 1 {
			parsed, err := parseWeek(args[0])
			if err != nil {
				return err
			}
			day = parsed
		}

		if err := st.EnsureWeeklyDir(); err != nil {
			return err
		}

		path := st.WeeklyPath(day)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			template := generateWeeklyTemplate(day)
			if err := os.WriteFile(path, []byte(template), 0644); err != nil {
				return err
			}
		}

		c := exec.Command(cfg.Editor, path)
		c.Stdin = os.Stdin
		c.Stdout = os.Stdout
		c.Stderr = os.Stderr
		return c.Run()
	},
}

func parseWeek(s string) (time.Time, error) {
	now := time.Now()
	switch strings.ToLower(s) {
	case "last", "prev":
		return now.AddDate(0, 0, -7), nil
	case "next":
		return now.AddDate(0, 0, 7), nil
	}

	// Try YYYY-Www format (e.g. 2026-W14)
	var year, week int
	if _, err := fmt.Sscanf(strings.ToUpper(s), "%d-W%d", &year, &week); err == nil {
		if week < 1 || week > 53 {
			return time.Time{}, fmt.Errorf("invalid week number %d", week)
		}
		jan4 := time.Date(year, 1, 4, 0, 0, 0, 0, now.Location())
		_, jan4Week := jan4.ISOWeek()
		return weekMonday(jan4).AddDate(0, 0, (week-jan4Week)*7), nil
	}

	return time.Time{}, fmt.Errorf("invalid week %q (use: last, next, or YYYY-Www)", s)
}

func weekMonday(t time.Time) time.Time {
	wd := t.Weekday()
	if wd == time.Sunday {
		return t.AddDate(0, 0, -6)
	}
	return t.AddDate(0, 0, -int(wd-time.Monday))
}

func generateWeeklyTemplate(day time.Time) string {
	_, week := day.ISOWeek()
	monday := weekMonday(day)
	sunday := monday.AddDate(0, 0, 6)

	var b strings.Builder
	b.WriteString(fmt.Sprintf("# Week %d — %s – %s\n\n",
		week,
		monday.Format("Jan 2"),
		sunday.Format("Jan 2, 2006"),
	))
	b.WriteString("## Goals\n\n")
	b.WriteString("## Notes\n\n")
	b.WriteString("## Reflections\n\n")

	return b.String()
}

func weeklyDashboard() {
	dim := color.New(color.Faint)
	headingColor := color.New(color.FgHiBlue, color.Bold)

	now := time.Now()
	_, week := now.ISOWeek()

	if !st.WeeklyExists(now) {
		headingColor.Print("📓 Weekly  ")
		dim.Printf("no note yet — `tk weekly` to start\n")
		fmt.Println()
		return
	}

	headingColor.Print("📓 Weekly  ")
	dim.Printf("W%d ✓\n", week)
	fmt.Println()
}

func init() {
	rootCmd.AddCommand(weeklyCmd)
}

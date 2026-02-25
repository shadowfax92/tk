package cmd

import (
	"fmt"
	"sort"
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

const usageTemplateWithAliases = `Usage:{{if .Runnable}}
  {{.UseLine}}{{end}}{{if .HasAvailableSubCommands}}
  {{.CommandPath}} [command]{{end}}{{if gt (len .Aliases) 0}}

Aliases:
  {{.NameAndAliases}}{{end}}{{if .HasExample}}

Examples:
{{.Example}}{{end}}{{if .HasAvailableSubCommands}}

Available Commands:{{range .Commands}}{{if (or .IsAvailableCommand (eq .Name "help"))}}
  {{rpad .Name .NamePadding }} {{.Short}}{{if .Aliases}} (aliases: {{range $i, $a := .Aliases}}{{if $i}}, {{end}}{{$a}}{{end}}){{end}}{{end}}{{end}}{{end}}{{if .HasAvailableLocalFlags}}

Flags:
{{.LocalFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}{{if .HasAvailableInheritedFlags}}

Global Flags:
{{.InheritedFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}{{if .HasHelpSubCommands}}

Additional help topics:{{range .Commands}}{{if .IsAdditionalHelpTopicCommand}}
  {{rpad .CommandPath .CommandPathPadding}} {{.Short}}{{end}}{{end}}{{end}}{{if .HasAvailableSubCommands}}

Use "{{.CommandPath}} [command] --help" for more information about a command.{{end}}
`

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

	// Next tasks (this week)
	nextTasks, _ := st.List(func(t *model.Task) bool {
		return t.Status == model.StatusNext
	})
	sortTasksByPriority(nextTasks)
	if len(nextTasks) > 0 {
		bold.Println("🔜 Next")
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

func init() {
	rootCmd.PersistentFlags().BoolVar(&jsonOutput, "json", false, "Output as JSON")
	rootCmd.SetUsageTemplate(usageTemplateWithAliases)
}

func Execute() error {
	return rootCmd.Execute()
}

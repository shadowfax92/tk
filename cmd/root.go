package cmd

import (
	"fmt"
	"strings"

	"github.com/fatih/color"
	"github.com/nickhudkins/tk/config"
	"github.com/nickhudkins/tk/store"
	"github.com/spf13/cobra"
)

var (
	jsonOutput bool
	cfg        *config.Config
	st         *store.Store
)

const (
	groupCore      = "core"
	groupStatus    = "status"
	groupView      = "view"
	groupInteract  = "interact"
	groupOrganize  = "organize"
)

var rootCmd = &cobra.Command{
	Use:   "tk",
	Short: "Terminal task manager backed by markdown",
	Long:  "A terminal-first task manager backed by plain markdown files.\nBare `tk` shows your dashboard. Tasks flow: inbox → todo → next → now → done.",
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

const usageTemplateWithAliases = `{{helpHeader "Usage:"}}{{if .Runnable}}
  {{.UseLine}}{{end}}{{if .HasAvailableSubCommands}}
  {{.CommandPath}} [command]{{end}}{{if gt (len .Aliases) 0}}

{{helpHeader "Aliases:"}}
  {{.NameAndAliases}}{{end}}{{if .HasExample}}

{{helpHeader "Examples:"}}
{{.Example}}{{end}}{{if .HasAvailableSubCommands}}
{{groupedHelp .}}{{end}}{{if .HasAvailableLocalFlags}}

{{helpHeader "Flags:"}}
{{.LocalFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}{{if .HasAvailableInheritedFlags}}

{{helpHeader "Global Flags:"}}
{{.InheritedFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}{{if .HasAvailableSubCommands}}

{{helpHint (printf "Use \"%s [command] --help\" for more information." .CommandPath)}}{{end}}
`

var (
	helpHeaderColor = color.New(color.Bold, color.FgCyan)
	helpCmdColor    = color.New(color.FgHiGreen)
	helpAliasColor  = color.New(color.FgYellow)
	helpHintColor   = color.New(color.Faint)
)

func helpHeader(s string) string {
	return helpHeaderColor.Sprint(s)
}

func helpCmdCol(s string) string {
	return helpCmdColor.Sprint(s)
}

func helpAliases(aliases []string) string {
	return helpAliasColor.Sprintf("(aliases: %s)", strings.Join(aliases, ", "))
}

func helpHint(s string) string {
	return helpHintColor.Sprint(s)
}

var groupOrder = []string{
	"Tasks:",
	"Status:",
	"Views:",
	"Interactive:",
	"Organize:",
	"Other:",
}

func groupedHelp(cmd *cobra.Command) string {
	groups := map[string][]*cobra.Command{}
	for _, c := range cmd.Commands() {
		if !c.IsAvailableCommand() && c.Name() != "help" {
			continue
		}
		g := c.Annotations["group"]
		if g == "" {
			g = "Other:"
		}
		groups[g] = append(groups[g], c)
	}

	var b strings.Builder
	for _, name := range groupOrder {
		cmds, ok := groups[name]
		if !ok {
			continue
		}
		b.WriteString("\n" + helpHeader(name) + "\n")
		for _, c := range cmds {
			line := "  " + helpCmdCol(fmt.Sprintf("%-12s", c.Name())) + " " + c.Short
			if len(c.Aliases) > 0 {
				line += " " + helpAliases(c.Aliases)
			}
			b.WriteString(line + "\n")
		}
	}
	return b.String()
}

func init() {
	cobra.AddTemplateFunc("helpHeader", helpHeader)
	cobra.AddTemplateFunc("helpCmdCol", helpCmdCol)
	cobra.AddTemplateFunc("helpAliases", helpAliases)
	cobra.AddTemplateFunc("helpHint", helpHint)
	cobra.AddTemplateFunc("groupedHelp", groupedHelp)

	rootCmd.PersistentFlags().BoolVar(&jsonOutput, "json", false, "Output as JSON")
	rootCmd.SetUsageTemplate(usageTemplateWithAliases)
}

func Execute() error {
	return rootCmd.Execute()
}

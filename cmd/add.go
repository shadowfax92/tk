package cmd

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/nickhudkins/tk/model"
	"github.com/spf13/cobra"
)

var addDesc string
var addNow bool
var addNext bool
var addTags []string
var addDue string

var addCmd = &cobra.Command{
	Use:         "add [title...]",
	Short:       "Add a task (defaults to inbox)",
	Aliases:     []string{"a"},
	Annotations: map[string]string{"group": "Tasks:"},
	Args:    cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		status := model.StatusInbox
		if addNow {
			status = model.StatusNow
		}
		if addNext {
			status = model.StatusNext
		}
		tags := normalizeTags(addTags)

		due, err := parseDue(addDue)
		if err != nil {
			return err
		}

		title := strings.Join(args, " ")
		t, err := st.AddFull(title, addDesc, status, tags, due)
		if err != nil {
			return err
		}
		msg := fmt.Sprintf("Added #%d: %s [%s]", t.ID, t.Title, t.Status)
		if t.HasDue() {
			msg += fmt.Sprintf(" (due %s)", t.Due)
		}
		fmt.Println(msg)
		return nil
	},
}

func parseDue(raw string) (string, error) {
	if raw == "" {
		return "", nil
	}
	if days, err := strconv.Atoi(raw); err == nil {
		return time.Now().AddDate(0, 0, days).Format("2006-01-02"), nil
	}
	if _, err := time.Parse("2006-01-02", raw); err == nil {
		return raw, nil
	}
	return "", fmt.Errorf("invalid --due value %q (use number of days or YYYY-MM-DD)", raw)
}

func init() {
	addCmd.Flags().StringVarP(&addDesc, "desc", "d", "", "Task description/body")
	addCmd.Flags().BoolVar(&addNow, "now", false, "Add task directly to now")
	addCmd.Flags().BoolVar(&addNext, "next", false, "Add task directly to next")
	addCmd.Flags().StringSliceVar(&addTags, "tags", nil, "Add tag(s), e.g. --tags \"#cli,#x\" or --tags cli --tags x")
	addCmd.Flags().StringVar(&addDue, "due", "", "Due date (number of days or YYYY-MM-DD)")
	addCmd.MarkFlagsMutuallyExclusive("now", "next")
	rootCmd.AddCommand(addCmd)
}

func normalizeTags(inputs []string) []string {
	seen := map[string]bool{}
	var out []string

	for _, input := range inputs {
		parts := strings.FieldsFunc(input, func(r rune) bool {
			return r == ',' || r == ' ' || r == '\t' || r == '\n'
		})
		for _, p := range parts {
			tag := strings.TrimSpace(strings.TrimPrefix(p, "#"))
			if tag == "" || seen[tag] {
				continue
			}
			seen[tag] = true
			out = append(out, tag)
		}
	}

	return out
}

package cmd

import (
	"fmt"
	"strings"

	"github.com/nickhudkins/tk/model"
	"github.com/spf13/cobra"
)

var addDesc string
var addNow bool
var addNext bool
var addTags []string

var addCmd = &cobra.Command{
	Use:     "add [title...]",
	Short:   "Add a task (defaults to inbox)",
	Aliases: []string{"a"},
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

		title := strings.Join(args, " ")
		t, err := st.AddWithStatusAndTags(title, addDesc, status, tags)
		if err != nil {
			return err
		}
		fmt.Printf("Added #%d: %s [%s]\n", t.ID, t.Title, t.Status)
		return nil
	},
}

func init() {
	addCmd.Flags().StringVarP(&addDesc, "desc", "d", "", "Task description/body")
	addCmd.Flags().BoolVar(&addNow, "now", false, "Add task directly to now")
	addCmd.Flags().BoolVar(&addNext, "next", false, "Add task directly to next")
	addCmd.Flags().StringSliceVar(&addTags, "tags", nil, "Add tag(s), e.g. --tags \"#cli,#x\" or --tags cli --tags x")
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

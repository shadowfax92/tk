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

var addCmd = &cobra.Command{
	Use:   "add [title...]",
	Short: "Add a task (defaults to inbox)",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		status := model.StatusInbox
		if addNow {
			status = model.StatusNow
		}
		if addNext {
			status = model.StatusNext
		}

		title := strings.Join(args, " ")
		t, err := st.AddWithStatus(title, addDesc, status)
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
	addCmd.MarkFlagsMutuallyExclusive("now", "next")
	rootCmd.AddCommand(addCmd)
}

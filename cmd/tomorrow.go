package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"time"

	"github.com/spf13/cobra"
)

var tomorrowCmd = &cobra.Command{
	Use:         "tomorrow",
	Short:       "Open tomorrow's daily note",
	Aliases:     []string{"tm"},
	Annotations: map[string]string{"group": "Organize:"},
	RunE: func(cmd *cobra.Command, args []string) error {
		day := time.Now().AddDate(0, 0, 1)

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

func init() {
	rootCmd.AddCommand(tomorrowCmd)
}

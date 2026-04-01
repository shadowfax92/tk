package cmd

import (
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
		if _, err := os.Stat(path); os.IsNotExist(err) {
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
		return c.Run()
	},
}

func init() {
	rootCmd.AddCommand(tomorrowCmd)
}

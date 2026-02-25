package cmd

import (
	"fmt"
	"strings"

	"github.com/nickhudkins/tk/config"
	"github.com/spf13/cobra"
)

var demoCmd = &cobra.Command{
	Use:   "demo [on|off]",
	Short: "Show or set dashboard demo mode",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			if cfg.Demo {
				fmt.Println("Demo mode: on")
			} else {
				fmt.Println("Demo mode: off")
			}
			return nil
		}

		var enabled bool
		switch strings.ToLower(strings.TrimSpace(args[0])) {
		case "on":
			enabled = true
		case "off":
			enabled = false
		default:
			return fmt.Errorf("invalid value %q (use: on|off)", args[0])
		}

		cfg.Demo = enabled
		if err := config.Save(cfg); err != nil {
			return err
		}

		if enabled {
			fmt.Println("Demo mode enabled.")
		} else {
			fmt.Println("Demo mode disabled.")
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(demoCmd)
}

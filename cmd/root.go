package cmd

import (
	"github.com/nickhudkins/tk/config"
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
}

func init() {
	rootCmd.PersistentFlags().BoolVar(&jsonOutput, "json", false, "Output as JSON")
}

func Execute() error {
	return rootCmd.Execute()
}

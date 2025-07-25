package cmd

import (
	"github.com/oschrenk/infuse/config"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(setupCmd)
}

var setupCmd = &cobra.Command{
	Use:   "setup",
	Short: "Setup command",
	Long:  "Setup command for infuse",
	RunE: func(cmd *cobra.Command, args []string) error {
		config.Setup()
		return nil
	},
}

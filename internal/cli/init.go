package cli

import (
	"github.com/oschrenk/infuse/internal/config"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(initCmd)
}

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize config file",
	Long:  "Initialize config file at $XDG_CONFIG_HOME/infuse/config.toml",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		return config.Init()
	},
}

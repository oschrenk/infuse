package cli

import (
	"fmt"

	"github.com/oschrenk/infuse/internal/config"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(dirCmd)
}

var dirCmd = &cobra.Command{
	Use:   "dir",
	Short: "Print the infuse repository path",
	Long:  "Print the base directory of the infuse repository. Use with: cd $(infuse dir)",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return fmt.Errorf("loading config: %w", err)
		}

		if cfg.Repo.Path == "" {
			return fmt.Errorf("repository path not configured")
		}

		fmt.Print(cfg.Repo.Path)
		return nil
	},
}

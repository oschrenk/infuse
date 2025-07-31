package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/oschrenk/infuse/config"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(cdCmd)
}

var cdCmd = &cobra.Command{
	Use:   "cd",
	Short: "Print the infuse repository path",
	Long:  "Load config and print the base directory of the infuse repository",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Load config
		cfg, err := config.Load()
		if err != nil {
			return fmt.Errorf("error loading config: %w", err)
		}

		// Check if repo path is configured
		if cfg.Repo.Path == "" {
			return fmt.Errorf("infuse repository path not configured")
		}

		// Change to the directory
		if err := os.Chdir(cfg.Repo.Path); err != nil {
			return fmt.Errorf("failed to change directory to %s: %w", cfg.Repo.Path, err)
		}

		fmt.Printf("Changed directory to: %s\n", cfg.Repo.Path)
		return nil
	},
}

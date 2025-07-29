package cmd

import (
	"fmt"
	"path/filepath"

	git "github.com/oschrenk/infuse/git"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(trackCmd)
}

var trackCmd = &cobra.Command{
	Use:   "track [path]",
	Short: "Track command",
	Long:  "Track command for infuse",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		path := args[0]
		absPath, err := filepath.Abs(path)
		if err != nil {
			return err
		}

		repo, err := git.Load(absPath)
		if err != nil {
			fmt.Printf("Path %s is not within a git repository\n", path)
			return nil
		}

		// Get remote info string
		remoteStr, err := repo.GetNormalizedRemote()
		if err != nil {
			return err
		}

		// Check if path is tracked
		isTracked, err := repo.IsTracked(absPath)
		if err != nil {
			return err
		}

		if isTracked {
			fmt.Printf("hello %s (git repository detected, tracked%s)\n", path, remoteStr)
		} else {
			fmt.Printf("hello %s (git repository detected, untracked%s)\n", path, remoteStr)
		}

		return nil
	},
}

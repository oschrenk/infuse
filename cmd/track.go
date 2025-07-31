package cmd

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/oschrenk/infuse/config"
	"github.com/oschrenk/infuse/core"
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
		// check path
		path := args[0]
		absPath, err := filepath.Abs(path)
		if err != nil {
			return err
		}

		// Check if path is in repo
		repo, err := git.Load(absPath)
		if err != nil {
			fmt.Printf("Path %s is not within a git repository\n", path)
			return nil
		}

		// Check if path is tracked
		isTracked, err := repo.IsTracked(absPath)
		if err != nil {
			return err
		}

		// Can't infuse if tracked
		if isTracked {
			return errors.New("Can't infuse tracked path: file is already tracked by git")
		}

		// Check if file exists
		_, err = os.Stat(absPath)
		exists := !os.IsNotExist(err)

		// Load config and construct a new infuse object
		cfg, err := config.Load()
		if err != nil {
			fmt.Printf("Error loading config: %v\n", err)
			return err
		}
		infuse := core.New(cfg)

		// Get remote info string
		remoteStr, err := repo.GetNormalizedRemote()
		if err != nil {
			fmt.Printf("No remote configured", path)
			return err
		}

		if exists {
			fmt.Printf("Moving %s", path)
			infuse.Move(absPath, remoteStr)
		}

		if isTracked {
			fmt.Printf("hello %s (git repository detected, tracked%s)\n", path, remoteStr)
		} else {
			fmt.Printf("hello %s (git repository detected, untrackedfd%s)\n", path, remoteStr)
		}

		return nil
	},
}

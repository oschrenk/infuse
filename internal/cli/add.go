package cli

import (
	"fmt"
	"os"
	"path/filepath"

	gogit "github.com/go-git/go-git/v5"
	"github.com/oschrenk/infuse/internal/config"
	"github.com/oschrenk/infuse/internal/git"
	"github.com/oschrenk/infuse/internal/infuse"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(addCmd)
}

var addCmd = &cobra.Command{
	Use:   "add [path...]",
	Short: "Add files to infuse",
	Long:  "Move files to the infuse repository, create symlinks, and exclude from git",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		// Load and validate config
		cfg, err := config.Load()
		if err != nil {
			return fmt.Errorf("loading config: %w\nRun: infuse init", err)
		}

		if cfg.Repo.Path == "" {
			return fmt.Errorf("repository path not configured\nRun: infuse init")
		}

		// Check $INFUSE_REPO is a git repository
		infuseRepoPath := cfg.Repo.Path
		if _, err := gogit.PlainOpen(infuseRepoPath); err != nil {
			return fmt.Errorf("infuse repository is not set up\nRun: infuse setup")
		}

		// Check working directory is in a git repository
		wd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("getting working directory: %w", err)
		}

		repo, err := git.Load(wd)
		if err != nil {
			return fmt.Errorf("not a git repository")
		}

		// Check remote is configured
		remote, err := repo.GetNormalizedRemote()
		if err != nil {
			return fmt.Errorf("getting remote: %w", err)
		}
		if remote == "" {
			return fmt.Errorf("no remote configured")
		}

		// Check <host>/<owner>/<repo> directory exists in $INFUSE_REPO
		remoteDir := filepath.Join(infuseRepoPath, remote)
		if _, err := os.Stat(remoteDir); os.IsNotExist(err) {
			return fmt.Errorf("directory %s does not exist\nRun: infuse setup", remoteDir)
		}

		repoRoot, err := repo.Root()
		if err != nil {
			return fmt.Errorf("getting repository root: %w", err)
		}

		inf := infuse.New(cfg)

		// Process each path
		for _, path := range args {
			absPath, err := filepath.Abs(path)
			if err != nil {
				return fmt.Errorf("resolving path %s: %w", path, err)
			}

			if err := addFile(inf, repo, repoRoot, remote, absPath, path); err != nil {
				return err
			}
		}

		return nil
	},
}

func addFile(inf *infuse.Infuse, repo *git.Repo, repoRoot, remote, absPath, displayPath string) error {
	// Check file exists
	fi, err := os.Lstat(absPath)
	if err != nil {
		return fmt.Errorf("%s: file not found", displayPath)
	}

	// Check if symlink
	if fi.Mode()&os.ModeSymlink != 0 {
		// Check if already tracked by infuse
		tracked, err := inf.IsTracked(absPath)
		if err != nil {
			return fmt.Errorf("%s: %w", displayPath, err)
		}
		if tracked {
			fmt.Printf("%s: already tracked by infuse\n", displayPath)
			return nil
		}
		return fmt.Errorf("%s: is a symlink", displayPath)
	}

	// Check not tracked by git
	isTracked, err := repo.IsTracked(absPath)
	if err != nil {
		return fmt.Errorf("%s: %w", displayPath, err)
	}
	if isTracked {
		return fmt.Errorf("%s: tracked by git", displayPath)
	}

	// Move file and create symlink
	relPath, err := inf.Add(absPath, repoRoot, remote)
	if err != nil {
		return fmt.Errorf("%s: %w", displayPath, err)
	}

	// Add to .git/info/exclude
	if err := repo.AddExclude(relPath); err != nil {
		return fmt.Errorf("%s: adding exclude entry: %w", displayPath, err)
	}

	fmt.Printf("%s: added\n", displayPath)
	return nil
}

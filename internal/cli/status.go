package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"text/tabwriter"

	gogit "github.com/go-git/go-git/v5"
	"github.com/oschrenk/infuse/internal/config"
	"github.com/oschrenk/infuse/internal/git"
	"github.com/oschrenk/infuse/internal/infuse"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(statusCmd)
}

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show managed files",
	Long:  "Show all files managed by infuse for the current working repository",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return fmt.Errorf("loading config: %w\nRun: infuse init", err)
		}

		if cfg.Repo.Path == "" {
			return fmt.Errorf("repository path not configured\nRun: infuse init")
		}

		infuseRepoPath := cfg.Repo.Path

		// Load infuse repo to check git status
		infuseGitRepo, err := gogit.PlainOpen(infuseRepoPath)
		if err != nil {
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

		remote, err := repo.GetNormalizedRemote()
		if err != nil {
			return fmt.Errorf("getting remote: %w", err)
		}
		if remote == "" {
			return fmt.Errorf("no remote configured")
		}

		repoRoot, err := repo.Root()
		if err != nil {
			return fmt.Errorf("getting repository root: %w", err)
		}

		inf := infuse.New(cfg)

		files, err := inf.ListFiles(remote)
		if err != nil {
			return fmt.Errorf("listing files: %w", err)
		}

		if len(files) == 0 {
			fmt.Println("No files managed by infuse")
			return nil
		}

		// Get infuse repo worktree status for modification checks
		wt, err := infuseGitRepo.Worktree()
		if err != nil {
			return fmt.Errorf("reading infuse repo worktree: %w", err)
		}

		wtStatus, err := wt.Status()
		if err != nil {
			return fmt.Errorf("reading infuse repo status: %w", err)
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "PATH\tSYMLINK\tGIT")
		for _, f := range files {
			symlinkStatus := checkSymlink(repoRoot, f)
			modifiedStatus := checkModified(infuseRepoPath, f, wtStatus)
			fmt.Fprintf(w, "%s\t%s\t%s\n", f.RelPath, symlinkStatus, modifiedStatus)
		}
		w.Flush()

		return nil
	},
}

func checkSymlink(repoRoot string, f infuse.FileStatus) string {
	symlinkPath := filepath.Join(repoRoot, f.RelPath)

	_, err := os.Readlink(symlinkPath)
	if err != nil {
		// Not a symlink or doesn't exist
		if _, statErr := os.Lstat(symlinkPath); os.IsNotExist(statErr) {
			return "missing"
		}
		return "not linked"
	}

	// Check if symlink points to the correct infuse path
	resolvedTarget, err := filepath.EvalSymlinks(symlinkPath)
	if err != nil {
		return "broken"
	}

	resolvedInfuse, err := filepath.EvalSymlinks(f.InfusePath)
	if err != nil {
		return "broken"
	}

	if resolvedTarget == resolvedInfuse {
		return "ok"
	}
	return "mismatch"
}

func checkModified(infuseRepoPath string, f infuse.FileStatus, wtStatus gogit.Status) string {
	relToInfuse, err := filepath.Rel(infuseRepoPath, f.InfusePath)
	if err != nil {
		return "?"
	}

	fileStatus, exists := wtStatus[relToInfuse]
	if !exists {
		return "clean"
	}

	if fileStatus.Worktree == gogit.Untracked {
		return "untracked"
	}

	if fileStatus.Worktree == gogit.Modified || fileStatus.Staging == gogit.Modified {
		return "modified"
	}

	return "clean"
}

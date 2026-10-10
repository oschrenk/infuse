package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	gogit "github.com/go-git/go-git/v5"
	"github.com/oschrenk/infuse/internal/config"
	"github.com/oschrenk/infuse/internal/git"
	"github.com/oschrenk/infuse/internal/infuse"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(rmCmd)
	rmCmd.Flags().BoolVar(&rmLocalOnly, "local-only", false, "Remove the symlink and exclude entry, keep the file in the infuse repository")
	rmCmd.Flags().BoolVar(&rmDryRun, "dry-run", false, "Show what would be done without making changes")
}

var (
	rmLocalOnly bool
	rmDryRun    bool
)

var rmCmd = &cobra.Command{
	Use:     "rm <path>",
	Aliases: []string{"remove"},
	Short:   "Delete a managed file from the working and infuse repositories",
	Long:    "Delete the symlink, the file in the infuse repository, and the .git/info/exclude entry.\nWith --local-only, keep the file in the infuse repository.",
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return fmt.Errorf("loading config: %w\nRun: infuse init", err)
		}

		if cfg.Repo.Path == "" {
			return fmt.Errorf("repository path not configured\nRun: infuse init")
		}

		if _, err := gogit.PlainOpen(cfg.Repo.Path); err != nil {
			return fmt.Errorf("infuse repository is not set up\nRun: infuse setup")
		}

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

		return rmFile(infuse.New(cfg), repo, repoRoot, remote, args[0])
	},
}

func rmFile(inf *infuse.Infuse, repo *git.Repo, repoRoot, remote, path string) error {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return fmt.Errorf("resolving path %s: %w", path, err)
	}
	// Root() resolves symlinks, so resolve the parent too (not the file, which is the symlink itself)
	if dir, err := filepath.EvalSymlinks(filepath.Dir(absPath)); err == nil {
		absPath = filepath.Join(dir, filepath.Base(absPath))
	}

	relPath, err := filepath.Rel(repoRoot, absPath)
	if err != nil || relPath == ".." || strings.HasPrefix(relPath, ".."+string(filepath.Separator)) {
		return fmt.Errorf("%s: outside the repository", path)
	}

	storePath := inf.DestPath(relPath, remote)
	_, err = os.Lstat(storePath)
	storeExists := err == nil

	fi, err := os.Lstat(absPath)
	localExists := err == nil

	switch {
	case !localExists && !storeExists:
		return fmt.Errorf("%s: not managed by infuse", path)
	case !localExists && rmLocalOnly:
		return fmt.Errorf("%s: no symlink to remove", path)
	case localExists && fi.Mode()&os.ModeSymlink == 0:
		return fmt.Errorf("%s: not a symlink, not managed by infuse", path)
	case localExists && !pointsTo(absPath, storePath, storeExists):
		return fmt.Errorf("%s: symlink does not point to %s", path, storePath)
	}

	deleteStore := storeExists && !rmLocalOnly
	if deleteStore {
		fmt.Printf("%s: delete %s\n", path, storePath)
	}
	if localExists {
		fmt.Printf("%s: remove symlink\n", path)
	}
	fmt.Printf("%s: remove exclude entry\n", path)

	if rmDryRun {
		return nil
	}

	if deleteStore {
		if err := os.Remove(storePath); err != nil {
			return fmt.Errorf("%s: deleting %s: %w", path, storePath, err)
		}
	}
	if localExists {
		if err := os.Remove(absPath); err != nil {
			return fmt.Errorf("%s: removing symlink: %w", path, err)
		}
	}
	if err := repo.RemoveExclude(relPath); err != nil {
		return fmt.Errorf("%s: removing exclude entry: %w", path, err)
	}

	fmt.Printf("%s: removed\n", path)
	return nil
}

// pointsTo reports whether the symlink at linkPath resolves to storePath.
// A dangling symlink matches only on the exact target that add and link write.
func pointsTo(linkPath, storePath string, storeExists bool) bool {
	if !storeExists {
		target, err := os.Readlink(linkPath)
		return err == nil && target == storePath
	}
	linkInfo, err := os.Stat(linkPath)
	if err != nil {
		return false
	}
	storeInfo, err := os.Stat(storePath)
	if err != nil {
		return false
	}
	return os.SameFile(linkInfo, storeInfo)
}

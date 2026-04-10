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
	rootCmd.AddCommand(linkCmd)
	linkCmd.Flags().BoolVar(&linkDryRun, "dry-run", false, "Show what would be done without making changes")
	linkCmd.Flags().BoolVar(&linkSkip, "skip", false, "Skip conflicting files, link the rest")
	linkCmd.Flags().BoolVar(&linkOverwrite, "overwrite", false, "Replace conflicting files/symlinks")
}

var (
	linkDryRun   bool
	linkSkip     bool
	linkOverwrite bool
)

type fileState int

const (
	stateMissing         fileState = iota // does not exist — will be linked
	stateLinked                           // correct symlink already in place
	stateConflictSymlink                  // symlink exists but wrong/broken target
	stateConflictFile                     // real file or directory exists
)

type fileClassification struct {
	relPath    string
	infusePath string
	state      fileState
}

var linkCmd = &cobra.Command{
	Use:   "link",
	Short: "Re-link infused files into the current repository",
	Long:  "Create symlinks for all infused files in the current repository (e.g. after a fresh clone)",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		if linkSkip && linkOverwrite {
			return fmt.Errorf("--skip and --overwrite are mutually exclusive")
		}

		cfg, err := config.Load()
		if err != nil {
			return fmt.Errorf("loading config: %w\nRun: infuse init", err)
		}

		if cfg.Repo.Path == "" {
			return fmt.Errorf("repository path not configured\nRun: infuse init")
		}

		infuseRepoPath := cfg.Repo.Path
		if _, err := gogit.PlainOpen(infuseRepoPath); err != nil {
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

		inf := infuse.New(cfg)

		files, err := inf.ListFiles(remote)
		if err != nil {
			return fmt.Errorf("listing infuse files: %w", err)
		}
		if len(files) == 0 {
			fmt.Println("no files managed by infuse for this repository")
			return nil
		}

		// Pre-flight: classify every file
		classifications := make([]fileClassification, 0, len(files))
		for _, f := range files {
			state, err := classifyFile(inf, repoRoot, f.RelPath)
			if err != nil {
				return fmt.Errorf("%s: %w", f.RelPath, err)
			}
			classifications = append(classifications, fileClassification{
				relPath:    f.RelPath,
				infusePath: f.InfusePath,
				state:      state,
			})
		}

		// Determine what action label to show per state
		actionLabel := func(s fileState) string {
			switch s {
			case stateMissing:
				return "will link"
			case stateLinked:
				return "skip (already linked)"
			case stateConflictSymlink, stateConflictFile:
				if linkDryRun {
					if linkSkip {
						return "skip (conflict)"
					}
					if linkOverwrite {
						return "overwrite"
					}
					return "fail (conflict)"
				}
				if linkSkip {
					return "skip (conflict)"
				}
				if linkOverwrite {
					return "overwrite"
				}
				return "fail (conflict)"
			}
			return "unknown"
		}

		stateLabel := func(s fileState) string {
			switch s {
			case stateMissing:
				return "missing"
			case stateLinked:
				return "linked"
			case stateConflictSymlink:
				return "conflict_symlink"
			case stateConflictFile:
				return "conflict_file"
			}
			return "unknown"
		}

		// Print pre-flight table
		for _, c := range classifications {
			fmt.Printf("%-40s %-16s → %s\n", c.relPath, stateLabel(c.state), actionLabel(c.state))
		}

		if linkDryRun {
			return nil
		}

		// Check for conflicts in default mode (no --skip, no --overwrite)
		if !linkSkip && !linkOverwrite {
			hasConflicts := false
			for _, c := range classifications {
				if c.state == stateConflictSymlink || c.state == stateConflictFile {
					hasConflicts = true
					break
				}
			}
			if hasConflicts {
				return fmt.Errorf("conflicts found — re-run with --skip or --overwrite")
			}
		}

		// Apply
		var linked, skipped, alreadyLinked int
		for _, c := range classifications {
			switch c.state {
			case stateLinked:
				alreadyLinked++

			case stateMissing:
				symlinkPath := filepath.Join(repoRoot, c.relPath)
				if err := os.MkdirAll(filepath.Dir(symlinkPath), 0755); err != nil {
					return fmt.Errorf("%s: creating parent dirs: %w", c.relPath, err)
				}
				if err := os.Symlink(c.infusePath, symlinkPath); err != nil {
					return fmt.Errorf("%s: creating symlink: %w", c.relPath, err)
				}
				if err := repo.AddExclude(c.relPath); err != nil {
					return fmt.Errorf("%s: adding exclude entry: %w", c.relPath, err)
				}
				linked++

			case stateConflictSymlink, stateConflictFile:
				if linkSkip {
					skipped++
					continue
				}
				// --overwrite
				symlinkPath := filepath.Join(repoRoot, c.relPath)
				if err := os.Remove(symlinkPath); err != nil {
					return fmt.Errorf("%s: removing existing: %w", c.relPath, err)
				}
				if err := os.Symlink(c.infusePath, symlinkPath); err != nil {
					return fmt.Errorf("%s: creating symlink: %w", c.relPath, err)
				}
				if err := repo.AddExclude(c.relPath); err != nil {
					return fmt.Errorf("%s: adding exclude entry: %w", c.relPath, err)
				}
				linked++
			}
		}

		fmt.Printf("Linked %d, skipped %d, already linked %d\n", linked, skipped, alreadyLinked)
		return nil
	},
}

func classifyFile(inf *infuse.Infuse, repoRoot, relPath string) (fileState, error) {
	symlinkPath := filepath.Join(repoRoot, relPath)

	fi, err := os.Lstat(symlinkPath)
	if err != nil {
		if os.IsNotExist(err) {
			return stateMissing, nil
		}
		return 0, fmt.Errorf("stat: %w", err)
	}

	if fi.Mode()&os.ModeSymlink != 0 {
		tracked, err := inf.IsTracked(symlinkPath)
		if err != nil {
			return 0, err
		}
		if tracked {
			return stateLinked, nil
		}
		return stateConflictSymlink, nil
	}

	return stateConflictFile, nil
}

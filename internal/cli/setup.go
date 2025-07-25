package cli

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	gogit "github.com/go-git/go-git/v5"
	"github.com/oschrenk/infuse/internal/config"
	"github.com/oschrenk/infuse/internal/git"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(setupCmd)
}

var setupCmd = &cobra.Command{
	Use:   "setup",
	Short: "Set up working repository",
	Long:  "Validate and prepare the current working directory for use with infuse",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Check config exists and is valid
		cfg, err := config.Load()
		if err != nil {
			return fmt.Errorf("loading config: %w", err)
		}

		if cfg.Repo.Path == "" {
			return fmt.Errorf("repository path not configured, run: infuse init")
		}

		// Ensure infuse repo directory exists
		infuseRepoPath := cfg.Repo.Path
		if _, err := os.Stat(infuseRepoPath); os.IsNotExist(err) {
			if !confirm(fmt.Sprintf("Directory %s does not exist. Create it?", infuseRepoPath)) {
				return fmt.Errorf("infuse repository directory does not exist")
			}
			if err := os.MkdirAll(infuseRepoPath, 0755); err != nil {
				return fmt.Errorf("creating directory: %w", err)
			}
			fmt.Printf("Created %s\n", infuseRepoPath)
		}

		// Ensure infuse repo is a git repository
		if _, err := gogit.PlainOpen(infuseRepoPath); err != nil {
			if !confirm(fmt.Sprintf("%s is not a git repository. Initialize it?", infuseRepoPath)) {
				return fmt.Errorf("infuse repository is not a git repository")
			}
			if _, err := gogit.PlainInit(infuseRepoPath, false); err != nil {
				return fmt.Errorf("initializing git repository: %w", err)
			}
			fmt.Printf("Initialized git repository at %s\n", infuseRepoPath)
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

		// Check git repository has a remote configured
		remote, err := repo.GetNormalizedRemote()
		if err != nil {
			return fmt.Errorf("getting remote: %w", err)
		}
		if remote == "" {
			return fmt.Errorf("no remote configured")
		}

		// Check working repo is not the infuse data repo
		repoRoot, err := repo.Root()
		if err != nil {
			return fmt.Errorf("getting repository root: %w", err)
		}

		dataPath, err := filepath.EvalSymlinks(infuseRepoPath)
		if err != nil {
			return fmt.Errorf("resolving data repository path: %w", err)
		}

		if repoRoot == dataPath {
			return fmt.Errorf("working repository cannot be the infuse data repository")
		}

		// Create directory structure inside infuse repo
		destDir := filepath.Join(infuseRepoPath, remote)
		if err := os.MkdirAll(destDir, 0755); err != nil {
			return fmt.Errorf("creating directory structure: %w", err)
		}

		fmt.Printf("Ready: %s\n", destDir)
		return nil
	},
}

func confirm(prompt string) bool {
	reader := bufio.NewReader(os.Stdin)
	fmt.Printf("%s [y/N] ", prompt)
	answer, _ := reader.ReadString('\n')
	answer = strings.TrimSpace(strings.ToLower(answer))
	return answer == "y" || answer == "yes"
}

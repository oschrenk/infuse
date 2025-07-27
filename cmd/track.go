package cmd

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/go-git/go-git/v5"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(trackCmd)
}

func normalizeRemoteURL(rawURL string) []string {
	var host, path string
	
	// Handle SSH URLs (git@host:repo format)
	if strings.HasPrefix(rawURL, "git@") {
		sshRegex := regexp.MustCompile(`git@([^:]+):(.+)`)
		matches := sshRegex.FindStringSubmatch(rawURL)
		if len(matches) == 3 {
			host = matches[1]
			path = strings.TrimSuffix(matches[2], ".git")
		}
	} else if strings.HasPrefix(rawURL, "http://") || strings.HasPrefix(rawURL, "https://") {
		// Handle HTTP/HTTPS URLs
		if u, err := url.Parse(rawURL); err == nil {
			host = u.Host
			path = strings.TrimPrefix(u.Path, "/")
			path = strings.TrimSuffix(path, ".git")
		}
	}
	
	if host != "" && path != "" {
		return []string{host, path}
	}
	
	return []string{rawURL}
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
		
		repo, err := git.PlainOpenWithOptions(absPath, &git.PlainOpenOptions{
			DetectDotGit: true,
		})
		if err != nil {
			fmt.Printf("Path %s is not within a git repository\n", path)
			return nil
		}
		
		// Get the repository root
		workTree, err := repo.Worktree()
		if err != nil {
			return err
		}
		repoRoot := workTree.Filesystem.Root()
		
		// Get remote URL
		remotes, err := repo.Remotes()
		if err != nil {
			return err
		}
		
		var remoteInfo []string
		if len(remotes) > 0 {
			// Try to find 'origin' remote first, otherwise use the first one
			for _, remote := range remotes {
				if remote.Config().Name == "origin" {
					if len(remote.Config().URLs) > 0 {
						remoteInfo = normalizeRemoteURL(remote.Config().URLs[0])
					}
					break
				}
			}
			// If no origin found, use first remote
			if len(remoteInfo) == 0 && len(remotes[0].Config().URLs) > 0 {
				remoteInfo = normalizeRemoteURL(remotes[0].Config().URLs[0])
			}
		}
		
		// Get relative path from repo root
		relPath, err := filepath.Rel(repoRoot, absPath)
		if err != nil {
			return err
		}
		
		// Check if file/directory is tracked
		status, err := workTree.Status()
		if err != nil {
			return err
		}
		
		// Prepare remote info string
		remoteStr := ""
		if len(remoteInfo) == 2 {
			remoteStr = fmt.Sprintf(", remote: [%s, %s]", remoteInfo[0], remoteInfo[1])
		} else if len(remoteInfo) == 1 {
			remoteStr = fmt.Sprintf(", remote: [%s]", remoteInfo[0])
		}
		
		// Check if it's a file and tracked
		if fileInfo, err := os.Stat(absPath); err == nil && !fileInfo.IsDir() {
			if fileStatus, exists := status[relPath]; exists {
				if fileStatus.Staging == git.Untracked && fileStatus.Worktree == git.Untracked {
					fmt.Printf("hello %s (git repository detected, file untracked%s)\n", path, remoteStr)
				} else {
					fmt.Printf("hello %s (git repository detected, file tracked%s)\n", path, remoteStr)
				}
			} else {
				fmt.Printf("hello %s (git repository detected, file tracked%s)\n", path, remoteStr)
			}
		} else {
			// For directories, check if any files within are tracked
			isTracked := false
			for filePath := range status {
				if strings.HasPrefix(filePath, relPath) {
					isTracked = true
					break
				}
			}
			if isTracked {
				fmt.Printf("hello %s (git repository detected, directory has tracked files%s)\n", path, remoteStr)
			} else {
				fmt.Printf("hello %s (git repository detected, directory untracked%s)\n", path, remoteStr)
			}
		}
		
		return nil
	},
}
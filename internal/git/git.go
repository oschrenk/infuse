package git

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/go-git/go-git/v5"
)

// Repo represents a git repository
type Repo struct {
	repo *git.Repository
}

// Load creates a new Repo instance by opening a git repository at the given path
func Load(path string) (*Repo, error) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}

	repo, err := git.PlainOpenWithOptions(absPath, &git.PlainOpenOptions{
		DetectDotGit: true,
	})
	if err != nil {
		return nil, err
	}

	return &Repo{
		repo: repo,
	}, nil
}

// Root returns the root directory of the git repository worktree
func (r *Repo) Root() (string, error) {
	wt, err := r.repo.Worktree()
	if err != nil {
		return "", err
	}
	return filepath.EvalSymlinks(wt.Filesystem.Root())
}

// AddExclude appends a path to .git/info/exclude if not already present.
func (r *Repo) AddExclude(relPath string) error {
	root, err := r.Root()
	if err != nil {
		return err
	}

	excludeDir := filepath.Join(root, ".git", "info")
	if err := os.MkdirAll(excludeDir, 0755); err != nil {
		return fmt.Errorf("creating .git/info directory: %w", err)
	}

	excludePath := filepath.Join(excludeDir, "exclude")

	// Check if entry already exists
	existing, err := os.ReadFile(excludePath)
	if err == nil {
		for _, line := range strings.Split(string(existing), "\n") {
			if line == relPath {
				return nil // already excluded
			}
		}
	}

	f, err := os.OpenFile(excludePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("opening exclude file: %w", err)
	}
	defer f.Close()

	if _, err := fmt.Fprintln(f, relPath); err != nil {
		return fmt.Errorf("writing exclude entry: %w", err)
	}

	return nil
}

// GetNormalizedRemote gets the normalized remote information and formats it as a string
func (r *Repo) GetNormalizedRemote() (string, error) {
	// Get remote URL
	remotes, err := r.repo.Remotes()
	if err != nil {
		return "", err
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

	// Prepare remote info string
	remoteStr := ""
	if len(remoteInfo) == 2 {
		remoteStr = fmt.Sprintf("%s/%s", remoteInfo[0], remoteInfo[1])
	} else if len(remoteInfo) == 1 {
		remoteStr = remoteInfo[0]
	}

	return remoteStr, nil
}

// Repository returns the underlying git repository
func (r *Repo) Repository() *git.Repository {
	return r.repo
}

// IsTracked checks if a file at the given path is tracked in the repository
func (r *Repo) IsTracked(path string) (bool, error) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return false, err
	}

	workTree, err := r.repo.Worktree()
	if err != nil {
		return false, err
	}

	repoRoot := workTree.Filesystem.Root()
	relPath, err := filepath.Rel(repoRoot, absPath)
	if err != nil {
		return false, err
	}

	// Check if file/directory is tracked
	status, err := workTree.Status()
	if err != nil {
		return false, err
	}

	// Check if it's a file and tracked
	if fileInfo, err := os.Stat(absPath); err == nil && !fileInfo.IsDir() {
		if fileStatus, exists := status[relPath]; exists {
			// If file exists in status and is not untracked, it's tracked
			return !(fileStatus.Staging == git.Untracked && fileStatus.Worktree == git.Untracked), nil
		} else {
			// If file doesn't exist in status, it means it's tracked (committed)
			return true, nil
		}
	} else {
		// For directories, check if any files within are tracked
		for filePath := range status {
			if strings.HasPrefix(filePath, relPath) {
				return true, nil
			}
		}
		return false, nil
	}
}

// normalizeRemoteURL parses a git remote URL and returns normalized components.
// Returns [host, path] for valid URLs, or [rawURL] if parsing fails.
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

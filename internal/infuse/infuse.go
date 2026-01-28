package infuse

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/oschrenk/infuse/internal/config"
)

type Infuse struct {
	config *config.Config
}

func New(config *config.Config) *Infuse {
	if config == nil {
		panic("config cannot be nil")
	}
	return &Infuse{
		config: config,
	}
}

// RepoPath returns the configured infuse repository path.
func (i *Infuse) RepoPath() string {
	return i.config.Repo.Path
}

// DestPath returns the destination path for a file in the infuse repository.
func (i *Infuse) DestPath(relPath, normalizedRemote string) string {
	return filepath.Join(i.config.Repo.Path, normalizedRemote, relPath)
}

// Add moves a file into the infuse repository and creates a symlink back.
// absPath is the absolute path to the file, repoRoot is the working repo root,
// normalizedRemote is the host/owner/repo string.
// Returns the relative path (for use in .git/info/exclude).
func (i *Infuse) Add(absPath, repoRoot, normalizedRemote string) (string, error) {
	relPath, err := filepath.Rel(repoRoot, absPath)
	if err != nil {
		return "", fmt.Errorf("computing relative path: %w", err)
	}

	destPath := i.DestPath(relPath, normalizedRemote)

	// Create destination directory structure
	if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
		return "", fmt.Errorf("creating destination directory: %w", err)
	}

	// Move the file
	if err := os.Rename(absPath, destPath); err != nil {
		return "", fmt.Errorf("moving %s to %s: %w", absPath, destPath, err)
	}

	// Create symlink from original location to infuse repo
	if err := os.Symlink(destPath, absPath); err != nil {
		// Attempt to move the file back on failure
		_ = os.Rename(destPath, absPath)
		return "", fmt.Errorf("creating symlink: %w", err)
	}

	return relPath, nil
}

// IsTracked checks if a path is already tracked by infuse (is a symlink
// pointing into the infuse repository).
func (i *Infuse) IsTracked(absPath string) (bool, error) {
	target, err := os.Readlink(absPath)
	if err != nil {
		return false, nil // not a symlink
	}

	// Resolve to absolute if relative
	if !filepath.IsAbs(target) {
		target = filepath.Join(filepath.Dir(absPath), target)
	}

	repoPath, err := filepath.EvalSymlinks(i.config.Repo.Path)
	if err != nil {
		return false, err
	}

	target, err = filepath.EvalSymlinks(target)
	if err != nil {
		return false, nil // broken symlink, not tracked
	}

	rel, err := filepath.Rel(repoPath, target)
	if err != nil {
		return false, nil
	}

	// If the relative path doesn't start with "..", the target is inside the repo
	return !filepath.IsAbs(rel) && rel != ".." && !hasPrefix(rel, ".."+string(filepath.Separator)), nil
}

// FileStatus represents the status of a file managed by infuse.
type FileStatus struct {
	// RelPath is the path relative to the working repo root
	RelPath string
	// InfusePath is the absolute path in the infuse repository
	InfusePath string
	// SymlinkExists is true if the symlink in the working repo exists and points to InfusePath
	SymlinkExists bool
	// SymlinkBroken is true if the symlink exists but the target is missing
	SymlinkBroken bool
	// Modified is true if the file in the infuse repo has uncommitted changes
	Modified bool
}

// ListFiles returns all files managed by infuse for the given normalized remote.
func (i *Infuse) ListFiles(normalizedRemote string) ([]FileStatus, error) {
	remoteDir := filepath.Join(i.config.Repo.Path, normalizedRemote)

	if _, err := os.Stat(remoteDir); os.IsNotExist(err) {
		return nil, nil
	}

	var files []FileStatus
	err := filepath.Walk(remoteDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}

		relPath, err := filepath.Rel(remoteDir, path)
		if err != nil {
			return err
		}

		files = append(files, FileStatus{
			RelPath:    relPath,
			InfusePath: path,
		})
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("walking infuse directory: %w", err)
	}

	return files, nil
}

func hasPrefix(s, prefix string) bool {
	return len(s) >= len(prefix) && s[:len(prefix)] == prefix
}

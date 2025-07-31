package core

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/oschrenk/infuse/config"
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

func (i *Infuse) Move(path string, normalizedRepo string) error {
	// Get the infuse repository path from config
	infuseRepoPath := i.config.Repo.Path
	if infuseRepoPath == "" {
		return fmt.Errorf("infuse repository path not configured")
	}

	// Create destination directory structure using normalized repo name
	destDir := filepath.Join(infuseRepoPath, normalizedRepo)
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return fmt.Errorf("failed to create destination directory: %w", err)
	}

	// Get the filename from the source path
	filename := filepath.Base(path)
	destPath := filepath.Join(destDir, filename)

	// Move the file
	if err := os.Rename(path, destPath); err != nil {
		return fmt.Errorf("failed to move file from %s to %s: %w", path, destPath, err)
	}

	return nil
}

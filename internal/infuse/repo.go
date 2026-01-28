package infuse

import (
	"fmt"

	"github.com/go-git/go-git/v5"
	"github.com/oschrenk/infuse/internal/config"
)

type Repo struct {
	Path string
}

func (r *Repo) Load(cfg *config.Config) error {
	r.Path = cfg.Repo.Path

	// check if it's a valid git repository
	_, err := git.PlainOpen(r.Path)
	if err != nil {
		return fmt.Errorf("invalid git repository at %s: %w", r.Path, err)
	}

	return nil
}

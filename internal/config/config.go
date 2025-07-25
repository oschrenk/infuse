package config

import (
	_ "embed"
	"fmt"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

//go:embed default.toml
var defaultConfig []byte

type Repo struct {
	Path string `toml:"path"`
}

type Config struct {
	Repo Repo `toml:"repo"`
}

func Load() (*Config, error) {
	var c Config
	configPath := configPath()

	if _, err := toml.DecodeFile(configPath, &c); err != nil {
		return nil, err
	}

	// expand env variables inside the repo path
	c.Repo.Path = os.ExpandEnv(c.Repo.Path)

	return &c, nil
}

func configPath() string {
	var configHome string

	if xdgConfigHome := os.Getenv("XDG_CONFIG_HOME"); xdgConfigHome != "" {
		configHome = xdgConfigHome
	} else {
		homeDir := os.Getenv("HOME")
		configHome = filepath.Join(homeDir, ".config")
	}

	return filepath.Join(configHome, "infuse", "config.toml")
}

func ensureConfig() error {
	configPath := configPath()

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		if err := os.MkdirAll(filepath.Dir(configPath), 0755); err != nil {
			return err
		}

		if err := os.WriteFile(configPath, defaultConfig, 0600); err != nil {
			return err
		}
		fmt.Printf("Created %s\n", configPath)
	} else {
		fmt.Printf("Config already exists: %s\n", configPath)
	}
	return nil
}

func Init() error {
	return ensureConfig()
}

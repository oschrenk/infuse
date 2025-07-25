package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

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

func upsertConfig() (*string, error) {
	configPath := configPath()

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		fmt.Printf("Creating %s\n", configPath)

		if err := os.MkdirAll(filepath.Dir(configPath), 0755); err != nil {
			return nil, err
		}

		file, err := os.OpenFile(configPath, os.O_CREATE|os.O_WRONLY, 0600)
		if err != nil {
			return nil, err
		}
		defer file.Close()
	}
	return &configPath, nil
}

func Setup() error {
	_, err := upsertConfig()
	if err != nil {
		return err
	}

	fmt.Printf("Done")
	return nil
}

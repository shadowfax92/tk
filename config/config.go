package config

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Root          string `yaml:"root"`
	Editor        string `yaml:"editor"`
	StaleWarnDays int    `yaml:"stale_warn_days"`
	StaleCritDays int    `yaml:"stale_critical_days"`
	FocusItems    int    `yaml:"focus_items"`
}

var DefaultConfig = Config{
	Root:          "~/notes/obsidian-browseros/2_tasks",
	Editor:        "nvim",
	StaleWarnDays: 28,
	StaleCritDays: 56,
	FocusItems:    3,
}

func configPath() string {
	if xdg := os.Getenv("XDG_CONFIG_HOME"); xdg != "" {
		return filepath.Join(xdg, "tk", "config.yaml")
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "tk", "config.yaml")
}

func Load() (*Config, error) {
	cfg := DefaultConfig

	data, err := os.ReadFile(configPath())
	if err != nil {
		if os.IsNotExist(err) {
			cfg.Root = expandHome(cfg.Root)
			return &cfg, nil
		}
		return nil, err
	}

	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	cfg.Root = expandHome(cfg.Root)
	if cfg.Editor == "" {
		cfg.Editor = "nvim"
	}
	if cfg.StaleWarnDays == 0 {
		cfg.StaleWarnDays = 28
	}
	if cfg.StaleCritDays == 0 {
		cfg.StaleCritDays = 56
	}
	if cfg.FocusItems <= 0 {
		cfg.FocusItems = 3
	}

	return &cfg, nil
}

func Init() error {
	p := configPath()
	if _, err := os.Stat(p); err == nil {
		return nil
	}

	if err := os.MkdirAll(filepath.Dir(p), 0755); err != nil {
		return err
	}

	data, err := yaml.Marshal(DefaultConfig)
	if err != nil {
		return err
	}

	return os.WriteFile(p, data, 0644)
}

func expandHome(path string) string {
	if len(path) > 1 && path[:2] == "~/" {
		home, _ := os.UserHomeDir()
		return filepath.Join(home, path[2:])
	}
	return path
}

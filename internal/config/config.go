package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/NaruseYuki/ramen-github-manager/internal/model"
	"gopkg.in/yaml.v3"
)

const (
	configDir  = "rgm"
	configFile = "config.yaml"
)

// DefaultConfig returns an empty configuration with sensible defaults.
func DefaultConfig() *model.Config {
	return &model.Config{
		Repositories: []model.RepoConfig{},
		Defaults: model.Defaults{
			Sort:  "updated",
			Limit: 30,
			State: "open",
		},
	}
}

// AddRepo adds a repository to the config. Returns error if already exists.
func AddRepo(cfg *model.Config, name, alias string) error {
	for _, r := range cfg.Repositories {
		if r.Name == name {
			return fmt.Errorf("repository %q already exists", name)
		}
	}
	cfg.Repositories = append(cfg.Repositories, model.RepoConfig{Name: name, Alias: alias})
	return nil
}

// RemoveRepo removes a repository from the config. Returns error if not found.
func RemoveRepo(cfg *model.Config, nameOrAlias string) error {
	for i, r := range cfg.Repositories {
		if r.Name == nameOrAlias || r.Alias == nameOrAlias {
			cfg.Repositories = append(cfg.Repositories[:i], cfg.Repositories[i+1:]...)
			return nil
		}
	}
	return fmt.Errorf("repository %q not found", nameOrAlias)
}

// ConfigPath returns the full path to the configuration file.
func ConfigPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}
	return filepath.Join(home, ".config", configDir, configFile), nil
}

// Load reads and parses the configuration file.
func Load() (*model.Config, error) {
	path, err := ConfigPath()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("config file not found at %s\nRun 'rgm config init' to create one", path)
		}
		return nil, fmt.Errorf("failed to read config: %w", err)
	}

	var cfg model.Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	return &cfg, nil
}

// Save writes the configuration to disk.
func Save(cfg *model.Config) error {
	path, err := ConfigPath()
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}

	return nil
}

// ResolveRepo resolves a repo name or alias to the actual repository name.
func ResolveRepo(cfg *model.Config, nameOrAlias string) (string, error) {
	for _, r := range cfg.Repositories {
		if r.Name == nameOrAlias || r.Alias == nameOrAlias {
			return r.Name, nil
		}
	}
	return "", fmt.Errorf("unknown repository: %s", nameOrAlias)
}

// RepoNames returns all configured repository names.
func RepoNames(cfg *model.Config) []string {
	names := make([]string, len(cfg.Repositories))
	for i, r := range cfg.Repositories {
		names[i] = r.Name
	}
	return names
}

// AliasFor returns the alias for a given repository name.
func AliasFor(cfg *model.Config, repoName string) string {
	for _, r := range cfg.Repositories {
		if r.Name == repoName {
			if r.Alias != "" {
				return r.Alias
			}
			return r.Name
		}
	}
	return repoName
}

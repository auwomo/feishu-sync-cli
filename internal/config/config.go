package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type Config struct {
	App struct {
		ID         string `yaml:"id"`
		SecretEnv  string `yaml:"secret_env"`
		SecretFile string `yaml:"secret_file"`
	} `yaml:"app"`

	Auth struct {
		// Mode: tenant (default) | user
		Mode string `yaml:"mode"`
	} `yaml:"auth"`

	Scope struct {
		Mode             string   `yaml:"mode"`
		DriveFolderTokens []string `yaml:"drive_folder_tokens"`
		WikiSpaceIDs     []string `yaml:"wiki_space_ids"`
	} `yaml:"scope"`

	Output struct {
		Dir string `yaml:"dir"`
	} `yaml:"output"`
}

func Load(path string) (*Config, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var c Config
	if err := yaml.Unmarshal(b, &c); err != nil {
		return nil, err
	}
	return &c, nil
}

func (c *Config) ValidateRelativeOutputDir() error {
	if c.Output.Dir == "" {
		return errors.New("output.dir is required")
	}
	if filepath.IsAbs(c.Output.Dir) {
		return fmt.Errorf("output.dir must be relative, got absolute: %s", c.Output.Dir)
	}
	if clean := filepath.Clean(c.Output.Dir); clean == "." || clean == string(filepath.Separator) {
		return fmt.Errorf("output.dir invalid: %s", c.Output.Dir)
	}
	return nil
}

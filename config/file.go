// Package config handles reading and writing the align configuration at ~/.align/config.
package config

import (
	"fmt"
	"os"
	"os/user"
	"path/filepath"

	"gopkg.in/yaml.v2"
)

const (
	confDir  = "/.align"
	confFile = "/config"
)

var defaultConfig = Config{
	Github: &GithubHost{
		Limits: &Limits{
			RequestsPerSecond: 10,
			Burst:             25,
		},
	},
}

// Config holds the top-level align configuration read from ~/.align/config.
type Config struct {
	Github *GithubHost `yaml:"github.com"`
}

// New returns a minimal Config with the given GitHub token set.
func New(tkn string) *Config {
	return &Config{Github: &GithubHost{Token: tkn}}
}

// WriteFile marshals and writes the config to ~/.align/config.
func (c *Config) WriteFile() error {
	b, err := yaml.Marshal(c)
	if err != nil {
		return fmt.Errorf("config: marshal: %v", err.Error())
	}

	usr, err := user.Current()
	if err != nil {
		return fmt.Errorf("config: get home directory: %v", err.Error())
	}

	err = os.WriteFile(filepath.Join(usr.HomeDir, confDir, confFile), b, 0600)
	if err != nil {
		return fmt.Errorf("config: write file: %v", err.Error())
	}

	return nil
}

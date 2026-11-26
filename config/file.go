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

// Config holds the top-level align configuration read from ~/.align/config.
type Config struct {
	Github *GithubHost `yaml:"github.com"`
}

// New returns a Config with the given GitHub token and default rate limits.
func New(tkn string) *Config {
	return &Config{Github: &GithubHost{
		Token: tkn,
		Limits: &Limits{
			RequestsPerSecond: 10,
			Burst:             25,
		},
	}}
}

// WithDefaults returns an independent copy with missing GitHub settings initialized.
func (c *Config) WithDefaults() *Config {
	defaults := New("")
	if c == nil {
		return defaults
	}

	result := *c
	if c.Github == nil {
		result.Github = defaults.Github
		return &result
	}

	host := *c.Github
	result.Github = &host
	if c.Github.Limits == nil {
		host.Limits = defaults.Github.Limits
		return &result
	}

	limits := *c.Github.Limits
	host.Limits = &limits
	return &result
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

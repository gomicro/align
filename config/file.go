package config

import (
	"fmt"
	"io/ioutil"
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

// Config represents the config file for align
type Config struct {
	Github *GithubHost `yaml:"github.com"`
}

// New takes a token string and creates the most basic config capable of being
// written.
func New(tkn string) *Config {
	return &Config{Github: &GithubHost{Token: tkn}}
}

// WriteFile writes the file to the defined location for the current user, and
// returns any errors encountered doing so.
func (c *Config) WriteFile() error {
	b, err := yaml.Marshal(c)
	if err != nil {
		return fmt.Errorf("config: marshal: %v", err.Error())
	}

	usr, err := user.Current()
	if err != nil {
		return fmt.Errorf("config: get home directory: %v", err.Error())
	}

	err = ioutil.WriteFile(filepath.Join(usr.HomeDir, confDir, confFile), b, 0600)
	if err != nil {
		return fmt.Errorf("config: write file: %v", err.Error())
	}

	return nil
}

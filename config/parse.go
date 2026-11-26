package config

import (
	"fmt"
	"os"
	"os/user"
	"path/filepath"

	"gopkg.in/yaml.v2"
)

// ParseFromFile reads ~/.align/config, creating the config directory if absent.
func ParseFromFile() (*Config, error) {
	path, err := FilePath()
	if err != nil {
		return nil, fmt.Errorf("finding config: %w", err)
	}

	conf, err := ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading config: %w", err)
	}

	return conf, nil
}

// ReadFile reads a config file, creating its parent directory if absent.
func ReadFile(path string) (*Config, error) {
	err := os.MkdirAll(filepath.Dir(path), 0700)
	if err != nil {
		return nil, fmt.Errorf("creating config directory: %w", err)
	}

	b, err := os.ReadFile(path)
	if err != nil && !os.IsNotExist(err) {
		return nil, fmt.Errorf("reading config file: %w", err)
	}

	conf, err := parseConfig(b)
	if err != nil {
		return nil, fmt.Errorf("parsing config file: %w", err)
	}

	return conf, nil
}

func parseConfig(b []byte) (*Config, error) {
	conf := New("")
	err := yaml.Unmarshal(b, conf)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal config file: %v", err.Error())
	}

	return conf.WithDefaults(), nil
}

// DirExists reports whether the ~/.align config directory exists.
func DirExists() (bool, error) {
	usr, err := user.Current()
	if err != nil {
		return false, fmt.Errorf("failed getting home directory: %v", err.Error())
	}

	_, err = os.Stat(filepath.Join(usr.HomeDir, confDir))
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, fmt.Errorf("failed to confirm config dir existence: %v", err.Error())
	}

	return true, nil
}

// FileExists reports whether the ~/.align/config file exists.
func FileExists() (bool, error) {
	usr, err := user.Current()
	if err != nil {
		return false, fmt.Errorf("failed getting home directory: %v", err.Error())
	}

	_, err = os.Stat(filepath.Join(usr.HomeDir, confDir, confFile))
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, fmt.Errorf("failed to confirm config file existence: %v", err.Error())
	}

	return true, nil
}

// CreateDir creates the ~/.align config directory, including any missing parents.
func CreateDir() error {
	usr, err := user.Current()
	if err != nil {
		return fmt.Errorf("config: get home directory: %v", err.Error())
	}

	err = os.MkdirAll(filepath.Join(usr.HomeDir, confDir), 0700)
	if err != nil {
		return fmt.Errorf("config: write file: %v", err.Error())
	}

	return nil
}

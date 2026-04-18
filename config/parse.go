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
	usr, err := user.Current()
	if err != nil {
		return nil, fmt.Errorf("failed getting home directory: %v", err.Error())
	}

	conf := defaultConfig
	dExists, err := DirExists()
	if err != nil {
		return nil, fmt.Errorf("config: parse from file: dir exists: %v", err.Error())
	}

	if !dExists {
		err := CreateDir()
		if err != nil {
			return nil, fmt.Errorf("config: parse from file: create config dir: %v", err.Error())
		}

		return &conf, nil
	}

	fExists, err := FileExists()
	if err != nil {
		return nil, fmt.Errorf("parse from file: file exists: %v", err.Error())
	}

	if !fExists {
		return &conf, nil
	}

	b, err := os.ReadFile(filepath.Join(usr.HomeDir, confDir, confFile))
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %v", err.Error())
	}

	err = yaml.Unmarshal(b, &conf)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal config file: %v", err.Error())
	}

	return &conf, nil
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

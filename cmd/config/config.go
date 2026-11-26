package configcmd

import (
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/gomicro/align/config"

	"github.com/spf13/cobra"
)

var configValidArgs = []string{
	"github.token\tGitHub access token",
	"github.username\tGitHub username",
	"github.private_key\tinline SSH private key",
	"github.private_key_file\tSSH private key file path",
	"github.limits.burst\tburstable rate for the GitHub client",
	"github.limits.requests_per_second\tmaximum requests per second for the GitHub client",
}

// ConfigCmd sets persisted configuration values.
var ConfigCmd = newConfigCommand("")

func newConfigCommand(path string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config <field> [value]",
		Short: "Set align configuration values",
		Long: `Set align configuration values in ~/.align/config.

Valid fields:

  github.token                        GitHub access token
  github.username                     GitHub username
  github.private_key                  inline SSH private key
  github.private_key_file             SSH private key file path
  github.limits.burst                 burstable rate for the GitHub client (default: 25)
  github.limits.requests_per_second   maximum requests per second for the GitHub client (default: 10)

Use --stdin to read the value exactly from standard input. For secrets, prefer
stdin over an argument to avoid shell history and process-list exposure.`,
		Args:      configArgs,
		ValidArgs: configValidArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			err := configFunc(cmd, args, path)
			if err != nil {
				return fmt.Errorf("updating config: %w", err)
			}

			return nil
		},
	}
	cmd.Flags().Bool("stdin", false, "read the value exactly from standard input")
	return cmd
}

func configArgs(cmd *cobra.Command, args []string) error {
	stdin, err := cmd.Flags().GetBool("stdin")
	if err != nil {
		return fmt.Errorf("reading stdin flag: %w", err)
	}

	if stdin && len(args) != 1 {
		return fmt.Errorf("provide exactly one field with --stdin")
	}

	if !stdin && len(args) != 2 {
		return fmt.Errorf("provide a field and value")
	}

	return nil
}

type configUpdate struct {
	field  string
	value  string
	number int
}

func parseUpdate(field, value string) (configUpdate, error) {
	update := configUpdate{field: strings.ToLower(field), value: value}
	switch update.field {
	case "github.token", "github.username", "github.private_key", "github.private_key_file":
		return update, nil

	case "github.limits.burst", "github.limits.requests_per_second":
		number, err := strconv.Atoi(value)
		if err != nil {
			return configUpdate{}, fmt.Errorf("config field requires an integer")
		}

		update.number = number
		return update, nil

	default:
		return configUpdate{}, fmt.Errorf("unrecognized config field")
	}
}

func (u configUpdate) apply(cfg *config.Config) {
	switch u.field {
	case "github.token":
		cfg.Github.Token = u.value

	case "github.username":
		cfg.Github.Username = u.value

	case "github.private_key":
		cfg.Github.PrivateKey = u.value

	case "github.private_key_file":
		cfg.Github.PrivateKeyFile = u.value

	case "github.limits.burst":
		cfg.Github.Limits.Burst = u.number

	case "github.limits.requests_per_second":
		cfg.Github.Limits.RequestsPerSecond = u.number
	}
}

func configFunc(cmd *cobra.Command, args []string, path string) error {
	stdin, err := cmd.Flags().GetBool("stdin")
	if err != nil {
		return fmt.Errorf("reading stdin flag: %w", err)
	}

	var value string
	if stdin {
		data, err := io.ReadAll(cmd.InOrStdin())
		if err != nil {
			return fmt.Errorf("reading stdin: %w", err)
		}

		value = string(data)
	} else {
		value = args[1]
	}

	update, err := parseUpdate(args[0], value)
	if err != nil {
		return fmt.Errorf("validating config update: %w", err)
	}

	if path == "" {
		path, err = config.FilePath()
		if err != nil {
			return fmt.Errorf("finding config: %w", err)
		}
	}

	cfg, err := config.ReadFile(path)
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	update.apply(cfg)
	info, err := os.Stat(path)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("checking config permissions: %w", err)
	}

	if err == nil && info.Mode().Perm()&0077 != 0 {
		// The byte count is not needed; the write error is checked.
		_, err = fmt.Fprintln(cmd.ErrOrStderr(), "warning: config file is accessible by other users")
		if err != nil {
			return fmt.Errorf("writing permission warning: %w", err)
		}
	}

	err = cfg.WriteFileAt(path)
	if err != nil {
		return fmt.Errorf("saving config: %w", err)
	}

	// The byte count is not needed; the write error is checked.
	_, err = fmt.Fprintln(cmd.OutOrStdout(), "Config file updated")
	if err != nil {
		return fmt.Errorf("writing confirmation: %w", err)
	}

	return nil
}

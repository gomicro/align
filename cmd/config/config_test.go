package configcmd

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gomicro/align/config"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v2"
)

const existingConfig = `github.com:
  token: original-token
  username: original-user
  private_key: original-key
  private_key_file: /original/key
  limits:
    request_per_second: 10
    burst: 25
`

func runConfig(t *testing.T, path, input string, args ...string) (string, string, error) {
	t.Helper()

	root := &cobra.Command{Use: "align", SilenceErrors: true, SilenceUsage: true}
	root.AddCommand(newConfigCommand(path))
	var out, stderr bytes.Buffer
	root.SetOut(&out)
	root.SetErr(&stderr)
	root.SetIn(strings.NewReader(input))
	root.SetArgs(append([]string{"config"}, args...))
	err := root.Execute()
	return out.String(), stderr.String(), err
}

func TestSetAllFields(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		field string
		value string
		want  config.GithubHost
	}{
		{name: "token", field: "github.token", value: "new-token", want: config.GithubHost{Token: "new-token", Username: "original-user", PrivateKey: "original-key", PrivateKeyFile: "/original/key", Limits: &config.Limits{RequestsPerSecond: 10, Burst: 25}}},
		{name: "username", field: "GITHUB.USERNAME", value: "new-user", want: config.GithubHost{Token: "original-token", Username: "new-user", PrivateKey: "original-key", PrivateKeyFile: "/original/key", Limits: &config.Limits{RequestsPerSecond: 10, Burst: 25}}},
		{name: "private key", field: "github.private_key", value: "new-key", want: config.GithubHost{Token: "original-token", Username: "original-user", PrivateKey: "new-key", PrivateKeyFile: "/original/key", Limits: &config.Limits{RequestsPerSecond: 10, Burst: 25}}},
		{name: "private key file", field: "github.private_key_file", value: "/new/key", want: config.GithubHost{Token: "original-token", Username: "original-user", PrivateKey: "original-key", PrivateKeyFile: "/new/key", Limits: &config.Limits{RequestsPerSecond: 10, Burst: 25}}},
		{name: "rate zero", field: "github.limits.requests_per_second", value: "0", want: config.GithubHost{Token: "original-token", Username: "original-user", PrivateKey: "original-key", PrivateKeyFile: "/original/key", Limits: &config.Limits{RequestsPerSecond: 0, Burst: 25}}},
		{name: "negative burst", field: "github.limits.burst", value: "-2", want: config.GithubHost{Token: "original-token", Username: "original-user", PrivateKey: "original-key", PrivateKeyFile: "/original/key", Limits: &config.Limits{RequestsPerSecond: 10, Burst: -2}}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "config")
			err := os.WriteFile(path, []byte(existingConfig), 0600)
			if !assert.NoError(t, err) {
				return
			}

			out, stderr, err := runConfig(t, path, "", tt.field, "--", tt.value)
			if !assert.NoError(t, err) {
				return
			}
			assert.NotContains(t, out+stderr, tt.value)

			data, err := os.ReadFile(path)
			if !assert.NoError(t, err) {
				return
			}
			var cfg config.Config
			err = yaml.Unmarshal(data, &cfg)
			if !assert.NoError(t, err) || !assert.NotNil(t, cfg.Github) {
				return
			}
			assert.Equal(t, tt.want, *cfg.Github)
		})
	}
}

func TestStdinAndEmptyValues(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input string
		args  []string
		want  string
	}{
		{name: "multiline stdin", input: "-----BEGIN KEY-----\nline\n-----END KEY-----\n", args: []string{"github.private_key", "--stdin"}, want: "-----BEGIN KEY-----\nline\n-----END KEY-----\n"},
		{name: "empty stdin", args: []string{"github.private_key", "--stdin"}},
		{name: "empty argument", args: []string{"github.private_key", ""}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "config")
			err := os.WriteFile(path, []byte(existingConfig), 0600)
			if !assert.NoError(t, err) {
				return
			}

			out, stderr, err := runConfig(t, path, tt.input, tt.args...)
			if !assert.NoError(t, err) {
				return
			}
			if tt.input != "" {
				assert.NotContains(t, out+stderr, tt.input)
			}

			data, err := os.ReadFile(path)
			if !assert.NoError(t, err) {
				return
			}
			var cfg config.Config
			err = yaml.Unmarshal(data, &cfg)
			if !assert.NoError(t, err) || !assert.NotNil(t, cfg.Github) {
				return
			}
			assert.Equal(t, tt.want, cfg.Github.PrivateKey)
			assert.Equal(t, "original-token", cfg.Github.Token)
		})
	}
}

func TestInvalidUpdatesDoNotWrite(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input string
		args  []string
	}{
		{name: "unknown field", args: []string{"github.unknown", "value"}},
		{name: "invalid rate", args: []string{"github.limits.requests_per_second", "not-a-number"}},
		{name: "missing value", args: []string{"github.token"}},
		{name: "missing field", args: []string{"--stdin"}},
		{name: "conflicting value", input: "secret", args: []string{"github.token", "other-secret", "--stdin"}},
		{name: "extra value", args: []string{"github.token", "one", "two"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "config")
			err := os.WriteFile(path, []byte(existingConfig), 0600)
			if !assert.NoError(t, err) {
				return
			}

			out, stderr, err := runConfig(t, path, tt.input, tt.args...)
			assert.Error(t, err)
			assert.NotContains(t, out+stderr, "other-secret")
			if tt.input != "" {
				assert.NotContains(t, out+stderr, tt.input)
			}
			if err != nil {
				assert.NotContains(t, err.Error(), "other-secret")
				if tt.input != "" {
					assert.NotContains(t, err.Error(), tt.input)
				}
			}
			data, readErr := os.ReadFile(path)
			if assert.NoError(t, readErr) {
				assert.Equal(t, existingConfig, string(data))
			}
		})
	}
}

func TestConfigReadErrorDoesNotWrite(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "config")
	invalid := []byte("github.com: [\n")
	err := os.WriteFile(path, invalid, 0600)
	if !assert.NoError(t, err) {
		return
	}

	out, stderr, err := runConfig(t, path, "", "github.token", "secret-value")
	assert.Error(t, err)
	assert.NotContains(t, out+stderr, "secret-value")
	if err != nil {
		assert.NotContains(t, err.Error(), "secret-value")
	}
	data, readErr := os.ReadFile(path)
	if assert.NoError(t, readErr) {
		assert.Equal(t, invalid, data)
	}
}

func TestInvalidUpdateDoesNotCreateConfig(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "missing", "config")
	out, stderr, err := runConfig(t, path, "", "github.limits.burst", "invalid")
	assert.Error(t, err)
	assert.Empty(t, out)
	assert.Empty(t, stderr)
	assert.NoDirExists(t, filepath.Dir(path))
}

func TestNumericStdin(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "new", "config")
	out, stderr, err := runConfig(t, path, "-12", "github.limits.requests_per_second", "--stdin")
	if !assert.NoError(t, err) {
		return
	}
	assert.NotContains(t, out+stderr, "-12")

	data, err := os.ReadFile(path)
	if !assert.NoError(t, err) {
		return
	}
	var cfg config.Config
	err = yaml.Unmarshal(data, &cfg)
	if !assert.NoError(t, err) || !assert.NotNil(t, cfg.Github) || !assert.NotNil(t, cfg.Github.Limits) {
		return
	}
	assert.Equal(t, -12, cfg.Github.Limits.RequestsPerSecond)
}

func TestConfigPermissions(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		mode    os.FileMode
		exists  bool
		warning bool
	}{
		{name: "new file", mode: 0600},
		{name: "private file", mode: 0600, exists: true},
		{name: "shared file", mode: 0644, exists: true, warning: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "config")
			if tt.exists {
				err := os.WriteFile(path, []byte(existingConfig), tt.mode)
				if !assert.NoError(t, err) {
					return
				}
				err = os.Chmod(path, tt.mode)
				if !assert.NoError(t, err) {
					return
				}
			}
			out, stderr, err := runConfig(t, path, "", "github.token", "secret-value")
			if !assert.NoError(t, err) {
				return
			}
			assert.NotContains(t, out+stderr, "secret-value")
			assert.Equal(t, tt.warning, strings.Contains(strings.ToLower(stderr), "warning"))
			info, err := os.Stat(path)
			if assert.NoError(t, err) {
				assert.Equal(t, tt.mode, info.Mode().Perm())
			}
		})
	}
}

func TestConfigHelpAndCompletion(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "config")
	help, _, err := runConfig(t, path, "", "--help")
	if !assert.NoError(t, err) {
		return
	}
	for _, field := range []string{"github.token", "github.username", "github.private_key", "github.private_key_file", "github.limits.requests_per_second", "github.limits.burst"} {
		assert.Contains(t, help, field)
	}
	assert.Contains(t, help, "--stdin")

	root := &cobra.Command{Use: "align", SilenceErrors: true, SilenceUsage: true}
	root.AddCommand(newConfigCommand(path))
	var out bytes.Buffer
	root.SetOut(&out)
	root.SetErr(&bytes.Buffer{})
	root.SetArgs([]string{"__complete", "config", "github."})
	err = root.Execute()
	if !assert.NoError(t, err) {
		return
	}
	for _, field := range []string{"github.token", "github.username", "github.private_key", "github.private_key_file", "github.limits.requests_per_second", "github.limits.burst"} {
		assert.Contains(t, out.String(), field+"\t")
	}
}

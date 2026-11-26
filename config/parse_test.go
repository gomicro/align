package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v2"
)

func TestParseConfigDefaults(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    string
		token    string
		username string
		rate     int
		burst    int
	}{
		{name: "empty", rate: 10, burst: 25},
		{name: "null document", input: "null\n", rate: 10, burst: 25},
		{name: "missing github", input: "{}\n", rate: 10, burst: 25},
		{name: "null github", input: "github.com: null\n", rate: 10, burst: 25},
		{name: "missing limits", input: "github.com:\n  token: example\n", token: "example", rate: 10, burst: 25},
		{name: "null limits", input: "github.com:\n  token: example\n  limits: null\n", token: "example", rate: 10, burst: 25},
		{name: "partial limits", input: "github.com:\n  username: example\n  limits:\n    burst: 7\n", username: "example", rate: 10, burst: 7},
		{name: "explicit zero", input: "github.com:\n  limits:\n    request_per_second: 0\n    burst: 0\n"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg, err := parseConfig([]byte(tt.input))
			if !assert.NoError(t, err) || !assert.NotNil(t, cfg.Github) || !assert.NotNil(t, cfg.Github.Limits) {
				return
			}
			assert.Equal(t, tt.token, cfg.Github.Token)
			assert.Equal(t, tt.username, cfg.Github.Username)
			assert.Equal(t, tt.rate, cfg.Github.Limits.RequestsPerSecond)
			assert.Equal(t, tt.burst, cfg.Github.Limits.Burst)
		})
	}
}

func TestParseConfigInvalidYAML(t *testing.T) {
	t.Parallel()

	_, err := parseConfig([]byte("github.com: [\n"))
	assert.Error(t, err)
}

func TestTokenUpdatePreservesConfiguration(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    string
		username string
		burst    int
	}{
		{name: "null github", input: "github.com: null\n", burst: 25},
		{name: "null limits", input: "github.com:\n  username: example\n  limits: null\n", username: "example", burst: 25},
		{name: "configured", input: "github.com:\n  username: example\n  limits:\n    burst: 7\n", username: "example", burst: 7},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg, err := parseConfig([]byte(tt.input))
			if !assert.NoError(t, err) || !assert.NotNil(t, cfg.Github) {
				return
			}

			cfg.Github.Token = "new-token"
			data, err := yaml.Marshal(cfg)
			if !assert.NoError(t, err) {
				return
			}

			restored, err := parseConfig(data)
			if !assert.NoError(t, err) || !assert.NotNil(t, restored.Github) || !assert.NotNil(t, restored.Github.Limits) {
				return
			}

			assert.Equal(t, "new-token", restored.Github.Token)
			assert.Equal(t, tt.username, restored.Github.Username)
			assert.Equal(t, 10, restored.Github.Limits.RequestsPerSecond)
			assert.Equal(t, tt.burst, restored.Github.Limits.Burst)
		})
	}
}

func TestConfigurationsDoNotShareDefaults(t *testing.T) {
	t.Parallel()

	first, err := parseConfig(nil)
	if !assert.NoError(t, err) || !assert.NotNil(t, first.Github) || !assert.NotNil(t, first.Github.Limits) {
		return
	}
	first.Github.Token = "example"
	first.Github.Limits.Burst = 1

	second, err := parseConfig(nil)
	if !assert.NoError(t, err) || !assert.NotNil(t, second.Github) || !assert.NotNil(t, second.Github.Limits) {
		return
	}
	assert.Empty(t, second.Github.Token)
	assert.Equal(t, 25, second.Github.Limits.Burst)

	created := New("another")
	if !assert.NotNil(t, created.Github) || !assert.NotNil(t, created.Github.Limits) {
		return
	}
	assert.Equal(t, "another", created.Github.Token)
	assert.Equal(t, 25, created.Github.Limits.Burst)
}

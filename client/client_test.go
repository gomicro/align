package client

import (
	"testing"

	"github.com/gomicro/align/config"

	"github.com/stretchr/testify/assert"
)

func TestNewWithMissingConfiguration(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		cfg  *config.Config
	}{
		{name: "nil config"},
		{name: "missing github", cfg: &config.Config{}},
		{name: "missing limits", cfg: &config.Config{Github: &config.GithubHost{Token: "example"}}},
		{name: "configured", cfg: &config.Config{Github: &config.GithubHost{Limits: &config.Limits{RequestsPerSecond: 2, Burst: 3}}}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clt, err := New(tt.cfg)
			if !assert.NoError(t, err) || !assert.NotNil(t, clt) || !assert.NotNil(t, clt.cfg.Github) || !assert.NotNil(t, clt.cfg.Github.Limits) {
				return
			}
			if tt.name == "configured" {
				assert.Equal(t, 2, clt.cfg.Github.Limits.RequestsPerSecond)
				assert.Equal(t, 3, clt.cfg.Github.Limits.Burst)
				return
			}
			assert.Equal(t, 10, clt.cfg.Github.Limits.RequestsPerSecond)
			assert.Equal(t, 25, clt.cfg.Github.Limits.Burst)
			if tt.name == "missing limits" {
				assert.Equal(t, "example", clt.cfg.Github.Token)
				assert.Nil(t, tt.cfg.Github.Limits)
			}
		})
	}
}

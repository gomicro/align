package cmd

import (
	"errors"
	"testing"

	"github.com/gomicro/align/client/testclient"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func TestPull(t *testing.T) {
	viper.Set("verbose", true)
	t.Cleanup(func() { viper.Set("verbose", false) })

	t.Run("calls expected commands", func(t *testing.T) {
		tc := testclient.New()
		clt = tc

		err := pullFunc(pullCmd, []string{})
		assert.NoError(t, err)

		tc.AssertCommandsCalled(t, "GetDirs", "PullRepos")
	})

	t.Run("returns error on get dirs failure", func(t *testing.T) {
		tc := testclient.New()
		tc.Errors["GetDirs"] = errors.New("some dirs error")
		clt = tc

		err := pullFunc(pullCmd, []string{})
		assert.ErrorContains(t, err, "get dirs")

		tc.AssertCommandsCalled(t, "GetDirs")
	})

	t.Run("returns error on pull repos failure", func(t *testing.T) {
		tc := testclient.New()
		tc.Errors["PullRepos"] = errors.New("some pull error")
		clt = tc

		err := pullFunc(pullCmd, []string{})
		assert.ErrorContains(t, err, "pull repos")

		tc.AssertCommandsCalled(t, "GetDirs", "PullRepos")
	})

	t.Run("passes --prune flag", func(t *testing.T) {
		prune = true
		t.Cleanup(func() { prune = false })

		tc := testclient.New()
		clt = tc

		err := pullFunc(pullCmd, []string{})
		assert.NoError(t, err)

		tc.AssertCommandsCalled(t, "GetDirs", "PullRepos")
	})
}

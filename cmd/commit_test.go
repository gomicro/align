package cmd

import (
	"errors"
	"testing"

	"github.com/gomicro/align/client/testclient"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func TestCommit(t *testing.T) {
	viper.Set("verbose", true)
	t.Cleanup(func() { viper.Set("verbose", false) })

	t.Run("calls expected commands", func(t *testing.T) {
		message = "test commit"
		t.Cleanup(func() { message = "" })

		tc := testclient.New()
		clt = tc

		err := commitFunc(commitCmd, []string{})
		assert.NoError(t, err)

		tc.AssertCommandsCalled(t, "GetDirs", "CommitRepos")
	})

	t.Run("returns error on get dirs failure", func(t *testing.T) {
		message = "test commit"
		t.Cleanup(func() { message = "" })

		tc := testclient.New()
		tc.Errors["GetDirs"] = errors.New("some dirs error")
		clt = tc

		err := commitFunc(commitCmd, []string{})
		assert.ErrorContains(t, err, "get dirs")

		tc.AssertCommandsCalled(t, "GetDirs")
	})

	t.Run("returns error on commit repos failure", func(t *testing.T) {
		message = "test commit"
		t.Cleanup(func() { message = "" })

		tc := testclient.New()
		tc.Errors["CommitRepos"] = errors.New("some commit error")
		clt = tc

		err := commitFunc(commitCmd, []string{})
		assert.ErrorContains(t, err, "commit repos")

		tc.AssertCommandsCalled(t, "GetDirs", "CommitRepos")
	})
}

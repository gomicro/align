package cmd

import (
	"errors"
	"testing"

	"github.com/gomicro/align/client/testclient"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func TestFetch(t *testing.T) {
	viper.Set("verbose", true)
	t.Cleanup(func() { viper.Set("verbose", false) })

	t.Run("calls expected commands", func(t *testing.T) {
		tc := testclient.New()
		clt = tc

		err := fetchFunc(fetchCmd, []string{})
		assert.NoError(t, err)

		tc.AssertCommandsCalled(t, "GetDirs", "FetchRepos")
	})

	t.Run("calls expected commands with remote arg", func(t *testing.T) {
		tc := testclient.New()
		clt = tc

		err := fetchFunc(fetchCmd, []string{"origin"})
		assert.NoError(t, err)

		tc.AssertCommandsCalled(t, "GetDirs", "FetchRepos")
	})

	t.Run("returns error on get dirs failure", func(t *testing.T) {
		tc := testclient.New()
		tc.Errors["GetDirs"] = errors.New("some dirs error")
		clt = tc

		err := fetchFunc(fetchCmd, []string{})
		assert.ErrorContains(t, err, "get dirs")

		tc.AssertCommandsCalled(t, "GetDirs")
	})

	t.Run("returns error on fetch repos failure", func(t *testing.T) {
		tc := testclient.New()
		tc.Errors["FetchRepos"] = errors.New("some fetch error")
		clt = tc

		err := fetchFunc(fetchCmd, []string{})
		assert.ErrorContains(t, err, "fetch repos")

		tc.AssertCommandsCalled(t, "GetDirs", "FetchRepos")
	})
}

package stash

import (
	"errors"
	"testing"

	"github.com/gomicro/align/client/testclient"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func TestStashDrop(t *testing.T) {
	viper.Set("verbose", true)
	t.Cleanup(func() { viper.Set("verbose", false) })

	t.Run("calls expected commands", func(t *testing.T) {
		tc := testclient.New()
		clt = tc

		err := dropFunc(dropCmd, []string{})
		assert.NoError(t, err)

		tc.AssertCommandsCalled(t, "GetDirs", "StashRepos")
	})

	t.Run("returns error on get dirs failure", func(t *testing.T) {
		tc := testclient.New()
		tc.Errors["GetDirs"] = errors.New("some dirs error")
		clt = tc

		err := dropFunc(dropCmd, []string{})
		assert.ErrorContains(t, err, "get dirs")

		tc.AssertCommandsCalled(t, "GetDirs")
	})

	t.Run("returns error on stash drop failure", func(t *testing.T) {
		tc := testclient.New()
		tc.Errors["StashRepos"] = errors.New("some stash drop error")
		clt = tc

		err := dropFunc(dropCmd, []string{})
		assert.ErrorContains(t, err, "stash drop")

		tc.AssertCommandsCalled(t, "GetDirs", "StashRepos")
	})
}

package stash

import (
	"errors"
	"testing"

	"github.com/gomicro/align/client/testclient"
	"github.com/stretchr/testify/assert"
)

func TestStashList(t *testing.T) {
	t.Run("calls expected commands", func(t *testing.T) {
	})

	t.Run("returns error on get dirs failure", func(t *testing.T) {
		tc := testclient.New()
		tc.Errors["GetDirs"] = errors.New("some dirs error")
		clt = tc

		err := listFunc(listCmd, []string{})
		assert.ErrorContains(t, err, "get dirs")

		tc.AssertCommandsCalled(t, "GetDirs")
	})

	t.Run("returns error on stash list failure", func(t *testing.T) {
		tc := testclient.New()
		tc.Errors["StashRepos"] = errors.New("some stash list error")
		clt = tc

		err := listFunc(listCmd, []string{})
		assert.ErrorContains(t, err, "stash list")

		tc.AssertCommandsCalled(t, "GetDirs", "StashRepos")
	})
}

package cmd

import (
	"errors"
	"testing"

	"github.com/gomicro/align/client/testclient"
	"github.com/stretchr/testify/assert"
)

func TestDiff(t *testing.T) {
	t.Run("calls expected commands", func(t *testing.T) {
		tc := testclient.New()
		clt = tc

		err := diffFunc(diffCmd, []string{})
		assert.NoError(t, err)

		tc.AssertCommandsCalled(t, "GetDirs", "DiffRepos")
	})

	t.Run("returns error on get dirs failure", func(t *testing.T) {
		tc := testclient.New()
		tc.Errors["GetDirs"] = errors.New("some dirs error")
		clt = tc

		err := diffFunc(diffCmd, []string{})
		assert.ErrorContains(t, err, "get dirs")

		tc.AssertCommandsCalled(t, "GetDirs")
	})

	t.Run("returns error on diff repos failure", func(t *testing.T) {
		tc := testclient.New()
		tc.Errors["DiffRepos"] = errors.New("some diff error")
		clt = tc

		err := diffFunc(diffCmd, []string{})
		assert.ErrorContains(t, err, "diff repos")

		tc.AssertCommandsCalled(t, "GetDirs", "DiffRepos")
	})
}

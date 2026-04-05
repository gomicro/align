package cmd

import (
	"errors"
	"testing"

	"github.com/gomicro/align/client/testclient"
	"github.com/stretchr/testify/assert"
)

func TestStatus(t *testing.T) {
	t.Run("calls expected commands", func(t *testing.T) {
		tc := testclient.New()
		clt = tc

		err := statusFunc(statusCmd, []string{})
		assert.NoError(t, err)

		tc.AssertCommandsCalled(t, "GetDirs", "StatusRepos")
	})

	t.Run("returns error on get dirs failure", func(t *testing.T) {
		tc := testclient.New()
		tc.Errors["GetDirs"] = errors.New("some dirs error")
		clt = tc

		err := statusFunc(statusCmd, []string{})
		assert.ErrorContains(t, err, "get dirs")

		tc.AssertCommandsCalled(t, "GetDirs")
	})

	t.Run("returns error on status repos failure", func(t *testing.T) {
		tc := testclient.New()
		tc.Errors["StatusRepos"] = errors.New("some status error")
		clt = tc

		err := statusFunc(statusCmd, []string{})
		assert.ErrorContains(t, err, "status repos")

		tc.AssertCommandsCalled(t, "GetDirs", "StatusRepos")
	})
}

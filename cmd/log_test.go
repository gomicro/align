package cmd

import (
	"errors"
	"testing"

	"github.com/gomicro/align/client/testclient"
	"github.com/stretchr/testify/assert"
)

func TestLog(t *testing.T) {
	t.Run("calls expected commands", func(t *testing.T) {
		tc := testclient.New()
		clt = tc

		err := logFunc(logCmd, []string{})
		assert.NoError(t, err)

		tc.AssertCommandsCalled(t, "GetDirs", "LogRepos")
	})

	t.Run("returns error on get dirs failure", func(t *testing.T) {
		tc := testclient.New()
		tc.Errors["GetDirs"] = errors.New("some dirs error")
		clt = tc

		err := logFunc(logCmd, []string{})
		assert.ErrorContains(t, err, "get dirs")

		tc.AssertCommandsCalled(t, "GetDirs")
	})

	t.Run("returns error on log repos failure", func(t *testing.T) {
		tc := testclient.New()
		tc.Errors["LogRepos"] = errors.New("some log error")
		clt = tc

		err := logFunc(logCmd, []string{})
		assert.ErrorContains(t, err, "log repos")

		tc.AssertCommandsCalled(t, "GetDirs", "LogRepos")
	})

	t.Run("passes --max-count flag", func(t *testing.T) {
		t.Parallel()

		maxCount = 10
		t.Cleanup(func() { maxCount = 0 })

		tc := testclient.New()
		clt = tc

		err := logFunc(logCmd, []string{})
		assert.NoError(t, err)

		tc.AssertCommandsCalled(t, "GetDirs", "LogRepos")
	})
}

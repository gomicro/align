package cmd

import (
	"errors"
	"testing"

	"github.com/gomicro/align/client/testclient"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func TestMerge(t *testing.T) {
	viper.Set("verbose", true)
	t.Cleanup(func() { viper.Set("verbose", false) })

	t.Run("calls expected commands", func(t *testing.T) {
		tc := testclient.New()
		clt = tc

		err := mergeFunc(mergeCmd, []string{"main"})
		assert.NoError(t, err)

		tc.AssertCommandsCalled(t, "GetDirs", "MergeRepos")
	})

	t.Run("aborts in-progress merge", func(t *testing.T) {
		abortMerge = true
		t.Cleanup(func() { abortMerge = false })

		tc := testclient.New()
		clt = tc

		err := mergeFunc(mergeCmd, []string{})
		assert.NoError(t, err)

		tc.AssertCommandsCalled(t, "GetDirs", "MergeRepos")
	})

	t.Run("passes --ff-only flag", func(t *testing.T) {
		ffOnly = true
		t.Cleanup(func() { ffOnly = false })

		tc := testclient.New()
		clt = tc

		err := mergeFunc(mergeCmd, []string{"main"})
		assert.NoError(t, err)

		tc.AssertCommandsCalled(t, "GetDirs", "MergeRepos")
	})

	t.Run("returns error on get dirs failure", func(t *testing.T) {
		tc := testclient.New()
		tc.Errors["GetDirs"] = errors.New("some dirs error")
		clt = tc

		err := mergeFunc(mergeCmd, []string{"main"})
		assert.ErrorContains(t, err, "get dirs")

		tc.AssertCommandsCalled(t, "GetDirs")
	})

	t.Run("returns error on merge repos failure", func(t *testing.T) {
		tc := testclient.New()
		tc.Errors["MergeRepos"] = errors.New("some merge error")
		clt = tc

		err := mergeFunc(mergeCmd, []string{"main"})
		assert.ErrorContains(t, err, "merge repos")

		tc.AssertCommandsCalled(t, "GetDirs", "MergeRepos")
	})
}

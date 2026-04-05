package cmd

import (
	"errors"
	"testing"

	"github.com/gomicro/align/client/testclient"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func TestTag(t *testing.T) {
	viper.Set("verbose", true)
	t.Cleanup(func() { viper.Set("verbose", false) })

	t.Run("lists tags when no args", func(t *testing.T) {
		tc := testclient.New()
		clt = tc

		err := tagFunc(tagCmd, []string{})
		assert.NoError(t, err)

		tc.AssertCommandsCalled(t, "GetDirs", "ListTags")
	})

	t.Run("lists tags with list flag", func(t *testing.T) {
		list = true
		t.Cleanup(func() { list = false })

		tc := testclient.New()
		clt = tc

		err := tagFunc(tagCmd, []string{})
		assert.NoError(t, err)

		tc.AssertCommandsCalled(t, "GetDirs", "ListTags")
	})

	t.Run("creates tag", func(t *testing.T) {
		message = "release"
		t.Cleanup(func() { message = "" })

		tc := testclient.New()
		clt = tc

		err := tagFunc(tagCmd, []string{"v1.0.0"})
		assert.NoError(t, err)

		tc.AssertCommandsCalled(t, "GetDirs", "TagRepos")
	})

	t.Run("deletes tag", func(t *testing.T) {
		del = true
		t.Cleanup(func() { del = false })

		tc := testclient.New()
		clt = tc

		err := tagFunc(tagCmd, []string{"v1.0.0"})
		assert.NoError(t, err)

		tc.AssertCommandsCalled(t, "GetDirs", "TagRepos")
	})

	t.Run("returns error on get dirs failure", func(t *testing.T) {
		tc := testclient.New()
		tc.Errors["GetDirs"] = errors.New("some dirs error")
		clt = tc

		err := tagFunc(tagCmd, []string{})
		assert.ErrorContains(t, err, "get dirs")

		tc.AssertCommandsCalled(t, "GetDirs")
	})

	t.Run("returns error on list tags failure", func(t *testing.T) {
		tc := testclient.New()
		tc.Errors["ListTags"] = errors.New("some list tags error")
		clt = tc

		err := tagFunc(tagCmd, []string{})
		assert.ErrorContains(t, err, "list tags")

		tc.AssertCommandsCalled(t, "GetDirs", "ListTags")
	})

	t.Run("returns error on tag repos failure", func(t *testing.T) {
		message = "release"
		t.Cleanup(func() { message = "" })

		tc := testclient.New()
		tc.Errors["TagRepos"] = errors.New("some tag error")
		clt = tc

		err := tagFunc(tagCmd, []string{"v1.0.0"})
		assert.ErrorContains(t, err, "tagging")

		tc.AssertCommandsCalled(t, "GetDirs", "TagRepos")
	})

	t.Run("returns error when sign requires message", func(t *testing.T) {
		sign = true
		message = ""
		t.Cleanup(func() {
			sign = false
			message = ""
		})

		tc := testclient.New()
		clt = tc

		err := tagFunc(tagCmd, []string{"v1.0.0"})
		assert.ErrorContains(t, err, "--message is required")

		tc.AssertCommandsCalled(t, "GetDirs")
	})
}

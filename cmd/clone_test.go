package cmd

import (
	"errors"
	"testing"

	"github.com/gomicro/align/client/testclient"
	"github.com/google/go-github/github"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func TestClone(t *testing.T) {
	viper.Set("verbose", true)
	t.Cleanup(func() { viper.Set("verbose", false) })

	t.Run("calls expected commands", func(t *testing.T) {
		tc := testclient.New()
		clt = tc

		err := cloneFunc(cloneCmd, []string{})
		assert.NoError(t, err)

		tc.AssertCommandsCalled(t, "GetRepos", "CloneRepos")
	})

	t.Run("calls expected commands with name arg", func(t *testing.T) {
		tc := testclient.New()
		clt = tc

		err := cloneFunc(cloneCmd, []string{"some-org"})
		assert.NoError(t, err)

		tc.AssertCommandsCalled(t, "GetRepos", "CloneRepos")
	})

	t.Run("returns error on get repos failure", func(t *testing.T) {
		tc := testclient.New()
		tc.Errors["GetRepos"] = errors.New("some get repos error")
		clt = tc

		err := cloneFunc(cloneCmd, []string{})
		assert.ErrorContains(t, err, "get repos")

		tc.AssertCommandsCalled(t, "GetRepos")
	})

	t.Run("returns error on clone repos failure", func(t *testing.T) {
		tc := testclient.New()
		tc.Errors["CloneRepos"] = errors.New("some clone error")
		clt = tc

		err := cloneFunc(cloneCmd, []string{})
		assert.ErrorContains(t, err, "clone repos")

		tc.AssertCommandsCalled(t, "GetRepos", "CloneRepos")
	})
}

func TestFilterByTopics(t *testing.T) {
	repoWithTopics := func(topics ...string) *github.Repository {
		return &github.Repository{Topics: topics}
	}

	t.Run("returns all repos when no topics filter", func(t *testing.T) {
		repos := []*github.Repository{
			repoWithTopics("go", "cli"),
			repoWithTopics("rust"),
		}

		result := filterByTopics(repos, []string{})
		assert.Equal(t, repos, result)
	})

	t.Run("filters repos by single topic", func(t *testing.T) {
		match := repoWithTopics("go", "cli")
		noMatch := repoWithTopics("rust")

		result := filterByTopics([]*github.Repository{match, noMatch}, []string{"go"})
		assert.Equal(t, []*github.Repository{match}, result)
	})

	t.Run("filters repos requiring all topics present", func(t *testing.T) {
		both := repoWithTopics("go", "cli")
		onlyOne := repoWithTopics("go")

		result := filterByTopics([]*github.Repository{both, onlyOne}, []string{"go", "cli"})
		assert.Equal(t, []*github.Repository{both}, result)
	})

	t.Run("returns nil when no repos match", func(t *testing.T) {
		repos := []*github.Repository{
			repoWithTopics("rust"),
		}

		result := filterByTopics(repos, []string{"go"})
		assert.Nil(t, result)
	})
}

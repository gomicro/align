package cmd

import (
	"errors"
	"testing"

	"github.com/gomicro/align/client/testclient"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func TestBranch(t *testing.T) {
	viper.Set("verbose", true)
	t.Cleanup(func() { viper.Set("verbose", false) })

	t.Run("lists branches", func(t *testing.T) {
		tc := testclient.New()
		clt = tc

		err := branchFunc(branchCmd, []string{})
		assert.NoError(t, err)

		tc.AssertCommandsCalled(t, "GetDirs", "Branches")
	})

	t.Run("deletes branch", func(t *testing.T) {
		del = true
		t.Cleanup(func() { del = false })

		tc := testclient.New()
		clt = tc

		err := branchFunc(branchCmd, []string{"main"})
		assert.NoError(t, err)

		tc.AssertCommandsCalled(t, "GetDirs", "Branches")
	})

	t.Run("force deletes branch", func(t *testing.T) {
		delForce = true
		t.Cleanup(func() { delForce = false })

		tc := testclient.New()
		clt = tc

		err := branchFunc(branchCmd, []string{"main"})
		assert.NoError(t, err)

		tc.AssertCommandsCalled(t, "GetDirs", "Branches")
	})

	t.Run("renames branch", func(t *testing.T) {
		moveBranch = true
		t.Cleanup(func() { moveBranch = false })

		tc := testclient.New()
		clt = tc

		err := branchFunc(branchCmd, []string{"old-name", "new-name"})
		assert.NoError(t, err)

		tc.AssertCommandsCalled(t, "GetDirs", "Branches")
	})

	t.Run("returns error when renaming without both branch names", func(t *testing.T) {
		moveBranch = true
		t.Cleanup(func() { moveBranch = false })

		tc := testclient.New()
		clt = tc

		err := branchFunc(branchCmd, []string{"only-one"})
		assert.ErrorContains(t, err, "old and new branch names are required")

		tc.AssertCommandsCalled(t, "GetDirs")
	})

	t.Run("returns error when deleting without branch name", func(t *testing.T) {
		del = true
		t.Cleanup(func() { del = false })

		tc := testclient.New()
		clt = tc

		err := branchFunc(branchCmd, []string{})
		assert.ErrorContains(t, err, "branch name is required")

		tc.AssertCommandsCalled(t, "GetDirs")
	})

	t.Run("returns error on get dirs failure", func(t *testing.T) {
		tc := testclient.New()
		tc.Errors["GetDirs"] = errors.New("some dirs error")
		clt = tc

		err := branchFunc(branchCmd, []string{})
		assert.ErrorContains(t, err, "get dirs")

		tc.AssertCommandsCalled(t, "GetDirs")
	})

	t.Run("returns error on branches list failure", func(t *testing.T) {
		tc := testclient.New()
		tc.Errors["Branches"] = errors.New("some branches error")
		clt = tc

		err := branchFunc(branchCmd, []string{})
		assert.ErrorContains(t, err, "list")

		tc.AssertCommandsCalled(t, "GetDirs", "Branches")
	})

	t.Run("returns error on branches delete failure", func(t *testing.T) {
		del = true
		t.Cleanup(func() { del = false })

		tc := testclient.New()
		tc.Errors["Branches"] = errors.New("some branches error")
		clt = tc

		err := branchFunc(branchCmd, []string{"main"})
		assert.ErrorContains(t, err, "delete")

		tc.AssertCommandsCalled(t, "GetDirs", "Branches")
	})

	t.Run("returns error on branches rename failure", func(t *testing.T) {
		moveBranch = true
		t.Cleanup(func() { moveBranch = false })

		tc := testclient.New()
		tc.Errors["Branches"] = errors.New("some branches error")
		clt = tc

		err := branchFunc(branchCmd, []string{"old-name", "new-name"})
		assert.ErrorContains(t, err, "move")

		tc.AssertCommandsCalled(t, "GetDirs", "Branches")
	})
}

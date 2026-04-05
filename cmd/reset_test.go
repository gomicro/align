package cmd

import (
	"errors"
	"testing"

	"github.com/gomicro/align/client/testclient"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func TestReset(t *testing.T) {
	viper.Set("verbose", true)
	t.Cleanup(func() { viper.Set("verbose", false) })

	t.Run("hard reset calls expected commands", func(t *testing.T) {
		hard = true
		t.Cleanup(func() { hard = false })

		tc := testclient.New()
		clt = tc

		err := resetFunc(resetCmd, []string{})
		assert.NoError(t, err)

		tc.AssertCommandsCalled(t, "GetDirs", "ResetRepos")
	})

	t.Run("soft reset calls expected commands", func(t *testing.T) {
		soft = true
		t.Cleanup(func() { soft = false })

		tc := testclient.New()
		clt = tc

		err := resetFunc(resetCmd, []string{})
		assert.NoError(t, err)

		tc.AssertCommandsCalled(t, "GetDirs", "ResetRepos")
	})

	t.Run("mixed reset calls expected commands", func(t *testing.T) {
		mixed = true
		t.Cleanup(func() { mixed = false })

		tc := testclient.New()
		clt = tc

		err := resetFunc(resetCmd, []string{})
		assert.NoError(t, err)

		tc.AssertCommandsCalled(t, "GetDirs", "ResetRepos")
	})

	t.Run("returns error when no mode flag provided", func(t *testing.T) {
		tc := testclient.New()
		clt = tc

		err := resetFunc(resetCmd, []string{})
		assert.ErrorContains(t, err, "--hard, --soft, or --mixed is required")

		tc.AssertNoCommandsCalled(t)
	})

	t.Run("returns error on get dirs failure", func(t *testing.T) {
		hard = true
		t.Cleanup(func() { hard = false })

		tc := testclient.New()
		tc.Errors["GetDirs"] = errors.New("some dirs error")
		clt = tc

		err := resetFunc(resetCmd, []string{})
		assert.ErrorContains(t, err, "get dirs")

		tc.AssertCommandsCalled(t, "GetDirs")
	})

	t.Run("returns error on reset repos failure", func(t *testing.T) {
		hard = true
		t.Cleanup(func() { hard = false })

		tc := testclient.New()
		tc.Errors["ResetRepos"] = errors.New("some reset error")
		clt = tc

		err := resetFunc(resetCmd, []string{})
		assert.ErrorContains(t, err, "reset repos")

		tc.AssertCommandsCalled(t, "GetDirs", "ResetRepos")
	})
}

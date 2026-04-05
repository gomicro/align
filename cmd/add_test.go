package cmd

import (
	"errors"
	"testing"

	"github.com/gomicro/align/client/testclient"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func TestAdd(t *testing.T) {
	viper.Set("verbose", true)
	t.Cleanup(func() { viper.Set("verbose", false) })

	t.Run("calls expected commands", func(t *testing.T) {
		tc := testclient.New()
		clt = tc

		err := addFunc(addCmd, []string{})
		assert.NoError(t, err)

		tc.AssertCommandsCalled(t, "GetDirs", "StageFiles")
	})

	t.Run("calls expected commands with file args", func(t *testing.T) {
		tc := testclient.New()
		clt = tc

		err := addFunc(addCmd, []string{"file.go"})
		assert.NoError(t, err)

		tc.AssertCommandsCalled(t, "GetDirs", "StageFiles")
	})

	t.Run("returns error on get dirs failure", func(t *testing.T) {
		tc := testclient.New()
		tc.Errors["GetDirs"] = errors.New("some dirs error")
		clt = tc

		err := addFunc(addCmd, []string{})
		assert.ErrorContains(t, err, "get dirs")

		tc.AssertCommandsCalled(t, "GetDirs")
	})

	t.Run("returns error on stage files failure", func(t *testing.T) {
		tc := testclient.New()
		tc.Errors["StageFiles"] = errors.New("some stage error")
		clt = tc

		err := addFunc(addCmd, []string{})
		assert.ErrorContains(t, err, "stage files")

		tc.AssertCommandsCalled(t, "GetDirs", "StageFiles")
	})
}

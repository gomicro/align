package remote

import (
	"errors"
	"testing"

	"github.com/gomicro/align/client/testclient"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func TestRemote(t *testing.T) {
	viper.Set("verbose", true)
	t.Cleanup(func() { viper.Set("verbose", false) })

	t.Run("calls expected commands", func(t *testing.T) {
		tc := testclient.New()
		clt = tc

		err := remoteFunc(RemoteCmd, []string{})
		assert.NoError(t, err)

		tc.AssertCommandsCalled(t, "GetDirs", "Remotes")
	})

	t.Run("returns error on get dirs failure", func(t *testing.T) {
		tc := testclient.New()
		tc.Errors["GetDirs"] = errors.New("some dirs error")
		clt = tc

		err := remoteFunc(RemoteCmd, []string{})
		assert.ErrorContains(t, err, "get dirs")

		tc.AssertCommandsCalled(t, "GetDirs")
	})

	t.Run("returns error on remotes failure", func(t *testing.T) {
		tc := testclient.New()
		tc.Errors["Remotes"] = errors.New("some remotes error")
		clt = tc

		err := remoteFunc(RemoteCmd, []string{})
		assert.ErrorContains(t, err, "remotes")

		tc.AssertCommandsCalled(t, "GetDirs", "Remotes")
	})
}

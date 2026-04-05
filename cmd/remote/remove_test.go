package remote

import (
	"errors"
	"testing"

	"github.com/gomicro/align/client/testclient"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func TestRemoteRemove(t *testing.T) {
	viper.Set("verbose", true)
	t.Cleanup(func() { viper.Set("verbose", false) })

	t.Run("calls expected commands", func(t *testing.T) {
		tc := testclient.New()
		clt = tc

		err := removeFunc(removeCmd, []string{"origin"})
		assert.NoError(t, err)

		tc.AssertCommandsCalled(t, "GetDirs", "Remove")
	})

	t.Run("returns error on get dirs failure", func(t *testing.T) {
		tc := testclient.New()
		tc.Errors["GetDirs"] = errors.New("some dirs error")
		clt = tc

		err := removeFunc(removeCmd, []string{"origin"})
		assert.ErrorContains(t, err, "get dirs")

		tc.AssertCommandsCalled(t, "GetDirs")
	})

	t.Run("returns error on remove failure", func(t *testing.T) {
		tc := testclient.New()
		tc.Errors["Remove"] = errors.New("some remove error")
		clt = tc

		err := removeFunc(removeCmd, []string{"origin"})
		assert.ErrorContains(t, err, "remove")

		tc.AssertCommandsCalled(t, "GetDirs", "Remove")
	})
}

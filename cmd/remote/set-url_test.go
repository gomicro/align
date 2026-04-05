package remote

import (
	"errors"
	"testing"

	"github.com/gomicro/align/client/testclient"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func TestRemoteSetURL(t *testing.T) {
	viper.Set("verbose", true)
	t.Cleanup(func() { viper.Set("verbose", false) })

	t.Run("calls expected commands", func(t *testing.T) {
		tc := testclient.New()
		clt = tc

		err := setURLFunc(setURLCmd, []string{"origin", "https://github.com/org"})
		assert.NoError(t, err)

		tc.AssertCommandsCalled(t, "GetDirs", "SetURLs")
	})

	t.Run("returns error on get dirs failure", func(t *testing.T) {
		tc := testclient.New()
		tc.Errors["GetDirs"] = errors.New("some dirs error")
		clt = tc

		err := setURLFunc(setURLCmd, []string{"origin", "https://github.com/org"})
		assert.ErrorContains(t, err, "get dirs")

		tc.AssertCommandsCalled(t, "GetDirs")
	})

	t.Run("returns error on set url failure", func(t *testing.T) {
		tc := testclient.New()
		tc.Errors["SetURLs"] = errors.New("some set url error")
		clt = tc

		err := setURLFunc(setURLCmd, []string{"origin", "https://github.com/org"})
		assert.ErrorContains(t, err, "set url")

		tc.AssertCommandsCalled(t, "GetDirs", "SetURLs")
	})
}

package remote

import (
	"errors"
	"testing"

	"github.com/gomicro/align/client/testclient"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func TestRemoteAdd(t *testing.T) {
	viper.Set("verbose", true)
	t.Cleanup(func() { viper.Set("verbose", false) })

	t.Run("calls expected commands", func(t *testing.T) {
		tc := testclient.New()
		clt = tc

		err := addFunc(addCmd, []string{"origin", "https://github.com/org"})
		assert.NoError(t, err)

		tc.AssertCommandsCalled(t, "GetDirs", "Add")
	})

	t.Run("returns error on get dirs failure", func(t *testing.T) {
		tc := testclient.New()
		tc.Errors["GetDirs"] = errors.New("some dirs error")
		clt = tc

		err := addFunc(addCmd, []string{"origin", "https://github.com/org"})
		assert.ErrorContains(t, err, "get dirs")

		tc.AssertCommandsCalled(t, "GetDirs")
	})

	t.Run("returns error on add failure", func(t *testing.T) {
		tc := testclient.New()
		tc.Errors["Add"] = errors.New("some add error")
		clt = tc

		err := addFunc(addCmd, []string{"origin", "https://github.com/org"})
		assert.ErrorContains(t, err, "add")

		tc.AssertCommandsCalled(t, "GetDirs", "Add")
	})
}

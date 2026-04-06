package remote

import (
	"errors"
	"testing"

	"github.com/gomicro/align/client/testclient"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func TestRenameRemote(t *testing.T) {
	viper.Set("verbose", true)
	t.Cleanup(func() { viper.Set("verbose", false) })

	t.Run("calls expected commands", func(t *testing.T) {
		tc := testclient.New()
		clt = tc

		err := renameFunc(renameCmd, []string{"origin", "upstream"})
		assert.NoError(t, err)

		tc.AssertCommandsCalled(t, "GetDirs", "Rename")
	})

	t.Run("returns error on get dirs failure", func(t *testing.T) {
		tc := testclient.New()
		tc.Errors["GetDirs"] = errors.New("some dirs error")
		clt = tc

		err := renameFunc(renameCmd, []string{"origin", "upstream"})
		assert.ErrorContains(t, err, "get dirs")

		tc.AssertCommandsCalled(t, "GetDirs")
	})

	t.Run("returns error on rename failure", func(t *testing.T) {
		tc := testclient.New()
		tc.Errors["Rename"] = errors.New("some rename error")
		clt = tc

		err := renameFunc(renameCmd, []string{"origin", "upstream"})
		assert.ErrorContains(t, err, "rename")

		tc.AssertCommandsCalled(t, "GetDirs", "Rename")
	})
}

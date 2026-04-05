package cmd

import (
	"errors"
	"testing"

	"github.com/gomicro/align/client/testclient"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func TestCheckout(t *testing.T) {
	// Set verbose=true to skip uiprogress global state across subtests.
	viper.Set("verbose", true)
	t.Cleanup(func() { viper.Set("verbose", false) })

	t.Run("calls expected commands", func(t *testing.T) {
		tc := testclient.New()
		clt = tc

		err := checkoutFunc(checkoutCmd, []string{"main"})
		assert.NoError(t, err)

		tc.AssertCommandsCalled(t, "GetDirs", "CheckoutRepos")
	})

	t.Run("passes args to checkout", func(t *testing.T) {
		tc := testclient.New()
		clt = tc

		err := checkoutFunc(checkoutCmd, []string{"my-feature-branch"})
		assert.NoError(t, err)

		tc.AssertCommandsCalled(t, "GetDirs", "CheckoutRepos")
	})

	t.Run("returns error on get dirs failure", func(t *testing.T) {
		tc := testclient.New()
		tc.Errors["GetDirs"] = errors.New("some dirs error")
		clt = tc

		err := checkoutFunc(checkoutCmd, []string{"main"})
		assert.ErrorContains(t, err, "get dirs")

		tc.AssertCommandsCalled(t, "GetDirs")
	})

	t.Run("returns error on checkout failure", func(t *testing.T) {
		tc := testclient.New()
		tc.Errors["CheckoutRepos"] = errors.New("some checkout error")
		clt = tc

		err := checkoutFunc(checkoutCmd, []string{"main"})
		assert.ErrorContains(t, err, "checkout repos")

		tc.AssertCommandsCalled(t, "GetDirs", "CheckoutRepos")
	})
}

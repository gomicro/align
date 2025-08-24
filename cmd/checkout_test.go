package cmd

import (
	"testing"

	"github.com/gomicro/align/client/testclient"
	"github.com/stretchr/testify/assert"
)

func TestCheckout(t *testing.T) {
	tc := testclient.New()
	clt = tc

	args := []string{}

	err := checkoutFunc(nil, args)
	assert.NoError(t, err)

	tc.AssertCommandsCalled(t, "GetDirs", "CheckoutRepos")
	tc.ResetCommandsCalled()
}

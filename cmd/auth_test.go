package cmd

import (
	"context"
	"os"
	"os/exec"
	"testing"
	"time"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func TestAuthMissingCredentials(t *testing.T) {
	t.Parallel()

	switch os.Getenv("ALIGN_AUTH_TEST_CASE") {
	case "id":
		clientID, clientSecret = "", "example"
	case "secret":
		clientID, clientSecret = "example", ""
	default:
		for _, name := range []string{"id", "secret"} {
			t.Run(name, func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
				defer cancel()
				process := exec.CommandContext(ctx, os.Args[0], "-test.run=^TestAuthMissingCredentials$")
				process.Env = append(os.Environ(), "ALIGN_AUTH_TEST_CASE="+name, "PATH="+t.TempDir())
				output, err := process.CombinedOutput()
				assert.NoError(t, err, "%s", output)
			})
		}
		return
	}

	command := &cobra.Command{}
	err := authFunc(command, nil)
	assert.ErrorContains(t, err, "client id and secret must be baked into the binary")
	assert.True(t, command.SilenceUsage)
}

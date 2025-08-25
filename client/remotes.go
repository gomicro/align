package client

import (
	"bytes"
	"context"
	"os/exec"
)

func (c *Client) Remotes(ctx context.Context, dirs []string, args ...string) error {
	args = append([]string{"remote"}, args...)

	for _, dir := range dirs {
		out := &bytes.Buffer{}
		errout := &bytes.Buffer{}

		cmd := exec.CommandContext(ctx, "git", args...)
		cmd.Stdout = out
		cmd.Stderr = errout
		cmd.Dir = dir

		c.scrb.BeginDescribe(dir)

		err := cmd.Run()
		if err != nil {
			c.scrb.Error(err)
			c.scrb.PrintLines(errout)
		} else {
			c.scrb.PrintLines(out)
		}

		c.scrb.EndDescribe()
	}

	return nil
}

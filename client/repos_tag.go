package client

import (
	"bytes"
	"context"
	"os/exec"
	"strings"
)

func (c *Client) ListTags(ctx context.Context, dirs []string, args ...string) error {
	args = append([]string{"tag"}, args...)

	verbose := Verbose(ctx)
	if verbose {
		c.scrb.BeginDescribe("Running with command:")
		c.scrb.Print(strings.Join(append([]string{"git"}, args...), " "))
		c.scrb.EndDescribe()
	}

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

func (c *Client) TagRepos(ctx context.Context, dirs []string, args ...string) error {
	return nil
}

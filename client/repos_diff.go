package client

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strings"
)

func (c *Client) DiffRepos(ctx context.Context, dirs []string, args ...string) error {
	args = append([]string{"diff"}, args...)

	c.scrb.BeginDescribe("Command")
	defer c.scrb.EndDescribe()

	c.scrb.Print(fmt.Sprintf("git %s", strings.Join(args, " ")))

	c.scrb.BeginDescribe("directories")
	defer c.scrb.EndDescribe()

	for _, dir := range dirs {
		out := &bytes.Buffer{}
		errout := &bytes.Buffer{}

		cmd := exec.CommandContext(ctx, "git", args...)
		cmd.Stdout = out
		cmd.Stderr = errout
		cmd.Dir = dir

		err := cmd.Run()

		c.scrb.BeginDescribe(dir)
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

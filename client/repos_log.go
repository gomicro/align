package client

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strings"
)

func (c *Client) LogRepos(ctx context.Context, dirs []string, ignoreEmpty bool, args ...string) error {
	args = append([]string{"log"}, args...)

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

		if ignoreEmpty && out.Len() == 0 && err == nil {
			continue
		}

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

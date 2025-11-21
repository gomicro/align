package client

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strings"
)

type DiffConfig struct {
	IgnoreEmpty      bool
	Args             []string
}

func (c *Client) DiffRepos(ctx context.Context, dirs []string, cfg *DiffConfig) error {
	args := append([]string{"diff"}, cfg.Args...)

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

		if cfg.IgnoreEmpty && out.Len() == 0 && err == nil {
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

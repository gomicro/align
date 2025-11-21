package client

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strings"
)

type DiffConfig struct {
	IgnoreEmpty      bool
	IgnoreFilePrefix []string
	MatchExtension   []string
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

		out = matchExtensions(out, cfg.MatchExtension)

		// filter first to have empty check accurate
		out = filterLines(out, cfg.IgnoreFilePrefix)

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

func filterLines(buf *bytes.Buffer, prefixes []string) *bytes.Buffer {
	if len(prefixes) == 0 {
		return buf
	}

	scanner := bufio.NewScanner(buf)
	out := &bytes.Buffer{}

	for scanner.Scan() {
		line := scanner.Text()

		ignore := false
		for _, prefix := range prefixes {
			if strings.HasPrefix(line, prefix) {
				ignore = true
				break
			}
		}

		if !ignore {
			out.WriteString(line + "\n")
		}
	}

	return out
}

func matchExtensions(buf *bytes.Buffer, extensions []string) *bytes.Buffer {
	if len(extensions) == 0 {
		return buf
	}

	for i := range extensions {
		if !strings.HasPrefix(extensions[i], ".") {
			extensions[i] = "." + extensions[i]
		}
	}

	scanner := bufio.NewScanner(buf)
	out := &bytes.Buffer{}

	for scanner.Scan() {
		line := scanner.Text()

		matched := false
		for _, ext := range extensions {
			if strings.HasSuffix(line, ext) {
				matched = true
				break
			}
		}

		if matched {
			out.WriteString(line + "\n")
		}
	}

	return out
}

package client

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os/exec"
	"strings"

	"github.com/gosuri/uiprogress"
)

var (
	ErrUnstagedChanges = errors.New("unstanged changes")
)

func (c *Client) CheckoutRepos(ctx context.Context, dirs []string, args ...string) error {
	count := len(dirs)
	args = append([]string{"checkout"}, args...)

	verbose := Verbose(ctx)

	var bar *uiprogress.Bar
	currRepo := ""

	if verbose {
		c.scrb.BeginDescribe("Command")
		defer c.scrb.EndDescribe()

		c.scrb.Print(fmt.Sprintf("git %s", strings.Join(args, " ")))

		c.scrb.BeginDescribe("directories")
		defer c.scrb.EndDescribe()
	} else {
		bar = uiprogress.AddBar(count).
			AppendCompleted().
			PrependElapsed().
			PrependFunc(func(b *uiprogress.Bar) string {
				return fmt.Sprintf("Checkingout (%d/%d)", b.Current(), count)
			}).
			AppendFunc(func(b *uiprogress.Bar) string {
				return currRepo
			})
	}

	unstagedRepos := []string{}
	for _, dir := range dirs {
		currRepo = fmt.Sprintf("\nCurrent Repo: %v", dir)

		out := &bytes.Buffer{}
		errout := &bytes.Buffer{}

		cmd := exec.CommandContext(ctx, "git", args...)
		cmd.Stdout = out
		cmd.Stderr = errout
		cmd.Dir = dir

		err := cmd.Run()
		if err != nil && !verbose {
			if errors.Is(err, ErrUnstagedChanges) {
				unstagedRepos = append(unstagedRepos, dir)
				continue
			}

			return fmt.Errorf("checkout repo: %w", err) //TODO: collect errors and return them all
		}

		if verbose {
			c.scrb.BeginDescribe(dir)
			if err != nil {
				c.scrb.Error(err)
				c.scrb.PrintLines(errout)
			} else {
				c.scrb.PrintLines(out)
			}

			c.scrb.EndDescribe()
		} else {
			bar.Incr()
		}
	}

	currRepo = ""

	if len(unstagedRepos) > 0 {
		return fmt.Errorf("unstaged repos needing manual attention: [%s]", strings.Join(unstagedRepos, ", "))
	}

	return nil
}

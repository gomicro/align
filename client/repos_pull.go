package client

import (
	"context"
	"errors"
	"fmt"

	"github.com/go-git/go-git/v5"
	"github.com/gosuri/uiprogress"
)

func (c *Client) PullRepos(ctx context.Context, dirs []string) error {
	count := len(dirs)

	currRepo := ""
	bar := uiprogress.AddBar(count).
		AppendCompleted().
		PrependElapsed().
		PrependFunc(func(b *uiprogress.Bar) string {
			return fmt.Sprintf("Pulling (%d/%d)", b.Current(), count)
		}).
		AppendFunc(func(b *uiprogress.Bar) string {
			return currRepo
		})

	for i := range dirs {
		currRepo = fmt.Sprintf("\nCurrent Repo: %v", dirs[i])
		err := c.PullRepo(ctx, dirs[i])
		if err != nil {
			return fmt.Errorf("pull repo: %w", err)
		}
		bar.Incr()
	}

	currRepo = ""

	return nil
}

func (c *Client) PullRepo(ctx context.Context, dir string) error {
	r, err := git.PlainOpen(dir)
	if err != nil {
		return fmt.Errorf("open dir: %w", err)
	}

	w, err := r.Worktree()
	if err != nil {
		return fmt.Errorf("worktree: %w", err)
	}

	opts := &git.PullOptions{
		RemoteName: "origin",
		Auth:       c.ghSSHAuth,
	}

	err = w.PullContext(ctx, opts)
	if err != nil && !errors.Is(err, git.NoErrAlreadyUpToDate) {
		return fmt.Errorf("pull: %w", err)
	}

	return nil
}

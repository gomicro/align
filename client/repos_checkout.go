package client

import (
	"context"
	"fmt"

	"github.com/go-git/go-git/v5"
	"github.com/gosuri/uiprogress"
)

func (c *Client) CheckoutRepos(ctx context.Context, dirs []string, branch string) error {
	count := len(dirs)

	currRepo := ""
	bar := uiprogress.AddBar(count).
		AppendCompleted().
		PrependElapsed().
		PrependFunc(func(b *uiprogress.Bar) string {
			return fmt.Sprintf("Checkingout (%d/%d)", b.Current(), count)
		}).
		AppendFunc(func(b *uiprogress.Bar) string {
			return currRepo
		})

	for i := range dirs {
		currRepo = fmt.Sprintf("\nCurrent Repo: %v", dirs[i])
		err := c.CheckoutRepo(ctx, dirs[i], branch)
		if err != nil {
			return fmt.Errorf("checkout repo: %w", err)
		}
		bar.Incr()
	}

	currRepo = ""

	return nil
}

func (c *Client) CheckoutRepo(ctx context.Context, dir, branch string) error {
	r, err := git.PlainOpen(dir)
	if err != nil {
		return fmt.Errorf("open dir: %w", err)
	}

	w, err := r.Worktree()
	if err != nil {
		return fmt.Errorf("worktree: %w", err)
	}

	opts := &git.CheckoutOptions{}

	err = w.Checkout(opts)
	if err != nil {
		return fmt.Errorf("checkout: %w", err)
	}

	return nil
}

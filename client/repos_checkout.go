package client

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
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

	unstagedRepos := []string{}
	for i := range dirs {
		currRepo = fmt.Sprintf("\nCurrent Repo: %v", dirs[i])
		err := c.CheckoutRepo(ctx, dirs[i], branch)
		if err != nil {
			if errors.Is(err, git.ErrUnstagedChanges) {
				unstagedRepos = append(unstagedRepos, dirs[i])
				continue
			}

			if errors.Is(err, plumbing.ErrReferenceNotFound) {
				continue
			}

			return fmt.Errorf("checkout repo: %w", err)
		}
		bar.Incr()
	}

	currRepo = ""

	if len(unstagedRepos) > 0 {
		return fmt.Errorf("unstaged repos needing manual attention: [%s]", strings.Join(unstagedRepos, ", "))
	}

	return nil
}

func (c *Client) CheckoutRepo(ctx context.Context, dir, branch string) error {
	r, err := git.PlainOpen(dir)
	if err != nil {
		return fmt.Errorf("open dir: %w: %s", err, dir)
	}

	w, err := r.Worktree()
	if err != nil {
		return fmt.Errorf("worktree: %w", err)
	}

	opts := &git.CheckoutOptions{
		Branch: plumbing.NewBranchReferenceName(branch),
	}

	err = w.Checkout(opts)
	if err != nil {
		return fmt.Errorf("checkout: %w: %s", err, dir)
	}

	return nil
}

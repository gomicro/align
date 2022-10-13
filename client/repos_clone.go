package client

import (
	"context"
	"fmt"
	"path"

	"github.com/gomicro/align/git"
	"github.com/gosuri/uiprogress"
)

func (c *Client) CloneRepos(ctx context.Context) ([]*repository, error) {
	dirRepos, err := RepoMap(ctx)
	if err != nil {
		return nil, fmt.Errorf("repomap context: %w", err)
	}

	dirRepos, err = removeExcludes(ctx, dirRepos)
	if err != nil {
		return nil, fmt.Errorf("remove excludes: %w", err)
	}

	count := 0
	for rs := range dirRepos {
		count += len(rs)
	}

	currRepo := ""
	bar := uiprogress.AddBar(count).
		AppendCompleted().
		PrependElapsed().
		PrependFunc(func(b *uiprogress.Bar) string {
			return fmt.Sprintf("Cloning (%d/%d)", b.Current(), count)
		}).
		AppendFunc(func(b *uiprogress.Bar) string {
			return currRepo
		})

	cloned := []*repository{}
	var errs error
	for dir, rs := range dirRepos {
		for i := range rs {
			currRepo = fmt.Sprintf("\nCurrent Repo: %v/%v", dir, rs[i].name)

			opts := &git.CloneOptions{
				URL:         rs[i].url,
				Destination: path.Join(".", dir, rs[i].name),
			}

			err := c.gitClient.Clone(ctx, opts)
			if err != nil {
				errs = fmt.Errorf("%w; ", fmt.Errorf("clone repo: %w", err))
			}

			cloned = append(cloned, rs[i])
			bar.Incr()
		}
	}

	currRepo = ""

	if errs != nil {
		return nil, errs
	}

	return cloned, nil
}

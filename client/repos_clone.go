package client

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"path"

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
	for _, rs := range dirRepos {
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

			dest := path.Join(".", dir, rs[i].name)
			cmd := exec.Command("git", "clone", rs[i].url, dest)

			buf := bytes.Buffer{}
			cmd.Stdout = &buf

			err := cmd.Run()
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

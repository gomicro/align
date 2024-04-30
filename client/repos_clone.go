package client

import (
	"bytes"
	"context"
	"fmt"
	"os"
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

			exists, err := dirExists(dest)
			if err != nil {
				errs = fmt.Errorf("repo dir: %w; ", err)
				continue
			}

			if exists {
				exists, err = dirExists(path.Join(dest, ".git"))
				if err != nil {
					errs = fmt.Errorf("repo git dir: %w; ", err)
					continue
				}
			}

			if !exists {
				cmd := exec.CommandContext(ctx, "git", "clone", rs[i].url, dest)

				buf := bytes.Buffer{}
				cmd.Stdout = &buf

				err := cmd.Run()
				if err != nil {
					errs = fmt.Errorf("%w; ", fmt.Errorf("clone repo: %w", err))
				}

				cloned = append(cloned, rs[i])
			}

			bar.Incr()
		}
	}

	currRepo = ""

	if errs != nil {
		return nil, errs
	}

	return cloned, nil
}

func dirExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}

	if os.IsNotExist(err) {
		return false, nil
	}

	return false, fmt.Errorf("exists check: %w", err)
}

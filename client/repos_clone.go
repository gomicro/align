package client

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/go-git/go-git/v5"
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
			err := c.CloneRepo(ctx, dir, rs[i].name, rs[i].url, false)
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

// CloneRepo takes a context, base directory to clone individual repos into, the
// name to call the repo, the url to clone the repo from, and a boolean to show
// the output. It attempts to clone the repo into the directory structure of
// "baseDir/name". If the repo already exists it will skip it, and otherwise
// returns any errors it encounters.
func (c *Client) CloneRepo(ctx context.Context, baseDir, name, url string, show bool) error {
	dir := path.Join(".", baseDir, name)

	opts := &git.CloneOptions{
		URL:  url,
		Auth: c.ghHTTPSAuth,
	}

	if show {
		opts.Progress = os.Stdout
	}

	if strings.HasPrefix(url, "git@") {
		opts.Auth = c.ghSSHAuth
	}

	_, err := git.PlainCloneContext(ctx, dir, false, opts)
	if err != nil && !errors.Is(err, git.ErrRepositoryAlreadyExists) {
		return fmt.Errorf("plain clone: %w", err)
	}

	return nil
}

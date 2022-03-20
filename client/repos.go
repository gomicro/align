package client

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"path"

	"github.com/go-git/go-git/v5"
	"github.com/google/go-github/github"
	"github.com/gosuri/uiprogress"
)

func (c *Client) GetRepos(ctx context.Context, name string) ([]*github.Repository, error) {
	count := 0
	orgFound := true

	c.rate.Wait(ctx) //nolint: errcheck
	org, resp, err := c.ghClient.Organizations.Get(ctx, name)
	if resp == nil && err != nil {
		if _, ok := err.(*github.RateLimitError); ok {
			return nil, fmt.Errorf("github: hit rate limit")
		}

		return nil, fmt.Errorf("get org: %w", err)
	}

	if resp.StatusCode == http.StatusNotFound {
		orgFound = false

		c.rate.Wait(ctx) //nolint: errcheck
		user, _, err := c.ghClient.Users.Get(ctx, name)
		if err != nil {
			if _, ok := err.(*github.RateLimitError); ok {
				return nil, fmt.Errorf("github: hit rate limit")
			}

			return nil, fmt.Errorf("get user: %w", err)
		}

		count = user.GetPublicRepos() + user.GetTotalPrivateRepos()
	} else {
		count = org.GetPublicRepos() + org.GetTotalPrivateRepos()
	}

	if count < 1 {
		return nil, fmt.Errorf("no repos found")
	}

	repoBar := uiprogress.AddBar(count).
		AppendCompleted().
		PrependElapsed().
		PrependFunc(func(b *uiprogress.Bar) string {
			return fmt.Sprintf("Fetching (%d/%d)", b.Current(), count)
		})

	orgOpts := &github.RepositoryListByOrgOptions{
		Type: "all",
		ListOptions: github.ListOptions{
			Page:    0,
			PerPage: 100,
		},
	}

	userOpts := &github.RepositoryListOptions{
		Type: "all",
		ListOptions: github.ListOptions{
			Page:    0,
			PerPage: 100,
		},
	}

	var repos []*github.Repository
	for {
		var rs []*github.Repository
		c.rate.Wait(ctx) //nolint: errcheck
		if orgFound {
			rs, resp, err = c.ghClient.Repositories.ListByOrg(ctx, name, orgOpts)
		} else {
			rs, resp, err = c.ghClient.Repositories.List(ctx, name, userOpts)
		}

		if err != nil {
			if _, ok := err.(*github.RateLimitError); ok {
				return nil, fmt.Errorf("github: hit rate limit")
			}

			return nil, fmt.Errorf("list repos: %w", err)
		}

		for i := range rs {
			repoBar.Incr()

			if rs[i].GetArchived() {
				continue
			}

			repos = append(repos, rs[i])
		}

		if resp.NextPage == 0 {
			break
		}

		if orgFound {
			orgOpts.Page = resp.NextPage
		} else {
			userOpts.Page = resp.NextPage
		}
	}

	return repos, nil
}

func (c *Client) CloneRepo(ctx context.Context, name, baseDir, sshUrl string) error {
	dir := path.Join(".", baseDir, name)
	_, err := git.PlainClone(dir, false, &git.CloneOptions{
		Auth:     c.publicKeys,
		URL:      sshUrl,
		Progress: os.Stdout,
	})

	return err
}

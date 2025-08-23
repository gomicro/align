package client

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/go-github/github"
)

type reposContext int

var (
	reposContextKey    reposContext = 0
	excludesContextKey reposContext = 1
	verboseContextKey  reposContext = 2

	ErrReposNotFoundInContext = errors.New("repos map not found in context")
)

func WithRepos(ctx context.Context, repos []*github.Repository) context.Context {
	repoMap := parseDirRepoMap(repos)

	return context.WithValue(ctx, reposContextKey, repoMap)
}

func RepoMap(ctx context.Context) (map[string][]*Repository, error) {
	v := ctx.Value(reposContextKey)
	repoMap, ok := v.(map[string][]*Repository)
	if !ok {
		return nil, ErrReposNotFoundInContext
	}

	return repoMap, nil
}

func WithExcludes(ctx context.Context, repos []*Repository) context.Context {
	return context.WithValue(ctx, excludesContextKey, repos)
}

func Excludes(ctx context.Context) ([]*Repository, error) {
	v := ctx.Value(excludesContextKey)
	excludes, ok := v.([]*Repository)
	if !ok {
		return nil, nil
	}

	return excludes, nil
}

type Repository struct {
	name string
	url  string
}

func parseDirRepoMap(repos []*github.Repository) map[string][]*Repository {
	var dirRepo = map[string][]*Repository{}
	for _, repo := range repos {
		parts := strings.Split(*repo.SSHURL, "/")

		dir := strings.Split(parts[0], ":")[1]
		name := strings.TrimSuffix(parts[1], ".git")

		r := &Repository{
			name: name,
			url:  *repo.SSHURL,
		}

		dirRepo[dir] = append(dirRepo[dir], r)
	}

	return dirRepo
}

func removeExcludes(ctx context.Context, repoMap map[string][]*Repository) (map[string][]*Repository, error) {
	newMap := map[string][]*Repository{}

	excludes, err := Excludes(ctx)
	if err != nil {
		return nil, fmt.Errorf("excludes context: %w", err)
	}

	for dir, rs := range repoMap {
		for i := range rs {
			keep := true
			for j := range excludes {
				if rs[i].url == excludes[j].url {
					keep = false
					break
				}
			}

			if keep {
				newMap[dir] = append(newMap[dir], rs[i])
			}
		}
	}

	return newMap, nil
}

func WithVerbose(ctx context.Context, verbose bool) context.Context {
	return context.WithValue(ctx, verboseContextKey, verbose)
}

func Verbose(ctx context.Context) bool {
	v := ctx.Value(verboseContextKey)
	verbose, ok := v.(bool)
	if !ok {
		return false
	}

	return verbose
}

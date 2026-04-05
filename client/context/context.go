package context

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/go-github/github"
)

type reposContext int
type verboseContextKey struct{}

var (
	reposContextKey    reposContext = 0
	excludesContextKey reposContext = 1
	verboseKey                      = verboseContextKey{}

	ErrReposNotFoundInContext = errors.New("repos map not found in context")
)

func WithVerbose(ctx context.Context, verbose bool) context.Context {
	return context.WithValue(ctx, verboseKey, verbose)
}

func Verbose(ctx context.Context) bool {
	v := ctx.Value(verboseKey)
	verbose, ok := v.(bool)
	if !ok {
		return false
	}
	return verbose
}

type Repository struct {
	Name string
	URL  string
}

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

func parseDirRepoMap(repos []*github.Repository) map[string][]*Repository {
	dirRepo := map[string][]*Repository{}
	for _, repo := range repos {
		parts := strings.Split(*repo.SSHURL, "/")

		dir := strings.Split(parts[0], ":")[1]
		name := strings.TrimSuffix(parts[1], ".git")

		r := &Repository{
			Name: name,
			URL:  *repo.SSHURL,
		}

		dirRepo[dir] = append(dirRepo[dir], r)
	}
	return dirRepo
}

func RemoveExcludes(ctx context.Context, repoMap map[string][]*Repository) (map[string][]*Repository, error) {
	newMap := map[string][]*Repository{}

	excludes, err := Excludes(ctx)
	if err != nil {
		return nil, fmt.Errorf("excludes context: %w", err)
	}

	for dir, rs := range repoMap {
		for i := range rs {
			keep := true
			for j := range excludes {
				if rs[i].URL == excludes[j].URL {
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

package context

import (
	"context"
	"slices"
)

type ignoredReposKey struct{}

// WithIgnoredRepos stores exact local directory names to exclude from discovery.
func WithIgnoredRepos(ctx context.Context, names []string) context.Context {
	return context.WithValue(ctx, ignoredReposKey{}, slices.Clone(names))
}

// IgnoredRepos returns the local directory exclusions, or nil when unset.
func IgnoredRepos(ctx context.Context) []string {
	names, ok := ctx.Value(ignoredReposKey{}).([]string)
	if !ok {
		return nil
	}

	return slices.Clone(names)
}

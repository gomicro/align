package client

import (
	"context"

	"github.com/google/go-github/github"
)

type Clienter interface {
	CheckoutRepos(ctx context.Context, repoDirs []string, args ...string) error
	CloneRepos(ctx context.Context) ([]*Repository, error)
	GetDirs(ctx context.Context, dir string) ([]string, error)
	GetLogins(ctx context.Context) ([]string, error)
	GetRepos(ctx context.Context, name string) ([]*github.Repository, error)
	PullRepos(ctx context.Context, repoDirs []string, args ...string) error
	PushRepos(ctx context.Context, repoDirs []string, args ...string) error
	Remotes(ctx context.Context, repoDirs []string, args ...string) error
	SetURLs(ctx context.Context, repoDirs []string, name, baseURL string) error
	Add(ctx context.Context, dirs []string, name, baseURL string) error
	Remove(ctx context.Context, dirs []string, name string) error
}

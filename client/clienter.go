package client

import (
	"context"

	clientctx "github.com/gomicro/align/client/context"
	"github.com/gomicro/align/client/repos"
	"github.com/google/go-github/github"
)

type Clienter interface {
	Add(ctx context.Context, dirs []string, name, baseURL string) error
	Branches(ctx context.Context, repoDirs []string, args ...string) error
	CheckoutRepos(ctx context.Context, repoDirs []string, args ...string) error
	CloneRepos(ctx context.Context, dir string) ([]*clientctx.Repository, error)
	CommitRepos(ctx context.Context, dirs []string, args ...string) error
	DiffRepos(ctx context.Context, repoDirs []string, cfg *repos.DiffConfig) error
	FetchRepos(ctx context.Context, repoDirs []string, args ...string) error
	GetBranchAndTagNames(ctx context.Context, dirs []string) ([]string, error)
	GetBranchNames(ctx context.Context, dirs []string) ([]string, error)
	GetDirs(ctx context.Context, dir string) ([]string, error)
	GetLogins(ctx context.Context) ([]string, error)
	GetRemoteNames(ctx context.Context, dirs []string) ([]string, error)
	GetRepos(ctx context.Context, name string) ([]*github.Repository, error)
	GetTagNames(ctx context.Context, dirs []string) ([]string, error)
	ListTags(ctx context.Context, repoDirs []string, args ...string) error
	LogRepos(ctx context.Context, repoDirs []string, ignoreEmtpy bool, args ...string) error
	MergeRepos(ctx context.Context, repoDirs []string, args ...string) error
	PullRepos(ctx context.Context, repoDirs []string, args ...string) error
	PushRepos(ctx context.Context, repoDirs []string, args ...string) error
	Remotes(ctx context.Context, repoDirs []string, args ...string) error
	Remove(ctx context.Context, dirs []string, name string) error
	ResetRepos(ctx context.Context, repoDirs []string, args ...string) error
	SetURLs(ctx context.Context, repoDirs []string, name, baseURL string) error
	StageFiles(ctx context.Context, dirs []string, args ...string) error
	StashRepos(ctx context.Context, dirs []string, args ...string) error
	StatusRepos(ctx context.Context, dirs []string, ignoreEmpty bool, args ...string) error
	TagRepos(ctx context.Context, repoDirs []string, args ...string) error
}

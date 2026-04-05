package testclient

import (
	"context"

	clientctx "github.com/gomicro/align/client/context"
	"github.com/gomicro/align/client/repos"
	"github.com/google/go-github/github"
)

type TestClient struct {
	CommandsCalled []string
	Errors         map[string]error
}

func New() *TestClient {
	return &TestClient{
		Errors: map[string]error{},
	}
}

func (c *TestClient) CheckoutRepos(ctx context.Context, repoDirs []string, args ...string) error {
	c.CommandsCalled = append(c.CommandsCalled, "CheckoutRepos")

	return c.Errors["CheckoutRepos"]
}

func (c *TestClient) CommitRepos(ctx context.Context, dirs []string, args ...string) error {
	c.CommandsCalled = append(c.CommandsCalled, "CommitRepos")

	return c.Errors["CommitRepos"]
}

func (c *TestClient) CloneRepos(ctx context.Context, baseDir string) ([]*clientctx.Repository, error) {
	c.CommandsCalled = append(c.CommandsCalled, "CloneRepos")

	return nil, c.Errors["CloneRepos"]
}

func (c *TestClient) GetBranchNames(ctx context.Context, dirs []string) ([]string, error) {
	c.CommandsCalled = append(c.CommandsCalled, "GetBranchNames")

	return nil, c.Errors["GetBranchNames"]
}

func (c *TestClient) GetBranchAndTagNames(ctx context.Context, dirs []string) ([]string, error) {
	c.CommandsCalled = append(c.CommandsCalled, "GetBranchAndTagNames")

	return nil, c.Errors["GetBranchAndTagNames"]
}

func (c *TestClient) GetDirs(ctx context.Context, dir string) ([]string, error) {
	c.CommandsCalled = append(c.CommandsCalled, "GetDirs")

	return nil, c.Errors["GetDirs"]
}

func (c *TestClient) GetLogins(ctx context.Context) ([]string, error) {
	c.CommandsCalled = append(c.CommandsCalled, "GetLogins")

	return nil, c.Errors["GetLogins"]
}

func (c *TestClient) GetRemoteNames(ctx context.Context, dirs []string) ([]string, error) {
	c.CommandsCalled = append(c.CommandsCalled, "GetRemoteNames")

	return nil, c.Errors["GetRemoteNames"]
}

func (c *TestClient) GetTagNames(ctx context.Context, dirs []string) ([]string, error) {
	c.CommandsCalled = append(c.CommandsCalled, "GetTagNames")

	return nil, c.Errors["GetTagNames"]
}

func (c *TestClient) GetRepos(ctx context.Context, name string) ([]*github.Repository, error) {
	c.CommandsCalled = append(c.CommandsCalled, "GetRepos")

	return nil, c.Errors["GetRepos"]
}

func (c *TestClient) Branches(ctx context.Context, repoDirs []string, args ...string) error {
	c.CommandsCalled = append(c.CommandsCalled, "Branches")

	return c.Errors["Branches"]
}

func (c *TestClient) ListTags(ctx context.Context, repoDirs []string, args ...string) error {
	c.CommandsCalled = append(c.CommandsCalled, "ListTags")

	return c.Errors["ListTags"]
}

func (c *TestClient) PullRepos(ctx context.Context, repoDirs []string, args ...string) error {
	c.CommandsCalled = append(c.CommandsCalled, "PullRepos")

	return c.Errors["PullRepos"]
}

func (c *TestClient) PushRepos(ctx context.Context, repoDirs []string, args ...string) error {
	c.CommandsCalled = append(c.CommandsCalled, "PushRepos")

	return c.Errors["PushRepos"]
}

func (c *TestClient) Remotes(ctx context.Context, repoDirs []string, args ...string) error {
	c.CommandsCalled = append(c.CommandsCalled, "Remotes")

	return c.Errors["Remotes"]
}

func (c *TestClient) SetURLs(ctx context.Context, repoDirs []string, name, baseURL string) error {
	c.CommandsCalled = append(c.CommandsCalled, "SetURLs")

	return c.Errors["SetURLs"]
}

func (c *TestClient) Add(ctx context.Context, dirs []string, name, baseURL string) error {
	c.CommandsCalled = append(c.CommandsCalled, "Add")

	return c.Errors["Add"]
}

func (c *TestClient) Remove(ctx context.Context, dirs []string, name string) error {
	c.CommandsCalled = append(c.CommandsCalled, "Remove")

	return c.Errors["Remove"]
}

func (c *TestClient) StatusRepos(ctx context.Context, dirs []string, ignoreEmpty bool, args ...string) error {
	c.CommandsCalled = append(c.CommandsCalled, "StatusRepos")

	return c.Errors["StatusRepos"]
}

func (c *TestClient) StageFiles(ctx context.Context, dirs []string, args ...string) error {
	c.CommandsCalled = append(c.CommandsCalled, "StageFiles")

	return c.Errors["StageFiles"]
}

func (c *TestClient) TagRepos(ctx context.Context, repoDirs []string, args ...string) error {
	c.CommandsCalled = append(c.CommandsCalled, "TagRepos")

	return c.Errors["TagRepos"]
}

func (c *TestClient) DiffRepos(ctx context.Context, repoDirs []string, cfg *repos.DiffConfig) error {
	c.CommandsCalled = append(c.CommandsCalled, "DiffRepos")

	return c.Errors["DiffRepos"]
}

func (c *TestClient) LogRepos(ctx context.Context, repoDirs []string, ignoreEmtpy bool, args ...string) error {
	c.CommandsCalled = append(c.CommandsCalled, "LogRepos")

	return c.Errors["LogRepos"]
}

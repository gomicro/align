package testclient

import (
	"context"

	"github.com/gomicro/align/client"
	"github.com/google/go-github/github"
)

type TestClient struct {
	CommandsCalled []string
}

func New() *TestClient {
	return &TestClient{}
}

func (c *TestClient) CheckoutRepos(ctx context.Context, repoDirs []string, args ...string) error {
	c.CommandsCalled = append(c.CommandsCalled, "CheckoutRepos")

	return nil
}

func (c *TestClient) CloneRepos(ctx context.Context, baseDir string) ([]*client.Repository, error) {
	c.CommandsCalled = append(c.CommandsCalled, "CloneRepos")

	return nil, nil
}

func (c *TestClient) GetDirs(ctx context.Context, dir string) ([]string, error) {
	c.CommandsCalled = append(c.CommandsCalled, "GetDirs")

	return nil, nil
}

func (c *TestClient) GetLogins(ctx context.Context) ([]string, error) {
	c.CommandsCalled = append(c.CommandsCalled, "GetLogins")

	return nil, nil
}

func (c *TestClient) GetRepos(ctx context.Context, name string) ([]*github.Repository, error) {
	c.CommandsCalled = append(c.CommandsCalled, "GetRepos")

	return nil, nil
}

func (c *TestClient) Branches(ctx context.Context, repoDirs []string, args ...string) error {
	c.CommandsCalled = append(c.CommandsCalled, "Branches")

	return nil
}

func (c *TestClient) ListTags(ctx context.Context, repoDirs []string, args ...string) error {
	c.CommandsCalled = append(c.CommandsCalled, "ListTags")

	return nil
}

func (c *TestClient) PullRepos(ctx context.Context, repoDirs []string, args ...string) error {
	c.CommandsCalled = append(c.CommandsCalled, "PullRepos")

	return nil
}

func (c *TestClient) PushRepos(ctx context.Context, repoDirs []string, args ...string) error {
	c.CommandsCalled = append(c.CommandsCalled, "PushRepos")

	return nil
}

func (c *TestClient) Remotes(ctx context.Context, repoDirs []string, args ...string) error {
	c.CommandsCalled = append(c.CommandsCalled, "Remotes")

	return nil
}

func (c *TestClient) SetURLs(ctx context.Context, repoDirs []string, name, baseURL string) error {
	c.CommandsCalled = append(c.CommandsCalled, "SetURLs")

	return nil
}

func (c *TestClient) Add(ctx context.Context, dirs []string, name, baseURL string) error {
	c.CommandsCalled = append(c.CommandsCalled, "Add")

	return nil
}

func (c *TestClient) Remove(ctx context.Context, dirs []string, name string) error {
	c.CommandsCalled = append(c.CommandsCalled, "Remove")

	return nil
}

func (c *TestClient) TagRepos(ctx context.Context, repoDirs []string, args ...string) error {
	c.CommandsCalled = append(c.CommandsCalled, "TagRepos")

	return nil
}

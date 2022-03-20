package client

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/go-git/go-git/v5/plumbing/transport/ssh"
	"github.com/gomicro/align/config"
	"github.com/gomicro/trust"
	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
	"golang.org/x/time/rate"
)

type Client struct {
	cfg        *config.Config
	ghClient   *github.Client
	rate       *rate.Limiter
	publicKeys *ssh.PublicKeys
}

func New(cfg *config.Config) (*Client, error) {
	pool := trust.New()

	certs, err := pool.CACerts()
	if err != nil {
		return nil, fmt.Errorf("failed to create cert pool: %w", err)
	}

	httpClient := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{RootCAs: certs},
		},
	}

	ctx := context.Background()
	ctx = context.WithValue(ctx, oauth2.HTTPClient, httpClient)

	ts := oauth2.StaticTokenSource(
		&oauth2.Token{
			AccessToken: cfg.Github.Token,
		},
	)

	rl := rate.NewLimiter(
		rate.Limit(cfg.Github.Limits.RequestsPerSecond),
		cfg.Github.Limits.Burst,
	)

	_, err = os.Stat(cfg.Github.PrivateKeyFile)
	if err != nil {
		return nil, fmt.Errorf("private key file: %w", err)
	}

	publicKeys, err := ssh.NewPublicKeysFromFile("git", cfg.Github.PrivateKeyFile, "")
	if err != nil {
		return nil, fmt.Errorf("public keys: %w", err)
	}

	return &Client{
		cfg:        cfg,
		ghClient:   github.NewClient(oauth2.NewClient(ctx, ts)),
		rate:       rl,
		publicKeys: publicKeys,
	}, nil
}

func (c *Client) GetLogins(ctx context.Context) ([]string, error) {
	logins := []string{}

	user, _, err := c.ghClient.Users.Get(ctx, "")
	if err != nil {
		if _, ok := err.(*github.RateLimitError); ok {
			return nil, fmt.Errorf("github: hit rate limit")
		}

		return nil, fmt.Errorf("get user: %v", err.Error())
	}

	logins = append(logins, strings.ToLower(user.GetLogin()))

	opts := &github.ListOptions{
		Page:    0,
		PerPage: 100,
	}

	orgs, _, err := c.ghClient.Organizations.List(ctx, "", opts)
	if err != nil {
		if _, ok := err.(*github.RateLimitError); ok {
			return nil, fmt.Errorf("github: hit rate limit")
		}

		return nil, fmt.Errorf("list orgs: %v", err.Error())
	}

	for i := range orgs {
		o := orgs[i].GetLogin()
		logins = append(logins, strings.ToLower(o))
	}

	return logins, nil
}

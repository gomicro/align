// Package client wires together the repos and remotes layers and exposes a unified Client.
package client

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	sshgit "github.com/go-git/go-git/v5/plumbing/transport/ssh"
	"github.com/gomicro/align/client/remotes"
	"github.com/gomicro/align/client/repos"
	"github.com/gomicro/align/config"
	"github.com/gomicro/scribe"
	"github.com/gomicro/scribe/color"
	"github.com/gomicro/trust"
	"github.com/google/go-github/github"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/knownhosts"
	"golang.org/x/oauth2"
	"golang.org/x/time/rate"
)

// Client provides the full set of cross-repo git operations backed by the GitHub API.
type Client struct {
	*repos.Repos
	remoteMgr   *remotes.Remotes
	cfg         *config.Config
	ghClient    *github.Client
	ghSSHAuth   *sshgit.PublicKeys
	ghHTTPSAuth *sshgit.Password
}

// Option configures a Client.
type Option func(*clientOptions)

type clientOptions struct {
	noColor bool
}

// WithNoColor disables ANSI color codes in verbose output.
func WithNoColor() Option {
	return func(o *clientOptions) {
		o.noColor = true
	}
}

// New constructs a Client from cfg, initialising SSH auth, HTTPS auth, the GitHub API client, and rate limiter.
func New(cfg *config.Config, opts ...Option) (*Client, error) {
	co := &clientOptions{}
	for _, o := range opts {
		o(co)
	}

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

	var publicKeys *sshgit.PublicKeys
	if cfg.Github.PrivateKey != "" && cfg.Github.Username != "" {
		pem := []byte(cfg.Github.PrivateKey)

		publicKeys, err = sshgit.NewPublicKeys(cfg.Github.Username, pem, "")
		if err != nil {
			return nil, fmt.Errorf("public keys: %w", err)
		}

		publicKeys.HostKeyCallback, err = knownHostsCallback()
		if err != nil {
			return nil, fmt.Errorf("known hosts: %w", err)
		}
	} else if cfg.Github.PrivateKeyFile != "" {
		_, err := os.Stat(cfg.Github.PrivateKeyFile)
		if err != nil {
			return nil, fmt.Errorf("private key file: %w", err)
		}

		publicKeys, err = sshgit.NewPublicKeysFromFile("git", cfg.Github.PrivateKeyFile, "")
		if err != nil {
			return nil, fmt.Errorf("public keys file: %w", err)
		}

		publicKeys.HostKeyCallback, err = knownHostsCallback()
		if err != nil {
			return nil, fmt.Errorf("known hosts: %w", err)
		}
	}

	var pass *sshgit.Password
	if cfg.Github.Username != "" && cfg.Github.Token != "" {
		hostKeyCallback, err := knownHostsCallback()
		if err != nil {
			return nil, fmt.Errorf("known hosts: %w", err)
		}

		pass = &sshgit.Password{
			User:     cfg.Github.Username,
			Password: cfg.Github.Token,
			HostKeyCallbackHelper: sshgit.HostKeyCallbackHelper{
				HostKeyCallback: hostKeyCallback,
			},
		}
	}

	var t *scribe.Theme
	if co.noColor {
		t = scribe.DefaultTheme
	} else {
		t = &scribe.Theme{
			Describe: func(s string) string {
				return color.CyanFg(s)
			},
			Print: scribe.NoopDecorator,
			Error: func(err error) string {
				return fmt.Sprintf("%s %s\n", color.RedFg("Error:"), err)
			},
		}
	}

	scrb, err := scribe.NewScribe(os.Stdout, t)
	if err != nil {
		return nil, fmt.Errorf("scribe: %w", err)
	}

	ghClient := github.NewClient(oauth2.NewClient(ctx, ts))

	return &Client{
		Repos:       repos.New(scrb, ghClient, rl),
		remoteMgr:   remotes.New(scrb),
		cfg:         cfg,
		ghClient:    ghClient,
		ghSSHAuth:   publicKeys,
		ghHTTPSAuth: pass,
	}, nil
}

func knownHostsCallback() (ssh.HostKeyCallback, error) {
	usr, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("home dir: %w", err)
	}

	cb, err := knownhosts.New(filepath.Join(usr, ".ssh", "known_hosts"))
	if err != nil {
		return nil, fmt.Errorf("parse known_hosts: %w", err)
	}

	return cb, nil
}

// GetLogins returns the authenticated user's login and all org logins, lowercased.
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

// Remotes runs git remote across all dirs.
func (c *Client) Remotes(ctx context.Context, dirs []string, args ...string) error {
	return c.remoteMgr.Remotes(ctx, dirs, args...)
}

// Add runs git remote add across all dirs, building the remote URL from baseURL and each dir basename.
func (c *Client) Add(ctx context.Context, dirs []string, name, baseURL string) error {
	return c.remoteMgr.Add(ctx, dirs, name, baseURL)
}

// Remove runs git remote remove across all dirs.
func (c *Client) Remove(ctx context.Context, dirs []string, name string) error {
	return c.remoteMgr.Remove(ctx, dirs, name)
}

// Rename runs git remote rename across all dirs.
func (c *Client) Rename(ctx context.Context, dirs []string, oldName, newName string) error {
	return c.remoteMgr.Rename(ctx, dirs, oldName, newName)
}

// SetURLs runs git remote set-url across all dirs, building the URL from baseURL and each dir basename.
func (c *Client) SetURLs(ctx context.Context, dirs []string, name, baseURL string) error {
	return c.remoteMgr.SetURLs(ctx, dirs, name, baseURL)
}

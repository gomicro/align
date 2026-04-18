// Package repos implements git operations fanned out across multiple repository directories.
//
// The simple fan-out methods (CheckoutRepos, CommitRepos, FetchRepos, MergeRepos, PullRepos,
// PushRepos, ResetRepos, StageFiles, StashRepos) each prepend their git sub-command and
// delegate to fanOut. They carry no individual doc comments.
package repos

import (
	"github.com/gomicro/scribe"
	"github.com/google/go-github/github"
	"golang.org/x/time/rate"
)

// Repos holds the scribe, GitHub client, and rate limiter shared by all operations.
type Repos struct {
	scrb     scribe.Scriber
	ghClient *github.Client
	rate     *rate.Limiter
}

// New returns a Repos using the provided scribe, GitHub client, and rate limiter.
func New(scrb scribe.Scriber, ghClient *github.Client, rate *rate.Limiter) *Repos {
	return &Repos{
		scrb:     scrb,
		ghClient: ghClient,
		rate:     rate,
	}
}

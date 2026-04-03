package repos

import (
	"github.com/gomicro/scribe"
	"github.com/google/go-github/github"
	"golang.org/x/time/rate"
)

type Repos struct {
	scrb     scribe.Scriber
	ghClient *github.Client
	rate     *rate.Limiter
}

func New(scrb scribe.Scriber, ghClient *github.Client, rate *rate.Limiter) *Repos {
	return &Repos{
		scrb:     scrb,
		ghClient: ghClient,
		rate:     rate,
	}
}

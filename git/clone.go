package git

import (
	"context"
	"fmt"
)

const (
	CloneCmd = "clone"
)

type CloneOptions struct {
	URL         string
	Destination string
	//ShowOutput bool
}

func (g *Git) Clone(ctx context.Context, opts *CloneOptions) error {
	err := g.Run(CloneCmd, opts.URL, opts.Destination)
	if err != nil {
		return fmt.Errorf("git: clone: %w", err)
	}

	return nil
}

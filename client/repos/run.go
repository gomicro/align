package repos

import (
	"context"
	"fmt"

	ctxhelper "github.com/gomicro/align/client/context"
	"github.com/gomicro/align/client/fanout"
)

// Default limits for operations run through the worker pool when no override is set.
const (
	DefaultCheckoutConcurrency = 50
	DefaultPullConcurrency     = 25
	DefaultFetchConcurrency    = 25
	DefaultPushConcurrency     = 25
	DefaultCloneConcurrency    = 10
	DefaultLocalConcurrency    = 50
)

func (r *Repos) fanOut(ctx context.Context, dirs []string, label string, args []string, fallback int) error {
	limit, err := ctxhelper.Concurrency(ctx, fallback)
	if err != nil {
		return fmt.Errorf("resolving concurrency: %w", err)
	}

	// Per-job results are already printed and joined into the error.
	_, err = fanout.Run(ctx, r.scrb, label, fanout.Jobs(dirs, args), limit)
	return err
}

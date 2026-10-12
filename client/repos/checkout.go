package repos

import (
	"context"
	"fmt"

	ctxhelper "github.com/gomicro/align/client/context"
)

// DefaultCheckoutConcurrency bounds simultaneous checkouts when no override is set.
const DefaultCheckoutConcurrency = 50

func (r *Repos) CheckoutRepos(ctx context.Context, dirs []string, args ...string) error {
	limit, err := ctxhelper.Concurrency(ctx, DefaultCheckoutConcurrency)
	if err != nil {
		return fmt.Errorf("checkout concurrency: %w", err)
	}

	return r.fanOutConcurrent(ctx, dirs, "Checking Out", append([]string{"checkout"}, args...), limit)
}

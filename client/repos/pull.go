package repos

import (
	"context"
	"fmt"

	ctxhelper "github.com/gomicro/align/client/context"
)

// DefaultPullConcurrency bounds simultaneous pulls when no override is set.
const DefaultPullConcurrency = 5

func (r *Repos) PullRepos(ctx context.Context, dirs []string, args ...string) error {
	limit, err := ctxhelper.Concurrency(ctx, DefaultPullConcurrency)
	if err != nil {
		return fmt.Errorf("pull concurrency: %w", err)
	}

	return r.fanOutConcurrent(ctx, dirs, "Pulling", append([]string{"pull"}, args...), limit)
}

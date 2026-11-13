package repos

import (
	"context"

	"github.com/gomicro/align/client/fanout"
)

// RunAlias runs one repository at a time because aliases may be order-sensitive
// and their arguments leave no room for a concurrency override.
func (r *Repos) RunAlias(ctx context.Context, dirs []string, name string, args ...string) error {
	// Per-job results are already printed and joined into the error.
	_, err := fanout.Run(ctx, r.scrb, name, fanout.Jobs(dirs, append([]string{name}, args...)), 1)
	return err
}

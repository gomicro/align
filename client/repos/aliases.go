package repos

import "context"

// RunAlias delegates expansion to Git in each repository's effective configuration.
func (r *Repos) RunAlias(ctx context.Context, dirs []string, name string, args ...string) error {
	return r.fanOut(ctx, dirs, name, append([]string{name}, args...))
}

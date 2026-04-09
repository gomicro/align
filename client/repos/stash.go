package repos

import "context"

func (r *Repos) StashRepos(ctx context.Context, dirs []string, args ...string) error {
	return r.fanOut(ctx, dirs, "Stashing", append([]string{"stash"}, args...))
}

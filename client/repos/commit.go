package repos

import "context"

func (r *Repos) CommitRepos(ctx context.Context, dirs []string, args ...string) error {
	return r.fanOut(ctx, dirs, "Committing", append([]string{"commit"}, args...))
}

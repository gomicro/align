package repos

import "context"

func (r *Repos) MergeRepos(ctx context.Context, dirs []string, args ...string) error {
	return r.fanOut(ctx, dirs, "Merging", append([]string{"merge"}, args...))
}

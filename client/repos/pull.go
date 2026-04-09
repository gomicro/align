package repos

import "context"

func (r *Repos) PullRepos(ctx context.Context, dirs []string, args ...string) error {
	return r.fanOut(ctx, dirs, "Pulling", append([]string{"pull"}, args...))
}

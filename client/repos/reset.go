package repos

import "context"

func (r *Repos) ResetRepos(ctx context.Context, dirs []string, args ...string) error {
	return r.fanOut(ctx, dirs, "Resetting", append([]string{"reset"}, args...))
}

package repos

import "context"

func (r *Repos) PushRepos(ctx context.Context, dirs []string, args ...string) error {
	return r.fanOut(ctx, dirs, "Pushing", append([]string{"push"}, args...))
}

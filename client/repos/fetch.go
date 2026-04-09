package repos

import "context"

func (r *Repos) FetchRepos(ctx context.Context, dirs []string, args ...string) error {
	return r.fanOut(ctx, dirs, "Fetching", append([]string{"fetch"}, args...))
}

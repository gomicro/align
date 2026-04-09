package repos

import "context"

func (r *Repos) CheckoutRepos(ctx context.Context, dirs []string, args ...string) error {
	return r.fanOut(ctx, dirs, "Checking Out", append([]string{"checkout"}, args...))
}

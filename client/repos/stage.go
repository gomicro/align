package repos

import "context"

func (r *Repos) StageFiles(ctx context.Context, dirs []string, args ...string) error {
	return r.fanOut(ctx, dirs, "Staging", append([]string{"add"}, args...))
}

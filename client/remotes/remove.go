package remotes

import "context"

func (r *Remotes) Remove(ctx context.Context, dirs []string, name string) error {
	args := []string{"remote", "remove", name}
	perDirArgs := make([][]string, len(dirs))
	for i := range dirs {
		perDirArgs[i] = args
	}
	return r.fanOut(ctx, dirs, "Removing Remote", perDirArgs)
}

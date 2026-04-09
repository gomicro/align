package remotes

import "context"

func (r *Remotes) Rename(ctx context.Context, dirs []string, oldName, newName string) error {
	args := []string{"remote", "rename", oldName, newName}
	perDirArgs := make([][]string, len(dirs))
	for i := range dirs {
		perDirArgs[i] = args
	}
	return r.fanOut(ctx, dirs, "Renaming Remote", perDirArgs)
}

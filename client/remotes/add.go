package remotes

import "context"

func (r *Remotes) Add(ctx context.Context, dirs []string, name, baseURL string) error {
	perDirArgs := make([][]string, len(dirs))
	for i, dir := range dirs {
		perDirArgs[i] = []string{"remote", "add", name, buildURL(baseURL, dir)}
	}
	return r.fanOut(ctx, dirs, "Adding Remote", perDirArgs)
}

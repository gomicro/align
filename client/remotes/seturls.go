package remotes

import (
	"context"
	"fmt"
	"strings"
)

func (r *Remotes) SetURLs(ctx context.Context, dirs []string, name, baseURL string) error {
	perDirArgs := make([][]string, len(dirs))
	for i, dir := range dirs {
		perDirArgs[i] = []string{"remote", "set-url", name, buildURL(baseURL, dir)}
	}
	return r.fanOut(ctx, dirs, "Setting URL", perDirArgs)
}

func buildURL(baseURL, dir string) string {
	baseURL = strings.TrimSuffix(baseURL, "/")

	return fmt.Sprintf("%s/%s.git", baseURL, dir)
}

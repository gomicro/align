package client

import (
	"context"
	"fmt"
	"os"
	"path"
	"slices"

	ctxhelper "github.com/gomicro/align/client/context"
)

// GetDirs discovers local repositories and validates invocation-scoped exclusions.
func (c *Client) GetDirs(ctx context.Context, baseDir string) ([]string, error) {
	files, err := os.ReadDir(baseDir)
	if err != nil {
		return nil, fmt.Errorf("read base dir: %w", err)
	}

	dirs := make([]string, 0, len(files))
	for _, f := range files {
		if !f.IsDir() {
			continue
		}

		p := path.Join(baseDir, f.Name())

		_, err := os.Stat(path.Join(p, ".git"))
		if err != nil {
			continue
		}

		dirs = append(dirs, p)
	}

	ignored := ctxhelper.IgnoredRepos(ctx)
	if len(ignored) == 0 {
		return dirs, nil
	}

	available := make(map[string]bool, len(dirs))
	for _, dir := range dirs {
		available[path.Base(dir)] = true
	}

	excluded := make(map[string]bool, len(ignored))
	var missing []string
	for _, name := range ignored {
		if excluded[name] {
			continue
		}

		excluded[name] = true
		if !available[name] {
			missing = append(missing, name)
		}
	}

	if len(missing) > 0 {
		slices.Sort(missing)
		return nil, fmt.Errorf("get dirs: ignored repositories not found: %q", missing)
	}

	selected := make([]string, 0, len(dirs))
	for _, dir := range dirs {
		if !excluded[path.Base(dir)] {
			selected = append(selected, dir)
		}
	}

	if len(selected) == 0 {
		return nil, fmt.Errorf("get dirs: no repositories selected")
	}

	return selected, nil
}

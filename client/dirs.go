package client

import (
	"context"
	"fmt"
	"os"
	"path"
)

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

	return dirs, nil
}

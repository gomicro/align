package client

import (
	"context"
	"fmt"
	"io/ioutil"
	"path"
)

func (c *Client) GetDirs(ctx context.Context, baseDir string) ([]string, error) {
	files, err := ioutil.ReadDir(baseDir)
	if err != nil {
		return nil, fmt.Errorf("read base dir: %w", err)
	}

	dirs := make([]string, 0, len(files))
	for _, f := range files {
		if f.IsDir() {
			dirs = append(dirs, path.Join(baseDir, f.Name()))
		}
	}

	return dirs, nil
}

package client

import (
	"bytes"
	"context"
	"os/exec"
	"strings"
)

func (c *Client) GetRemoteNames(ctx context.Context, dirs []string) ([]string, error) {
	seen := map[string]struct{}{}
	names := []string{}

	for _, dir := range dirs {
		out := &bytes.Buffer{}

		cmd := exec.CommandContext(ctx, "git", "remote")
		cmd.Stdout = out
		cmd.Dir = dir

		if err := cmd.Run(); err != nil {
			continue
		}

		for _, name := range strings.Split(strings.TrimSpace(out.String()), "\n") {
			if name == "" {
				continue
			}

			if _, ok := seen[name]; !ok {
				seen[name] = struct{}{}
				names = append(names, name)
			}
		}
	}

	return names, nil
}

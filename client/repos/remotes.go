package repos

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"sort"
	"strings"
)

func (r *Repos) GetRemoteNames(ctx context.Context, dirs []string) ([]string, error) {
	seen := map[string]struct{}{}
	names := []string{}

	for _, dir := range dirs {
		out := &bytes.Buffer{}
		cmd := exec.CommandContext(ctx, "git", "remote")
		cmd.Stdout = out
		cmd.Dir = dir

		err := cmd.Run()
		if err != nil {
			return nil, fmt.Errorf("get remote names: %w", err)
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

	sort.Strings(names)
	return names, nil
}

package client

import (
	"context"
)

func (c *Client) GetBranchAndTagNames(ctx context.Context, dirs []string) ([]string, error) {
	branches, err := c.GetBranchNames(ctx, dirs)
	if err != nil {
		return nil, err
	}

	tags, err := c.GetTagNames(ctx, dirs)
	if err != nil {
		return nil, err
	}

	seen := map[string]struct{}{}
	names := make([]string, 0, len(branches)+len(tags))

	for _, name := range append(branches, tags...) {
		if _, ok := seen[name]; !ok {
			seen[name] = struct{}{}
			names = append(names, name)
		}
	}

	return names, nil
}

package client

import (
	"context"
)

func (c *Client) StashRepos(ctx context.Context) error {
	return nil
}

func (c *Client) StashRepo(ctx, baseDir, name string) error {
	return nil
}

func (c *Client) CheckoutMains(ctx context.Context) error {
	return nil
}

func (c *Client) CheckoutMain(ctx, baseDir, name string) error {
	return nil
}

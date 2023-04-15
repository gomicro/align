package client

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"

	"github.com/gosuri/uiprogress"
)

func (c *Client) PullRepos(ctx context.Context, dirs []string) error {
	count := len(dirs)

	currRepo := ""
	bar := uiprogress.AddBar(count).
		AppendCompleted().
		PrependElapsed().
		PrependFunc(func(b *uiprogress.Bar) string {
			return fmt.Sprintf("Pulling (%d/%d)", b.Current(), count)
		}).
		AppendFunc(func(b *uiprogress.Bar) string {
			return currRepo
		})

	for _, dir := range dirs {
		currRepo = fmt.Sprintf("\nCurrent Repo: %v", dir)

		cmd := exec.CommandContext(ctx, "git", "pull")

		buf := bytes.Buffer{}
		cmd.Stdout = &buf

		cmd.Dir = dir

		err := cmd.Run()
		if err != nil {
			return fmt.Errorf("pull repo: %w", err)
		}

		bar.Incr()
	}

	currRepo = ""

	return nil
}

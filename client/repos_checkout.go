package client

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os/exec"
	"strings"

	"github.com/gosuri/uiprogress"
)

var (
	ErrUnstagedChanges = errors.New("unstanged changes")
)

func (c *Client) CheckoutRepos(ctx context.Context, dirs []string, args []string) error {
	count := len(dirs)
	args = append([]string{"checkout"}, args...)

	currRepo := ""
	bar := uiprogress.AddBar(count).
		AppendCompleted().
		PrependElapsed().
		PrependFunc(func(b *uiprogress.Bar) string {
			return fmt.Sprintf("Checkingout (%d/%d)", b.Current(), count)
		}).
		AppendFunc(func(b *uiprogress.Bar) string {
			return currRepo
		})

	unstagedRepos := []string{}
	for _, dir := range dirs {
		currRepo = fmt.Sprintf("\nCurrent Repo: %v", dir)

		cmd := exec.CommandContext(ctx, "git", args...)

		buf := bytes.Buffer{}
		cmd.Stdout = &buf

		cmd.Dir = dir

		err := cmd.Run()
		if err != nil {
			if errors.Is(err, ErrUnstagedChanges) {
				unstagedRepos = append(unstagedRepos, dir)
				continue
			}

			return fmt.Errorf("checkout repo: %w", err)
		}

		bar.Incr()
	}

	currRepo = ""

	if len(unstagedRepos) > 0 {
		return fmt.Errorf("unstaged repos needing manual attention: [%s]", strings.Join(unstagedRepos, ", "))
	}

	return nil
}

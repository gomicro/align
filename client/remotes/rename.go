package remotes

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strings"

	ctxhelper "github.com/gomicro/align/client/context"
	"github.com/gosuri/uiprogress"
)

func (r *Remotes) Rename(ctx context.Context, dirs []string, oldName, newName string) error {
	count := len(dirs)

	args := []string{"remote", "rename", oldName, newName}

	verbose := ctxhelper.Verbose(ctx)

	var bar *uiprogress.Bar
	currRepo := ""

	if verbose {
		r.scrb.BeginDescribe("Command")
		defer r.scrb.EndDescribe()

		r.scrb.Print(fmt.Sprintf("git %s", strings.Join(args, " ")))

		r.scrb.BeginDescribe("directories")
		defer r.scrb.EndDescribe()
	} else {
		bar = uiprogress.AddBar(count).
			AppendCompleted().
			PrependElapsed().
			PrependFunc(func(b *uiprogress.Bar) string {
				return fmt.Sprintf("Renaming Remote (%d/%d)", b.Current(), count)
			}).
			AppendFunc(func(b *uiprogress.Bar) string {
				return currRepo
			})
	}

	for _, dir := range dirs {
		currRepo = fmt.Sprintf("\nCurrent Repo: %v", dir)

		out := &bytes.Buffer{}
		errout := &bytes.Buffer{}

		cmd := exec.CommandContext(ctx, "git", args...)
		cmd.Stdout = out
		cmd.Stderr = errout
		cmd.Dir = dir

		err := cmd.Run()
		if err != nil && !verbose {
			return fmt.Errorf("run: %w", err) // TODO: collect errors and return them all
		}

		if verbose {
			r.scrb.BeginDescribe(dir)
			if err != nil {
				r.scrb.Error(err)
				r.scrb.PrintLines(errout)
			} else {
				r.scrb.PrintLines(out)
			}

			r.scrb.EndDescribe()
		} else {
			bar.Incr()
		}
	}

	currRepo = ""

	return nil
}

package repos

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os/exec"
	"strings"
	"sync/atomic"
	"time"

	ctxhelper "github.com/gomicro/align/client/context"

	"github.com/gosuri/uiprogress"
)

func (r *Repos) fanOut(ctx context.Context, dirs []string, label string, args []string) error {
	count := len(dirs)
	verbose := ctxhelper.Verbose(ctx)

	var bar *uiprogress.Bar
	var currRepo atomic.Value
	currRepo.Store("")

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
				return fmt.Sprintf("%s (%d/%d)", label, b.Current(), count)
			}).
			AppendFunc(func(b *uiprogress.Bar) string {
				return currRepo.Load().(string)
			})
	}

	var errs []error

	for _, dir := range dirs {
		currRepo.Store(fmt.Sprintf("\nCurrent Repo: %v", dir))
		result := runGit(ctx, dir, args, false)
		if result.err != nil {
			errs = append(errs, result.failure())
		}

		if verbose {
			r.printResult(result)
		} else {
			bar.Incr()
		}
	}

	currRepo.Store("")

	return errors.Join(errs...)
}

type gitResult struct {
	dir    string
	stdout bytes.Buffer
	stderr bytes.Buffer
	err    error
}

func runGit(ctx context.Context, dir string, args []string, cancelTree bool) *gitResult {
	result := &gitResult{dir: dir}
	cmd := exec.CommandContext(ctx, "git", args...)
	cmd.Stdout = &result.stdout
	cmd.Stderr = &result.stderr
	cmd.Dir = dir
	if cancelTree {
		cmd.WaitDelay = time.Second
		configureCancellation(cmd)
	}

	result.err = cmd.Run()
	return result
}

func (result *gitResult) failure() error {
	detail := result.dir
	stderr := strings.TrimSpace(result.stderr.String())
	if stderr != "" {
		detail += ": " + stderr
	}

	return fmt.Errorf("%s: %w", detail, result.err)
}

func (r *Repos) printResult(result *gitResult) {
	r.scrb.BeginDescribe(result.dir)
	if result.err != nil {
		r.scrb.Error(result.err)
	}

	r.scrb.PrintLines(&result.stdout)
	r.scrb.PrintLines(&result.stderr)
	r.scrb.EndDescribe()
}

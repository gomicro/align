package repos

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os/exec"
	"sort"
	"strings"

	ctxhelper "github.com/gomicro/align/client/context"
	"github.com/gosuri/uiprogress"
)

func (r *Repos) ListTags(ctx context.Context, dirs []string, args ...string) error {
	args = append([]string{"tag"}, args...)

	verbose := ctxhelper.Verbose(ctx)
	if verbose {
		r.scrb.BeginDescribe("Command")
		defer r.scrb.EndDescribe()

		r.scrb.Print(fmt.Sprintf("git %s", strings.Join(args, " ")))

		r.scrb.BeginDescribe("directories")
		defer r.scrb.EndDescribe()
	}

	for _, dir := range dirs {
		out := &bytes.Buffer{}
		errout := &bytes.Buffer{}

		cmd := exec.CommandContext(ctx, "git", args...)
		cmd.Stdout = out
		cmd.Stderr = errout
		cmd.Dir = dir

		r.scrb.BeginDescribe(dir)

		err := cmd.Run()
		if err != nil {
			r.scrb.Error(err)
			r.scrb.PrintLines(errout)
		} else {
			r.scrb.PrintLines(out)
		}

		r.scrb.EndDescribe()
	}

	return nil
}

func (r *Repos) TagRepos(ctx context.Context, dirs []string, args ...string) error {
	count := len(dirs)
	args = append([]string{"tag"}, args...)

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
				return fmt.Sprintf("Tagging (%d/%d)", b.Current(), count)
			}).
			AppendFunc(func(b *uiprogress.Bar) string {
				return currRepo
			})
	}

	var errs []error

	for _, dir := range dirs {
		currRepo = fmt.Sprintf("\nCurrent Repo: %v", dir)

		out := &bytes.Buffer{}
		errout := &bytes.Buffer{}

		cmd := exec.CommandContext(ctx, "git", args...)
		cmd.Stdout = out
		cmd.Stderr = errout
		cmd.Dir = dir

		err := cmd.Run()
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
			if err != nil {
				errs = append(errs, fmt.Errorf("%s: %w: %s", dir, err, strings.TrimSpace(errout.String())))
			}

			bar.Incr()
		}
	}

	currRepo = ""

	return errors.Join(errs...)
}

func (r *Repos) GetTagNames(ctx context.Context, dirs []string) ([]string, error) {
	seen := map[string]struct{}{}
	names := []string{}

	for _, dir := range dirs {
		out := &bytes.Buffer{}
		cmd := exec.CommandContext(ctx, "git", "tag", "--list")
		cmd.Stdout = out
		cmd.Dir = dir

		err := cmd.Run()
		if err != nil {
			return nil, fmt.Errorf("get tag names: %w", err)
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

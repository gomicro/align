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

func (r *Repos) ListBranches(ctx context.Context, dirs []string, args ...string) error {
	args = append([]string{"branch"}, args...)

	verbose := ctxhelper.Verbose(ctx)
	if verbose {
		r.scrb.BeginDescribe("Running with command:")
		r.scrb.Print(strings.Join(append([]string{"git"}, args...), " "))
		r.scrb.EndDescribe()
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

func (r *Repos) Branches(ctx context.Context, dirs []string, args ...string) error {
	count := len(dirs)
	args = append([]string{"branch"}, args...)

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
				return fmt.Sprintf("Branches (%d/%d)", b.Current(), count)
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

func (r *Repos) GetBranchNames(ctx context.Context, dirs []string) ([]string, error) {
	seen := map[string]struct{}{}
	names := []string{}

	for _, dir := range dirs {
		out := &bytes.Buffer{}
		cmd := exec.CommandContext(ctx, "git", "branch", "--list", "--format=%(refname:short)")
		cmd.Stdout = out
		cmd.Dir = dir

		err := cmd.Run()
		if err != nil {
			return nil, fmt.Errorf("get branch names: %w", err)
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

func (r *Repos) GetBranchAndTagNames(ctx context.Context, dirs []string) ([]string, error) {
	branches, err := r.GetBranchNames(ctx, dirs)
	if err != nil {
		return nil, fmt.Errorf("get branches: %w", err)
	}

	tags, err := r.GetTagNames(ctx, dirs)
	if err != nil {
		return nil, fmt.Errorf("get tags: %w", err)
	}

	seen := map[string]struct{}{}
	names := []string{}
	for _, n := range append(branches, tags...) {
		if _, ok := seen[n]; !ok {
			seen[n] = struct{}{}
			names = append(names, n)
		}
	}

	return names, nil
}

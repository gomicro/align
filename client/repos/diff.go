package repos

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"os/exec"
	"strings"
)

// DiffConfig controls filtering and output options for DiffRepos.
type DiffConfig struct {
	IgnoreEmpty      bool
	IgnoreFilePrefix []string
	MatchExtension   []string
	Args             []string
}

// DiffRepos runs git diff across all dirs using cfg to filter output.
func (r *Repos) DiffRepos(ctx context.Context, dirs []string, cfg *DiffConfig) error {
	args := append([]string{"diff"}, cfg.Args...)

	r.scrb.BeginDescribe("Command")
	defer r.scrb.EndDescribe()

	r.scrb.Print(fmt.Sprintf("git %s", strings.Join(args, " ")))

	r.scrb.BeginDescribe("directories")
	defer r.scrb.EndDescribe()

	var errs []error

	for _, dir := range dirs {
		out := &bytes.Buffer{}
		errout := &bytes.Buffer{}

		cmd := exec.CommandContext(ctx, "git", args...)
		cmd.Stdout = out
		cmd.Stderr = errout
		cmd.Dir = dir

		err := cmd.Run()

		out = matchExtensions(out, cfg.MatchExtension)

		// filter first to have empty check accurate
		out = filterLines(out, cfg.IgnoreFilePrefix)

		if cfg.IgnoreEmpty && out.Len() == 0 && err == nil {
			continue
		}

		r.scrb.BeginDescribe(dir)
		if err != nil {
			r.scrb.Error(err)
			r.scrb.PrintLines(errout)
			errs = append(errs, fmt.Errorf("%s: %w", dir, err))
		} else {
			r.scrb.PrintLines(out)
		}

		r.scrb.EndDescribe()
	}

	return errors.Join(errs...)
}

func filterLines(buf *bytes.Buffer, prefixes []string) *bytes.Buffer {
	if len(prefixes) == 0 {
		return buf
	}

	scanner := bufio.NewScanner(buf)
	out := &bytes.Buffer{}

	for scanner.Scan() {
		line := scanner.Text()

		ignore := false
		for _, prefix := range prefixes {
			if strings.HasPrefix(line, prefix) {
				ignore = true
				break
			}
		}

		if !ignore {
			out.WriteString(line + "\n")
		}
	}

	return out
}

func matchExtensions(buf *bytes.Buffer, extensions []string) *bytes.Buffer {
	if len(extensions) == 0 {
		return buf
	}

	for i := range extensions {
		if !strings.HasPrefix(extensions[i], ".") {
			extensions[i] = "." + extensions[i]
		}
	}

	scanner := bufio.NewScanner(buf)
	out := &bytes.Buffer{}

	for scanner.Scan() {
		line := scanner.Text()

		matched := false
		for _, ext := range extensions {
			if strings.HasSuffix(line, ext) {
				matched = true
				break
			}
		}

		if matched {
			out.WriteString(line + "\n")
		}
	}

	return out
}

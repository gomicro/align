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

func (r *Remotes) SetURLs(ctx context.Context, dirs []string, name, baseURL string) error {
	count := len(dirs)

	args := append([]string{"remote", "set-url"}, name)

	verbose := ctxhelper.Verbose(ctx)

	var bar *uiprogress.Bar
	currRepo := ""

	if verbose {
		r.scrb.BeginDescribe("Command")
		defer r.scrb.EndDescribe()

		var vargs []string
		vargs = append(vargs, args...)

		url := buildURL(baseURL, "<dir>")
		vargs = append(vargs, url)

		r.scrb.Print(fmt.Sprintf("git %s", strings.Join(vargs, " ")))

		r.scrb.BeginDescribe("directories")
		defer r.scrb.EndDescribe()
	} else {
		bar = uiprogress.AddBar(count).
			AppendCompleted().
			PrependElapsed().
			PrependFunc(func(b *uiprogress.Bar) string {
				return fmt.Sprintf("Setting URL (%d/%d)", b.Current(), count)
			}).
			AppendFunc(func(b *uiprogress.Bar) string {
				return currRepo
			})
	}

	for _, dir := range dirs {
		currRepo = fmt.Sprintf("\nCurrent Repo: %v", dir)

		var cargs []string
		cargs = append(cargs, args...)

		url := buildURL(baseURL, dir)
		cargs = append(cargs, url)

		out := &bytes.Buffer{}
		errout := &bytes.Buffer{}

		cmd := exec.CommandContext(ctx, "git", cargs...)
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

func buildURL(baseURL, dir string) string {
	baseURL = strings.TrimSuffix(baseURL, "/")

	return fmt.Sprintf("%s/%s.git", baseURL, dir)
}

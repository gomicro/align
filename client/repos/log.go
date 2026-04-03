package repos

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strings"

	ctxhelper "github.com/gomicro/align/client/context"
)

func (r *Repos) LogRepos(ctx context.Context, dirs []string, ignoreEmpty bool, args ...string) error {
	args = append([]string{"log"}, args...)

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

		err := cmd.Run()

		if ignoreEmpty && out.Len() == 0 && err == nil {
			continue
		}

		r.scrb.BeginDescribe(dir)
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

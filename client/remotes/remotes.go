package remotes

import (
	"bytes"
	"context"
	"os/exec"

	"github.com/gomicro/scribe"
)

type Remotes struct {
	scrb scribe.Scriber
}

func New(scrb scribe.Scriber) *Remotes {
	return &Remotes{scrb: scrb}
}

func (r *Remotes) Remotes(ctx context.Context, dirs []string, args ...string) error {
	args = append([]string{"remote"}, args...)

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

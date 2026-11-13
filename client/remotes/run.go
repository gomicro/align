package remotes

import (
	"context"
	"fmt"

	ctxhelper "github.com/gomicro/align/client/context"
	"github.com/gomicro/align/client/fanout"
)

// DefaultConcurrency bounds simultaneous remote changes when no override is set.
const DefaultConcurrency = 50

func (r *Remotes) fanOut(ctx context.Context, dirs []string, label string, perDirArgs [][]string) error {
	limit, err := ctxhelper.Concurrency(ctx, DefaultConcurrency)
	if err != nil {
		return fmt.Errorf("resolving concurrency: %w", err)
	}

	jobs := make([]fanout.Job, len(dirs))
	for i, dir := range dirs {
		jobs[i] = fanout.Job{Dir: dir, Args: perDirArgs[i]}
	}

	// Per-job results are already printed and joined into the error.
	_, err = fanout.Run(ctx, r.scrb, label, jobs, limit)
	return err
}

package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"

	ctxhelper "github.com/gomicro/align/client/context"

	"github.com/spf13/cobra"
)

const concurrencyHelp = `

Repositories run in a rolling worker pool: each completed Git call frees a slot
for the next repository. --concurrency sets the maximum number of active calls
for this invocation; it must be positive, and 1 runs sequentially. Overrides may
exceed the default. This option is not forwarded to Git or saved to configuration.

Verbose output is buffered and printed in complete per-repo blocks in completion
order. Normal progress shows completed/total attempts. Repository failures do not
stop other repositories; any failure returns a nonzero exit status. Ctrl-C cancels
active Git calls and stops queued work. Interactive Git input is not supported.`

func concurrencyContext(cmd *cobra.Command) (context.Context, context.CancelFunc, error) {
	limit, err := cmd.Flags().GetInt("concurrency")
	if err != nil {
		return nil, nil, fmt.Errorf("reading concurrency: %w", err)
	}

	if limit < 1 {
		return nil, nil, fmt.Errorf("--concurrency must be a positive integer")
	}

	ctx := cmd.Context()
	if ctx == nil {
		ctx = context.Background()
	}

	ctx, stop := signal.NotifyContext(ctx, os.Interrupt)
	ctx = ctxhelper.WithConcurrency(ctx, limit)
	return ctx, stop, nil
}

package stash

import (
	"fmt"

	ctxhelper "github.com/gomicro/align/client/context"
	"github.com/gomicro/align/client/repos"
	"github.com/gomicro/align/cmd/concurrency"
	"github.com/gomicro/align/cmd/selection"
	"github.com/spf13/cobra"
)

func init() {
	StashCmd.AddCommand(listCmd)
	concurrency.AddFlag(listCmd, repos.DefaultLocalConcurrency)
}

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List stash entries across all repos",
	Long:  "List stash entries across all repos in a directory." + concurrency.Help,
	RunE:  listFunc,
}

func listFunc(cmd *cobra.Command, args []string) error {
	ctx, stop, err := concurrency.Context(cmd)
	if err != nil {
		cmd.SilenceUsage = true
		return fmt.Errorf("stash list options: %w", err)
	}

	defer stop()

	// list output is always verbose — force it so scribe prints each repo's output
	ctx = ctxhelper.WithVerbose(ctx, true)

	repoDirs, err := selection.Dirs(ctx, cmd, clt, ".")
	if err != nil {
		cmd.SilenceUsage = true
		return fmt.Errorf("get dirs: %w", err)
	}

	err = clt.StashRepos(ctx, repoDirs, "list")
	if err != nil {
		cmd.SilenceUsage = true
		return fmt.Errorf("stash list: %w", err)
	}

	return nil
}

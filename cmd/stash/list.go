package stash

import (
	"context"
	"fmt"

	ctxhelper "github.com/gomicro/align/client/context"
	"github.com/spf13/cobra"
)

func init() {
	StashCmd.AddCommand(listCmd)
}

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List stash entries across all repos",
	Long:  "List stash entries across all repos in a directory.",
	RunE:  listFunc,
}

func listFunc(cmd *cobra.Command, args []string) error {
	// list output is always verbose — force it so scribe prints each repo's output
	ctx := ctxhelper.WithVerbose(context.Background(), true)

	repoDirs, err := clt.GetDirs(ctx, ".")
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

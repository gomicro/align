package cmd

import (
	"context"
	"fmt"

	"github.com/gosuri/uiprogress"
	"github.com/spf13/cobra"
)

var pullCmd = &cobra.Command{
	Use:              "pull [dir]",
	Short:            "Pull all repos in a directory",
	Long:             `Pull all repos in a directory.`,
	Args:             cobra.MaximumNArgs(1),
	PersistentPreRun: setupClient,
	RunE:             pullFunc,
}

func pullFunc(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	dir := "."
	if len(args) > 0 {
		dir = args[0]
	}

	uiprogress.Start()
	defer uiprogress.Stop()

	repoDirs, err := clt.GetDirs(ctx, dir)
	if err != nil {
		cmd.SilenceUsage = true
		return fmt.Errorf("get dirs: %w", err)
	}

	err = clt.PullRepos(ctx, repoDirs)
	if err != nil {
		cmd.SilenceUsage = true
		return fmt.Errorf("pull repos: %w", err)
	}

	return nil
}

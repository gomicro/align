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
	Args:             cobra.ExactArgs(1),
	PersistentPreRun: setupClient,
	RunE:             pullFunc,
}

func pullFunc(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	uiprogress.Start()
	defer uiprogress.Stop()

	repoDirs, err := clt.GetDirs(ctx, args[0])
	if err != nil {
		return fmt.Errorf("get dirs: %w", err)
	}

	err = clt.PullRepos(ctx, repoDirs)
	if err != nil {
		return fmt.Errorf("pull repos: %w", err)
	}

	return nil
}

package cmd

import (
	"context"
	"fmt"

	"github.com/gosuri/uiprogress"
	"github.com/spf13/cobra"
)

var checkoutCmd = &cobra.Command{
	Use:              "checkout [branch] (dir)",
	Short:            "checkout the desired branch",
	Long:             `checkout the desired branch`,
	Args:             cobra.RangeArgs(1, 2),
	PersistentPreRun: setupClient,
	RunE:             checkoutFunc,
}

func checkoutFunc(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	branch := args[0]

	dir := "."
	if len(args) > 1 {
		dir = args[1]
	}

	uiprogress.Start()
	defer uiprogress.Stop()

	repoDirs, err := clt.GetDirs(ctx, dir)
	if err != nil {
		return fmt.Errorf("get dirs: %w", err)
	}

	err = clt.CheckoutRepos(ctx, repoDirs, branch)
	if err != nil {
		return fmt.Errorf("checkout repos: %w", err)
	}

	return nil
}

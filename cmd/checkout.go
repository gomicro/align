package cmd

import (
	"context"
	"fmt"

	ctxhelper "github.com/gomicro/align/client/context"
	"github.com/gosuri/uiprogress"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func init() {
	RootCmd.AddCommand(checkoutCmd)
}

var checkoutCmd = &cobra.Command{
	Use:               "checkout [branch]",
	Short:             "Checkout a branch across all repos in a directory",
	Long:              `Switch to the specified branch across all repos in a directory.`,
	Args:              cobra.RangeArgs(1, 2),
	ValidArgsFunction: checkoutCmdValidArgsFunc,
	PersistentPreRun:  setupClient,
	RunE:              checkoutFunc,
}

func checkoutCmdValidArgsFunc(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	if len(args) >= 2 {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	setupClient(cmd, args)

	checkoutDir := "."

	ctx := context.Background()

	repoDirs, err := clt.GetDirs(ctx, checkoutDir)
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	names, err := clt.GetBranchAndTagNames(ctx, repoDirs)
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	return names, cobra.ShellCompDirectiveNoFileComp
}

func checkoutFunc(cmd *cobra.Command, args []string) error {
	verbose := viper.GetBool("verbose")
	ctx := ctxhelper.WithVerbose(context.Background(), verbose)

	if !verbose {
		uiprogress.Start()
		defer uiprogress.Stop()
	}

	repoDirs, err := clt.GetDirs(ctx, ".")
	if err != nil {
		cmd.SilenceUsage = true
		return fmt.Errorf("get dirs: %w", err)
	}

	err = clt.CheckoutRepos(ctx, repoDirs, args...)
	if err != nil {
		cmd.SilenceUsage = true
		return fmt.Errorf("checkout repos: %w", err)
	}

	return nil
}

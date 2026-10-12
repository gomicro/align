package cmd

import (
	"context"
	"fmt"

	ctxhelper "github.com/gomicro/align/client/context"
	"github.com/gomicro/align/client/repos"
	"github.com/gomicro/align/cmd/selection"

	"github.com/gosuri/uiprogress"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func init() {
	RootCmd.AddCommand(checkoutCmd)
	checkoutCmd.Flags().Int("concurrency", repos.DefaultCheckoutConcurrency, "maximum simultaneous Git checkouts (1 runs sequentially)")
}

var checkoutCmd = &cobra.Command{
	Use:               "checkout [branch]",
	Short:             "Checkout a branch across all repos in a directory",
	Long:              `Switch to the specified branch across all repos in a directory.` + concurrencyHelp,
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

	repoDirs, err := selection.Dirs(ctx, cmd, clt, checkoutDir)
	if err != nil {
		cobra.CompErrorln(err.Error())
		return nil, cobra.ShellCompDirectiveError
	}

	names, err := clt.GetBranchAndTagNames(ctx, repoDirs)
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	return names, cobra.ShellCompDirectiveNoFileComp
}

func checkoutFunc(cmd *cobra.Command, args []string) error {
	ctx, stop, err := concurrencyContext(cmd)
	if err != nil {
		cmd.SilenceUsage = true
		return fmt.Errorf("checkout options: %w", err)
	}

	defer stop()

	verbose := viper.GetBool("verbose")
	ctx = ctxhelper.WithVerbose(ctx, verbose)

	if !verbose {
		uiprogress.Start()
		defer uiprogress.Stop()
	}

	repoDirs, err := selection.Dirs(ctx, cmd, clt, ".")
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

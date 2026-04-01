package cmd

import (
	"context"
	"fmt"

	"github.com/gomicro/align/client"
	"github.com/gosuri/uiprogress"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func init() {
	RootCmd.AddCommand(checkoutCmd)

	checkoutCmd.Flags().StringVar(&dir, "dir", ".", "directory to checkout repos from")
}

var checkoutCmd = &cobra.Command{
	Use:               "checkout [branch]",
	Short:             "checkout the desired branch",
	Long:              `checkout the desired branch`,
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

	checkoutDir, err := cmd.Flags().GetString("dir")
	if err != nil {
		checkoutDir = "."
	}

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
	ctx := client.WithVerbose(context.Background(), verbose)

	if !verbose {
		uiprogress.Start()
		defer uiprogress.Stop()
	}

	repoDirs, err := clt.GetDirs(ctx, dir)
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

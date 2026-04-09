package cmd

import (
	"context"
	"fmt"

	ctxhelper "github.com/gomicro/align/client/context"
	"github.com/gosuri/uiprogress"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var prune bool

func init() {
	RootCmd.AddCommand(fetchCmd)

	fetchCmd.Flags().BoolVar(&tags, "tags", false, "fetch all tags")
	fetchCmd.Flags().BoolVarP(&prune, "prune", "p", false, "remove stale remote-tracking refs after fetching")
	fetchCmd.Flags().BoolVar(&all, "all", false, "fetch from all configured remotes")
}

var fetchCmd = &cobra.Command{
	Use:               "fetch [remote]",
	Short:             "Fetch from remotes across all repos in a directory",
	Long:              `Fetch from remotes across all repos in a directory without merging into the working tree.`,
	Args:              cobra.MaximumNArgs(1),
	ValidArgsFunction: fetchCmdValidArgsFunc,
	PersistentPreRun:  setupClient,
	RunE:              fetchFunc,
}

func fetchCmdValidArgsFunc(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	if len(args) >= 1 {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	setupClient(cmd, args)

	ctx := context.Background()

	repoDirs, err := clt.GetDirs(ctx, ".")
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	names, err := clt.GetRemoteNames(ctx, repoDirs)
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	return names, cobra.ShellCompDirectiveNoFileComp
}

func fetchFunc(cmd *cobra.Command, args []string) error {
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

	if all {
		args = append(args, "--all")
	}

	if prune {
		args = append(args, "--prune")
	}

	if tags {
		args = append(args, "--tags")
	}

	err = clt.FetchRepos(ctx, repoDirs, args...)
	if err != nil {
		cmd.SilenceUsage = true
		return fmt.Errorf("fetch repos: %w", err)
	}

	return nil
}

package cmd

import (
	"context"
	"fmt"
	"slices"

	ctxhelper "github.com/gomicro/align/client/context"
	"github.com/gosuri/uiprogress"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	all         bool
	force       bool
	setUpstream bool
	followTags  bool
)

func init() {
	RootCmd.AddCommand(pushCmd)

	pushCmd.Flags().StringVar(&dir, "dir", ".", "directory to push repos from")
	pushCmd.Flags().BoolVar(&all, "all", false, "all branches")
	pushCmd.Flags().BoolVar(&force, "force", false, "force push")
	pushCmd.Flags().BoolVar(&tags, "tags", false, "push all tags")
	pushCmd.Flags().BoolVar(&followTags, "follow-tags", false, "push annotated tags reachable from pushed commits")
	pushCmd.Flags().BoolVarP(&setUpstream, "set-upstream", "u", false, "set upstream tracking reference")
}

var pushCmd = &cobra.Command{
	Use:               "push",
	Short:             "Push all repos in a directory",
	Long:              `Push local commits to the remote across all repos in a directory.`,
	ValidArgsFunction: pushCmdValidArgsFunc,
	PersistentPreRun:  setupClient,
	RunE:              pushFunc,
}

func pushCmdValidArgsFunc(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	setupClient(cmd, args)

	pushDir, err := cmd.Flags().GetString("dir")
	if err != nil {
		pushDir = "."
	}

	ctx := context.Background()

	repoDirs, err := clt.GetDirs(ctx, pushDir)
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	switch len(args) {
	case 0:
		names, err := clt.GetRemoteNames(ctx, repoDirs)
		if err != nil {
			return nil, cobra.ShellCompDirectiveNoFileComp
		}

		return names, cobra.ShellCompDirectiveNoFileComp

	case 1:
		names, err := clt.GetBranchAndTagNames(ctx, repoDirs)
		if err != nil {
			return nil, cobra.ShellCompDirectiveNoFileComp
		}

		return names, cobra.ShellCompDirectiveNoFileComp
	}

	return nil, cobra.ShellCompDirectiveNoFileComp
}

func pushFunc(cmd *cobra.Command, args []string) error {
	verbose := viper.GetBool("verbose")
	ctx := ctxhelper.WithVerbose(context.Background(), verbose)

	if !verbose {
		uiprogress.Start()
		defer uiprogress.Stop()
	}

	repoDirs, err := clt.GetDirs(ctx, dir)
	if err != nil {
		cmd.SilenceUsage = true
		return fmt.Errorf("get dirs: %w", err)
	}

	if setUpstream {
		args = slices.Insert(args, 0, "--set-upstream")
	}

	if all {
		args = slices.Insert(args, 0, "--all")
	}

	if tags {
		args = slices.Insert(args, 0, "--tags")
	}

	if followTags {
		args = slices.Insert(args, 0, "--follow-tags")
	}

	if force {
		args = slices.Insert(args, 0, "--force")
	}

	err = clt.PushRepos(ctx, repoDirs, args...)
	if err != nil {
		cmd.SilenceUsage = true
		return fmt.Errorf("push repos: %w", err)
	}

	return nil
}

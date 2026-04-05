package cmd

import (
	"context"
	"fmt"

	ctxhelper "github.com/gomicro/align/client/context"
	"github.com/gosuri/uiprogress"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	del      bool
	delForce bool
)

func init() {
	RootCmd.AddCommand(branchCmd)

	branchCmd.Flags().BoolVarP(&all, "all", "a", false, "list all branches")
	branchCmd.Flags().BoolVarP(&del, "delete", "d", false, "delete the branch from the repos")
	branchCmd.Flags().BoolVarP(&delForce, "force-delete", "D", false, "force delete the branch from the repos")
	branchCmd.Flags().BoolVarP(&force, "force", "f", false, "force the desired action")

	branchCmd.MarkFlagsMutuallyExclusive("all", "delete", "force-delete")
}

var branchCmd = &cobra.Command{
	Use:               "branch",
	Short:             "manage branches for a set of repositories",
	Long:              `manage branches for a set of repositories`,
	ValidArgsFunction: branchCmdValidArgsFunc,
	PersistentPreRun:  setupClient,
	RunE:              branchFunc,
}

func branchCmdValidArgsFunc(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	if len(args) >= 1 {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	isDelete, _ := cmd.Flags().GetBool("delete")
	isForceDelete, _ := cmd.Flags().GetBool("force-delete")

	if !isDelete && !isForceDelete {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	setupClient(cmd, args)

	ctx := context.Background()

	repoDirs, err := clt.GetDirs(ctx, dir)
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	names, err := clt.GetBranchNames(ctx, repoDirs)
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	return names, cobra.ShellCompDirectiveNoFileComp
}

func branchFunc(cmd *cobra.Command, args []string) error {
	verbose := viper.GetBool("verbose")
	ctx := ctxhelper.WithVerbose(context.Background(), verbose)

	repoDirs, err := clt.GetDirs(ctx, dir)
	if err != nil {
		cmd.SilenceUsage = true
		return fmt.Errorf("get dirs: %w", err)
	}

	if del || delForce {
		if len(args) == 0 {
			cmd.SilenceUsage = true
			return fmt.Errorf("branch name is required when deleting a branch")
		}

		if !verbose {
			uiprogress.Start()
			defer uiprogress.Stop()
		}

		name := args[0]

		args = []string{"--delete"}

		if delForce || force {
			args = append(args, "--force")
		}

		args = append(args, name)

		err := clt.Branches(ctx, repoDirs, args...)
		if err != nil {
			cmd.SilenceUsage = true
			return fmt.Errorf("delete: %w", err)
		}

		return nil
	}

	if all {
		args = append(args, "--all")
	}

	// This must be verbose to show anything
	ctx = ctxhelper.WithVerbose(ctx, true)

	err = clt.Branches(ctx, repoDirs, args...)
	if err != nil {
		cmd.SilenceUsage = true
		return fmt.Errorf("list: %w", err)
	}

	return nil
}

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
	del        bool
	delForce   bool
	moveBranch bool
)

func init() {
	RootCmd.AddCommand(branchCmd)

	branchCmd.Flags().BoolVarP(&all, "all", "a", false, "list all branches")
	branchCmd.Flags().BoolVarP(&del, "delete", "d", false, "delete the branch from the repos")
	branchCmd.Flags().BoolVarP(&delForce, "force-delete", "D", false, "force delete the branch from the repos")
	branchCmd.Flags().BoolVarP(&force, "force", "f", false, "force the desired action")
	branchCmd.Flags().BoolVarP(&moveBranch, "move", "m", false, "rename a branch: align branch --move <old> <new>")

	branchCmd.MarkFlagsMutuallyExclusive("all", "delete", "force-delete", "move")
}

var branchCmd = &cobra.Command{
	Use:               "branch",
	Short:             "manage branches for a set of repositories",
	Long:              `manage branches for a set of repositories`,
	Args:              cobra.MaximumNArgs(2),
	ValidArgsFunction: branchCmdValidArgsFunc,
	PersistentPreRun:  setupClient,
	RunE:              branchFunc,
}

func branchCmdValidArgsFunc(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	isDelete, _ := cmd.Flags().GetBool("delete")
	isForceDelete, _ := cmd.Flags().GetBool("force-delete")
	isMove, _ := cmd.Flags().GetBool("move")

	// delete/force-delete: complete first arg only
	if (isDelete || isForceDelete) && len(args) >= 1 {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	// move: complete the old branch name (first arg); new name is freeform
	if isMove && len(args) >= 1 {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	if !isDelete && !isForceDelete && !isMove {
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

	if moveBranch {
		if len(args) != 2 {
			return fmt.Errorf("old and new branch names are required when renaming a branch")
		}

		if !verbose {
			uiprogress.Start()
			defer uiprogress.Stop()
		}

		err := clt.Branches(ctx, repoDirs, "--move", args[0], args[1])
		if err != nil {
			cmd.SilenceUsage = true
			return fmt.Errorf("move: %w", err)
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

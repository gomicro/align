package cmd

import (
	"context"
	"fmt"

	"github.com/gomicro/align/client"
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
	Use:              "branch",
	Short:            "manage branches for a set of repositories",
	Long:             `manage branches for a set of repositories`,
	PersistentPreRun: setupClient,
	RunE:             branchFunc,
}

func branchFunc(cmd *cobra.Command, args []string) error {
	verbose := viper.GetBool("verbose")
	ctx := client.WithVerbose(context.Background(), verbose)

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
	ctx = client.WithVerbose(ctx, true)

	err = clt.Branches(ctx, repoDirs, args...)
	if err != nil {
		cmd.SilenceUsage = true
		return fmt.Errorf("list: %w", err)
	}

	return nil
}

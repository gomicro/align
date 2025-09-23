package cmd

import (
	"context"
	"fmt"
	"slices"

	"github.com/gomicro/align/client"
	"github.com/gosuri/uiprogress"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	all         bool
	force       bool
	setUpstream bool
)

func init() {
	RootCmd.AddCommand(pushCmd)

	pushCmd.Flags().StringVar(&dir, "dir", ".", "directory to push repos from")
	pushCmd.Flags().BoolVar(&all, "all", false, "all branches")
	pushCmd.Flags().BoolVar(&force, "force", false, "force push")
	pushCmd.Flags().BoolVarP(&setUpstream, "set-upstream", "u", false, "set upstream tracking reference")
}

var pushCmd = &cobra.Command{
	Use:              "push",
	Short:            "Push all repos in a directory",
	Long:             `Push all repos in a directory.`,
	PersistentPreRun: setupClient,
	RunE:             pushFunc,
}

func pushFunc(cmd *cobra.Command, args []string) error {
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

	if setUpstream {
		args = slices.Insert(args, 0, "--set-upstream")
	}

	if all {
		args = slices.Insert(args, 0, "--all")
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

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
	setUpstream bool
)

func init() {
	RootCmd.AddCommand(pushCmd)

	pushCmd.Flags().StringVarP(&dir, "dir", "d", ".", "directory to push repos from")
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

	err = clt.PushRepos(ctx, repoDirs, args...)
	if err != nil {
		cmd.SilenceUsage = true
		return fmt.Errorf("push repos: %w", err)
	}

	return nil
}

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
	dir  string
	tags bool
)

func init() {
	RootCmd.AddCommand(pullCmd)

	pullCmd.Flags().StringVar(&dir, "dir", ".", "directory to pull repos from")
	pullCmd.Flags().BoolVar(&tags, "tags", false, "pull tags")
	pullCmd.Flags().BoolVar(&prune, "prune", false, "remove stale remote-tracking refs after pulling")
}

var pullCmd = &cobra.Command{
	Use:              "pull",
	Short:            "Pull all repos in a directory",
	Long:             `Pull all repos in a directory.`,
	PersistentPreRun: setupClient,
	RunE:             pullFunc,
}

func pullFunc(cmd *cobra.Command, args []string) error {
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

	if tags {
		args = append(args, "--tags")
	}

	if prune {
		args = append(args, "--prune")
	}

	err = clt.PullRepos(ctx, repoDirs, args...)
	if err != nil {
		cmd.SilenceUsage = true
		return fmt.Errorf("pull repos: %w", err)
	}

	return nil
}

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
	dir  string
	tags bool
)

func init() {
	RootCmd.AddCommand(pullCmd)

	pullCmd.Flags().StringVarP(&dir, "dir", "d", ".", "directory to pull repos from")
	pullCmd.Flags().BoolVar(&tags, "tags", false, "pull tags")
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

	if tags {
		args = append(args, "--tags")
	}

	err = clt.PullRepos(ctx, repoDirs, args...)
	if err != nil {
		cmd.SilenceUsage = true
		return fmt.Errorf("pull repos: %w", err)
	}

	return nil
}

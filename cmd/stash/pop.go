package stash

import (
	"context"
	"fmt"

	ctxhelper "github.com/gomicro/align/client/context"
	"github.com/gosuri/uiprogress"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func init() {
	StashCmd.AddCommand(popCmd)
}

var popCmd = &cobra.Command{
	Use:   "pop",
	Short: "Apply the most recent stash across all repos",
	Long:  "Apply the most recent stash entry and remove it from the stash list across all repos.",
	RunE:  popFunc,
}

func popFunc(cmd *cobra.Command, args []string) error {
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

	err = clt.StashRepos(ctx, repoDirs, "pop")
	if err != nil {
		cmd.SilenceUsage = true
		return fmt.Errorf("stash pop: %w", err)
	}

	return nil
}

package stash

import (
	"fmt"

	ctxhelper "github.com/gomicro/align/client/context"
	"github.com/gomicro/align/client/repos"
	"github.com/gomicro/align/cmd/concurrency"
	"github.com/gomicro/align/cmd/selection"
	"github.com/gosuri/uiprogress"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func init() {
	StashCmd.AddCommand(popCmd)
	concurrency.AddFlag(popCmd, repos.DefaultLocalConcurrency)
}

var popCmd = &cobra.Command{
	Use:   "pop",
	Short: "Apply the most recent stash across all repos",
	Long:  "Apply the most recent stash entry and remove it from the stash list across all repos." + concurrency.Help,
	RunE:  popFunc,
}

func popFunc(cmd *cobra.Command, args []string) error {
	ctx, stop, err := concurrency.Context(cmd)
	if err != nil {
		cmd.SilenceUsage = true
		return fmt.Errorf("stash pop options: %w", err)
	}

	defer stop()

	verbose := viper.GetBool("verbose")
	ctx = ctxhelper.WithVerbose(ctx, verbose)

	if !verbose {
		uiprogress.Start()
		defer uiprogress.Stop()
	}

	repoDirs, err := selection.Dirs(ctx, cmd, clt, ".")
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

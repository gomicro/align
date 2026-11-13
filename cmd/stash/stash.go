package stash

import (
	"fmt"
	"os"

	"github.com/gomicro/align/client"
	ctxhelper "github.com/gomicro/align/client/context"
	"github.com/gomicro/align/client/repos"
	"github.com/gomicro/align/cmd/concurrency"
	"github.com/gomicro/align/cmd/selection"
	"github.com/gomicro/align/config"
	"github.com/gosuri/uiprogress"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	clt client.Clienter
)

func init() {
	concurrency.AddFlag(StashCmd, repos.DefaultLocalConcurrency)
}

var StashCmd = &cobra.Command{
	Use:              "stash",
	Short:            "Stash changes across all repos in a directory",
	Long:             "Stash uncommitted changes across all repos in a directory." + concurrency.Help,
	Args:             cobra.NoArgs,
	PersistentPreRun: setupClient,
	RunE:             stashFunc,
}

func stashFunc(cmd *cobra.Command, args []string) error {
	ctx, stop, err := concurrency.Context(cmd)
	if err != nil {
		cmd.SilenceUsage = true
		return fmt.Errorf("stash options: %w", err)
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

	err = clt.StashRepos(ctx, repoDirs)
	if err != nil {
		cmd.SilenceUsage = true
		return fmt.Errorf("stash repos: %w", err)
	}

	return nil
}

func setupClient(cmd *cobra.Command, args []string) {
	c, err := config.ParseFromFile()
	if err != nil {
		fmt.Printf("Error: %s\n", err)
		os.Exit(1)
	}

	clt, err = client.New(c)
	if err != nil {
		fmt.Printf("Error: %s\n", err)
		os.Exit(1)
	}
}

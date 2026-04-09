package stash

import (
	"context"
	"fmt"
	"os"

	"github.com/gomicro/align/client"
	ctxhelper "github.com/gomicro/align/client/context"
	"github.com/gomicro/align/config"
	"github.com/gosuri/uiprogress"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	clt client.Clienter
)

var StashCmd = &cobra.Command{
	Use:              "stash",
	Short:            "Stash changes across all repos in a directory",
	Long:             "Stash uncommitted changes across all repos in a directory.",
	PersistentPreRun: setupClient,
	RunE:             stashFunc,
}

func stashFunc(cmd *cobra.Command, args []string) error {
	verbose := viper.GetBool("verbose")
	ctx := ctxhelper.WithVerbose(context.Background(), verbose)

	if !verbose {
		uiprogress.Start()
		defer uiprogress.Stop()
	}

	repoDirs, err := clt.GetDirs(ctx, ".")
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

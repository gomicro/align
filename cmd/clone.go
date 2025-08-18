package cmd

import (
	"context"
	"fmt"

	"github.com/gomicro/align/client"
	"github.com/gosuri/uiprogress"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cloneCmd = &cobra.Command{
	Use:               "clone [user|org]",
	Short:             "Clone all active repos from an org or user.",
	Long:              `Clone all active repos from an org or user.`,
	Args:              cobra.MaximumNArgs(1),
	ValidArgsFunction: createCmdValidArgsFunc,
	PersistentPreRun:  setupClient,
	RunE:              cloneFunc,
}

func cloneFunc(cmd *cobra.Command, args []string) error {
	verbose := viper.GetBool("verbose")
	ctx := client.WithVerbose(context.Background(), verbose)

	if !verbose {
		uiprogress.Start()
		defer uiprogress.Stop()
	}

	name := ""
	if len(args) > 0 {
		name = args[0]
	}

	repos, err := clt.GetRepos(ctx, name)
	if err != nil {
		cmd.SilenceUsage = true
		return fmt.Errorf("get repos: %w", err)
	}

	ctx = client.WithRepos(ctx, repos)

	_, err = clt.CloneRepos(ctx)
	if err != nil {
		cmd.SilenceUsage = true
		return fmt.Errorf("clone repos: %w", err)
	}

	return nil
}

func createCmdValidArgsFunc(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	setupClient(cmd, args)

	valid, err := clt.GetLogins(context.Background())
	if err != nil {
		valid = []string{"error fetching"}
	}

	return valid, cobra.ShellCompDirectiveNoFileComp
}

package cmd

import (
	"context"

	"github.com/gosuri/uiprogress"
	"github.com/spf13/cobra"
)

func init() {
}

var downloadCmd = &cobra.Command{
	Use:               "download [user|org]",
	Short:             "Download all active repos from an org or user.",
	Long:              `Download all active repos from an org or user.`,
	Args:              cobra.MaximumNArgs(1),
	ValidArgsFunction: createCmdValidArgsFunc,
	PersistentPreRun:  setupClient,
	RunE:              downloadFunc,
}

func downloadFunc(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	uiprogress.Start()

	name := ""
	if len(args) > 0 {
		name = args[0]
	}

	repos, err := clt.GetRepos(ctx, name)
	if err != nil {
		return err
	}

	err = clt.CloneRepos(ctx, repos)
	if err != nil {
		return err
	}

	uiprogress.Stop()

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

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
	list bool
)

func init() {
	RootCmd.AddCommand(tagCmd)

	tagCmd.Flags().BoolVarP(&list, "list", "l", false, "list tags in repositories with optional pattern")
}

var tagCmd = &cobra.Command{
	Use:              "tag",
	Short:            "Create, list, or delete tags in repositories",
	Long:             `Create, list, or delete tags in repositories`,
	PersistentPreRun: setupClient,
	RunE:             tagFunc,
}

func tagFunc(cmd *cobra.Command, args []string) error {
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

	if list || len(args) == 0 {
		args = append([]string{"--list"}, args...)

		err = clt.ListTags(ctx, repoDirs, args...)
		if err != nil {
			cmd.SilenceUsage = true
			return fmt.Errorf("list tags: %w", err)
		}

		return nil
	}

	return nil
}

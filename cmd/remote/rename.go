package remote

import (
	"context"
	"fmt"

	ctxhelper "github.com/gomicro/align/client/context"
	"github.com/gosuri/uiprogress"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func init() {
	RemoteCmd.AddCommand(renameCmd)
}

var renameCmd = &cobra.Command{
	Use:               "rename <old_name> <new_name>",
	Short:             "Rename a remote across all repos.",
	Long:              "Rename a remote across all repos by specifying its current name and the desired new name.",
	Args:              cobra.ExactArgs(2),
	ValidArgsFunction: renameCmdValidArgsFunc,
	RunE:              renameFunc,
}

func renameCmdValidArgsFunc(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	if len(args) >= 1 {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	setupClient(cmd, args)

	ctx := context.Background()

	repoDirs, err := clt.GetDirs(ctx, dir)
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	names, err := clt.GetRemoteNames(ctx, repoDirs)
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	return names, cobra.ShellCompDirectiveNoFileComp
}

func renameFunc(cmd *cobra.Command, args []string) error {
	verbose := viper.GetBool("verbose")
	ctx := ctxhelper.WithVerbose(context.Background(), verbose)

	if !verbose {
		uiprogress.Start()
		defer uiprogress.Stop()
	}

	oldName, newName := args[0], args[1]

	repoDirs, err := clt.GetDirs(ctx, dir)
	if err != nil {
		cmd.SilenceUsage = true
		return fmt.Errorf("get dirs: %w", err)
	}

	err = clt.Rename(ctx, repoDirs, oldName, newName)
	if err != nil {
		cmd.SilenceUsage = true
		return fmt.Errorf("rename: %w", err)
	}

	return nil
}

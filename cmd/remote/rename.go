package remote

import (
	"context"
	"fmt"

	ctxhelper "github.com/gomicro/align/client/context"
	"github.com/gomicro/align/client/remotes"
	"github.com/gomicro/align/cmd/concurrency"
	"github.com/gomicro/align/cmd/selection"
	"github.com/gosuri/uiprogress"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func init() {
	RemoteCmd.AddCommand(renameCmd)
	concurrency.AddFlag(renameCmd, remotes.DefaultConcurrency)
}

var renameCmd = &cobra.Command{
	Use:               "rename <old_name> <new_name>",
	Short:             "Rename a remote across all repos.",
	Long:              "Rename a remote across all repos by specifying its current name and the desired new name." + concurrency.Help,
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

	repoDirs, err := selection.Dirs(ctx, cmd, clt, ".")
	if err != nil {
		cobra.CompErrorln(err.Error())
		return nil, cobra.ShellCompDirectiveError
	}

	names, err := clt.GetRemoteNames(ctx, repoDirs)
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	return names, cobra.ShellCompDirectiveNoFileComp
}

func renameFunc(cmd *cobra.Command, args []string) error {
	ctx, stop, err := concurrency.Context(cmd)
	if err != nil {
		cmd.SilenceUsage = true
		return fmt.Errorf("remote rename options: %w", err)
	}

	defer stop()

	verbose := viper.GetBool("verbose")
	ctx = ctxhelper.WithVerbose(ctx, verbose)

	if !verbose {
		uiprogress.Start()
		defer uiprogress.Stop()
	}

	oldName, newName := args[0], args[1]

	repoDirs, err := selection.Dirs(ctx, cmd, clt, ".")
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

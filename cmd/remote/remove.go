package remote

import (
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
	RemoteCmd.AddCommand(removeCmd)
	concurrency.AddFlag(removeCmd, remotes.DefaultConcurrency)
}

var removeCmd = &cobra.Command{
	Use:   "remove <remote_name>",
	Short: "Remove a remote repository.",
	Long:  "Remove a remote repository by its name." + concurrency.Help,
	Args:  cobra.ExactArgs(1),
	RunE:  removeFunc,
}

func removeFunc(cmd *cobra.Command, args []string) error {
	ctx, stop, err := concurrency.Context(cmd)
	if err != nil {
		cmd.SilenceUsage = true
		return fmt.Errorf("remote remove options: %w", err)
	}

	defer stop()

	verbose := viper.GetBool("verbose")
	ctx = ctxhelper.WithVerbose(ctx, verbose)

	if !verbose {
		uiprogress.Start()
		defer uiprogress.Stop()
	}

	name := args[0]

	repoDirs, err := selection.Dirs(ctx, cmd, clt, ".")
	if err != nil {
		cmd.SilenceUsage = true
		return fmt.Errorf("get dirs: %w", err)
	}

	err = clt.Remove(ctx, repoDirs, name)
	if err != nil {
		cmd.SilenceUsage = true
		return fmt.Errorf("remove: %w", err)
	}

	return nil
}

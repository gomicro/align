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
	RemoteCmd.AddCommand(setURLCmd)
	concurrency.AddFlag(setURLCmd, remotes.DefaultConcurrency)
}

var setURLCmd = &cobra.Command{
	Use:   "set-url <remote_name> <base_url>",
	Short: "Set the URL for a remote across all repos.",
	Long:  "Update the URL of a named remote across all repos in a directory." + concurrency.Help,
	Args:  cobra.ExactArgs(2),
	RunE:  setURLFunc,
}

func setURLFunc(cmd *cobra.Command, args []string) error {
	ctx, stop, err := concurrency.Context(cmd)
	if err != nil {
		cmd.SilenceUsage = true
		return fmt.Errorf("remote set-url options: %w", err)
	}

	defer stop()

	verbose := viper.GetBool("verbose")
	ctx = ctxhelper.WithVerbose(ctx, verbose)

	if !verbose {
		uiprogress.Start()
		defer uiprogress.Stop()
	}

	name, baseURL := args[0], args[1]

	repoDirs, err := selection.Dirs(ctx, cmd, clt, ".")
	if err != nil {
		cmd.SilenceUsage = true
		return fmt.Errorf("get dirs: %w", err)
	}

	err = clt.SetURLs(ctx, repoDirs, name, baseURL)
	if err != nil {
		cmd.SilenceUsage = true
		return fmt.Errorf("set url: %w", err)
	}

	return nil
}

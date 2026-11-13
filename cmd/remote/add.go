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
	RemoteCmd.AddCommand(addCmd)
	concurrency.AddFlag(addCmd, remotes.DefaultConcurrency)
}

var addCmd = &cobra.Command{
	Use:   "add <remote_name> <base_url>",
	Short: "Add a remote repository.",
	Long:  "Add a remote repository by specifying its name and base URL. Additional arguments can be provided for configuration." + concurrency.Help,
	Args:  cobra.ExactArgs(2),
	RunE:  addFunc,
}

func addFunc(cmd *cobra.Command, args []string) error {
	ctx, stop, err := concurrency.Context(cmd)
	if err != nil {
		cmd.SilenceUsage = true
		return fmt.Errorf("remote add options: %w", err)
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

	err = clt.Add(ctx, repoDirs, name, baseURL)
	if err != nil {
		cmd.SilenceUsage = true
		return fmt.Errorf("add: %w", err)
	}

	return nil
}

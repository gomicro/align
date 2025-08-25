package remote

import (
	"context"
	"fmt"

	"github.com/gomicro/align/client"
	"github.com/gosuri/uiprogress"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func init() {
	RemoteCmd.AddCommand(removeCmd)
}

var removeCmd = &cobra.Command{
	Use:   "remove <remote_name>",
	Short: "Remove a remote repository.",
	Long:  "Remove a remote repository by its name.",
	Args:  cobra.ExactArgs(1),
	RunE:  removeFunc,
}

func removeFunc(cmd *cobra.Command, args []string) error {
	verbose := viper.GetBool("verbose")
	ctx := client.WithVerbose(context.Background(), verbose)

	if !verbose {
		uiprogress.Start()
		defer uiprogress.Stop()
	}

	name := args[0]

	repoDirs, err := clt.GetDirs(ctx, dir)
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

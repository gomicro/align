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
	RemoteCmd.AddCommand(setURLCmd)
}

var setURLCmd = &cobra.Command{
	Use:   "set-url <remote_name> <base_url>",
	Short: "Set the URL for a remote repository.",
	Long:  "Set the URL for a remote repository.",
	Args:  cobra.ExactArgs(2),
	RunE:  setURLFunc,
}

func setURLFunc(cmd *cobra.Command, args []string) error {
	verbose := viper.GetBool("verbose")
	ctx := client.WithVerbose(context.Background(), verbose)

	if !verbose {
		uiprogress.Start()
		defer uiprogress.Stop()
	}

	name, baseURL := args[0], args[1]

	repoDirs, err := clt.GetDirs(ctx, dir)
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

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
	RemoteCmd.AddCommand(addCmd)
}

var addCmd = &cobra.Command{
	Use:   "add <remote_name> <base_url>",
	Short: "Add a remote repository.",
	Long:  "Add a remote repository by specifying its name and base URL. Additional arguments can be provided for configuration.",
	Args:  cobra.ExactArgs(2),
	RunE:  addFunc,
}

func addFunc(cmd *cobra.Command, args []string) error {
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

	err = clt.Add(ctx, repoDirs, name, baseURL)
	if err != nil {
		cmd.SilenceUsage = true
		return fmt.Errorf("add: %w", err)
	}

	return nil
}

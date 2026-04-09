package remote

import (
	"context"
	"fmt"
	"os"

	"github.com/gomicro/align/client"
	ctxhelper "github.com/gomicro/align/client/context"
	"github.com/gomicro/align/config"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var clt client.Clienter

var RemoteCmd = &cobra.Command{
	Use:              "remote",
	Short:            "Manage remotes across all repos in a directory",
	Long:             "Add, remove, rename, or update remotes across all repos in a directory.",
	PersistentPreRun: setupClient,
	RunE:             remoteFunc,
}

func remoteFunc(cmd *cobra.Command, args []string) error {
	verbose := viper.GetBool("verbose")
	ctx := ctxhelper.WithVerbose(context.Background(), verbose)

	repoDirs, err := clt.GetDirs(ctx, ".")
	if err != nil {
		cmd.SilenceUsage = true
		return fmt.Errorf("get dirs: %w", err)
	}

	if verbose {
		args = append(args, "--verbose")
	}

	err = clt.Remotes(ctx, repoDirs, args...)
	if err != nil {
		cmd.SilenceUsage = true
		return fmt.Errorf("remotes: %w", err)
	}

	return nil
}

func setupClient(cmd *cobra.Command, args []string) {
	c, err := config.ParseFromFile()
	if err != nil {
		fmt.Printf("Error: %s\n", err)
		os.Exit(1)
	}

	clt, err = client.New(c)
	if err != nil {
		fmt.Printf("Error: %s\n", err)
		os.Exit(1)
	}
}

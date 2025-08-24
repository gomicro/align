package remote

import (
	"context"
	"fmt"
	"os"

	"github.com/gomicro/align/client"
	"github.com/gomicro/align/config"
	"github.com/spf13/cobra"
)

var (
	clt     client.Clienter
	dir     string
	verbose bool
)

func init() {
	RemoteCmd.Flags().StringVarP(&dir, "dir", "d", ".", "directory to pull repos from")
	RemoteCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "enable verbose output")
}

var RemoteCmd = &cobra.Command{
	Use:              "remote",
	Short:            "manage repositories remotes",
	Long:             "manage repositories remotes",
	PersistentPreRun: setupClient,
	RunE:             remoteFunc,
}

func remoteFunc(cmd *cobra.Command, args []string) error {
	ctx := client.WithVerbose(context.Background(), verbose)

	repoDirs, err := clt.GetDirs(ctx, dir)
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

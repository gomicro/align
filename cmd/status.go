package cmd

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
)

func init() {
	RootCmd.AddCommand(statusCmd)

	statusCmd.Flags().StringVar(&dir, "dir", ".", "directory to show status of repos in")
	statusCmd.Flags().BoolVarP(&short, "short", "s", false, "show status in short format")
	statusCmd.Flags().BoolVar(&ignoreEmtpy, "ignore-empty", false, "ignore repos with no changes (most useful with --short)")
}

var statusCmd = &cobra.Command{
	Use:              "status",
	Short:            "Show working tree status across all repos in a directory",
	Long:             `Show working tree status across all repos in a directory.`,
	PersistentPreRun: setupClient,
	RunE:             statusFunc,
}

func statusFunc(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	repoDirs, err := clt.GetDirs(ctx, dir)
	if err != nil {
		cmd.SilenceUsage = true
		return fmt.Errorf("get dirs: %w", err)
	}

	if short {
		args = append(args, "--short")
	}

	err = clt.StatusRepos(ctx, repoDirs, ignoreEmtpy, args...)
	if err != nil {
		cmd.SilenceUsage = true
		return fmt.Errorf("status repos: %w", err)
	}

	return nil
}

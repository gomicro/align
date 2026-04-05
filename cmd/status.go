package cmd

import (
	"context"
	"fmt"

	ctxhelper "github.com/gomicro/align/client/context"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var showBranch bool

func init() {
	RootCmd.AddCommand(statusCmd)

	statusCmd.Flags().StringVar(&dir, "dir", ".", "directory to show status of repos in")
	statusCmd.Flags().BoolVarP(&short, "short", "s", false, "show status in short format")
	statusCmd.Flags().BoolVarP(&showBranch, "branch", "b", false, "show branch and upstream tracking info")
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
	verbose := viper.GetBool("verbose")
	ctx := ctxhelper.WithVerbose(context.Background(), verbose)

	repoDirs, err := clt.GetDirs(ctx, dir)
	if err != nil {
		cmd.SilenceUsage = true
		return fmt.Errorf("get dirs: %w", err)
	}

	if short {
		args = append(args, "--short")
	}

	if showBranch {
		args = append(args, "--branch")
	}

	err = clt.StatusRepos(ctx, repoDirs, ignoreEmtpy, args...)
	if err != nil {
		cmd.SilenceUsage = true
		return fmt.Errorf("status repos: %w", err)
	}

	return nil
}

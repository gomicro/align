package cmd

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
)

var (
	oneline bool
)

func init() {
	RootCmd.AddCommand(logCmd)

	logCmd.Flags().BoolVar(&oneline, "oneline", false, "Show each commit on a single line")
	logCmd.Flags().BoolVar(&ignoreEmtpy, "ignore-empty", false, "Ignore empty repositories")
}

var logCmd = &cobra.Command{
	Use:   "log",
	Short: "Show commit logs for all repos in a directory",
	Long: `Show commit logs all repos in a directory. Since commit hashes would not be the same between
multiple repos this command really only makes sense when used with two branch names or two tags.`,
	PersistentPreRun: setupClient,
	RunE:             logFunc,
}

func logFunc(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	repoDirs, err := clt.GetDirs(ctx, dir)
	if err != nil {
		cmd.SilenceUsage = true
		return fmt.Errorf("get dirs: %w", err)
	}

	if oneline {
		args = append(args, "--oneline")
	}

	// TODO: Add the option to disable color as well
	args = append(args, "--color")

	err = clt.LogRepos(ctx, repoDirs, ignoreEmtpy, args...)
	if err != nil {
		cmd.SilenceUsage = true
		return fmt.Errorf("log repos: %w", err)
	}

	return nil
}

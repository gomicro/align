package cmd

import (
	"context"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

var (
	oneline bool
	noColor bool
)

func init() {
	RootCmd.AddCommand(logCmd)

	logCmd.Flags().BoolVar(&oneline, "oneline", false, "Show each commit on a single line")
	logCmd.Flags().BoolVar(&noColor, "no-color", false, "Disable color output")
	logCmd.Flags().BoolVar(&ignoreEmtpy, "ignore-empty", false, "Ignore empty repositories")
}

var logCmd = &cobra.Command{
	Use:   "log",
	Short: "Show commit logs for all repos in a directory",
	Long: `Show commit logs all repos in a directory. Since commit hashes would not be the same between
multiple repos this command really only makes sense when used with two branch names or two tags.`,
	ValidArgsFunction: logCmdValidArgsFunc,
	PersistentPreRun:  setupClient,
	RunE:              logFunc,
}

func logCmdValidArgsFunc(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	if len(args) >= 1 {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	setupClient(cmd, args)

	ctx := context.Background()

	repoDirs, err := clt.GetDirs(ctx, dir)
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	names, err := clt.GetBranchAndTagNames(ctx, repoDirs)
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	// When the user has typed a range separator, complete only the ref after it
	// and return completions with the already-typed prefix preserved.
	var rangePrefix string
	if idx := strings.LastIndex(toComplete, "..."); idx != -1 {
		rangePrefix = toComplete[:idx+3]
	} else if idx := strings.LastIndex(toComplete, ".."); idx != -1 {
		rangePrefix = toComplete[:idx+2]
	}

	if rangePrefix != "" {
		completions := make([]string, len(names))
		for i, name := range names {
			completions[i] = rangePrefix + name
		}

		return completions, cobra.ShellCompDirectiveNoFileComp
	}

	return names, cobra.ShellCompDirectiveNoFileComp
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

	if !noColor {
		args = append(args, "--color")
	}

	err = clt.LogRepos(ctx, repoDirs, ignoreEmtpy, args...)
	if err != nil {
		cmd.SilenceUsage = true
		return fmt.Errorf("log repos: %w", err)
	}

	return nil
}

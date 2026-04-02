package cmd

import (
	"context"
	"fmt"

	"github.com/gomicro/align/client"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	short            bool
	nameOnly         bool
	ignoreEmtpy      bool
	ignoreFilePrefix []string
	matchExtension   []string
)

func init() {
	RootCmd.AddCommand(diffCmd)

	diffCmd.Flags().StringVar(&dir, "dir", ".", "directory to diff repos from")

	diffCmd.Flags().StringArrayVar(&ignoreFilePrefix, "ignore-file-prefix", []string{}, "ignore files in diffs with the given prefix(es)")
	diffCmd.Flags().StringArrayVar(&matchExtension, "match-extension", []string{}, "only include files in diffs with the given extension(s)")

	diffCmd.Flags().BoolVar(&ignoreEmtpy, "ignore-empty", false, "ignore empty diffs")
	diffCmd.Flags().BoolVar(&noColor, "no-color", false, "disable color output")
	diffCmd.Flags().BoolVar(&short, "shortstat", false, "show only the number of changed files, insertions, and deletions")
	diffCmd.Flags().BoolVar(&nameOnly, "name-only", false, "show only names of changed files")

	diffCmd.MarkFlagsMutuallyExclusive("shortstat", "name-only")
	diffCmd.MarkFlagsMutuallyExclusive("shortstat", "ignore-file-prefix")
}

var diffCmd = &cobra.Command{
	Use:   "diff [flags] [<branch|tag> [<branch|tag>]]",
	Short: "Diff all repos in a directory",
	Long: `Diff all repos in a directory. With no arguments, shows unstaged working tree changes
across all repos, equivalent to a bare 'git diff'. With two arguments, diffs between
two branches or tags. Since commit hashes would not be the same between multiple repos,
two-argument usage really only makes sense with branch names or tags.`,
	Args:              cobra.RangeArgs(0, 2),
	ValidArgsFunction: diffCmdValidArgsFunc,
	PersistentPreRun:  setupClient,
	RunE:              diffFunc,
}

func diffCmdValidArgsFunc(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	if len(args) >= 2 {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	setupClient(cmd, args)

	diffDir, err := cmd.Flags().GetString("dir")
	if err != nil {
		diffDir = "."
	}

	ctx := context.Background()

	repoDirs, err := clt.GetDirs(ctx, diffDir)
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	names, err := clt.GetBranchAndTagNames(ctx, repoDirs)
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	return names, cobra.ShellCompDirectiveNoFileComp
}

func diffFunc(cmd *cobra.Command, args []string) error {
	verbose := viper.GetBool("verbose")
	ctx := client.WithVerbose(context.Background(), verbose)

	repoDirs, err := clt.GetDirs(ctx, dir)
	if err != nil {
		cmd.SilenceUsage = true
		return fmt.Errorf("get dirs: %w", err)
	}

	switch {
	case short:
		args = append(args, "--shortstat")
	case nameOnly:
		args = append(args, "--name-only")
	}

	if !noColor {
		args = append([]string{"--color"}, args...)
	}

	cfg := &client.DiffConfig{
		IgnoreEmpty:      ignoreEmtpy,
		IgnoreFilePrefix: ignoreFilePrefix,
		MatchExtension:   matchExtension,
		Args:             args,
	}

	err = clt.DiffRepos(ctx, repoDirs, cfg)
	if err != nil {
		cmd.SilenceUsage = true
		return fmt.Errorf("diff repos: %w", err)
	}

	return nil
}

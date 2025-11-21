package cmd

import (
	"context"
	"fmt"

	"github.com/gomicro/align/client"
	"github.com/spf13/cobra"
)

var (
	short            bool
	nameOnly         bool
	ignoreEmtpy      bool
	ignoreFilePrefix []string
)

func init() {
	RootCmd.AddCommand(diffCmd)

	diffCmd.Flags().StringVar(&dir, "dir", ".", "directory to diff repos from")

	diffCmd.Flags().BoolVar(&ignoreEmtpy, "ignore-empty", false, "ignore empty diffs")
	diffCmd.Flags().StringArrayVar(&ignoreFilePrefix, "ignore-file-prefix", []string{}, "ignore files in diffs with the given prefix")

	diffCmd.Flags().BoolVar(&short, "shortstat", false, "show only the number of changed files, insertions, and deletions")
	diffCmd.Flags().BoolVar(&nameOnly, "name-only", false, "show only names of changed files")

	diffCmd.MarkFlagsMutuallyExclusive("shortstat", "name-only")
	diffCmd.MarkFlagsMutuallyExclusive("shortstat", "ignore-file-prefix")
}

var diffCmd = &cobra.Command{
	Use:   "diff [flags] <branch|tag> <branch|tag>",
	Short: "Diff all repos in a directory",
	Long: `Diff all repos in a directory. Since commit hashes would not be the same between multiple repos,
this command really only makes sense when used with two branch names or two tags.`,
	Args:             cobra.ExactArgs(2),
	PersistentPreRun: setupClient,
	RunE:             diffFunc,
}

func diffFunc(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

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

	cfg := &client.DiffConfig{
		IgnoreEmpty:      ignoreEmtpy,
		IgnoreFilePrefix: ignoreFilePrefix,
		Args:             args,
	}

	err = clt.DiffRepos(ctx, repoDirs, cfg)
	if err != nil {
		cmd.SilenceUsage = true
		return fmt.Errorf("diff repos: %w", err)
	}

	return nil
}

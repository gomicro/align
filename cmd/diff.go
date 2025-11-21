package cmd

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
)

var (
	short    bool
	nameOnly bool
)

func init() {
	RootCmd.AddCommand(diffCmd)

	diffCmd.Flags().StringVar(&dir, "dir", ".", "directory to diff repos from")
	diffCmd.Flags().BoolVar(&short, "shortstat", false, "show only the number of changed files, insertions, and deletions")
	diffCmd.Flags().BoolVar(&nameOnly, "name-only", false, "show only names of changed files")

	diffCmd.MarkFlagsMutuallyExclusive("shortstat", "name-only")
}

var diffCmd = &cobra.Command{
	Use:              "diff",
	Short:            "Diff all repos in a directory",
	Long:             `Diff all repos in a directory. Since commit hashes would not be the same between multiple repos, this command really only makes sense when used with two branch names or two tags.`,
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

	err = clt.DiffRepos(ctx, repoDirs, args...)
	if err != nil {
		cmd.SilenceUsage = true
		return fmt.Errorf("diff repos: %w", err)
	}

	return nil
}

package cmd

import (
	"context"
	"fmt"

	ctxhelper "github.com/gomicro/align/client/context"
	"github.com/gosuri/uiprogress"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	noFF          bool
	ffOnly        bool
	squash        bool
	abortMerge    bool
	continueMerge bool
)

func init() {
	RootCmd.AddCommand(mergeCmd)

	mergeCmd.Flags().BoolVar(&noFF, "no-ff", false, "create a merge commit even when fast-forward is possible")
	mergeCmd.Flags().BoolVar(&ffOnly, "ff-only", false, "refuse to merge unless the result is a fast-forward")
	mergeCmd.Flags().BoolVar(&squash, "squash", false, "squash commits from the branch into a single commit")
	mergeCmd.Flags().BoolVar(&abortMerge, "abort", false, "abort an in-progress merge")
	mergeCmd.Flags().BoolVar(&continueMerge, "continue", false, "continue an in-progress merge after resolving conflicts")

	mergeCmd.MarkFlagsMutuallyExclusive("squash", "no-ff")
	mergeCmd.MarkFlagsMutuallyExclusive("abort", "squash")
	mergeCmd.MarkFlagsMutuallyExclusive("abort", "no-ff")
	mergeCmd.MarkFlagsMutuallyExclusive("ff-only", "squash", "no-ff", "abort")
	mergeCmd.MarkFlagsMutuallyExclusive("continue", "squash", "no-ff", "ff-only", "abort")
}

var mergeCmd = &cobra.Command{
	Use:               "merge [branch]",
	Short:             "Merge a branch into the current branch across all repos",
	Long:              `Merge a branch into the current branch across all repos in a directory.`,
	Args:              cobra.MaximumNArgs(1),
	ValidArgsFunction: mergeCmdValidArgsFunc,
	PersistentPreRun:  setupClient,
	RunE:              mergeFunc,
}

func mergeCmdValidArgsFunc(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	if len(args) >= 1 {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	setupClient(cmd, args)

	ctx := context.Background()

	repoDirs, err := clt.GetDirs(ctx, ".")
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	names, err := clt.GetBranchNames(ctx, repoDirs)
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	return names, cobra.ShellCompDirectiveNoFileComp
}

func mergeFunc(cmd *cobra.Command, args []string) error {
	verbose := viper.GetBool("verbose")
	ctx := ctxhelper.WithVerbose(context.Background(), verbose)

	if !verbose {
		uiprogress.Start()
		defer uiprogress.Stop()
	}

	repoDirs, err := clt.GetDirs(ctx, ".")
	if err != nil {
		cmd.SilenceUsage = true
		return fmt.Errorf("get dirs: %w", err)
	}

	if abortMerge {
		args = []string{"--abort"}
	} else if continueMerge {
		args = []string{"--continue"}
	} else {
		if noFF {
			args = append(args, "--no-ff")
		}

		if ffOnly {
			args = append(args, "--ff-only")
		}

		if squash {
			args = append(args, "--squash")
		}
	}

	err = clt.MergeRepos(ctx, repoDirs, args...)
	if err != nil {
		cmd.SilenceUsage = true
		return fmt.Errorf("merge repos: %w", err)
	}

	return nil
}

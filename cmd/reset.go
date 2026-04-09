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
	hard  bool
	soft  bool
	mixed bool
)

func init() {
	RootCmd.AddCommand(resetCmd)

	resetCmd.Flags().BoolVar(&hard, "hard", false, "reset index and working tree, discarding all uncommitted changes")
	resetCmd.Flags().BoolVar(&soft, "soft", false, "reset HEAD only, keeping all changes staged")
	resetCmd.Flags().BoolVar(&mixed, "mixed", false, "reset HEAD and index, keeping changes in the working tree (unstaged)")

	resetCmd.MarkFlagsMutuallyExclusive("hard", "soft", "mixed")
}

var resetCmd = &cobra.Command{
	Use:               "reset [ref]",
	Short:             "Reset the current branch across all repos",
	Long:              `Reset the current branch across all repos in a directory. One of --hard, --soft, or --mixed must be provided.`,
	Args:              cobra.MaximumNArgs(1),
	ValidArgsFunction: resetCmdValidArgsFunc,
	PersistentPreRun:  setupClient,
	RunE:              resetFunc,
}

func resetCmdValidArgsFunc(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	if len(args) >= 1 {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	setupClient(cmd, args)

	ctx := context.Background()

	repoDirs, err := clt.GetDirs(ctx, ".")
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	names, err := clt.GetBranchAndTagNames(ctx, repoDirs)
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	return names, cobra.ShellCompDirectiveNoFileComp
}

func resetFunc(cmd *cobra.Command, args []string) error {
	if !hard && !soft && !mixed {
		return fmt.Errorf("one of --hard, --soft, or --mixed is required")
	}

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

	var modeFlag string
	switch {
	case hard:
		modeFlag = "--hard"
	case soft:
		modeFlag = "--soft"
	case mixed:
		modeFlag = "--mixed"
	}

	args = append([]string{modeFlag}, args...)

	err = clt.ResetRepos(ctx, repoDirs, args...)
	if err != nil {
		cmd.SilenceUsage = true
		return fmt.Errorf("reset repos: %w", err)
	}

	return nil
}

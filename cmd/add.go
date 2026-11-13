package cmd

import (
	"fmt"

	ctxhelper "github.com/gomicro/align/client/context"
	"github.com/gomicro/align/client/repos"
	"github.com/gomicro/align/cmd/concurrency"
	"github.com/gomicro/align/cmd/selection"
	"github.com/gosuri/uiprogress"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var update bool

func init() {
	RootCmd.AddCommand(addCmd)
	concurrency.AddFlag(addCmd, repos.DefaultLocalConcurrency)

	addCmd.Flags().BoolVarP(&update, "update", "u", false, "stage modified and deleted files but not new untracked files")
}

var addCmd = &cobra.Command{
	Use:   "add [files...]",
	Short: "Stage changes across all repos in a directory",
	Long: `Stage changes across all repos in a directory. Without arguments, stages all changes
in each repo (equivalent to 'git add -A'). With arguments, stages only the specified files.` + concurrency.Help,
	PersistentPreRun: setupClient,
	RunE:             addFunc,
}

func addFunc(cmd *cobra.Command, args []string) error {
	ctx, stop, err := concurrency.Context(cmd)
	if err != nil {
		cmd.SilenceUsage = true
		return fmt.Errorf("add options: %w", err)
	}

	defer stop()

	verbose := viper.GetBool("verbose")
	ctx = ctxhelper.WithVerbose(ctx, verbose)

	if !verbose {
		uiprogress.Start()
		defer uiprogress.Stop()
	}

	repoDirs, err := selection.Dirs(ctx, cmd, clt, ".")
	if err != nil {
		cmd.SilenceUsage = true
		return fmt.Errorf("get dirs: %w", err)
	}

	if update {
		args = []string{"--update"}
	} else if len(args) == 0 {
		args = []string{"-A"}
	}

	err = clt.StageFiles(ctx, repoDirs, args...)
	if err != nil {
		cmd.SilenceUsage = true
		return fmt.Errorf("stage files: %w", err)
	}

	return nil
}

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
	message    string
	amend      bool
	noEdit     bool
	allowEmpty bool
)

func init() {
	RootCmd.AddCommand(commitCmd)

	commitCmd.Flags().StringVar(&dir, "dir", ".", "directory to commit repos in")
	commitCmd.Flags().StringVarP(&message, "message", "m", "", "commit message")
	commitCmd.Flags().BoolVarP(&all, "all", "a", false, "stage all tracked modified and deleted files before committing")
	commitCmd.Flags().BoolVar(&amend, "amend", false, "amend the last commit")
	commitCmd.Flags().BoolVar(&noEdit, "no-edit", false, "use the existing commit message when amending")
	commitCmd.Flags().BoolVar(&allowEmpty, "allow-empty", false, "allow a commit with no staged changes (useful for triggering CI)")

	commitCmd.MarkFlagRequired("message") //nolint:errcheck
}

var commitCmd = &cobra.Command{
	Use:              "commit",
	Short:            "Commit staged changes across all repos in a directory",
	Long:             `Commit staged changes across all repos in a directory.`,
	PersistentPreRun: setupClient,
	RunE:             commitFunc,
}

func commitFunc(cmd *cobra.Command, args []string) error {
	verbose := viper.GetBool("verbose")
	ctx := ctxhelper.WithVerbose(context.Background(), verbose)

	if !verbose {
		uiprogress.Start()
		defer uiprogress.Stop()
	}

	repoDirs, err := clt.GetDirs(ctx, dir)
	if err != nil {
		cmd.SilenceUsage = true
		return fmt.Errorf("get dirs: %w", err)
	}

	if all {
		args = append(args, "--all")
	}

	if allowEmpty {
		args = append(args, "--allow-empty")
	}

	args = append(args, "-m", message)

	if amend {
		args = append(args, "--amend")
	}

	if noEdit {
		args = append(args, "--no-edit")
	}

	err = clt.CommitRepos(ctx, repoDirs, args...)
	if err != nil {
		cmd.SilenceUsage = true
		return fmt.Errorf("commit repos: %w", err)
	}

	return nil
}

package cmd

import (
	"context"
	"fmt"
	"os"
	"os/exec"

	"github.com/gomicro/align/client"
	cfgCmd "github.com/gomicro/align/cmd/config"
	"github.com/gomicro/align/cmd/remote"
	"github.com/gomicro/align/cmd/selection"
	"github.com/gomicro/align/cmd/stash"
	"github.com/gomicro/align/config"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	clt client.Clienter
)

func init() {
	cobra.OnInitialize(initEnvs)

	RootCmd.AddCommand(cfgCmd.ConfigCmd)
	RootCmd.AddCommand(remote.RemoteCmd)
	RootCmd.AddCommand(stash.StashCmd)

	RootCmd.PersistentFlags().BoolP("verbose", "v", false, "show more verbose output")
	RootCmd.PersistentFlags().Bool("no-color", false, "disable color in verbose output")
	RootCmd.PersistentFlags().StringArray("ignore-repo", nil, "exclude an exact local repository directory name (repeatable)")
	err := RootCmd.RegisterFlagCompletionFunc("ignore-repo", selection.Complete)
	if err != nil {
		fmt.Printf("Error setting up: %s\n", err)
		os.Exit(1)
	}

	err = viper.BindPFlag("verbose", RootCmd.PersistentFlags().Lookup("verbose"))
	if err != nil {
		fmt.Printf("Error setting up: %s\n", err)
		os.Exit(1)
	}

	err = viper.BindPFlag("no-color", RootCmd.PersistentFlags().Lookup("no-color"))
	if err != nil {
		fmt.Printf("Error setting up: %s\n", err)
		os.Exit(1)
	}
}

func initEnvs() {
}

var RootCmd = &cobra.Command{
	Use:   "align [flags]",
	Short: "Tool for managing git repos together",
	Long: `align fans out git operations across every git repository in the current directory.

Each sub-command mirrors its git equivalent and runs it against all repos found in the working directory.

Use --verbose (-v) when scripting or using align with AI tools. Verbose mode replaces the
interactive progress bars with structured per-repo output written to stdout, which is easier
to parse and redirect.

` + selectionHelp + "\n\n" + aliasHelp,
	TraverseChildren: true,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		err := selection.Reject(cmd, args)
		if err != nil {
			return err
		}

		if _, err := exec.LookPath("git"); err != nil {
			return fmt.Errorf("git is not installed or not on PATH — it is required for align to function")
		}

		return nil
	},
}

const selectionHelp = `Exclude local repositories for this invocation with --ignore-repo:
  align checkout main --ignore-repo userauth
  align --ignore-repo userauth --ignore-repo other cleanup

Repeat the option for multiple exact, case-sensitive directory names. Paths and
wildcard patterns are not interpreted; commas are part of a name. Unknown names
and exclusions that leave no repositories selected fail before Git execution.
For aliases, put this option BEFORE the alias name; subsequent tokens belong to
Git. Ignored repos do not contribute output or progress totals. Nothing is saved.
This option is unsupported for clone, auth, config, version, and completion
generation.`

// Execute registers aliases when needed and runs the command selected by os.Args.
func Execute() error {
	err := prepareAliases(context.Background(), RootCmd, os.Args[1:])
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		if isCompletionRequest(os.Args[1:]) {
			fmt.Fprintf(os.Stdout, ":%d\n", cobra.ShellCompDirectiveError)
		}

		return fmt.Errorf("preparing command: %w", err)
	}

	err = RootCmd.Execute()
	if err != nil {
		return fmt.Errorf("executing command: %w", err)
	}

	return nil
}

func setupClient(cmd *cobra.Command, args []string) {
	c, err := config.ParseFromFile()
	if err != nil {
		fmt.Printf("Error: %s\n", err)
		os.Exit(1)
	}

	var opts []client.Option
	if viper.GetBool("no-color") {
		opts = append(opts, client.WithNoColor())
	}

	clt, err = client.New(c, opts...)
	if err != nil {
		fmt.Printf("Error: %s\n", err)
		os.Exit(1)
	}
}

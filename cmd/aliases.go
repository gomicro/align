package cmd

import (
	"context"
	"fmt"

	"github.com/gomicro/align/client"
	ctxhelper "github.com/gomicro/align/client/context"

	"github.com/gosuri/uiprogress"
	"github.com/spf13/cobra"
)

const aliasHelp = `Git aliases from global and system configuration can run across all repos.

Use align [align flags] <alias> [alias arguments], for example:
  align cleanup
  align -v lg1
  align -v <alias> --help

Put align options (-v, --verbose, --no-color) BEFORE the alias name.
All arguments after the name go unchanged to Git. To view align's help for an
alias, use align help <alias>; <alias> --help is passed to Git.
Built-in align commands always take precedence over aliases.

Git expands ordinary and shell (!) aliases using each repository's effective
configuration, including local overrides. Only global/system alias names are
discovered and completed; local-only aliases are not exposed. Discovery uses
the invocation directory, so conditional includes may differ in target repos.
Git's own command/alias precedence also applies. Alias discovery requires
Git 2.26 or newer.

Normal mode shows progress and suppresses successful output. Verbose mode shows
buffered stdout and stderr grouped by repo, not a live terminal. Every repo is
attempted, and any failure produces a nonzero exit status in either mode.

Aliases are not checked for safety or useful output. They can delete data or run
arbitrary shell commands. Use only trusted aliases suitable for every target
repo. Graphs and large output may be unreadable; interactive input, editors,
terminal UIs, and pagers are not supported. Shell aliases must work with Git's
shell and have their required programs available.

Fetch/pull pruning removes stale remote-tracking refs; it does not replace
deleting merged local branches.`

func prepareAliases(ctx context.Context, root *cobra.Command, args []string) error {
	root.InitDefaultHelpCmd()
	target, remaining, err := root.Find(args)
	aliasHelpRequest := err == nil && target.Name() == "help" && len(remaining) > 0 && !aliasReserved(root, remaining[0])
	if err == nil && !aliasHelpRequest {
		return nil
	}

	names, err := client.AliasNames(ctx)
	if err != nil {
		return fmt.Errorf("discovering aliases: %w", err)
	}

	for _, name := range names {
		if aliasReserved(root, name) {
			continue
		}

		root.AddCommand(newAliasCommand(root, name))
	}

	if isCompletionRequest(args) {
		return nil
	}

	target, remaining, err = root.Find(args)
	if err != nil {
		return fmt.Errorf("resolving command: %w", err)
	}

	if target == root && len(remaining) > 0 {
		return fmt.Errorf("unknown command %q for %q", remaining[0], root.Name())
	}

	return nil
}

func isCompletionRequest(args []string) bool {
	if len(args) == 0 {
		return false
	}

	return args[0] == cobra.ShellCompRequestCmd || args[0] == cobra.ShellCompNoDescRequestCmd
}

func aliasReserved(root *cobra.Command, name string) bool {
	switch name {
	case "help", cobra.ShellCompRequestCmd, cobra.ShellCompNoDescRequestCmd:
		return true
	}

	for _, command := range root.Commands() {
		if command.Name() == name || command.HasAlias(name) {
			return true
		}
	}

	return false
}

func newAliasCommand(root *cobra.Command, name string) *cobra.Command {
	command := &cobra.Command{
		Use:                name + " [alias arguments]",
		Short:              "Run Git alias " + name + " across all repos",
		Long:               aliasHelp,
		DisableFlagParsing: true,
		PersistentPreRun:   setupClient,
		RunE:               aliasFunc,
		ValidArgsFunction:  aliasArgs,
	}

	// Cobra completes inherited flags even when flag parsing is disabled.
	for _, name := range []string{"verbose", "no-color"} {
		flag := root.PersistentFlags().Lookup(name)
		if flag == nil {
			continue
		}

		shadow := *flag
		shadow.Hidden = true
		command.Flags().AddFlag(&shadow)
	}

	command.Flags().BoolP("help", "h", false, "")
	command.Flags().Lookup("help").Hidden = true
	return command
}

func aliasArgs(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	return nil, cobra.ShellCompDirectiveNoFileComp
}

func aliasFunc(cmd *cobra.Command, args []string) error {
	verbose, err := cmd.Root().PersistentFlags().GetBool("verbose")
	if err != nil {
		return fmt.Errorf("reading verbosity: %w", err)
	}

	ctx := ctxhelper.WithVerbose(cmd.Context(), verbose)
	dirs, err := clt.GetDirs(ctx, ".")
	if err != nil {
		cmd.SilenceUsage = true
		return fmt.Errorf("discovering repositories: %w", err)
	}

	if len(dirs) == 0 {
		cmd.SilenceUsage = true
		return fmt.Errorf("no repositories found in the current directory")
	}

	if !verbose {
		uiprogress.Start()
		defer uiprogress.Stop()
	}

	err = clt.RunAlias(ctx, dirs, cmd.Name(), args...)
	if err != nil {
		cmd.SilenceUsage = true
		return fmt.Errorf("running alias %s: %w", cmd.Name(), err)
	}

	return nil
}

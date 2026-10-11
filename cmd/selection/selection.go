// Package selection connects invocation flags to local repository discovery.
package selection

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/gomicro/align/client"
	ctxhelper "github.com/gomicro/align/client/context"

	"github.com/spf13/cobra"
)

// Dirs applies parsed exclusions while preserving the caller's context state.
func Dirs(ctx context.Context, cmd *cobra.Command, clt client.Clienter, base string) ([]string, error) {
	flags := cmd.Flags()
	flags.AddFlagSet(cmd.InheritedFlags())
	if flags.Lookup("ignore-repo") == nil {
		dirs, err := clt.GetDirs(ctx, base)
		if err != nil {
			return nil, fmt.Errorf("discovering repositories: %w", err)
		}

		return dirs, nil
	}

	names, err := flags.GetStringArray("ignore-repo")
	if err != nil {
		return nil, fmt.Errorf("reading repository exclusions: %w", err)
	}

	changed := flags.Lookup("ignore-repo").Changed
	rootFlag := cmd.Root().PersistentFlags().Lookup("ignore-repo")
	if rootFlag != nil {
		changed = changed || rootFlag.Changed
	}

	if changed && len(names) == 0 {
		return nil, fmt.Errorf("ignored repository not found: empty name")
	}

	ctx = ctxhelper.WithIgnoredRepos(ctx, names)
	dirs, err := clt.GetDirs(ctx, base)
	if err != nil {
		return nil, fmt.Errorf("selecting repositories: %w", err)
	}

	return dirs, nil
}

// Reject prevents unsupported commands from silently accepting exclusions.
func Reject(cmd *cobra.Command, args []string) error {
	flags := cmd.Flags()
	flags.AddFlagSet(cmd.InheritedFlags())
	flag := flags.Lookup("ignore-repo")
	if flag != nil && flag.Changed {
		return fmt.Errorf("--ignore-repo is not supported by %s", cmd.CommandPath())
	}

	return nil
}

// Complete lists raw repo names without applying partially typed exclusions.
func Complete(cmd *cobra.Command, args []string, prefix string) ([]string, cobra.ShellCompDirective) {
	clt := &client.Client{}
	ctx := ctxhelper.WithIgnoredRepos(cmd.Context(), nil)
	dirs, err := clt.GetDirs(ctx, ".")
	if err != nil {
		cobra.CompErrorln(err.Error())
		return nil, cobra.ShellCompDirectiveError
	}

	names := make([]string, 0, len(dirs))
	for _, dir := range dirs {
		name := filepath.Base(dir)
		if strings.HasPrefix(name, prefix) {
			names = append(names, name)
		}
	}

	return names, cobra.ShellCompDirectiveNoFileComp
}

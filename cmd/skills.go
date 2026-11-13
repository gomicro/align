package cmd

import (
	"fmt"

	"github.com/gomicro/align/cmd/selection"
	"github.com/gomicro/align/skills"

	"github.com/spf13/cobra"
)

func init() {
	RootCmd.AddCommand(newSkillsCommand())
}

func newSkillsCommand() *cobra.Command {
	command := &cobra.Command{
		Use:               "skills",
		Short:             "Export portable agent skills bundled with align",
		Long:              "Export bundled guidance for agents using align. This built-in takes precedence over a Git alias named skills.",
		Args:              cobra.NoArgs,
		PersistentPreRunE: selection.Reject,
		RunE:              skillsHelp,
	}

	dump := &cobra.Command{
		Use:   "dump --output-dir <dir>",
		Short: "Write bundled skills to an explicit destination",
		Long: `Write the bundled skill to <output-dir>/align/SKILL.md.

The destination is required; relative paths use the current directory. Identical
content is left unchanged. Differing content is refused unless --force is given.
Symlink and non-regular skill entries are refused, even with --force. Unrelated
files are preserved.

Export works offline without Git, align configuration, credentials, or repositories.
It does not install or register skills with any agent. Choose a destination your
agent supports and follow that agent's skill discovery instructions.`,
		Args: cobra.NoArgs,
		RunE: skillsDump,
	}
	dump.Flags().String("output-dir", "", "destination directory (required)")
	dump.Flags().Bool("force", false, "replace differing exported skill content")
	command.AddCommand(dump)
	return command
}

func skillsHelp(cmd *cobra.Command, args []string) error {
	err := cmd.Help()
	if err != nil {
		return fmt.Errorf("showing skills help: %w", err)
	}

	return nil
}

func skillsDump(cmd *cobra.Command, args []string) error {
	outputDir, err := cmd.Flags().GetString("output-dir")
	if err != nil {
		return fmt.Errorf("reading output directory: %w", err)
	}

	force, err := cmd.Flags().GetBool("force")
	if err != nil {
		return fmt.Errorf("reading replacement option: %w", err)
	}

	path, err := skills.Dump(cmd.Context(), outputDir, force)
	if err != nil {
		cmd.SilenceUsage = true
		return fmt.Errorf("dumping skills: %w", err)
	}

	n, err := fmt.Fprintf(cmd.OutOrStdout(), "Skill available at %s\n", path)
	if err != nil {
		return fmt.Errorf("reporting skill destination after %d bytes: %w", n, err)
	}

	return nil
}

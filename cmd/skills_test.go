package cmd

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func runSkillsCommand(t *testing.T, args ...string) (string, error) {
	t.Helper()
	root := &cobra.Command{
		Use: "align",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			return fmt.Errorf("git setup must not run")
		},
	}
	root.PersistentFlags().StringArray("ignore-repo", nil, "")
	root.PersistentFlags().BoolP("verbose", "v", false, "")
	root.PersistentFlags().Bool("no-color", false, "")
	root.AddCommand(newSkillsCommand())
	var output bytes.Buffer
	root.SetOut(&output)
	root.SetErr(&output)
	root.SetArgs(args)
	err := root.Execute()
	return output.String(), err
}

func TestSkillsDump(t *testing.T) {
	t.Parallel()
	outputDir := filepath.Join(t.TempDir(), "export")
	out, err := runSkillsCommand(t, "-v", "--no-color", "skills", "dump", "--output-dir", outputDir)
	if !assert.NoError(t, err, out) {
		return
	}

	path := filepath.Join(outputDir, "align", "SKILL.md")
	source, err := os.ReadFile("../skills/align/SKILL.md")
	assert.NoError(t, err)
	exported, err := os.ReadFile(path)
	assert.NoError(t, err)
	assert.Equal(t, source, exported)
	assert.Contains(t, out, path)

	err = os.WriteFile(path, []byte("customized"), 0644)
	assert.NoError(t, err)
	out, err = runSkillsCommand(t, "skills", "dump", "--output-dir", outputDir)
	assert.ErrorContains(t, err, "--force", out)
	out, err = runSkillsCommand(t, "skills", "dump", "--output-dir", outputDir, "--force")
	assert.NoError(t, err, out)
	exported, err = os.ReadFile(path)
	assert.NoError(t, err)
	assert.Equal(t, source, exported)
}

func TestSkillsNoWrites(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name      string
		args      []string
		outputDir bool
		wantError bool
	}{
		{name: "group help", args: []string{"skills"}},
		{name: "explicit group help", args: []string{"skills", "--help"}},
		{name: "dump help", args: []string{"skills", "dump", "--help"}, outputDir: true},
		{name: "missing destination", args: []string{"skills", "dump"}, wantError: true},
		{name: "empty destination", args: []string{"skills", "dump", "--output-dir="}, wantError: true},
		{name: "blank destination", args: []string{"skills", "dump", "--output-dir", " "}, wantError: true},
		{name: "unexpected argument", args: []string{"skills", "dump", "extra"}, outputDir: true, wantError: true},
		{name: "unknown flag", args: []string{"skills", "dump", "--unknown"}, outputDir: true, wantError: true},
		{name: "unknown subcommand", args: []string{"skills", "unknown"}, wantError: true},
		{name: "excluded group", args: []string{"--ignore-repo", "repo", "skills"}, wantError: true},
		{name: "excluded dump", args: []string{"skills", "dump", "--ignore-repo", "repo"}, outputDir: true, wantError: true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			base := t.TempDir()
			outputDir := filepath.Join(base, "absent")
			args := slices.Clone(tc.args)
			if tc.outputDir {
				args = append(args, "--output-dir", outputDir)
			}

			out, err := runSkillsCommand(t, args...)
			if tc.wantError {
				assert.Error(t, err, out)
			} else {
				assert.NoError(t, err, out)
				assert.Contains(t, out, "dump")
			}

			entries, err := os.ReadDir(base)
			assert.NoError(t, err)
			assert.Empty(t, entries)
		})
	}
}

func TestSkillsCLIProcess(t *testing.T) {
	t.Parallel()
	if os.Getenv("ALIGN_SKILLS_TEST_PROCESS") != "1" {
		return
	}

	index := slices.Index(os.Args, "--")
	if index < 0 {
		os.Exit(1)
	}

	os.Args = append([]string{os.Args[0]}, os.Args[index+1:]...)
	err := Execute()
	if err != nil {
		os.Exit(1)
	}

	os.Exit(0)
}

func runSkillsCLI(t *testing.T, f aliasFixture, args ...string) (string, string, error) {
	t.Helper()
	testArgs := append([]string{"-test.run=^TestSkillsCLIProcess$", "--"}, args...)
	cmd := exec.Command(os.Args[0], testArgs...)
	cmd.Dir = f.base
	cmd.Env = append(slices.Clone(f.env), "ALIGN_SKILLS_TEST_PROCESS=1")
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	return stdout.String(), stderr.String(), err
}

func TestSkillsWithoutGit(t *testing.T) {
	t.Parallel()
	f := newAliasFixture(t)
	f.env = append(f.env, "PATH="+t.TempDir())
	out, stderr, err := runSkillsCLI(t, f, "skills", "dump", "--output-dir", "exported")
	assert.NoError(t, err, stderr)
	assert.Contains(t, out, filepath.Join("exported", "align", "SKILL.md"))
	got, err := os.ReadFile(filepath.Join(f.base, "exported", "align", "SKILL.md"))
	assert.NoError(t, err)
	source, err := os.ReadFile("../skills/align/SKILL.md")
	assert.NoError(t, err)
	assert.Equal(t, source, got)

	for _, args := range [][]string{{"skills"}, {"skills", "--help"}, {"skills", "dump", "--help"}, {"help", "skills", "dump"}} {
		out, stderr, err = runSkillsCLI(t, f, args...)
		assert.NoError(t, err, stderr)
		assert.Contains(t, out, "dump")
	}

	out, stderr, err = runSkillsCLI(t, f, "--ignore-repo", "repo", "skills", "dump", "--output-dir", "absent")
	assert.Error(t, err, out)
	assert.Contains(t, stderr, "not supported")
	info, err := os.Stat(filepath.Join(f.base, "absent"))
	assert.Nil(t, info)
	assert.True(t, os.IsNotExist(err))
}

func TestSkillsAliasPrecedence(t *testing.T) {
	t.Parallel()
	f := newAliasFixture(t)
	dir := f.repo(t, "repo")
	f.alias(t, "skills", "!printf alias > marker")
	out, stderr, err := runSkillsCLI(t, f, "skills", "dump", "--output-dir", "exported")
	assert.NoError(t, err, stderr)
	assert.Contains(t, out, "Skill available")
	info, err := os.Stat(filepath.Join(dir, "marker"))
	assert.Nil(t, info)
	assert.True(t, os.IsNotExist(err))
}

type skillsFailedWriter struct{}

func (skillsFailedWriter) Write(data []byte) (int, error) {
	return 0, fmt.Errorf("writing output: %w", io.ErrClosedPipe)
}

func TestSkillsOutputFailure(t *testing.T) {
	t.Parallel()
	command := newSkillsCommand()
	command.SetOut(skillsFailedWriter{})
	command.SetErr(io.Discard)
	command.SetArgs([]string{"dump", "--output-dir", t.TempDir()})
	err := command.Execute()
	assert.ErrorIs(t, err, io.ErrClosedPipe)
}

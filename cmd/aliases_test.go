package cmd

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/gomicro/align/client"
	"github.com/gomicro/align/client/repos"

	"github.com/gomicro/scribe"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

// Subprocesses isolate Cobra/Viper globals, Git configuration, and progress bars.
func TestAliasCLIProcess(t *testing.T) {
	t.Parallel()
	if os.Getenv("ALIGN_ALIAS_TEST_PROCESS") != "1" {
		return
	}

	index := slices.Index(os.Args, "--")
	if index < 0 {
		fmt.Fprintln(os.Stderr, "missing subprocess arguments")
		os.Exit(1)
	}

	args := os.Args[index+1:]
	err := prepareAliases(context.Background(), RootCmd, args)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	scrb, err := scribe.NewScribe(os.Stdout, scribe.DefaultTheme)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	clt = &client.Client{Repos: repos.New(scrb, nil, nil)}
	for _, command := range RootCmd.Commands() {
		if command.DisableFlagParsing {
			command.PersistentPreRun = nil
		}
	}

	RootCmd.SetArgs(args)
	err = RootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}

	os.Exit(0)
}

type aliasFixture struct {
	base   string
	env    []string
	global string
	system string
}

func newAliasFixture(t *testing.T) aliasFixture {
	t.Helper()
	base := t.TempDir()
	global := filepath.Join(base, "global.config")
	system := filepath.Join(base, "system.config")
	env := make([]string, 0, len(os.Environ()))
	for _, entry := range os.Environ() {
		key, value, found := strings.Cut(entry, "=")
		if !found || strings.HasPrefix(key, "GIT_") || key == "HOME" || key == "XDG_CONFIG_HOME" || key == "ALIGN_ALIAS_TEST_PROCESS" {
			continue
		}

		env = append(env, key+"="+value)
	}

	env = append(env, "HOME="+base, "XDG_CONFIG_HOME="+base, "GIT_CONFIG_GLOBAL="+global, "GIT_CONFIG_SYSTEM="+system, "ALIGN_ALIAS_TEST_PROCESS=1")
	return aliasFixture{base: base, env: env, global: global, system: system}
}

func (f aliasFixture) git(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	cmd.Env = f.env
	out, err := cmd.CombinedOutput()
	assert.NoError(t, err, "%s", out)
}

func (f aliasFixture) alias(t *testing.T, name, value string) {
	t.Helper()
	f.git(t, f.base, "config", "--file", f.global, "alias."+name, value)
}

func (f aliasFixture) repo(t *testing.T, name string) string {
	t.Helper()
	dir := filepath.Join(f.base, name)
	err := os.Mkdir(dir, 0700)
	if !assert.NoError(t, err) {
		return dir
	}

	f.git(t, dir, "init", "--quiet", "--initial-branch=main")
	return dir
}

func (f aliasFixture) run(t *testing.T, args ...string) (string, string, error) {
	t.Helper()
	testArgs := append([]string{"-test.run=^TestAliasCLIProcess$", "--"}, args...)
	cmd := exec.Command(os.Args[0], testArgs...)
	cmd.Dir = f.base
	cmd.Env = f.env
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	return stdout.String(), stderr.String(), err
}

func TestAliasArgumentBoundary(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name    string
		args    []string
		verbose bool
		want    []string
	}{
		{name: "root short flag", args: []string{"-v", "sample", "-v", "space value", "--help", "--no-color", "--"}, verbose: true, want: []string{"-v", "space value", "--help", "--no-color", "--"}},
		{name: "root long flags", args: []string{"--verbose", "--no-color", "sample", "--unknown"}, verbose: true, want: []string{"--unknown"}},
		{name: "alias flag only", args: []string{"sample", "-v"}, want: []string{"-v"}},
		{name: "literal separator", args: []string{"sample", "--", "-v"}, want: []string{"--", "-v"}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			root := &cobra.Command{Use: "align", TraverseChildren: true}
			root.PersistentFlags().BoolP("verbose", "v", false, "")
			root.PersistentFlags().Bool("no-color", false, "")
			command := newAliasCommand(root, "sample")
			command.PersistentPreRun = nil
			var got []string
			command.RunE = func(cmd *cobra.Command, args []string) error {
				got = args
				return nil
			}
			root.AddCommand(command)
			root.SetArgs(tc.args)

			err := root.Execute()
			assert.NoError(t, err)
			assert.Equal(t, tc.want, got)
			verbose, err := root.PersistentFlags().GetBool("verbose")
			assert.NoError(t, err)
			assert.Equal(t, tc.verbose, verbose)
		})
	}
}

func TestAliasesExecute(t *testing.T) {
	t.Parallel()
	f := newAliasFixture(t)
	first := f.repo(t, "a")
	second := f.repo(t, "b")
	f.alias(t, "echoargs", `!f() { printf 'ARG=<%s>\n' "$@"; }; f`)
	f.alias(t, "streams", `!printf 'stdout-marker\n'; printf 'stderr-marker\n' >&2`)
	f.alias(t, "state", "status --short")
	f.alias(t, "touchhelp", "!printf executed > root-help-executed")
	f.alias(t, "mixed", `!if test "$(basename "$PWD")" = a; then printf 'failure-stdout\n'; printf 'failure-stderr\n' >&2; exit 7; fi; printf 'later-stdout\n'; printf done > continued`)

	out, stderr, err := f.run(t, "-v", "--no-color", "echoargs", "space value", "-v", "--help", "--no-color", "--")
	assert.NoError(t, err, stderr)
	assert.Equal(t, 2, strings.Count(out, "ARG=<space value>"))
	assert.Equal(t, 2, strings.Count(out, "ARG=<-v>"))
	assert.Equal(t, 2, strings.Count(out, "ARG=<--help>"))
	assert.Equal(t, 2, strings.Count(out, "ARG=<--no-color>"))
	assert.Equal(t, 2, strings.Count(out, "ARG=<-->"))

	out, stderr, err = f.run(t, "--help", "touchhelp")
	assert.NoError(t, err, stderr)
	assert.Contains(t, out, "Git aliases")
	helpInfo, helpErr := os.Stat(filepath.Join(first, "root-help-executed"))
	assert.Nil(t, helpInfo)
	assert.True(t, os.IsNotExist(helpErr))

	out, stderr, err = f.run(t, "-v", "streams")
	assert.NoError(t, err, stderr)
	assert.Equal(t, 2, strings.Count(out, "stdout-marker\n"))
	assert.Equal(t, 2, strings.Count(out, "stderr-marker\n"))
	assert.NotContains(t, out, "%]")

	out, stderr, err = f.run(t, "streams")
	assert.NoError(t, err, stderr)
	assert.NotContains(t, out, "stdout-marker\n")
	assert.NotContains(t, out, "stderr-marker\n")
	assert.Contains(t, out, "streams")

	err = os.WriteFile(filepath.Join(first, "untracked"), []byte("fixture"), 0600)
	assert.NoError(t, err)
	out, stderr, err = f.run(t, "-v", "state")
	assert.NoError(t, err, stderr)
	assert.Contains(t, out, "?? untracked")

	f.git(t, second, "config", "alias.streams", "!printf 'local-override\\n'")
	out, stderr, err = f.run(t, "-v", "streams")
	assert.NoError(t, err, stderr)
	assert.Contains(t, out, "local-override")
	assert.Equal(t, 1, strings.Count(out, "stdout-marker\n"))

	for _, flags := range [][]string{{"-v", "mixed"}, {"mixed"}} {
		out, stderr, err = f.run(t, flags...)
		assert.Error(t, err)
		assert.Contains(t, stderr, "failure-stderr")
		if flags[0] == "-v" {
			assert.Contains(t, out, "failure-stdout\n")
			assert.Contains(t, out, "failure-stderr\n")
			assert.Contains(t, out, "later-stdout\n")
		}

		data, readErr := os.ReadFile(filepath.Join(second, "continued"))
		assert.NoError(t, readErr)
		assert.Equal(t, "done", string(data))
		removeErr := os.Remove(filepath.Join(second, "continued"))
		assert.NoError(t, removeErr)
	}
}

func TestAliasCompletionAndDiscovery(t *testing.T) {
	t.Parallel()
	f := newAliasFixture(t)
	f.repo(t, "a")
	f.alias(t, "zglobal", "!printf executed")
	f.alias(t, "shared", "status")
	f.alias(t, "branch", "!printf collision")
	f.alias(t, "help", "!printf collision")
	f.git(t, f.base, "config", "--file", f.system, "alias.asystem", "status")
	f.git(t, f.base, "config", "--file", f.system, "alias.shared", "log")
	included := filepath.Join(f.base, "included.config")
	f.git(t, f.base, "config", "--file", included, "alias.included", "status")
	f.git(t, f.base, "config", "--file", f.global, "include.path", included)

	out, stderr, err := f.run(t, "__complete", "")
	assert.NoError(t, err, stderr)
	assert.Contains(t, out, "asystem\t")
	assert.Contains(t, out, "included\t")
	assert.Contains(t, out, "zglobal\t")
	assert.Equal(t, 1, strings.Count(out, "shared\t"))
	assert.NotContains(t, out, "Run Git alias branch")
	assert.NotContains(t, out, "Run Git alias help")
	assert.NotContains(t, out, "Run Git alias __complete")
	assert.NotContains(t, out, "executed")
	assert.Less(t, strings.Index(out, "asystem\t"), strings.Index(out, "zglobal\t"))

	out, stderr, err = f.run(t, "__complete", "zg")
	assert.NoError(t, err, stderr)
	assert.Contains(t, out, "zglobal\t")
	assert.NotContains(t, out, "asystem\t")

	out, stderr, err = f.run(t, "__complete", "zglobal", "")
	assert.NoError(t, err, stderr)
	assert.Equal(t, ":4\n", out)
	out, stderr, err = f.run(t, "__complete", "zglobal", "--v")
	assert.NoError(t, err, stderr)
	assert.Equal(t, ":4\n", out)

	f.alias(t, "added", "status")
	out, stderr, err = f.run(t, "__complete", "add")
	assert.NoError(t, err, stderr)
	assert.Contains(t, out, "added\t")

	f.env = append(f.env, "GIT_CONFIG_NOSYSTEM=1")
	out, stderr, err = f.run(t, "__complete", "")
	assert.NoError(t, err, stderr)
	assert.NotContains(t, out, "asystem\t")

	out, stderr, err = f.run(t, "help", "branch")
	assert.NoError(t, err, stderr)
	assert.Contains(t, out, "List, create, delete")
	out, stderr, err = f.run(t, "help", "zglobal")
	assert.NoError(t, err, stderr)
	assert.Contains(t, out, "BEFORE")
}

func TestAliasDiscoveryErrors(t *testing.T) {
	t.Parallel()
	f := newAliasFixture(t)
	dir := f.repo(t, "a")
	f.git(t, dir, "config", "alias.localonly", "status")
	f.git(t, f.base, "init", "--quiet")
	f.git(t, f.base, "config", "alias.parentlocal", "status")

	out, stderr, err := f.run(t, "__complete", "")
	assert.NoError(t, err, stderr)
	assert.NotContains(t, out, "localonly")
	assert.NotContains(t, out, "parentlocal")
	out, stderr, err = f.run(t, "localonly")
	assert.Error(t, err)
	assert.Empty(t, out)
	assert.Contains(t, stderr, "unknown command")

	f.alias(t, "empty", "status")
	emptyBase := filepath.Join(f.base, "empty-directory")
	err = os.Mkdir(emptyBase, 0700)
	assert.NoError(t, err)
	emptyFixture := f
	emptyFixture.base = emptyBase
	out, stderr, err = emptyFixture.run(t, "empty")
	assert.Error(t, err)
	assert.Empty(t, out)
	assert.Contains(t, stderr, "no repositories found")

	err = os.WriteFile(f.global, []byte("[broken\n"), 0600)
	assert.NoError(t, err)
	for _, args := range [][]string{{"empty"}, {"__complete", ""}} {
		out, stderr, err = f.run(t, args...)
		assert.Error(t, err)
		assert.Empty(t, out)
		assert.Contains(t, stderr, "reading git aliases")
		assert.Contains(t, stderr, "bad config")
	}

	for _, args := range [][]string{{"--help"}, {"help", "branch"}, {"version"}} {
		out, stderr, err = f.run(t, args...)
		assert.NoError(t, err, stderr)
		assert.NotEmpty(t, out)
		assert.NotContains(t, stderr, "reading git aliases")
	}
}

func TestAliasCleanupAndGraph(t *testing.T) {
	t.Parallel()
	f := newAliasFixture(t)
	dir := f.repo(t, "a")
	f.git(t, dir, "-c", "user.name=Fixture", "-c", "user.email=fixture@example.invalid", "-c", "commit.gpgsign=false", "commit", "--quiet", "--allow-empty", "-m", "fixture-commit")
	f.git(t, dir, "branch", "merged")
	f.alias(t, "cleanup", `!git branch --format='%(refname:short)' --merged | grep -v '^main$' | xargs git branch -d`)
	f.alias(t, "graph", "log --graph --oneline --all -1")

	out, stderr, err := f.run(t, "cleanup")
	assert.NoError(t, err, stderr)
	assert.Contains(t, out, "cleanup")
	assert.NotContains(t, out, "Deleted branch")
	cmd := exec.Command("git", "show-ref", "--verify", "refs/heads/merged")
	cmd.Dir = dir
	cmd.Env = f.env
	err = cmd.Run()
	assert.Error(t, err)
	f.git(t, dir, "show-ref", "--verify", "refs/heads/main")

	out, stderr, err = f.run(t, "-v", "graph")
	assert.NoError(t, err, stderr)
	assert.Contains(t, out, "* ")
	assert.Contains(t, out, "fixture-commit")
}

func TestAliasCompletionScripts(t *testing.T) {
	t.Parallel()
	for _, shellName := range []string{"bash", "zsh", "fish", "powershell"} {
		t.Run(shellName, func(t *testing.T) {
			t.Parallel()
			f := newAliasFixture(t)
			out, stderr, err := f.run(t, "completion", "--shell", shellName)
			assert.NoError(t, err, stderr)
			assert.Contains(t, out, "__complete")
			assert.NotContains(t, out, "Run Git alias")
		})
	}
}

func TestAliasGeneratedShellCompletion(t *testing.T) {
	t.Parallel()
	cases := []struct {
		shell  string
		prefix string
		invoke string
	}{
		{
			shell:  "bash",
			prefix: "_init_completion() { words=(\"${COMP_WORDS[@]}\"); cword=$COMP_CWORD; cur=${words[cword]}; prev=${words[cword-1]}; }\n",
			invoke: "COMP_WORDS=(align sample); COMP_CWORD=1; COMP_LINE='align sample'; COMP_POINT=${#COMP_LINE}; __start_align; printf '%s\\n' \"${COMPREPLY[@]}\"",
		},
		{
			shell:  "zsh",
			prefix: "compdef() { :; }; _describe() { printf '%s\\n' \"${completions[@]}\"; }\n",
			invoke: "words=(align sample); CURRENT=2; _align",
		},
	}

	for _, tc := range cases {
		t.Run(tc.shell, func(t *testing.T) {
			t.Parallel()
			path, err := exec.LookPath(tc.shell)
			if err != nil {
				t.Skipf("%s is unavailable: %v", tc.shell, err)
			}

			f := newAliasFixture(t)
			f.alias(t, "samplealias", "status")
			script, stderr, err := f.run(t, "completion", "--shell", tc.shell)
			if !assert.NoError(t, err, stderr) {
				return
			}

			wrapper := "testbinary=$1\nalign() { \"$testbinary\" -test.run=^TestAliasCLIProcess$ -- \"$@\"; }\n"
			cmd := exec.Command(path, "-s", "--", os.Args[0])
			cmd.Dir = f.base
			cmd.Env = f.env
			cmd.Stdin = strings.NewReader(tc.prefix + wrapper + script + "\n" + tc.invoke + "\n")
			out, err := cmd.CombinedOutput()
			assert.NoError(t, err, "%s", out)
			assert.Contains(t, string(out), "samplealias")
		})
	}
}

func TestBuiltinFlagTraversal(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name string
		args []string
	}{
		{name: "before command", args: []string{"-v", "branch", "--all"}},
		{name: "after command", args: []string{"branch", "--all", "-v"}},
		{name: "between nested commands", args: []string{"remote", "-v", "add", "origin"}},
		{name: "after nested command", args: []string{"remote", "add", "-v", "origin"}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			root := &cobra.Command{Use: "align", TraverseChildren: true}
			root.PersistentFlags().BoolP("verbose", "v", false, "")
			branch := &cobra.Command{Use: "branch"}
			branch.Flags().Bool("all", false, "")
			remote := &cobra.Command{Use: "remote"}
			add := &cobra.Command{Use: "add"}
			called := false
			for _, leaf := range []*cobra.Command{branch, add} {
				leaf.RunE = func(cmd *cobra.Command, args []string) error {
					called = true
					return nil
				}
			}

			remote.AddCommand(add)
			root.AddCommand(branch, remote)
			root.SetArgs(tc.args)
			err := root.Execute()
			assert.NoError(t, err)
			assert.True(t, called)
			verbose, err := root.PersistentFlags().GetBool("verbose")
			assert.NoError(t, err)
			assert.True(t, verbose)
		})
	}
}

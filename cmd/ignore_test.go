package cmd

import (
	"bytes"
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

func TestIgnoreCLIProcess(t *testing.T) {
	t.Parallel()
	if os.Getenv("ALIGN_IGNORE_TEST_PROCESS") != "1" {
		return
	}

	index := slices.Index(os.Args, "--")
	if index < 0 {
		os.Exit(1)
	}

	scrb, err := scribe.NewScribe(os.Stdout, scribe.DefaultTheme)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	clt = &client.Client{Repos: repos.New(scrb, nil, nil)}
	for _, command := range RootCmd.Commands() {
		if command.Name() != "clone" {
			command.PersistentPreRun = nil
		}
	}

	target, remaining, err := RootCmd.Find(os.Args[index+1:])
	if err == nil && remaining != nil && target == RootCmd {
		fmt.Fprintln(os.Stderr, "unexpected root command")
		os.Exit(1)
	}

	if err == nil && (target.Name() == "checkout" || target.Name() == "status" || target.Name() == "diff") {
		target.PersistentPreRunE = func(cmd *cobra.Command, args []string) error { return nil }
	}

	os.Args = append([]string{os.Args[0]}, os.Args[index+1:]...)
	err = Execute()
	if err != nil {
		os.Exit(1)
	}

	os.Exit(0)
}

func runIgnore(t *testing.T, f aliasFixture, args ...string) (string, string, error) {
	t.Helper()
	testArgs := append([]string{"-test.run=^TestIgnoreCLIProcess$", "--"}, args...)
	cmd := exec.Command(os.Args[0], testArgs...)
	cmd.Dir = f.base
	cmd.Env = append(f.env, "ALIGN_IGNORE_TEST_PROCESS=1")
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	return stdout.String(), stderr.String(), err
}

func currentBranch(t *testing.T, f aliasFixture, dir string) string {
	t.Helper()
	cmd := exec.Command("git", "branch", "--show-current")
	cmd.Dir = dir
	cmd.Env = f.env
	out, err := cmd.Output()
	assert.NoError(t, err)
	return strings.TrimSpace(string(out))
}

func TestIgnoreCheckout(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name string
		args []string
	}{
		{name: "after command", args: []string{"checkout", "target", "--ignore-repo", "userauth"}},
		{name: "before command", args: []string{"--ignore-repo", "userauth", "checkout", "target"}},
		{name: "equals verbose", args: []string{"-v", "checkout", "target", "--ignore-repo=userauth"}},
		{name: "duplicates", args: []string{"checkout", "target", "--ignore-repo", "userauth", "--ignore-repo", "userauth"}},
		{name: "concurrency override", args: []string{"checkout", "target", "--concurrency=2", "--ignore-repo", "userauth"}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			f := newAliasFixture(t)
			selected := f.repo(t, "selected")
			ignored := f.repo(t, "userauth")
			for _, dir := range []string{selected, ignored} {
				f.git(t, dir, "-c", "user.name=Fixture", "-c", "user.email=fixture@example.invalid", "-c", "commit.gpgsign=false", "commit", "--quiet", "--allow-empty", "-m", "fixture")
			}

			f.git(t, selected, "branch", "target")
			out, stderr, err := runIgnore(t, f, tc.args...)
			assert.NoError(t, err, stderr)
			assert.Equal(t, "target", currentBranch(t, f, selected))
			assert.Equal(t, "main", currentBranch(t, f, ignored))
			assert.NotContains(t, out, "userauth")
			if tc.args[0] != "-v" {
				assert.Contains(t, out, "(1/1)")
			}

			// Selection is not retained by the next process.
			out, stderr, err = runIgnore(t, f, "checkout", "main")
			assert.NoError(t, err, stderr)
			assert.Contains(t, out, "(2/2)")
		})
	}
}

func TestIgnoreAliases(t *testing.T) {
	t.Parallel()
	f := newAliasFixture(t)
	selected := f.repo(t, "selected")
	ignored := f.repo(t, "userauth")
	f.alias(t, "mark", `!printf 'selected-output\n'; printf done > alias-marker`)
	for _, flags := range [][]string{{"--ignore-repo", "userauth", "mark"}, {"-v", "--ignore-repo=userauth", "mark"}} {
		out, stderr, err := f.run(t, flags...)
		assert.NoError(t, err, stderr)
		assert.NotContains(t, out, "userauth")
		marker, readErr := os.ReadFile(filepath.Join(selected, "alias-marker"))
		assert.NoError(t, readErr)
		assert.Equal(t, "done", string(marker))
		info, statErr := os.Stat(filepath.Join(ignored, "alias-marker"))
		assert.Nil(t, info)
		assert.True(t, os.IsNotExist(statErr))
	}

	f.alias(t, "argcheck", `!f() { printf '<%s>\n' "$@"; }; f`)
	out, stderr, err := f.run(t, "-v", "argcheck", "--ignore-repo", "userauth")
	assert.NoError(t, err, stderr)
	assert.Equal(t, 2, strings.Count(out, "<--ignore-repo>"))
	assert.Equal(t, 2, strings.Count(out, "<userauth>"))
}

func TestIgnoreInvalidSelection(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name string
		args []string
		want string
	}{
		{name: "typo", args: []string{"--ignore-repo", "typo", "mark"}, want: "not found"},
		{name: "all", args: []string{"--ignore-repo", "userauth", "mark"}, want: "no repositories selected"},
		{name: "empty", args: []string{"--ignore-repo=", "mark"}, want: "not found"},
		{name: "missing flag value", args: []string{"--ignore-repo"}, want: "needs an argument"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			f := newAliasFixture(t)
			dir := f.repo(t, "userauth")
			f.alias(t, "mark", "!printf changed > marker")
			out, stderr, err := f.run(t, tc.args...)
			assert.Error(t, err)
			assert.Contains(t, stderr, tc.want)
			assert.NotContains(t, out, "(1/1)")
			info, statErr := os.Stat(filepath.Join(dir, "marker"))
			assert.Nil(t, info)
			assert.True(t, os.IsNotExist(statErr))
		})
	}
}

func TestIgnoreFlagCompletion(t *testing.T) {
	t.Parallel()
	f := newAliasFixture(t)
	f.repo(t, "selected")
	f.repo(t, "userauth")
	f.alias(t, "mark", "status")
	for _, args := range [][]string{
		{"__complete", "--ignore-repo", "user"},
		{"__complete", "--ignore-repo", "selected", "--ignore-repo", "user"},
		{"__complete", "--ignore-repo", "selected", "--ignore-repo", "userauth", "--ignore-repo", "user"},
		{"__complete", "checkout", "--ignore-repo=user"},
	} {
		out, stderr, err := f.run(t, args...)
		assert.NoError(t, err, stderr)
		assert.Contains(t, out, "userauth")
		assert.NotContains(t, out, "selected\n")
	}

	out, stderr, err := f.run(t, "__complete", "mark", "--ignore")
	assert.NoError(t, err, stderr)
	assert.Equal(t, ":4\n", out)
}

func TestIgnoreUnsupported(t *testing.T) {
	t.Parallel()
	for _, args := range [][]string{
		{"version", "--ignore-repo", "userauth"},
		{"auth", "--ignore-repo", "userauth"},
		{"config", "github.limits.burst", "10", "--ignore-repo", "userauth"},
		{"completion", "--shell", "bash", "--ignore-repo", "userauth"},
		{"clone", "fixture", "--ignore-repo", "userauth"},
	} {
		t.Run(args[0], func(t *testing.T) {
			t.Parallel()
			f := newAliasFixture(t)
			out, stderr, err := runIgnore(t, f, args...)
			assert.Error(t, err)
			assert.Contains(t, stderr, "not supported")
			assert.NotContains(t, out, "Config file updated")
		})
	}
}

func TestIgnoreOperationCompletion(t *testing.T) {
	t.Parallel()
	f := newAliasFixture(t)
	selected := f.repo(t, "selected")
	ignored := f.repo(t, "userauth")
	for _, dir := range []string{selected, ignored} {
		f.git(t, dir, "-c", "user.name=Fixture", "-c", "user.email=fixture@example.invalid", "-c", "commit.gpgsign=false", "commit", "--quiet", "--allow-empty", "-m", "fixture")
	}

	f.git(t, selected, "branch", "selectedbranch")
	f.git(t, ignored, "branch", "ignoredbranch")
	for _, args := range [][]string{
		{"__complete", "checkout", "--ignore-repo", "userauth", ""},
		{"__complete", "branch", "--delete", "--ignore-repo", "userauth", ""},
	} {
		out, stderr, err := runIgnore(t, f, args...)
		assert.NoError(t, err, stderr)
		assert.Contains(t, out, "selectedbranch")
		assert.NotContains(t, out, "ignoredbranch")
	}

	out, stderr, err := runIgnore(t, f, "__complete", "checkout", "--ignore-repo", "typo", "")
	assert.NoError(t, err, stderr) // Cobra reports completion errors via the directive.
	assert.Contains(t, out, ":1")
	assert.Contains(t, stderr, "not found")
}

func TestIgnoreInformationalOutput(t *testing.T) {
	t.Parallel()
	f := newAliasFixture(t)
	for _, name := range []string{"selected", "userauth"} {
		dir := f.repo(t, name)
		err := os.WriteFile(filepath.Join(dir, name+"-file"), []byte("original"), 0600)
		assert.NoError(t, err)
		f.git(t, dir, "add", ".")
		f.git(t, dir, "-c", "user.name=Fixture", "-c", "user.email=fixture@example.invalid", "-c", "commit.gpgsign=false", "commit", "--quiet", "-m", "fixture")
		err = os.WriteFile(filepath.Join(dir, name+"-file"), []byte("modified"), 0600)
		assert.NoError(t, err)
	}

	for _, command := range []string{"status", "diff"} {
		for _, flags := range [][]string{{}, {"-v"}} {
			args := append([]string{}, flags...)
			args = append(args, command, "--ignore-repo", "userauth")
			out, stderr, err := runIgnore(t, f, args...)
			assert.NoError(t, err, stderr)
			assert.Contains(t, out, "selected-file")
			assert.NotContains(t, out, "userauth-file")
			assert.NotContains(t, out, "\n    userauth")
		}
	}
}

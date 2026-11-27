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

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func TestCompletionGeneratorProcess(t *testing.T) {
	t.Parallel()

	name := os.Getenv("ALIGN_TEST_COMPLETION_GENERATOR")
	if name == "" {
		return
	}

	shell = name
	completionFunc(completionCmd, nil)
	os.Exit(0)
}

func TestCompletionShellProcess(t *testing.T) {
	t.Parallel()

	if os.Getenv("ALIGN_TEST_COMPLETION_SHELL") != "1" {
		return
	}

	index := slices.Index(os.Args, "--")
	if index < 0 {
		fmt.Fprintln(os.Stderr, "missing completion arguments")
		os.Exit(1)
	}

	fixture := &cloneCompletionFixture{accounts: []client.CloneAccount{
		{Login: "acme", Organization: true, Name: "Acme Inc", Description: "Example organization"},
		{Login: "other", Organization: true},
		{Login: "org-a", Organization: true, Description: "Shared"},
		{Login: "org-b", Organization: true, Description: "Shared"},
	}}
	root := &cobra.Command{Use: "align", SilenceErrors: true, SilenceUsage: true}
	root.AddCommand(&cobra.Command{
		Use: "clone [user|org] [directory]",
		ValidArgsFunction: func(cmd *cobra.Command, args []string, prefix string) ([]string, cobra.ShellCompDirective) {
			completions, directive := completeClone(cmd.Context(), args, prefix, fixture)
			return completions, directive
		},
	})
	root.SetArgs(os.Args[index+1:])
	err := root.Execute()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	os.Exit(0)
}

func TestCompletionGenerators(t *testing.T) {
	t.Parallel()

	for _, name := range []string{"bash", "zsh", "fish", "powershell"} {
		t.Run(name, func(t *testing.T) {
			process := exec.Command(os.Args[0], "-test.run=^TestCompletionGeneratorProcess$")
			process.Env = append(os.Environ(), "ALIGN_TEST_COMPLETION_GENERATOR="+name)
			script, err := process.Output()
			if !assert.NoError(t, err) {
				return
			}
			assert.Contains(t, string(script), "__complete")
			assert.NotContains(t, string(script), "__completeNoDesc")

			if name != "bash" && name != "zsh" {
				return
			}

			shellPath, err := exec.LookPath(name)
			if err != nil {
				t.Skipf("%s is unavailable", name)
			}
			scriptPath := filepath.Join(t.TempDir(), "completion")
			err = os.WriteFile(scriptPath, script, 0600)
			if !assert.NoError(t, err) {
				return
			}
			output, err := exec.Command(shellPath, "-n", scriptPath).CombinedOutput()
			assert.NoError(t, err, "%s", output)

			if name == "bash" {
				testBashCompletion(t, shellPath, scriptPath)
			}
		})
	}
}

func testBashCompletion(t *testing.T, shellPath, scriptPath string) {
	t.Helper()

	dir := t.TempDir()
	wrapper := "#!/bin/sh\nexec \"$ALIGN_TEST_BINARY\" -test.run=^TestCompletionShellProcess$ -- \"$@\"\n"
	err := os.WriteFile(filepath.Join(dir, "align"), []byte(wrapper), 0700)
	if !assert.NoError(t, err) {
		return
	}

	err = os.Mkdir(filepath.Join(dir, "destination"), 0700)
	if !assert.NoError(t, err) {
		return
	}
	err = os.WriteFile(filepath.Join(dir, "data-file"), []byte("fixture"), 0600)
	if !assert.NoError(t, err) {
		return
	}

	program := `source "$ALIGN_TEST_SCRIPT"
_init_completion() {
  cur="${COMP_WORDS[COMP_CWORD]}"
  prev="${COMP_WORDS[COMP_CWORD-1]}"
  words=("${COMP_WORDS[@]}")
  cword=$COMP_CWORD
}
_filedir() {
  while IFS= read -r entry; do COMPREPLY+=("$entry"); done < <(compgen -d -- "$cur")
}
COMP_WORDS=(align clone ac)
COMP_CWORD=2
COMP_LINE='align clone ac'
COMP_POINT=${#COMP_LINE}
__start_align
printf 'account:%s\n' "${COMPREPLY[@]}"
COMP_WORDS=(align clone org)
COMP_CWORD=2
COMP_LINE='align clone org'
COMP_POINT=${#COMP_LINE}
__start_align
printf 'matching:%s\n' "${COMPREPLY[@]}"
COMP_WORDS=(align clone acme d)
COMP_CWORD=3
COMP_LINE='align clone acme d'
COMP_POINT=${#COMP_LINE}
__start_align
printf 'directory:%s\n' "${COMPREPLY[@]}"`
	process := exec.Command(shellPath, "-c", program)
	process.Dir = dir
	process.Env = append(os.Environ(),
		"PATH="+dir+string(os.PathListSeparator)+os.Getenv("PATH"),
		"ALIGN_TEST_BINARY="+os.Args[0],
		"ALIGN_TEST_SCRIPT="+scriptPath,
		"ALIGN_TEST_COMPLETION_SHELL=1",
	)
	var stderr bytes.Buffer
	process.Stderr = &stderr
	output, err := process.Output()
	if !assert.NoError(t, err, "%s", stderr.String()) {
		return
	}
	result := strings.TrimSpace(string(output))
	assert.Contains(t, strings.Split(result, "\n"), "account:acme", "%s", stderr.String())
	assert.Contains(t, strings.Split(result, "\n"), "matching:org-a", "%s", stderr.String())
	assert.Contains(t, strings.Split(result, "\n"), "matching:org-b", "%s", stderr.String())
	assert.Contains(t, result, "directory:destination", "%s", stderr.String())
	assert.NotContains(t, result, "data-file")
	assert.NotContains(t, result, "Example organization")
}

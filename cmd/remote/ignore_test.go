package remote

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/gomicro/align/client"
	"github.com/gomicro/align/cmd/selection"
	"github.com/gomicro/align/config"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func TestIgnoreRemoteTraversal(t *testing.T) {
	t.Parallel()
	cases := [][]string{
		{"--ignore-repo", "userauth", "remote", "add"},
		{"remote", "--ignore-repo", "userauth", "add"},
		{"remote", "add", "--ignore-repo", "userauth"},
	}
	for _, args := range cases {
		t.Run(args[0], func(t *testing.T) {
			t.Parallel()
			base := t.TempDir()
			for _, name := range []string{"selected", "userauth"} {
				err := os.MkdirAll(filepath.Join(base, name, ".git"), 0700)
				assert.NoError(t, err)
			}

			root := &cobra.Command{Use: "align", TraverseChildren: true}
			root.PersistentFlags().StringArray("ignore-repo", nil, "")
			remote := &cobra.Command{Use: "remote"}
			var got []string
			add := &cobra.Command{Use: "add", RunE: func(cmd *cobra.Command, args []string) error {
				dirs, err := selection.Dirs(context.Background(), cmd, &client.Client{}, base)
				got = dirs
				return err
			}}
			root.AddCommand(remote)
			remote.AddCommand(add)
			root.SetArgs(args)
			err := root.Execute()
			assert.NoError(t, err)
			assert.Equal(t, []string{filepath.Join(base, "selected")}, got)
		})
	}
}

func TestRemoteIgnoreProcess(t *testing.T) {
	t.Parallel()
	if os.Getenv("ALIGN_REMOTE_IGNORE_PROCESS") != "1" {
		return
	}

	cfg := &config.Config{Github: &config.GithubHost{Limits: &config.Limits{RequestsPerSecond: 10, Burst: 25}}}
	c, err := client.New(cfg, client.WithNoColor())
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	clt = c
	RemoteCmd.PersistentPreRun = nil
	root := &cobra.Command{Use: "align", TraverseChildren: true}
	root.PersistentFlags().StringArray("ignore-repo", nil, "")
	root.AddCommand(RemoteCmd)
	root.SetArgs([]string{"remote", "add", "origin", "https://example.invalid", "--ignore-repo", "userauth"})
	err = root.Execute()
	if err != nil {
		os.Exit(1)
	}

	os.Exit(0)
}

func TestRemoteIgnoreExecution(t *testing.T) {
	t.Parallel()
	base := t.TempDir()
	for _, name := range []string{"selected", "userauth"} {
		dir := filepath.Join(base, name)
		err := os.Mkdir(dir, 0700)
		assert.NoError(t, err)
		git := exec.Command("git", "init", "--quiet", dir)
		out, err := git.CombinedOutput()
		assert.NoError(t, err, "%s", out)
	}

	cmd := exec.Command(os.Args[0], "-test.run=^TestRemoteIgnoreProcess$")
	cmd.Dir = base
	cmd.Env = append(os.Environ(), "ALIGN_REMOTE_IGNORE_PROCESS=1")
	out, err := cmd.CombinedOutput()
	assert.NoError(t, err, "%s", out)
	for _, name := range []string{"selected", "userauth"} {
		git := exec.Command("git", "remote")
		git.Dir = filepath.Join(base, name)
		out, err := git.Output()
		assert.NoError(t, err)
		if name == "selected" {
			assert.Equal(t, "origin\n", string(out))
		} else {
			assert.Empty(t, out)
		}
	}
}

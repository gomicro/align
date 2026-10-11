package stash

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/gomicro/align/client"
	"github.com/gomicro/align/config"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func TestStashIgnoreProcess(t *testing.T) {
	t.Parallel()
	if os.Getenv("ALIGN_STASH_IGNORE_PROCESS") != "1" {
		return
	}

	cfg := &config.Config{Github: &config.GithubHost{Limits: &config.Limits{RequestsPerSecond: 10, Burst: 25}}}
	c, err := client.New(cfg, client.WithNoColor())
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	clt = c
	StashCmd.PersistentPreRun = nil
	root := &cobra.Command{Use: "align", TraverseChildren: true}
	root.PersistentFlags().StringArray("ignore-repo", nil, "")
	root.AddCommand(StashCmd)
	viper.Set("verbose", os.Getenv("ALIGN_STASH_VERBOSE") == "1")
	root.SetArgs([]string{"stash", "--ignore-repo", "userauth"})
	err = root.Execute()
	if err != nil {
		os.Exit(1)
	}

	os.Exit(0)
}

func TestStashIgnoreExecution(t *testing.T) {
	t.Parallel()
	for _, verbose := range []bool{false, true} {
		t.Run(fmt.Sprintf("verbose=%t", verbose), func(t *testing.T) {
			t.Parallel()
			base := t.TempDir()
			for _, name := range []string{"selected", "userauth"} {
				dir := filepath.Join(base, name)
				err := os.Mkdir(dir, 0700)
				assert.NoError(t, err)
				stashFixtureGit(t, dir, "init", "--quiet")
				err = os.WriteFile(filepath.Join(dir, "tracked"), []byte("original"), 0600)
				assert.NoError(t, err)
				stashFixtureGit(t, dir, "add", "tracked")
				stashFixtureGit(t, dir, "-c", "user.name=Fixture", "-c", "user.email=fixture@example.invalid", "-c", "commit.gpgsign=false", "commit", "--quiet", "-m", "fixture")
				err = os.WriteFile(filepath.Join(dir, "tracked"), []byte("modified"), 0600)
				assert.NoError(t, err)
			}

			cmd := exec.Command(os.Args[0], "-test.run=^TestStashIgnoreProcess$")
			cmd.Dir = base
			cmd.Env = append(os.Environ(), "ALIGN_STASH_IGNORE_PROCESS=1")
			if verbose {
				cmd.Env = append(cmd.Env, "ALIGN_STASH_VERBOSE=1")
			}

			out, err := cmd.CombinedOutput()
			assert.NoError(t, err, "%s", out)
			assert.NotContains(t, string(out), "userauth")
			for _, name := range []string{"selected", "userauth"} {
				data, err := os.ReadFile(filepath.Join(base, name, "tracked"))
				assert.NoError(t, err)
				if name == "selected" {
					assert.Equal(t, "original", string(data))
				} else {
					assert.Equal(t, "modified", string(data))
				}
			}
		})
	}
}

func stashFixtureGit(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), "GIT_CONFIG_GLOBAL=/dev/null", "GIT_CONFIG_SYSTEM=/dev/null")
	out, err := cmd.CombinedOutput()
	assert.NoError(t, err, "%s", out)
}

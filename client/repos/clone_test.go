package repos

import (
	"bytes"
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	ctxhelper "github.com/gomicro/align/client/context"

	"github.com/gomicro/scribe"
	"github.com/google/go-github/github"
	"github.com/stretchr/testify/assert"
)

// setupCloneFixture creates bare remotes for the given "org/name" paths and
// rewrites GitHub SSH URLs to them. It changes the working directory because
// clone destinations are relative to it, so callers cannot run in parallel.
func setupCloneFixture(t *testing.T, remotes ...string) *Repos {
	t.Helper()
	base := t.TempDir()
	remoteDir := filepath.Join(base, "remotes")
	for _, remote := range remotes {
		out, err := exec.Command("git", "init", "--bare", "--quiet", filepath.Join(remoteDir, remote+".git")).CombinedOutput()
		if !assert.NoError(t, err, string(out)) {
			t.FailNow()
		}
	}

	t.Setenv("GIT_CONFIG_GLOBAL", os.DevNull)
	t.Setenv("GIT_CONFIG_NOSYSTEM", "1")
	t.Setenv("GIT_CONFIG_COUNT", "1")
	t.Setenv("GIT_CONFIG_KEY_0", "url."+remoteDir+"/.insteadOf")
	t.Setenv("GIT_CONFIG_VALUE_0", "git@github.com:")

	work := filepath.Join(base, "work")
	err := os.Mkdir(work, 0700)
	if !assert.NoError(t, err) {
		t.FailNow()
	}

	t.Chdir(work)
	scrb, err := scribe.NewScribe(&bytes.Buffer{}, scribe.DefaultTheme)
	if !assert.NoError(t, err) {
		t.FailNow()
	}

	return New(scrb, nil, nil)
}

func cloneContext(names ...string) context.Context {
	var ghRepos []*github.Repository
	for _, name := range names {
		ghRepos = append(ghRepos, &github.Repository{SSHURL: github.String("git@github.com:" + name + ".git")})
	}

	ctx := ctxhelper.WithVerbose(context.Background(), true)
	return ctxhelper.WithRepos(ctx, ghRepos)
}

func clonedNames(cloned []*ctxhelper.Repository) []string {
	names := make([]string, len(cloned))
	for i := range cloned {
		names[i] = cloned[i].Name
	}

	return names
}

func TestCloneReposSkipsExistingAndReportsSuccesses(t *testing.T) {
	r := setupCloneFixture(t, "org/fresh", "org/other", "org/existing")
	err := os.MkdirAll(filepath.Join("org", "existing", ".git"), 0700)
	if !assert.NoError(t, err) {
		return
	}

	ctx := ctxhelper.WithConcurrency(cloneContext("org/fresh", "org/other", "org/existing", "org/missing"), 2)
	cloned, err := r.CloneRepos(ctx, "")
	assert.ErrorContains(t, err, filepath.Join("org", "missing"))
	assert.ElementsMatch(t, []string{"fresh", "other"}, clonedNames(cloned))
	assert.DirExists(t, filepath.Join("org", "fresh", ".git"))
	assert.DirExists(t, filepath.Join("org", "other", ".git"))
	assert.NoFileExists(t, filepath.Join("org", "existing", ".git", "HEAD"))
}

func TestCloneReposDeduplicatesDestinations(t *testing.T) {
	r := setupCloneFixture(t, "first/dup", "second/dup")

	cloned, err := r.CloneRepos(cloneContext("first/dup", "second/dup"), "dest")
	assert.NoError(t, err)
	assert.Len(t, cloned, 1)
	assert.DirExists(t, filepath.Join("dest", "dup", ".git"))
}

package remotes

import (
	"bytes"
	"context"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	ctxhelper "github.com/gomicro/align/client/context"

	"github.com/gomicro/scribe"
	"github.com/stretchr/testify/assert"
)

func gitIn(t *testing.T, dir string, args ...string) {
	t.Helper()
	out, err := exec.Command("git", append([]string{"-C", dir}, args...)...).CombinedOutput()
	if !assert.NoError(t, err, string(out)) {
		t.FailNow()
	}
}

func TestAddReportsFailuresInVerboseMode(t *testing.T) {
	t.Parallel()
	base := t.TempDir()
	fresh := filepath.Join(base, "fresh")
	taken := filepath.Join(base, "taken")
	for _, dir := range []string{fresh, taken} {
		gitIn(t, base, "init", "--quiet", dir)
	}

	gitIn(t, taken, "remote", "add", "upstream", "https://example.invalid/taken.git")

	var output bytes.Buffer
	scrb, err := scribe.NewScribe(&output, scribe.DefaultTheme)
	if !assert.NoError(t, err) {
		return
	}

	ctx := ctxhelper.WithConcurrency(ctxhelper.WithVerbose(context.Background(), true), 2)
	err = New(scrb).Add(ctx, []string{fresh, taken}, "upstream", "https://example.invalid")
	assert.ErrorContains(t, err, taken)
	assert.NotContains(t, err.Error(), fresh+":")

	out, err := exec.Command("git", "-C", fresh, "remote").Output()
	assert.NoError(t, err)
	assert.Equal(t, "upstream", strings.TrimSpace(string(out)))
	assert.Contains(t, output.String(), "git remote add upstream")
}

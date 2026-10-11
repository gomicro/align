package selection

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/gomicro/align/client"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func TestDirsInvalidFlag(t *testing.T) {
	t.Parallel()
	command := &cobra.Command{Use: "test"}
	command.Flags().Bool("ignore-repo", false, "")
	dirs, err := Dirs(context.Background(), command, &client.Client{}, t.TempDir())
	assert.Nil(t, dirs)
	assert.ErrorContains(t, err, "reading repository exclusions")
}

func TestDirsCommaAndSpaceNames(t *testing.T) {
	t.Parallel()
	base := t.TempDir()
	for _, name := range []string{"comma,name", "space name", "selected"} {
		err := os.MkdirAll(filepath.Join(base, name, ".git"), 0700)
		assert.NoError(t, err)
	}

	command := &cobra.Command{Use: "test"}
	command.Flags().StringArray("ignore-repo", nil, "")
	err := command.ParseFlags([]string{"--ignore-repo", "comma,name", "--ignore-repo", "space name"})
	assert.NoError(t, err)
	dirs, err := Dirs(context.Background(), command, &client.Client{}, base)
	assert.NoError(t, err)
	assert.Equal(t, []string{filepath.Join(base, "selected")}, dirs)
}

package skills

import (
	"context"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v2"
)

func TestDump(t *testing.T) {
	t.Parallel()
	outputDir := filepath.Join(t.TempDir(), "nested", "skills")
	path, err := Dump(context.Background(), outputDir, false)
	if !assert.NoError(t, err) {
		return
	}

	got, err := os.ReadFile(path)
	assert.NoError(t, err)
	source, err := os.ReadFile("align/SKILL.md")
	assert.NoError(t, err)
	assert.Equal(t, source, got)

	parts := strings.SplitN(string(got), "---", 3)
	if !assert.Len(t, parts, 3) {
		return
	}

	var metadata struct {
		Name        string `yaml:"name"`
		Description string `yaml:"description"`
	}
	err = yaml.UnmarshalStrict([]byte(parts[1]), &metadata)
	assert.NoError(t, err)
	assert.Equal(t, filepath.Base(filepath.Dir(path)), metadata.Name)
	assert.NotEmpty(t, metadata.Description)
	assert.LessOrEqual(t, len(metadata.Description), 1024)
	assert.Contains(t, parts[2], "<HandleResults>")
}

func TestDumpExisting(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name      string
		identical bool
		force     bool
		wantError bool
	}{
		{name: "identical", identical: true},
		{name: "identical forced", identical: true, force: true},
		{name: "different refused", wantError: true},
		{name: "different replaced", force: true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			outputDir := t.TempDir()
			path, err := Dump(context.Background(), outputDir, false)
			if !assert.NoError(t, err) {
				return
			}

			original, err := os.ReadFile(path)
			assert.NoError(t, err)
			unrelated := filepath.Join(filepath.Dir(path), "notes.txt")
			err = os.WriteFile(unrelated, []byte("keep"), 0644)
			assert.NoError(t, err)
			if !tc.identical {
				err = os.WriteFile(path, []byte("customized"), 0644)
				assert.NoError(t, err)
			}

			oldTime := time.Unix(1000000000, 0)
			err = os.Chtimes(path, oldTime, oldTime)
			assert.NoError(t, err)
			exported, err := Dump(context.Background(), outputDir, tc.force)
			if tc.wantError {
				assert.ErrorContains(t, err, "--force")
				assert.Empty(t, exported)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, path, exported)
			}

			got, readErr := os.ReadFile(path)
			assert.NoError(t, readErr)
			expected := original
			if tc.wantError {
				expected = []byte("customized")
			}

			assert.Equal(t, expected, got)
			info, statErr := os.Stat(path)
			if assert.NoError(t, statErr) && (tc.identical || tc.wantError) {
				assert.Equal(t, oldTime, info.ModTime())
			}

			notes, readErr := os.ReadFile(unrelated)
			assert.NoError(t, readErr)
			assert.Equal(t, "keep", string(notes))
			entries, readErr := os.ReadDir(filepath.Dir(path))
			assert.NoError(t, readErr)
			assert.Len(t, entries, 2)
		})
	}
}

func TestDumpInvalidDestination(t *testing.T) {
	t.Parallel()
	for _, outputDir := range []string{"", " ", "\t"} {
		path, err := Dump(context.Background(), outputDir, false)
		assert.ErrorContains(t, err, "--output-dir")
		assert.Empty(t, path)
	}

	file := filepath.Join(t.TempDir(), "file")
	err := os.WriteFile(file, []byte("keep"), 0644)
	assert.NoError(t, err)
	path, err := Dump(context.Background(), filepath.Join(file, "nested"), true)
	assert.Error(t, err)
	assert.Empty(t, path)
	got, err := os.ReadFile(file)
	assert.NoError(t, err)
	assert.Equal(t, "keep", string(got))
}

func TestDumpCanceled(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	outputDir := filepath.Join(t.TempDir(), "absent")
	path, err := Dump(ctx, outputDir, false)
	assert.ErrorIs(t, err, context.Canceled)
	assert.Empty(t, path)
	info, err := os.Stat(outputDir)
	assert.Nil(t, info)
	assert.ErrorIs(t, err, fs.ErrNotExist)
}

func TestDumpConflictingEntries(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name     string
		symlink  bool
		skillDir bool
	}{
		{name: "directory symlink", symlink: true, skillDir: true},
		{name: "file symlink", symlink: true},
		{name: "directory is file", skillDir: true},
		{name: "file is directory"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			outputDir := t.TempDir()
			external := t.TempDir()
			externalFile := filepath.Join(external, "SKILL.md")
			err := os.WriteFile(externalFile, []byte("keep"), 0644)
			assert.NoError(t, err)
			entry := filepath.Join(outputDir, "align")
			target := external
			if !tc.skillDir {
				err = os.Mkdir(entry, 0755)
				assert.NoError(t, err)
				entry = filepath.Join(entry, "SKILL.md")
				target = externalFile
			}

			switch {
			case tc.symlink:
				err = os.Symlink(target, entry)
			case tc.skillDir:
				err = os.WriteFile(entry, []byte("keep"), 0644)
			default:
				err = os.Mkdir(entry, 0755)
			}
			if !assert.NoError(t, err) {
				return
			}

			for _, force := range []bool{false, true} {
				path, dumpErr := Dump(context.Background(), outputDir, force)
				assert.Error(t, dumpErr)
				assert.Empty(t, path)
				got, readErr := os.ReadFile(externalFile)
				assert.NoError(t, readErr)
				assert.Equal(t, "keep", string(got))
			}
		})
	}
}

func TestWriteFailureCleanup(t *testing.T) {
	t.Parallel()
	for _, force := range []bool{false, true} {
		outputDir := t.TempDir()
		skillDir := filepath.Join(outputDir, "align")
		err := os.MkdirAll(filepath.Join(skillDir, "SKILL.md"), 0755)
		assert.NoError(t, err)
		root, err := os.OpenRoot(outputDir)
		if !assert.NoError(t, err) {
			return
		}

		err = write(context.Background(), root, []byte("replacement"), force)
		assert.ErrorContains(t, err, "installing skill")
		closeErr := root.Close()
		assert.NoError(t, closeErr)
		entries, readErr := os.ReadDir(skillDir)
		if assert.NoError(t, readErr) && assert.Len(t, entries, 1) {
			assert.Equal(t, "SKILL.md", entries[0].Name())
			assert.True(t, entries[0].IsDir())
		}
	}
}

func TestInstallDoesNotReplaceWithoutForce(t *testing.T) {
	t.Parallel()
	outputDir := t.TempDir()
	skillDir := filepath.Join(outputDir, "align")
	err := os.Mkdir(skillDir, 0755)
	assert.NoError(t, err)
	path := filepath.Join(skillDir, "SKILL.md")
	err = os.WriteFile(path, []byte("concurrent content"), 0644)
	assert.NoError(t, err)
	root, err := os.OpenRoot(outputDir)
	if !assert.NoError(t, err) {
		return
	}

	err = write(context.Background(), root, []byte("replacement"), false)
	assert.ErrorIs(t, err, fs.ErrExist)
	closeErr := root.Close()
	assert.NoError(t, closeErr)
	got, err := os.ReadFile(path)
	assert.NoError(t, err)
	assert.Equal(t, "concurrent content", string(got))
	entries, err := os.ReadDir(skillDir)
	assert.NoError(t, err)
	assert.Len(t, entries, 1)
}

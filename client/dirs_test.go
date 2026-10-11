package client

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	ctxhelper "github.com/gomicro/align/client/context"

	"github.com/stretchr/testify/assert"
)

func TestGetDirsSelection(t *testing.T) {
	t.Parallel()
	base := t.TempDir()
	for _, name := range []string{"a", "userauth", "space name", "comma,name"} {
		err := os.MkdirAll(filepath.Join(base, name, ".git"), 0700)
		assert.NoError(t, err)
	}

	err := os.Mkdir(filepath.Join(base, "notrepo"), 0700)
	assert.NoError(t, err)
	cases := []struct {
		name    string
		ignored []string
		want    []string
		failure string
	}{
		{name: "unchanged", want: []string{"a", "comma,name", "space name", "userauth"}},
		{name: "single", ignored: []string{"userauth"}, want: []string{"a", "comma,name", "space name"}},
		{name: "multiple duplicate", ignored: []string{"userauth", "comma,name", "userauth", "space name"}, want: []string{"a"}},
		{name: "all", ignored: []string{"a", "userauth", "space name", "comma,name"}, failure: "no repositories selected"},
		{name: "case", ignored: []string{"Userauth"}, failure: "Userauth"},
		{name: "typo", ignored: []string{"missing"}, failure: "missing"},
		{name: "empty", ignored: []string{""}, failure: "not found"},
		{name: "path", ignored: []string{"./userauth"}, failure: "./userauth"},
		{name: "pattern", ignored: []string{"user*"}, failure: "user*"},
		{name: "not repository", ignored: []string{"notrepo"}, failure: "notrepo"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			ctx := ctxhelper.WithIgnoredRepos(context.Background(), tc.ignored)
			clt := &Client{}
			dirs, err := clt.GetDirs(ctx, base)
			if tc.failure != "" {
				assert.ErrorContains(t, err, tc.failure)
				assert.Nil(t, dirs)
				return
			}

			assert.NoError(t, err)
			want := make([]string, 0, len(tc.want))
			for _, name := range tc.want {
				want = append(want, filepath.Join(base, name))
			}

			assert.Equal(t, want, dirs)
		})
	}
}

func TestGetDirsEmptySelection(t *testing.T) {
	t.Parallel()
	clt := &Client{}
	base := t.TempDir()
	dirs, err := clt.GetDirs(context.Background(), base)
	assert.NoError(t, err)
	assert.Empty(t, dirs)
	dirs, err = clt.GetDirs(ctxhelper.WithIgnoredRepos(context.Background(), []string{"a"}), base)
	assert.ErrorContains(t, err, "not found")
	assert.Nil(t, dirs)
}

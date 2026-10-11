package context

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIgnoredRepos(t *testing.T) {
	t.Parallel()
	assert.Nil(t, IgnoredRepos(context.Background()))
	names := []string{"a"}
	ctx := WithIgnoredRepos(context.Background(), names)
	names[0] = "changed"
	got := IgnoredRepos(ctx)
	assert.Equal(t, []string{"a"}, got)
	got[0] = "changed"
	assert.Equal(t, []string{"a"}, IgnoredRepos(ctx))
}

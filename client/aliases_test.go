package client

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAliasNamesCancellation(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	names, err := AliasNames(ctx)
	assert.Nil(t, names)
	assert.ErrorIs(t, err, context.Canceled)
}

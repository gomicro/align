package repos

import (
	"bytes"
	"context"
	"testing"

	ctxhelper "github.com/gomicro/align/client/context"

	"github.com/gomicro/scribe"
	"github.com/stretchr/testify/assert"
)

func TestRunAliasCancellation(t *testing.T) {
	t.Parallel()
	var output bytes.Buffer
	scrb, err := scribe.NewScribe(&output, scribe.DefaultTheme)
	if !assert.NoError(t, err) {
		return
	}

	r := New(scrb, nil, nil)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	ctx = ctxhelper.WithVerbose(ctx, true)

	err = r.RunAlias(ctx, []string{t.TempDir(), t.TempDir()}, "anything")
	assert.ErrorIs(t, err, context.Canceled)
	assert.Contains(t, err.Error(), "context canceled")
}

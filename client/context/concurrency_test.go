package context

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConcurrency(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name     string
		override int
		present  bool
		want     int
		wantErr  bool
	}{
		{name: "fallback", want: 5},
		{name: "sequential", override: 1, present: true, want: 1},
		{name: "above default", override: 100, present: true, want: 100},
		{name: "zero", present: true, wantErr: true},
		{name: "negative", override: -1, present: true, wantErr: true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			ctx := WithVerbose(context.Background(), true)
			if tc.present {
				ctx = WithConcurrency(ctx, tc.override)
			}

			limit, err := Concurrency(ctx, 5)
			if tc.wantErr {
				assert.ErrorContains(t, err, "positive integer")
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.want, limit)
			}
			assert.True(t, Verbose(ctx))
		})
	}
}

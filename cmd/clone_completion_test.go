package cmd

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/gomicro/align/client"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

type cloneCompletionFixture struct {
	accounts []client.CloneAccount
	err      error
	calls    int
}

func (f *cloneCompletionFixture) CloneAccounts(context.Context) ([]client.CloneAccount, error) {
	f.calls++
	if f.err != nil {
		return nil, fmt.Errorf("fetching accounts: %w", f.err)
	}

	return f.accounts, nil
}

func TestCloneCompletion(t *testing.T) {
	t.Parallel()

	fixture := &cloneCompletionFixture{accounts: []client.CloneAccount{
		{Login: "user"},
		{Login: "acme", Organization: true, Name: "ACME\tIndustries\n\x1b[31m", Description: "Tools\tand\nservices \x1b[31m"},
		{Login: "empty", Organization: true},
		{Login: "empty-two", Organization: true, Name: " \t\n", Description: "\n"},
		{Login: "name-only", Organization: true, Name: "Name Only"},
		{Login: "description-only", Organization: true, Description: "Tools"},
		{Login: "alpha", Organization: true, Name: "Shared", Description: "shared"},
		{Login: "beta", Organization: true, Name: "Shared", Description: "shared"},
		{Login: "bad\ncandidate", Organization: true},
		{Login: "bad\u202ecandidate", Organization: true},
	}}
	completions, directive := completeClone(context.Background(), nil, "", fixture)
	assert.Equal(t, []string{
		"user\tuser",
		"acme\torganization - ACME Industries [31m - Tools and services [31m",
		"empty\torganization",
		"empty-two\torganization",
		"name-only\torganization - Name Only",
		"description-only\torganization - Tools",
		"alpha\torganization - Shared - shared",
		"beta\torganization - Shared - shared",
	}, completions)
	assert.Equal(t, cobra.ShellCompDirectiveNoFileComp, directive)
	assert.Equal(t, 1, fixture.calls)

	completions, directive = completeClone(context.Background(), []string{"acme"}, "", fixture)
	assert.Empty(t, completions)
	assert.Equal(t, cobra.ShellCompDirectiveFilterDirs, directive)
	assert.Equal(t, 1, fixture.calls)

	completions, directive = completeClone(context.Background(), []string{"acme", "dest"}, "", fixture)
	assert.Empty(t, completions)
	assert.Equal(t, cobra.ShellCompDirectiveNoFileComp, directive)
	assert.Equal(t, 1, fixture.calls)
}

func TestCloneCompletionFailure(t *testing.T) {
	t.Parallel()

	fixture := &cloneCompletionFixture{err: errors.New("unavailable")}
	completions, directive := completeClone(context.Background(), nil, "", fixture)
	assert.Empty(t, completions)
	assert.Equal(t, cobra.ShellCompDirectiveNoFileComp, directive)

	path := filepath.Join(t.TempDir(), "config")
	err := os.WriteFile(path, []byte("github.com: [\n"), 0600)
	if !assert.NoError(t, err) {
		return
	}

	completions, directive = completeClone(context.Background(), nil, "", liveCloneAccountSource{path: path})
	assert.Empty(t, completions)
	assert.Equal(t, cobra.ShellCompDirectiveNoFileComp, directive)
}

func TestCloneCompletionProtocol(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		args []string
		want string
	}{
		{name: "description", args: []string{"__complete", "clone", "ac"}, want: "acme\torganization - Acme Inc - Example organization\n:4\n"},
		{name: "without description", args: []string{"__completeNoDesc", "clone", "ac"}, want: "acme\n:4\n"},
		{name: "matching descriptions", args: []string{"__complete", "clone", "org"}, want: "org-a\torganization - Shared\norg-b\torganization - Shared\n:4\n"},
		{name: "matching descriptions without descriptions", args: []string{"__completeNoDesc", "clone", "org"}, want: "org-a\norg-b\n:4\n"},
		{name: "directory", args: []string{"__complete", "clone", "acme", ""}, want: ":16\n"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fixture := &cloneCompletionFixture{accounts: []client.CloneAccount{
				{Login: "acme", Organization: true, Name: "Acme Inc", Description: "Example organization"},
				{Login: "org-a", Organization: true, Description: "Shared"},
				{Login: "org-b", Organization: true, Description: "Shared"},
				{Login: "user"},
			}}
			root := &cobra.Command{Use: "align", SilenceErrors: true, SilenceUsage: true}
			clone := &cobra.Command{
				Use: "clone [user|org] [directory]",
				ValidArgsFunction: func(cmd *cobra.Command, args []string, prefix string) ([]string, cobra.ShellCompDirective) {
					result, directive := completeClone(cmd.Context(), args, prefix, fixture)
					return result, directive
				},
			}
			root.AddCommand(clone)
			var out, stderr bytes.Buffer
			root.SetOut(&out)
			root.SetErr(&stderr)
			root.SetArgs(tt.args)
			err := root.Execute()
			assert.NoError(t, err)
			assert.Equal(t, tt.want, out.String())
		})
	}
}

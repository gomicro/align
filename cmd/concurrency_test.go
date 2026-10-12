package cmd

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/gomicro/align/client"
	ctxhelper "github.com/gomicro/align/client/context"
	"github.com/gomicro/align/client/repos"
	"github.com/gomicro/align/client/testclient"

	"github.com/gomicro/scribe"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

type concurrencyTestKey struct{}

type concurrencyRecord struct {
	Calls   []string
	Args    []string
	Limit   int
	Parent  bool
	Verbose bool
}

type blockingConcurrencyClient struct {
	*testclient.TestClient
}

func (c *blockingConcurrencyClient) CheckoutRepos(ctx context.Context, dirs []string, args ...string) error {
	err := c.TestClient.CheckoutRepos(ctx, dirs, args...)
	if err != nil {
		return fmt.Errorf("recording checkout: %w", err)
	}

	fmt.Fprintln(os.Stdout, "ready")
	<-ctx.Done()
	return fmt.Errorf("checkout canceled: %w", ctx.Err())
}

func (c *blockingConcurrencyClient) PullRepos(ctx context.Context, dirs []string, args ...string) error {
	err := c.TestClient.PullRepos(ctx, dirs, args...)
	if err != nil {
		return fmt.Errorf("recording pull: %w", err)
	}

	fmt.Fprintln(os.Stdout, "ready")
	<-ctx.Done()
	return fmt.Errorf("pull canceled: %w", ctx.Err())
}

func TestConcurrencyCLIProcess(t *testing.T) {
	t.Parallel()
	mode := os.Getenv("ALIGN_CONCURRENCY_CLI_TEST")
	if mode == "" {
		return
	}

	index := slices.Index(os.Args, "--")
	if index < 0 {
		os.Exit(1)
	}

	tc := testclient.New()
	clt = tc
	if mode == "block" {
		clt = &blockingConcurrencyClient{TestClient: tc}
	}

	// Client construction is unrelated to flag/context wiring.
	checkoutCmd.PersistentPreRun = nil
	pullCmd.PersistentPreRun = nil
	RootCmd.SetContext(context.WithValue(context.Background(), concurrencyTestKey{}, true))
	os.Args = append([]string{os.Args[0]}, os.Args[index+1:]...)
	err := Execute()
	record := concurrencyRecord{Calls: tc.CommandsCalled}
	ctx := tc.CheckoutContext
	record.Args = tc.CheckoutArgs
	if tc.PullContext != nil {
		ctx = tc.PullContext
		record.Args = tc.PullArgs
	}

	if ctx != nil {
		limit, limitErr := ctxhelper.Concurrency(ctx, 1)
		if limitErr != nil {
			fmt.Fprintln(os.Stderr, limitErr)
			os.Exit(1)
		}

		record.Limit = limit
		record.Parent = ctx.Value(concurrencyTestKey{}) == true
		record.Verbose = ctxhelper.Verbose(ctx)
	}

	value, marshalErr := json.Marshal(record)
	if marshalErr != nil {
		fmt.Fprintln(os.Stderr, marshalErr)
		os.Exit(1)
	}

	fmt.Fprintf(os.Stdout, "record=%s\n", value)
	if err != nil {
		os.Exit(1)
	}

	os.Exit(0)
}

func concurrencyCLI(t *testing.T, mode string, args ...string) *exec.Cmd {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	t.Cleanup(cancel)
	testArgs := append([]string{"-test.run=^TestConcurrencyCLIProcess$", "--"}, args...)
	command := exec.CommandContext(ctx, os.Args[0], testArgs...)
	command.Env = append(os.Environ(), "ALIGN_CONCURRENCY_CLI_TEST="+mode)
	return command
}

func readConcurrencyRecord(t *testing.T, out string) concurrencyRecord {
	t.Helper()
	index := strings.Index(out, "record=")
	if !assert.GreaterOrEqual(t, index, 0, out) {
		t.FailNow()
	}

	var record concurrencyRecord
	err := json.Unmarshal([]byte(strings.TrimSpace(out[index+len("record="):])), &record)
	assert.NoError(t, err)
	return record
}

func TestConcurrencyFlags(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name  string
		args  []string
		op    string
		limit int
		git   []string
	}{
		{name: "checkout default", args: []string{"-v", "checkout", "target"}, op: "CheckoutRepos", limit: 50, git: []string{"target"}},
		{name: "pull default", args: []string{"-v", "pull"}, op: "PullRepos", limit: 5},
		{name: "checkout sequential", args: []string{"-v", "checkout", "--concurrency=1", "target"}, op: "CheckoutRepos", limit: 1, git: []string{"target"}},
		{name: "pull override and git flags", args: []string{"-v", "pull", "--concurrency", "7", "origin", "main", "--ff-only", "--prune", "--tags"}, op: "PullRepos", limit: 7, git: []string{"origin", "main", "--tags", "--ff-only", "--prune"}},
		{name: "checkout above default", args: []string{"-v", "checkout", "target", "--concurrency", "100"}, op: "CheckoutRepos", limit: 100, git: []string{"target"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			command := concurrencyCLI(t, "capture", tc.args...)
			out, err := command.CombinedOutput()
			assert.NoError(t, err, string(out))
			record := readConcurrencyRecord(t, string(out))
			assert.Equal(t, []string{"GetDirs", tc.op}, record.Calls)
			assert.Equal(t, tc.limit, record.Limit)
			assert.Equal(t, tc.git, record.Args)
			assert.True(t, record.Parent)
			assert.True(t, record.Verbose)
		})
	}
}

func TestConcurrencyInvalidFlags(t *testing.T) {
	t.Parallel()
	for _, op := range []string{"checkout", "pull"} {
		for _, value := range []string{"0", "-1", "bad", "999999999999999999999999999999"} {
			t.Run(op+"/"+value, func(t *testing.T) {
				t.Parallel()
				args := []string{"-v", op, "--concurrency=" + value}
				if op == "checkout" {
					args = append(args, "target")
				}
				command := concurrencyCLI(t, "capture", args...)
				out, err := command.CombinedOutput()
				assert.Error(t, err)
				record := readConcurrencyRecord(t, string(out))
				assert.Empty(t, record.Calls)
				assert.Contains(t, string(out), "concurrency")
			})
		}
	}
	t.Run("not inherited by fetch", func(t *testing.T) {
		t.Parallel()
		out, err := concurrencyCLI(t, "capture", "fetch", "--concurrency=2").CombinedOutput()
		assert.Error(t, err)
		assert.Contains(t, string(out), "unknown flag: --concurrency")
		record := readConcurrencyRecord(t, string(out))
		assert.Empty(t, record.Calls)
	})
}

func TestConcurrencyCLIInterrupt(t *testing.T) {
	t.Parallel()
	if runtime.GOOS == "windows" {
		t.Skip("os.Interrupt delivery to another process is unsupported on Windows")
	}
	for _, op := range []string{"checkout", "pull"} {
		t.Run(op, func(t *testing.T) {
			t.Parallel()
			args := []string{"-v", op}
			if op == "checkout" {
				args = append(args, "target")
			}
			command := concurrencyCLI(t, "block", args...)
			stdout, err := command.StdoutPipe()
			if !assert.NoError(t, err) {
				return
			}
			var stderr bytes.Buffer
			command.Stderr = &stderr
			err = command.Start()
			if !assert.NoError(t, err) {
				return
			}
			reader := bufio.NewReader(stdout)
			line, err := reader.ReadString('\n')
			assert.NoError(t, err)
			assert.Equal(t, "ready\n", line)
			assert.NoError(t, command.Process.Signal(os.Interrupt))
			out, err := io.ReadAll(reader)
			assert.NoError(t, err)
			assert.Error(t, command.Wait())
			assert.Contains(t, stderr.String(), "context canceled")
			record := readConcurrencyRecord(t, string(out))
			assert.True(t, record.Parent)
		})
	}
}

func TestConcurrencyRealCLIProcess(t *testing.T) {
	t.Parallel()
	if os.Getenv("ALIGN_CONCURRENCY_REAL_TEST") != "1" {
		return
	}

	index := slices.Index(os.Args, "--")
	if index < 0 {
		os.Exit(1)
	}
	scrb, err := scribe.NewScribe(os.Stdout, scribe.DefaultTheme)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	clt = &client.Client{Repos: repos.New(scrb, nil, nil)}
	for _, command := range []*cobra.Command{checkoutCmd, pullCmd} {
		command.PersistentPreRun = nil
		command.PersistentPreRunE = func(cmd *cobra.Command, args []string) error { return nil }
	}

	os.Args = append([]string{os.Args[0]}, os.Args[index+1:]...)
	err = Execute()
	if err != nil {
		os.Exit(1)
	}

	os.Exit(0)
}

func TestConcurrentRealPull(t *testing.T) {
	t.Parallel()
	f := newAliasFixture(t)
	source := t.TempDir()
	f.git(t, source, "init", "--quiet", "--initial-branch=main")
	file := filepath.Join(source, "tracked.txt")
	assert.NoError(t, os.WriteFile(file, []byte("initial\n"), 0600))
	f.git(t, source, "add", "tracked.txt")
	f.git(t, source, "-c", "user.name=Fixture", "-c", "user.email=fixture@example.invalid", "-c", "commit.gpgsign=false", "commit", "--quiet", "-m", "initial")
	dirs := []string{f.repo(t, "first"), f.repo(t, "second"), f.repo(t, "ignored")}
	for _, dir := range dirs {
		f.git(t, dir, "remote", "add", "origin", source)
		f.git(t, dir, "fetch", "--quiet", "origin")
		f.git(t, dir, "reset", "--quiet", "--hard", "origin/main")
	}

	assert.NoError(t, os.WriteFile(file, []byte("updated\n"), 0600))
	f.git(t, source, "-c", "user.name=Fixture", "-c", "user.email=fixture@example.invalid", "-c", "commit.gpgsign=false", "commit", "--quiet", "-am", "update")
	command := concurrencyCLI(t, "capture")
	command.Args = []string{os.Args[0], "-test.run=^TestConcurrencyRealCLIProcess$", "--",
		"-v", "--no-color", "pull", "origin", "main", "--ff-only", "--concurrency=2", "--ignore-repo=ignored"}
	command.Dir = f.base
	command.Env = append(f.env, "ALIGN_CONCURRENCY_REAL_TEST=1")
	out, err := command.CombinedOutput()
	assert.NoError(t, err, string(out))
	assert.NotContains(t, string(out), "ignored")
	assert.NotContains(t, string(out), "\x1b[")
	for i, dir := range dirs {
		data, err := os.ReadFile(filepath.Join(dir, "tracked.txt"))
		assert.NoError(t, err)
		want := "updated\n"
		if i == 2 {
			want = "initial\n"
		}
		assert.Equal(t, want, string(data))
	}
}

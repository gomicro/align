package repos

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"testing"
	"time"

	ctxhelper "github.com/gomicro/align/client/context"

	"github.com/gomicro/scribe"
	"github.com/gosuri/uiprogress"
	"github.com/stretchr/testify/assert"
)

type gitStart struct {
	Dir  string
	Args []string
	PID  int
}

type gitReply struct {
	Stdout string
	Stderr string
	Exit   int
	Spawn  bool
}

type driverConfig struct {
	Dirs     []string
	Op       string
	Limit    int
	Override bool
	Canceled bool
	Normal   bool
}

// TestMain lets the test executable double as a controlled Git subprocess.
func TestMain(m *testing.M) {
	name := strings.TrimSuffix(filepath.Base(os.Args[0]), ".exe")
	if name == "git" && os.Getenv("ALIGN_TEST_GIT_ADDRESS") != "" {
		os.Exit(controlledGit())
	}

	os.Exit(m.Run())
}

func controlledGit() int {
	conn, err := net.Dial("tcp", os.Getenv("ALIGN_TEST_GIT_ADDRESS"))
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}

	defer conn.Close()
	dir, err := os.Getwd()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}

	err = json.NewEncoder(conn).Encode(gitStart{Dir: dir, Args: os.Args[1:], PID: os.Getpid()})
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}

	var reply gitReply
	err = json.NewDecoder(conn).Decode(&reply)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}

	fmt.Fprint(os.Stdout, reply.Stdout)
	fmt.Fprint(os.Stderr, reply.Stderr)
	if reply.Spawn {
		child := exec.Command(os.Args[0], "child")
		child.Stdout = os.Stdout
		child.Stderr = os.Stderr
		err = child.Run()
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			return 1
		}
	}

	return reply.Exit
}

func TestConcurrentDriver(t *testing.T) {
	t.Parallel()
	value := os.Getenv("ALIGN_TEST_DRIVER")
	if value == "" {
		return
	}

	var cfg driverConfig
	err := json.Unmarshal([]byte(value), &cfg)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()
	if cfg.Canceled {
		stop()
	}

	if cfg.Override {
		ctx = ctxhelper.WithConcurrency(ctx, cfg.Limit)
	}

	ctx = ctxhelper.WithVerbose(ctx, !cfg.Normal)
	scrb, err := scribe.NewScribe(os.Stdout, scribe.DefaultTheme)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	r := New(scrb, nil, nil)
	if cfg.Normal {
		uiprogress.Start()
	}

	switch cfg.Op {
	case "checkout":
		err = r.CheckoutRepos(ctx, cfg.Dirs, "target")

	case "pull":
		err = r.PullRepos(ctx, cfg.Dirs, "origin", "main", "--ff-only")

	case "alias":
		err = r.RunAlias(ctx, cfg.Dirs, "custom", "--git-option")
	}

	if cfg.Normal {
		uiprogress.Stop()
	}

	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	os.Exit(0)
}

type gitCall struct {
	gitStart
	conn net.Conn
}

type concurrentFixture struct {
	cmd      *exec.Cmd
	listener net.Listener
	calls    chan gitCall
	done     chan struct{}
	stdout   bytes.Buffer
	stderr   bytes.Buffer
	wg       sync.WaitGroup
	waited   bool
	// connMu protects connections retained for cleanup if a test exits early.
	connMu sync.Mutex
	conns  []net.Conn
}

func startConcurrent(t *testing.T, cfg driverConfig) *concurrentFixture {
	t.Helper()
	base := t.TempDir()
	if cfg.Dirs == nil {
		cfg.Dirs = []string{}
	}
	for i := range cfg.Dirs {
		cfg.Dirs[i] = filepath.Join(base, cfg.Dirs[i])
		err := os.Mkdir(cfg.Dirs[i], 0700)
		if !assert.NoError(t, err) {
			t.FailNow()
		}
	}

	gitName := "git"
	if runtime.GOOS == "windows" {
		gitName += ".exe"
	}
	executable, err := os.ReadFile(os.Args[0])
	if !assert.NoError(t, err) {
		t.FailNow()
	}

	err = os.WriteFile(filepath.Join(base, gitName), executable, 0700)
	if !assert.NoError(t, err) {
		t.FailNow()
	}

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if !assert.NoError(t, err) {
		t.FailNow()
	}

	value, err := json.Marshal(cfg)
	if !assert.NoError(t, err) {
		t.FailNow()
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	cmd := exec.CommandContext(ctx, os.Args[0], "-test.run=^TestConcurrentDriver$")
	cmd.Env = append(os.Environ(),
		"PATH="+base+string(os.PathListSeparator)+os.Getenv("PATH"),
		"ALIGN_TEST_GIT_ADDRESS="+listener.Addr().String(),
		"ALIGN_TEST_DRIVER="+string(value))
	f := &concurrentFixture{
		cmd: cmd, listener: listener, calls: make(chan gitCall, len(cfg.Dirs)+1),
		done: make(chan struct{}),
	}
	cmd.Stdout = &f.stdout
	cmd.Stderr = &f.stderr
	err = cmd.Start()
	if !assert.NoError(t, err) {
		cancel()
		t.FailNow()
	}

	f.wg.Add(1)
	go f.acceptCalls()
	t.Cleanup(func() {
		close(f.done)
		assert.NoError(t, listener.Close())
		f.wg.Wait()
		for _, conn := range f.conns {
			err := conn.Close()
			if !errors.Is(err, net.ErrClosed) {
				assert.NoError(t, err)
			}
		}
		cancel()
		if !f.waited {
			err := f.cmd.Wait()
			if err != nil {
				t.Logf("cleaning up driver: %v", err)
			}
		}
	})
	return f
}

func (f *concurrentFixture) acceptCalls() {
	defer f.wg.Done()
	defer close(f.calls)
	for {
		conn, err := f.listener.Accept()
		if err != nil {
			return
		}

		f.connMu.Lock()
		f.conns = append(f.conns, conn)
		f.connMu.Unlock()

		var start gitStart
		err = json.NewDecoder(conn).Decode(&start)
		if err != nil {
			closeErr := conn.Close()
			if closeErr != nil {
				fmt.Fprintln(os.Stderr, closeErr)
			}
			continue
		}

		select {
		case f.calls <- gitCall{gitStart: start, conn: conn}:

		case <-f.done:
			closeErr := conn.Close()
			if closeErr != nil {
				fmt.Fprintln(os.Stderr, closeErr)
			}
			return
		}
	}
}

func (f *concurrentFixture) next(t *testing.T) gitCall {
	t.Helper()
	select {
	case call := <-f.calls:
		if !assert.NotNil(t, call.conn) {
			t.FailNow()
		}
		return call

	case <-time.After(15 * time.Second):
		t.Fatal("timed out waiting for a Git call")
		return gitCall{}
	}
}

func (f *concurrentFixture) noCall(t *testing.T) {
	t.Helper()
	select {
	case call := <-f.calls:
		assert.NoError(t, call.conn.Close())
		t.Fatalf("unexpected extra Git call: %s", call.Dir)

	case <-time.After(100 * time.Millisecond):
	}
}

func releaseGit(t *testing.T, call gitCall, reply gitReply) {
	t.Helper()
	assert.NoError(t, json.NewEncoder(call.conn).Encode(reply))
	assert.NoError(t, call.conn.Close())
}

func (f *concurrentFixture) wait(t *testing.T, wantErr bool) {
	t.Helper()
	err := f.cmd.Wait()
	f.waited = true
	if wantErr {
		assert.Error(t, err)
	} else {
		assert.NoError(t, err, f.stderr.String())
	}
}

func numberedDirs(count int) []string {
	dirs := make([]string, count)
	for i := range dirs {
		dirs[i] = fmt.Sprintf("repo-%03d", i)
	}
	return dirs
}

func TestConcurrentLimits(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name     string
		op       string
		limit    int
		override bool
		count    int
		want     int
	}{
		{name: "checkout default", op: "checkout", count: DefaultCheckoutConcurrency + 2, want: DefaultCheckoutConcurrency},
		{name: "pull default", op: "pull", count: DefaultPullConcurrency + 2, want: DefaultPullConcurrency},
		{name: "explicit checkout pool", op: "checkout", limit: 2, override: true, count: 4, want: 2},
		{name: "explicit pull pool", op: "pull", limit: 2, override: true, count: 4, want: 2},
		{name: "sequential checkout", op: "checkout", limit: 1, override: true, count: 3, want: 1},
		{name: "sequential pull", op: "pull", limit: 1, override: true, count: 3, want: 1},
		{name: "above checkout default", op: "checkout", limit: DefaultCheckoutConcurrency + 1, override: true, count: 3, want: min(3, DefaultCheckoutConcurrency+1)},
		{name: "above pull default", op: "pull", limit: DefaultPullConcurrency + 1, override: true, count: 3, want: min(3, DefaultPullConcurrency+1)},
		{name: "alias stays sequential", op: "alias", limit: 10, override: true, count: 3, want: 1},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			f := startConcurrent(t, driverConfig{
				Op: tc.op, Dirs: numberedDirs(tc.count), Limit: tc.limit, Override: tc.override,
			})
			active := make([]gitCall, tc.want)
			seen := make(map[string]bool)
			for i := range active {
				active[i] = f.next(t)
				assert.False(t, seen[active[i].Dir])
				seen[active[i].Dir] = true
			}
			f.noCall(t)

			// Keep all other slots blocked: one release must start the next job.
			for i := tc.want; i < tc.count; i++ {
				releaseGit(t, active[0], gitReply{})
				active[0] = f.next(t)
				assert.False(t, seen[active[0].Dir])
				seen[active[0].Dir] = true
				f.noCall(t)
			}
			for _, call := range active {
				if tc.op == "pull" {
					assert.Equal(t, []string{"pull", "origin", "main", "--ff-only"}, call.Args)
				}
				releaseGit(t, call, gitReply{})
			}
			f.wait(t, false)
			assert.Len(t, seen, tc.count)
		})
	}
}

func TestConcurrentOutputAndFailures(t *testing.T) {
	t.Parallel()
	f := startConcurrent(t, driverConfig{
		Op: "pull", Dirs: numberedDirs(4), Override: true, Limit: 2,
	})
	slow := f.next(t)
	second := f.next(t)
	releaseGit(t, second, gitReply{Stdout: "stdout-second\n", Stderr: "stderr-second\n", Exit: 1})
	third := f.next(t)
	releaseGit(t, third, gitReply{Stdout: "stdout-third\n", Stderr: "stderr-third\n", Exit: 1})
	fourth := f.next(t)
	releaseGit(t, fourth, gitReply{Stdout: "stdout-fourth\n", Stderr: "stderr-fourth\n"})
	releaseGit(t, slow, gitReply{Stdout: "stdout-slow\n", Stderr: "stderr-slow\n"})
	f.wait(t, true)
	out := f.stdout.String()
	secondAt := strings.Index(out, "stdout-second")
	thirdAt := strings.Index(out, "stdout-third")
	assert.GreaterOrEqual(t, secondAt, 0)
	assert.Greater(t, thirdAt, secondAt)
	secondStderrAt := strings.Index(out, "stderr-second")
	assert.Greater(t, secondStderrAt, secondAt)
	assert.Less(t, secondStderrAt, thirdAt)
	secondHeaderAt := strings.Index(out, second.Dir)
	thirdHeaderAt := strings.Index(out, third.Dir)
	assert.GreaterOrEqual(t, secondHeaderAt, 0)
	assert.Less(t, secondHeaderAt, secondAt)
	assert.Greater(t, thirdHeaderAt, secondStderrAt)
	assert.Less(t, thirdHeaderAt, thirdAt)
	for _, name := range []string{"second", "third", "fourth", "slow"} {
		assert.Contains(t, out, "stdout-"+name+"\n")
		assert.Contains(t, out, "stderr-"+name+"\n")
	}
	assert.Contains(t, f.stderr.String(), second.Dir)
	assert.Contains(t, f.stderr.String(), third.Dir)
	assert.Contains(t, f.stderr.String(), "stderr-second")
	assert.Contains(t, f.stderr.String(), "stderr-third")
}

func TestConcurrentEmptyInvalidAndCanceled(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name     string
		cfg      driverConfig
		wantErr  bool
		contains string
	}{
		{name: "empty", cfg: driverConfig{Op: "checkout"}},
		{name: "zero", cfg: driverConfig{Op: "pull", Override: true}, wantErr: true, contains: "positive integer"},
		{name: "negative", cfg: driverConfig{Op: "checkout", Override: true, Limit: -2}, wantErr: true, contains: "positive integer"},
		{name: "pre-canceled", cfg: driverConfig{Op: "pull", Canceled: true}, wantErr: true, contains: "context canceled"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if tc.wantErr {
				tc.cfg.Dirs = numberedDirs(2)
			}
			f := startConcurrent(t, tc.cfg)
			f.wait(t, tc.wantErr)
			f.noCall(t)
			if tc.contains != "" {
				assert.Contains(t, f.stderr.String(), tc.contains)
			}
		})
	}
}

func TestConcurrentInterrupt(t *testing.T) {
	t.Parallel()
	if runtime.GOOS == "windows" {
		t.Skip("os.Interrupt delivery to another process is unsupported on Windows")
	}
	f := startConcurrent(t, driverConfig{
		Op: "pull", Dirs: numberedDirs(6), Override: true, Limit: 2, Normal: true,
	})
	first := f.next(t)
	second := f.next(t)
	f.noCall(t)
	assert.NoError(t, f.cmd.Process.Signal(os.Interrupt))
	f.wait(t, true)
	assert.Contains(t, f.stderr.String(), "context canceled")
	assert.Contains(t, f.stdout.String(), "Pulling (2/6)")
	assert.NotContains(t, f.stdout.String(), "(6/6)")
	for _, call := range []gitCall{first, second} {
		assert.NoError(t, call.conn.SetReadDeadline(time.Now().Add(time.Second)))
		var reply gitReply
		err := json.NewDecoder(call.conn).Decode(&reply)
		assert.ErrorIs(t, err, io.EOF, "canceled Git process should close its connection")
		assert.NoError(t, call.conn.Close())
	}
	f.noCall(t)
}

func TestConcurrentInterruptDescendants(t *testing.T) {
	t.Parallel()
	if runtime.GOOS == "windows" {
		t.Skip("os.Interrupt delivery to another process is unsupported on Windows")
	}
	f := startConcurrent(t, driverConfig{
		Op: "checkout", Dirs: numberedDirs(3), Override: true, Limit: 1,
	})
	parent := f.next(t)
	assert.NoError(t, json.NewEncoder(parent.conn).Encode(gitReply{Spawn: true}))
	child := f.next(t)
	assert.Equal(t, []string{"child"}, child.Args)
	assert.NoError(t, f.cmd.Process.Signal(os.Interrupt))
	f.wait(t, true)
	assert.Contains(t, f.stderr.String(), "context canceled")
	for _, call := range []gitCall{parent, child} {
		assert.NoError(t, call.conn.SetReadDeadline(time.Now().Add(time.Second)))
		var reply gitReply
		err := json.NewDecoder(call.conn).Decode(&reply)
		assert.ErrorIs(t, err, io.EOF)
		assert.NoError(t, call.conn.Close())
	}
	f.noCall(t)
}

func TestConcurrentProgress(t *testing.T) {
	t.Parallel()
	f := startConcurrent(t, driverConfig{
		Op: "pull", Dirs: numberedDirs(3), Override: true, Limit: 2, Normal: true,
	})
	first := f.next(t)
	second := f.next(t)
	releaseGit(t, second, gitReply{Stderr: "failed-repo\n", Exit: 1})
	third := f.next(t)
	releaseGit(t, third, gitReply{Stdout: "successful-output\n"})
	releaseGit(t, first, gitReply{})
	f.wait(t, true)
	assert.Contains(t, f.stdout.String(), "Pulling (3/3)")
	assert.NotContains(t, f.stdout.String(), "Current Repo")
	assert.NotContains(t, f.stdout.String(), "successful-output")
	assert.Contains(t, f.stderr.String(), "failed-repo")
}

func TestConcurrentCanceledError(t *testing.T) {
	t.Parallel()
	var output bytes.Buffer
	scrb, err := scribe.NewScribe(&output, scribe.DefaultTheme)
	if !assert.NoError(t, err) {
		return
	}

	r := New(scrb, nil, nil)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	err = r.CheckoutRepos(ctx, []string{t.TempDir()}, "target")
	assert.ErrorIs(t, err, context.Canceled)
	err = r.PullRepos(ctx, []string{t.TempDir()})
	assert.ErrorIs(t, err, context.Canceled)
	assert.Empty(t, output.String())
}

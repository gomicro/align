// Package fanout runs Git jobs across repositories in a bounded rolling worker
// pool. Workers own each job's output buffers; only the calling goroutine writes
// scribe output and advances progress, in completion order.
package fanout

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os/exec"
	"slices"
	"strings"
	"sync"
	"time"

	ctxhelper "github.com/gomicro/align/client/context"

	"github.com/gomicro/scribe"
	"github.com/gosuri/uiprogress"
)

// Job is a single Git invocation. Dir is the working directory, empty for the
// current one; Name identifies the job in output and errors and defaults to Dir.
type Job struct {
	Dir  string
	Name string
	Args []string
}

// Result holds the captured output of a Job. Index is the job's position in
// the slice passed to Run.
type Result struct {
	Index  int
	Job    Job
	Stdout bytes.Buffer
	Stderr bytes.Buffer
	Err    error
}

// Jobs builds one job per directory, all running the same arguments.
func Jobs(dirs []string, args []string) []Job {
	jobs := make([]Job, len(dirs))
	for i, dir := range dirs {
		jobs[i] = Job{Dir: dir, Args: args}
	}

	return jobs
}

// Run executes jobs with at most limit Git processes active at once. Failures
// do not stop other jobs; they are joined into the returned error. Cancellation
// kills active Git process trees and skips queued jobs. Results are returned in
// completion order.
func Run(ctx context.Context, scrb scribe.Scriber, label string, jobs []Job, limit int) ([]*Result, error) {
	if ctx.Err() != nil {
		return nil, fmt.Errorf("running repositories: %w", ctx.Err())
	}

	count := len(jobs)
	if count == 0 {
		return nil, nil
	}

	if limit < 1 {
		return nil, fmt.Errorf("concurrency must be a positive integer")
	}

	limit = min(limit, count)
	shared := sharedArgs(jobs)
	verbose := ctxhelper.Verbose(ctx)
	var bar *uiprogress.Bar
	if verbose {
		header := label
		if shared {
			header = commandLine(jobs[0].Args)
		}

		scrb.BeginDescribe("Command")
		defer scrb.EndDescribe()

		scrb.Print(header)

		scrb.BeginDescribe("directories")
		defer scrb.EndDescribe()
	} else {
		bar = uiprogress.AddBar(count).
			AppendCompleted().
			PrependElapsed().
			PrependFunc(func(b *uiprogress.Bar) string {
				return fmt.Sprintf("%s (%d/%d)", label, b.Current(), count)
			})
	}

	queue := make(chan int)
	finished := make(chan *Result, limit)
	var workers sync.WaitGroup
	workers.Add(limit + 1)
	go submit(ctx, count, queue, &workers)
	for range limit {
		go work(ctx, jobs, queue, finished, &workers)
	}

	go func() {
		workers.Wait()
		close(finished)
	}()

	results := make([]*Result, 0, count)
	var errs []error
	for result := range finished {
		results = append(results, result)
		if result.Err != nil {
			errs = append(errs, result.failure())
		}

		if verbose {
			result.print(scrb, !shared)
		} else {
			bar.Incr()
		}
	}

	if ctx.Err() != nil {
		errs = append(errs, fmt.Errorf("running repositories: %w", ctx.Err()))
	}

	return results, errors.Join(errs...)
}

func submit(ctx context.Context, count int, queue chan<- int, workers *sync.WaitGroup) {
	defer workers.Done()
	defer close(queue)

	for i := range count {
		if ctx.Err() != nil {
			return
		}

		select {
		case <-ctx.Done():
			return

		case queue <- i:
		}
	}
}

func work(ctx context.Context, jobs []Job, queue <-chan int, finished chan<- *Result, workers *sync.WaitGroup) {
	defer workers.Done()

	for i := range queue {
		if ctx.Err() != nil {
			return
		}

		finished <- runGit(ctx, i, jobs[i])
	}
}

func runGit(ctx context.Context, index int, job Job) *Result {
	result := &Result{Index: index, Job: job}
	cmd := exec.CommandContext(ctx, "git", job.Args...)
	cmd.Stdout = &result.Stdout
	cmd.Stderr = &result.Stderr
	cmd.Dir = job.Dir
	cmd.WaitDelay = time.Second
	configureCancellation(cmd)

	result.Err = cmd.Run()
	return result
}

func sharedArgs(jobs []Job) bool {
	for i := range jobs {
		if !slices.Equal(jobs[i].Args, jobs[0].Args) {
			return false
		}
	}

	return true
}

func commandLine(args []string) string {
	return fmt.Sprintf("git %s", strings.Join(args, " "))
}

func (j Job) name() string {
	if j.Name != "" {
		return j.Name
	}

	return j.Dir
}

func (r *Result) failure() error {
	detail := r.Job.name()
	stderr := strings.TrimSpace(r.Stderr.String())
	if stderr != "" {
		detail += ": " + stderr
	}

	return fmt.Errorf("%s: %w", detail, r.Err)
}

func (r *Result) print(scrb scribe.Scriber, withCommand bool) {
	scrb.BeginDescribe(r.Job.name())
	if withCommand {
		scrb.Print(commandLine(r.Job.Args))
	}

	if r.Err != nil {
		scrb.Error(r.Err)
	}

	scrb.PrintLines(&r.Stdout)
	scrb.PrintLines(&r.Stderr)
	scrb.EndDescribe()
}

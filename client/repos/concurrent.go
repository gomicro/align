package repos

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"

	ctxhelper "github.com/gomicro/align/client/context"

	"github.com/gosuri/uiprogress"
)

func (r *Repos) fanOutConcurrent(ctx context.Context, dirs []string, label string, args []string, limit int) error {
	if ctx.Err() != nil {
		return fmt.Errorf("running repositories: %w", ctx.Err())
	}

	count := len(dirs)
	if count == 0 {
		return nil
	}

	limit = min(limit, count)
	verbose := ctxhelper.Verbose(ctx)
	var bar *uiprogress.Bar
	if verbose {
		r.scrb.BeginDescribe("Command")
		defer r.scrb.EndDescribe()

		r.scrb.Print(fmt.Sprintf("git %s", strings.Join(args, " ")))

		r.scrb.BeginDescribe("directories")
		defer r.scrb.EndDescribe()
	} else {
		bar = uiprogress.AddBar(count).
			AppendCompleted().
			PrependElapsed().
			PrependFunc(func(b *uiprogress.Bar) string {
				return fmt.Sprintf("%s (%d/%d)", label, b.Current(), count)
			})
	}

	jobs := make(chan string)
	results := make(chan *gitResult, limit)
	var workers sync.WaitGroup
	workers.Add(limit + 1)
	go submitRepos(ctx, dirs, jobs, &workers)
	for i := 0; i < limit; i++ {
		go runRepos(ctx, args, jobs, results, &workers)
	}

	go func() {
		workers.Wait()
		close(results)
	}()

	var errs []error
	for result := range results {
		if result.err != nil {
			errs = append(errs, result.failure())
		}

		if verbose {
			r.printResult(result)
		} else {
			bar.Incr()
		}
	}

	if ctx.Err() != nil {
		errs = append(errs, fmt.Errorf("running repositories: %w", ctx.Err()))
	}

	return errors.Join(errs...)
}

func submitRepos(ctx context.Context, dirs []string, jobs chan<- string, workers *sync.WaitGroup) {
	defer workers.Done()
	defer close(jobs)

	for _, dir := range dirs {
		if ctx.Err() != nil {
			return
		}

		select {
		case <-ctx.Done():
			return

		case jobs <- dir:
		}
	}
}

func runRepos(ctx context.Context, args []string, jobs <-chan string, results chan<- *gitResult, workers *sync.WaitGroup) {
	defer workers.Done()

	for dir := range jobs {
		if ctx.Err() != nil {
			return
		}

		result := runGit(ctx, dir, args, true)
		results <- result
	}
}

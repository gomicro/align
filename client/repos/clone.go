package repos

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path"

	ctxhelper "github.com/gomicro/align/client/context"
	"github.com/gomicro/align/client/fanout"
)

// CloneRepos clones every repository in the context's repo map that is not
// already present, returning those cloned successfully.
func (r *Repos) CloneRepos(ctx context.Context, baseDir string) ([]*ctxhelper.Repository, error) {
	limit, err := ctxhelper.Concurrency(ctx, DefaultCloneConcurrency)
	if err != nil {
		return nil, fmt.Errorf("resolving concurrency: %w", err)
	}

	dirRepos, err := ctxhelper.RepoMap(ctx)
	if err != nil {
		return nil, fmt.Errorf("repomap context: %w", err)
	}

	dirRepos, err = ctxhelper.RemoveExcludes(ctx, dirRepos)
	if err != nil {
		return nil, fmt.Errorf("remove excludes: %w", err)
	}

	var pending []*ctxhelper.Repository
	var jobs []fanout.Job
	var errs []error
	seen := map[string]bool{}
	for dir, rs := range dirRepos {
		if baseDir != "" {
			dir = baseDir
		}

		for i := range rs {
			dest := path.Join(".", dir, rs[i].Name)
			if seen[dest] {
				continue
			}

			seen[dest] = true

			exists, err := cloneExists(dest)
			if err != nil {
				errs = append(errs, err)
				continue
			}

			if exists {
				continue
			}

			pending = append(pending, rs[i])
			jobs = append(jobs, fanout.Job{Name: dest, Args: []string{"clone", rs[i].URL, dest}})
		}
	}

	results, err := fanout.Run(ctx, r.scrb, "Cloning", jobs, limit)
	if err != nil {
		errs = append(errs, err)
	}

	cloned := []*ctxhelper.Repository{}
	for _, result := range results {
		if result.Err == nil {
			cloned = append(cloned, pending[result.Index])
		}
	}

	return cloned, errors.Join(errs...)
}

// cloneExists reports whether dest already holds a Git repository.
func cloneExists(dest string) (bool, error) {
	exists, err := dirExists(dest)
	if err != nil {
		return false, fmt.Errorf("repo dir %s: %w", dest, err)
	}

	if !exists {
		return false, nil
	}

	exists, err = dirExists(path.Join(dest, ".git"))
	if err != nil {
		return false, fmt.Errorf("repo git dir %s: %w", dest, err)
	}

	return exists, nil
}

func dirExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}

	if os.IsNotExist(err) {
		return false, nil
	}

	return false, fmt.Errorf("exists check: %w", err)
}

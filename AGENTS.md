# Agent Instructions

## Overview

`align` is a Go CLI tool that fans out `git` commands across every repository in
a directory, letting a collection of repos be managed as a unit. Its interface
mirrors `git` exactly — same flags, same arguments — but operates on all `.git`
subdirectories at once. Example: `align push origin main` runs `git push origin
main` in every repo under the current directory.

## Tech Stack

- **Go 1.26**
- `cobra` — CLI framework
- `go-git` — used only for `clone`; all other operations shell out to `exec git`
- `scribe` — structured output / verbose logging
- `uiprogress` — progress bars in non-verbose mode
- `go-github` + `oauth2` + `rate` — GitHub API for clone and OAuth auth
- `golang.org/x/crypto/ssh/knownhosts` — SSH host key verification
- `gopkg.in/yaml.v2` — config file parsing
- `testify` — test assertions

## Build & Validate

```sh
go fmt ./...
go vet ./...
go build ./...
go test ./...
```

CI also runs `golangci-lint`. Releases use `goreleaser` (see `forge.yaml`).
The binary requires `ALIGN_CLIENT_ID` and `ALIGN_CLIENT_SECRET` set at build
time for the GitHub OAuth flow.

## Repository Layout

```
main.go              Entry point — calls cmd.Execute()
cmd/                 Cobra command definitions (one file per command)
  config/            align config subcommand
  remote/            align remote subcommands (add, remove, rename, set-url)
  stash/             align stash subcommands (pop, drop, list)
client/              Business logic layer called by cmd/
  client.go          Client struct, SSH auth, GitHub client construction
  clienter.go        Clienter interface (used for testing)
  dirs.go            GetDirs: discovers .git subdirs in a base path
  context/           Request-scoped state (verbose flag, repo map, excludes)
  repos/             Per-repo git operations (fanOut pattern)
    run.go           fanOut(ctx, dirs, label, args) — shared exec loop
    clone.go         CloneRepos: uses go-git, not exec
    status.go        StatusRepos: filters empty output, always verbose
    diff.go          DiffRepos: DiffConfig for filtering
  remotes/           Remote management operations
    run.go           fanOut variant with per-directory args slice
  testclient/        Fake client for cmd/ tests
config/              Config file parsing (~/.align/config)
vendor/              Vendored deps — never edit directly
```

## Architecture Notes

- **`fanOut` is the core pattern.** Every standard git operation assembles
  `args []string` and calls `repos.fanOut(ctx, dirs, label, args)`. It never
  aborts on first failure — all repos are attempted and errors are joined.
- **Non-verbose vs verbose:** non-verbose writes a `uiprogress` bar to stdout
  (not pipeable); verbose uses `scribe` for structured per-repo output. Always
  pass `--verbose` / `-v` when running non-interactively or in scripts.
- **Remote operations** use a parallel fanOut variant that accepts
  `perDirArgs [][]string` instead of a shared `args []string`.
- **Context package** (`client/context/`) carries the verbose flag, GitHub repo
  map, and exclude list. Use `ctxhelper.WithVerbose` / `ctxhelper.Verbose`;
  never pass verbose as an explicit parameter.
- **Shared cmd/ package-level vars:** `tags` declared in `pull.go` reused by
  fetch/push; `all`, `force`, `noColor` declared in push/diff; `short`,
  `ignoreEmpty` declared in `diff.go` reused in status.
- **SSH auth:** reads `~/.ssh/known_hosts` via `knownHostsCallback()`. Never use
  `ssh.InsecureIgnoreHostKey()`. Supports both inline PEM (`private_key`) and
  file path (`private_key_file`).

## Key Conventions

- **Error wrapping:** `fmt.Errorf("methodName: operation: %w", err)` — lowercase,
  no period. Top-level package functions omit the method segment.
- **`init()` only** for cobra flag registration — no logic or side effects.
- **No blank identifier:** use `for i := range` not `for i, _ := range`; never
  silently discard errors with `_`.
- **No callbacks:** prefer explicit data (e.g., `[][]string`) over function
  arguments passed as callbacks.
- **No nested control structures:** use early returns and guard clauses.
- **Sequential by default:** no goroutines unless there is a demonstrated need.
- **Vendor is read-only:** modify only via `go mod vendor`.
- **After any Go change:** run `go fmt`, `go vet`, `go build ./...`,
  `go test ./...` in that order before committing.

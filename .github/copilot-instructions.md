# align — Copilot Instructions

## Overview

`align` is a Go CLI tool that fans out git commands across a directory of git repositories,
allowing a set of repos to be managed together as a unit. Every `align` command mirrors its
`git` counterpart in interface — same flags, same positional arguments, same mental model — but
operates across all git repos found in the current directory.

Example: `align push origin main` behaves like running `git push origin main` in every repo
under the current directory.

## Tech Stack

- Go 1.26
- [cobra](https://github.com/spf13/cobra) — CLI framework
- [go-git](https://github.com/go-git/go-git) — used only for `clone` (all other ops use `exec git`)
- [scribe](https://github.com/gomicro/scribe) — structured output / verbose logging
- [uiprogress](https://github.com/gosuri/uiprogress) — progress bars in non-verbose mode
- [go-github](https://github.com/google/go-github) — GitHub API (clone, auth)
- [oauth2](https://golang.org/x/oauth2) + [rate](https://golang.org/x/time/rate) — GitHub API auth and rate limiting
- [golang.org/x/crypto/ssh/knownhosts](https://pkg.go.dev/golang.org/x/crypto/ssh/knownhosts) — SSH host key verification

## Build & Validate

```sh
go build ./...
go vet ./...
go test ./...
```

CI also runs `golangci-lint`. Releases use goreleaser (see `forge.yaml` for local build/install
with injected ldflags). The binary requires `ALIGN_CLIENT_ID` and `ALIGN_CLIENT_SECRET` env
vars to be set at build time for the GitHub OAuth flow.

## Repository Layout

```
main.go           Entry point; calls cmd.Execute()
cmd/              Cobra command definitions (one file per command)
  config/         align config subcommand
  remote/         align remote subcommands (add, remove, rename, set-url)
  stash/          align stash subcommands (pop, drop, list)
client/           Business logic layer; called by cmd/
  client.go       Client struct, SSH auth, GitHub client construction
  clienter.go     Clienter interface (used for testing)
  dirs.go         GetDirs: discovers .git subdirs in a base path
  context/        Context helpers: verbose flag, repo map, excludes list
  repos/          Per-repo git operations (fanOut pattern)
    run.go        fanOut(ctx, dirs, label, args) — shared exec loop
    clone.go      CloneRepos: uses go-git, not exec
    status.go     StatusRepos: filters empty output, always verbose
    diff.go       DiffRepos: has DiffConfig for filtering
  remotes/        Remote management operations
    run.go        fanOut variant with per-directory args slice
  testclient/     Fake client for cmd/ tests
config/           Config file parsing (~/.align/config)
  file.go         Config struct + WriteFile
  github.go       GithubHost sub-struct (token, username, keys, limits)
  parse.go        ParseFromFile
vendor/           Vendored dependencies (never edit directly)
```

## Architecture Notes

### fanOut — the core pattern
All standard git operations are one-liners that assemble `args` and call `fanOut`:
```go
func (r *Repos) fanOut(ctx context.Context, dirs []string, label string, args []string) error
```
- Non-verbose: shows a `uiprogress` bar; errors are accumulated, not surfaced inline
- Verbose: uses `scribe` to print structured per-repo output to stdout
- Never aborts on first failure — all repos are attempted; all errors returned via `errors.Join`

Remote operations use a variant with per-directory args:
```go
func (r *Remotes) fanOut(ctx context.Context, dirs []string, label string, perDirArgs [][]string) error
```

### SSH authentication
- `knownHostsCallback()` reads `~/.ssh/known_hosts` via `golang.org/x/crypto/ssh/knownhosts`
- Never use `ssh.InsecureIgnoreHostKey()` — it does not exist in this codebase for a reason
- SSH auth supports both inline PEM (`private_key`) and file path (`private_key_file`)

### Context package (`client/context/`)
Carries request-scoped state through the call chain: verbose flag, GitHub repo map (for clone),
and exclude list. Use `ctxhelper.WithVerbose` / `ctxhelper.Verbose` — do not pass verbose as a
parameter.

### Shared package-level vars (cmd/ package)
Several flags are declared in one file and referenced in others:
- `tags bool` — declared in `pull.go`, reused by fetch/push
- `all bool`, `force bool`, `noColor bool` — declared in push/diff, reused elsewhere
- `short bool`, `ignoreEmpty bool` — declared in diff.go, reused in status

### Tab completion
Implemented via cobra's `ValidArgsFunction`. Completion functions call into the `client` layer
to enumerate branch/tag/remote names from the repos in the current directory at completion time.

## Config File

`~/.align/config` (YAML, read with `gopkg.in/yaml.v2`):
```yaml
github.com:
  token: <oauth token>
  username: <github username>
  private_key: <optional inline SSH private key PEM>
  private_key_file: <optional path to SSH private key file>
  limits:
    request_per_second: 10
    burst: 25
```

## Key Conventions

- **Error wrapping**: `fmt.Errorf("methodName: operation: %w", err)` — lowercase, no period; top-level package functions omit the method segment
- **`init()` only for flag registration**: never use `init()` for side effects or logic
- **No blank identifier**: use `for i := range` not `for i, _ := range`; never silently discard errors
- **No callbacks**: prefer explicit data (e.g., `[][]string`) over function arguments
- **No nested control structures**: prefer early returns and guard clauses
- **Sequential by default**: no goroutines unless there is a demonstrated need
- **Vendor is read-only**: `go mod vendor` is the only permitted way to modify `vendor/`
- **After any Go change**: run `go fmt`, `go vet`, `go build ./...`, `go test ./...` in that order

## Scripting and AI Usage

Always use `--verbose` (`-v`) when running `align` non-interactively. Non-verbose mode writes
progress bars to stdout, which is not suitable for piping or parsing.

```sh
align status --verbose
align push origin main --verbose
```

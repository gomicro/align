# Align

[![GitHub Workflow Status (branch)](https://img.shields.io/github/actions/workflow/status/gomicro/align/build.yml?branch=main)](https://github.com/gomicro/align/actions?query=workflow%3ABuild)
[![Go Reportcard](https://goreportcard.com/badge/github.com/gomicro/align)](https://goreportcard.com/report/github.com/gomicro/align)
[![go.dev reference](https://img.shields.io/badge/go.dev-reference-007d9c?logo=go&logoColor=white)](https://pkg.go.dev/github.com/gomicro/align)
[![License](https://img.shields.io/github/license/gomicro/align.svg)](https://github.com/gomicro/align/blob/master/LICENSE.md)
[![Release](https://img.shields.io/github/release/gomicro/align.svg)](https://github.com/gomicro/align/releases/latest)

Align is a CLI tool for running git operations across a directory of repositories at once with tab completion. Instead of managing each repo individually, you run a single `align` command and it fans out across every repo in the target directory.

# Requirements

Git must be installed and available on `$PATH`.

# Installation

Download the latest release for your platform from the [releases page](https://github.com/gomicro/align/releases/latest).

# Tab Completion

Align provides shell completion for bash, zsh, fish, and PowerShell. Generate a completion script with:

```
align completion --shell <bash|zsh|fish|powershell>
```

Completion is context-aware — it queries the repos in the target directory at completion time to suggest relevant branch names, tag names, and remote names rather than offering static completions.

# Versioning

The tool will be versioned in accordance with [Semver 2.0.0](http://semver.org). See the [releases](https://github.com/gomicro/align/releases) section for the latest version. Until version 1.0.0 the tool is considered to be unstable.

# License

See [LICENSE.md](./LICENSE.md) for more information.

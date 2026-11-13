---
name: align
description: 'Use align for shared Git operations across sibling repositories. Use when asked to inspect, update, compare, or manage several repositories together, or replace repetitive per-repository Git loops. Establish target scope, choose noninteractive execution, and handle partial failures safely.'
---

# Align

<WhenToUse>
Prefer align when the same Git operation and arguments apply to several sibling
repositories. Use direct Git for a single repository, different per-repository
arguments, or operations the installed align command does not support.
Do not install align or change repository configuration without authorization.
</WhenToUse>

<EstablishScope>
1. Check that align is available. Read `align --help` and the relevant
   subcommand's help for the installed version's syntax and limitations.
2. Identify the parent directory and enumerate its immediate child repositories.
   Align discovers children with a `.git` entry; it does not recurse or operate
   on the parent repository itself.
3. Compare that set with the user's intended targets before running a mutation.
   Use the exclusion rules in align's help only when the remaining set is exactly
   the intended set. Otherwise use explicit per-repository Git operations.
4. Inspect working-tree and branch state before mutations. Account for differing
   branch names, remotes, and local changes; a shared command must fit every
   selected repository. Respect the user's approval and destructive-action rules.
</EstablishScope>

<Execute>
- Use the installed help's verbose and no-color options for noninteractive work.
  Align's ordinary progress display is not suitable for scripts. Verbose output
  is buffered, human-oriented per-repository text, not JSON or a live terminal.
- Treat command help as authoritative. Do not assume every Git flag passes
  through or invent a recursive, include-only, dry-run, or target-list option.
- Before running an alias, inspect its definition and each target repository's
  effective configuration. Global/system discovery does not make an alias safe;
  local overrides and shell aliases can change its behavior.
- Put align options before an alias name. All subsequent tokens belong to Git.
  Consult root help for discovery and precedence rules.
- Avoid commands requiring prompts, editors, pagers, or terminal UIs. If
  noninteractive authentication or input is unavailable, stop and report it
  rather than hanging or changing credentials.
</Execute>

<HandleResults>
Inspect the exit status and each repository's output. A failure may follow
successful mutations in other repositories; there is no fleet-wide rollback.
Report which repositories succeeded, failed, or were not attempted when known.
Do not infer success from a progress indicator or retry the entire fleet blindly.
Inspect failed or uncertain repositories first, then retry only the necessary
operations with direct Git or a verified selection. Check resulting state after
mutations, even if the command returned success.
</HandleResults>

<UpdateProtocol>
If asked to change these conventions, propose the exact change and whether it is
required or conditional, then obtain confirmation before editing the source
skill. Update the versioned align source rather than an exported copy; exported
edits can be replaced by a later dump.
</UpdateProtocol>

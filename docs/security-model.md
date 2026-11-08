# Security model

## Authority

Core policies define intended boundaries.
Native harness permissions enforce the boundaries where supported.
Prompts do not provide security isolation.

## Installation

AWM writes only beneath the selected home directory.
It rejects target symlinks and parent-symlink escapes.
It writes files atomically.
It backs up replaced files with restricted permissions.
It records hashes for rollback checks.
It preserves unmanaged guidance outside AWM markers.
It asks before replacing unmanaged agent files.

## Builder scope

OpenCode can constrain edit paths through native permission rules.
Shell commands still require Human approval.
Claude Code hooks validate Write and Edit targets.
Shell commands can still have side effects outside the file allowlist.
Codex and OMP cannot enforce the same exact per-file boundary.
Use a separate approved session and disposable worktree for those adapters.

## Read scope

OpenCode denies common secret path patterns for Read.
OpenCode asks before every Grep query because Grep rules match query text, not paths.
Claude Code blocks common sensitive paths and broad Grep or Glob requests.
The Claude read hook checks requested paths, not every file below a directory.
Other harnesses depend on native sandbox settings and explicit policies.
These controls reduce exposure but do not create a complete data-loss boundary.
Do not expose credentials to agents.

## Third-party skills and plugins

AWM installs only bundled skills by default.
It does not install plugins.
It does not download unpinned skill sources.
It reports missing external capabilities and explains the blocker.

## Memory

AWM audits memory integration paths without changing settings.
It does not alter capture mode or auto-improve scheduling.
It does not delete, merge, or promote saved memories.
Human approval controls durable memory changes.

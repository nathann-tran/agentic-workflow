# Claude Code adapter

AWM writes global agents under `~/.claude/agents/`.
It adds a managed block to `~/.claude/CLAUDE.md`.
The installer preserves text outside the managed block.

The Worker receives only Read, Grep, and Glob tools.
A pre-tool hook blocks common sensitive paths and broad searches.
Grep must specify a narrow path.
Glob must specify a narrow pattern.
This guard checks requested paths, not every file below a directory.
It remains best-effort and cannot replace OS-level isolation.

Write/Edit hooks enforce role output roots.
Task-scoped Builder hooks use an exact repository-relative file allowlist.
Bash may still alter files outside that list.
Use a disposable worktree and approve every consequential command.

Review hook availability and event payloads after installing the current Claude Code release.

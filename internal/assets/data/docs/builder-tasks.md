# Task-scoped Builder

Builder is the only role that may edit production source and tests.
A base Builder profile cannot write.
Create a task profile only after Human approval.

The task envelope must use schema version `1.0`.
It must assign `assignee_role` to `builder`.
It must set `constraints.source_write_allowed` to `true`.
It must set `constraints.read_only` to `false` or omit it.
It must provide exact repository-relative paths in `scope.write_allowlist`.
Do not use globs, parent traversal, absolute paths, or directory entries.

Example:

```sh
awm builder render \\
  --task examples/task-envelope-builder.json \\
  --harness opencode \\
  --human-approved
```

OpenCode and Claude Code receive task-scoped profiles.
Claude Code adds native worktree isolation for the task-scoped Builder.
Codex and OMP use a separate approved session and an isolated worktree.
The base Builder stays read-only on every adapter.
Human approval does not remove shell side-effect risks.

AWM records a task manifest with installed hashes and original-file backups.
Restore the profile and allowlist with this command:

```sh
awm builder rollback --task-id TASK-001
```

Rollback skips files changed after profile creation.
Use `--force` only after reviewing those changes.
Rollback the previous task before rendering another profile with the same task ID.

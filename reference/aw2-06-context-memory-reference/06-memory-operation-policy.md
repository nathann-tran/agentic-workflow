# AW2-06 — Agent Memory Operation Policy

## 1. Access boundaries

All agents may read memory within the task's authorized workspace and project scope.

Conductor owns durable memory curation.

Specialists return memory candidates in their task result.

Workers receive only the read-only memory capability required by their task.

## 2. Tool policy

| Tool family | Conductor | Plan / Scout / Architect / Builder | Worker |
|---|---|---|---|
| `memory_query`, `memory_read_page`, `memory_recent` | Allow | Read-only, scoped | Only when task requires it |
| `memory_briefing`, `memory_explore`, `memory_status` | Allow | Scoped read-only as needed | Normally deny |
| `memory_handoff_*` | Allow for session continuity | Deny by default | Deny |
| `memory_message_*` | Cross-project coordination | Deny by default | Deny |
| `memory_auto_improve` | Review proposals under Human control | Deny | Deny |
| `memory_write_page` | Human-gated durable promotion | Deny | Deny |
| `memory_feedback` | Evidence-based curation | Candidate only | Deny |
| `memory_lint` | Maintenance | Deny | Deny |
| `memory_forget_sweep` | Human-gated, dry-run first | Deny | Deny |
| `memory_delete_page`, `memory_restore_page` | Exact-path Human approval | Deny | Deny |
| `memory_read_session_observations` | Exceptional recovery only | Deny | Deny |

This is the logical policy.

Map it to each harness's actual tool and permission names.

Do not assume an agent-level MCP deny disables plugin lifecycle hooks.

## 3. Durable memory candidate

A specialist candidate must include:

- Category.
- Claim.
- Workspace and project.
- Source file or artifact.
- Commit or source revision when relevant.
- Confidence.
- Known uncertainty.
- Reason the information will help future work.
- Existing page that may overlap.

Conductor verifies the current source and checks for duplicates.

Human approves durable rules and consequential decisions.

Keep canonical specifications and ADRs in Git.

Do not copy a Git ADR into the wiki as a second authority.

## 4. Session capture and handoff

Use automatic lifecycle capture for supported harness sessions.

Use normal session-start handoff injection when the plugin delivers it.

Use manual `memory_handoff_begin` only for an explicit future-session baton.

Do not create a handoff for every Worker result.

Do not use ai-memory handoffs as live specialist messaging.

Use Conductor-mediated delegation for live work.

## 5. Cleanup

Do not run bulk cleanup during AW2-06.

For future maintenance:

1. Run `memory_lint` for a bounded scope.
2. Inspect exact findings.
3. Verify each finding against current sources.
4. Mark stale or wrong only with evidence.
5. Run retention sweep with `dry_run = true`.
6. Review exact candidate pages.
7. Obtain Human approval before destructive cleanup.

Do not delete raw observations because their count looks large.

## 6. Cross-project context

Use `memory_message_send` only for explicit cross-project coordination owned by Conductor.

Use `memory_query` with explicit workspace/project scopes for retrieval.

Do not set global recall as the default.

The generated OpenCode TypeScript integration does not support `server` profile routing in the current upstream documentation.

Keep all projects on the currently configured ai-memory server for this rollout.

Do not put server URLs or tokens in `.ai-memory.toml`.

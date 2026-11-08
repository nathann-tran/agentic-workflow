# Native adapters

AWM uses each harness's native agent and permission system.
The Core remains the source of truth.
An adapter converts Core roles into native definitions.

| Adapter | Agent format | Builder scope | Nested delegation |
|---|---|---|---|
| OpenCode V2 | Markdown agents | Exact edit paths; shell remains approval-gated | Native, subject to depth settings |
| Claude Code | Markdown subagents | Write/Edit hooks; shell side effects remain possible | Native, subject to runtime limits |
| Codex CLI | TOML agents | Not enforceable per file; use a separate approved session and worktree | Conductor-mediated fallback |
| Oh My Pi | Markdown task agents | Not enforceable by agent frontmatter alone; use worktree and approval | Native spawn allowlists and depth settings |

Do not treat a prompt as a security boundary.
Review `capabilities.json` before selecting an adapter.

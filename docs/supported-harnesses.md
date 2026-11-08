# Supported harnesses

AWM supports OpenCode V2, Claude Code, Codex CLI, and Oh My Pi.
The Core defines role behavior and workflow contracts.
Each adapter maps those contracts to native agent files.

| Harness | Global agent path | Scope note |
|---|---|---|
| OpenCode V2 | `~/.config/opencode/agents/awm/` | Nested IDs such as `awm/conductor`; exact edit paths; shell approval remains necessary |
| Claude Code | `~/.claude/agents/` | Write/Edit hooks; task Builder uses worktree isolation; shell effects remain outside the hook boundary |
| Codex CLI | `~/.codex/agents/` | Builder needs a separate approved writable session |
| Oh My Pi | `~/.omp/agent/agents/` | Builder needs an isolated worktree |

Select `awm/conductor` as the primary OpenCode agent after installation.
AWM does not overwrite your harness default-agent setting.
OpenCode asks before Grep because its V2 permission matcher uses query text.
Check the adapter capability manifest before work.
Do not infer support from an agent's prompt text.

## Runtime references

- OpenCode V2 agents: https://opencode.ai/v2/docs/agents
- Claude Code subagents: https://code.claude.com/docs/en/sub-agents
- Codex custom agents: https://learn.chatgpt.com/docs/agent-configuration/subagents
- Oh My Pi task agents: https://github.com/can1357/oh-my-pi/blob/main/docs/task-agent-discovery.md

The runtime documentation can change after this release.
Re-run `awm doctor` and review current harness documentation before upgrades.

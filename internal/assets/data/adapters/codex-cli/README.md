# Codex CLI adapter

AWM writes native TOML agents under `~/.codex/agents/`.
It adds a managed block to `~/.codex/AGENTS.md`.

Codex subagents inherit the parent sandbox and approval settings.
The adapter does not claim per-agent file-write isolation.
Run Builder in a separate Human-approved writable session.
Use a dedicated Git worktree for every implementation task.
Worker read-only behavior depends on the inherited sandbox.
Conductor-mediated delegation is the safe fallback when nested spawning is unavailable.

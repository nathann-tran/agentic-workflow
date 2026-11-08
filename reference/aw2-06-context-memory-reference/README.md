# Agentic Workflow 2.0 — AW2-06 Context & Memory Integration

- **Version:** 0.1
- **Status:** Integration package prepared; live configuration changes pending Human-run rollout
- **Target reference runtime:** OpenCode 2.2.6 + ai-memory OpenCode V2 integration
- **Date:** 2026-11-08

## Purpose

Complete the operational design for ai-memory and Understand Anything.

This package converts the approved context architecture into safe configuration and onboarding steps. It does not modify the user's machine.

## Key decisions

1. Keep ai-memory as the persistent cross-session memory service.
2. Keep Understand Anything as a repository-derived knowledge graph.
3. Group repositories by product, client, team, personal area, or trust boundary.
4. Use one ai-memory project scope per independently maintained Git repository.
5. Use Git remote identity first. Use `project_strategy = "repo-root"` as the remote-less fallback.
6. Use allowlist capture after the generated OpenCode integration is refreshed and its marker resolution is confirmed.
7. Set `auto_improve.require_approval = true`.
8. Disable the auto-improvement scheduler during initial rollout.
9. Keep assistant-text capture disabled.
10. Keep cross-project recall explicit. Do not enable global recall by default.
11. Keep `server` profile routing out of OpenCode marker files. The current generated TypeScript integration does not route server profiles.
12. Keep memory cleanup and durable memory promotion under Conductor and Human control.

## Runtime evidence supplied by the user

- OpenCode `v2.2.6`.
- ai-memory MCP server ID `ai-memory`, with 24 tools exposed.
- `memory_status` returned workspace `default`, project `semester-1-paper`, resolved by session.
- The SessionStart hook delivered a previously pending handoff.
- The local generated plugin was inspected. It currently has `CAPTURE_MODE = "denylist"` and `CAPTURE_ASSISTANT = false`.
- The plugin contains an allowlist gate and capture-policy checks before event spooling.
- `findMarker` / `findSettingsMarker` behavior in the installed generated plugin must be reconciled with the current upstream marker-walk contract.
- Understand Anything is installed at `~/.understand-anything-plugin`, through a symlink to the plugin repository. The user verified the internal agent definitions.

## Package contents

- `01-runtime-assessment.md` — Facts, gaps, and decisions.
- `02-ai-memory-server-config.md` — Server settings and effective-config discovery.
- `03-opencode-capture-rollout.md` — Safe regeneration of the OpenCode V2 plugin with allowlist capture.
- `04-workspace-onboarding.md` — Workspace taxonomy and marker templates.
- `05-understand-anything-operations.md` — Graph ownership, freshness, and worktree handling.
- `06-memory-operation-policy.md` — Agent access, writes, handoffs, and cleanup.
- `07-apply-and-rollback-runbook.md` — Ordered rollout procedure.
- `08-runtime-verification.md` — Read-only verification commands.
- `templates/` — TOML snippets and repository onboarding template.
- `scripts/inspect_runtime.sh` — Read-only runtime inspection. It does not print tokens or config bodies.

## Important limitation

The package is rollout-ready, not installed. No ai-memory configuration, generated plugin, marker, wiki page, graph, or repository was changed.
The current AWM release checks the reference OpenCode plugin path only. It does not validate native memory adapters for the other harnesses.

Apply only after reviewing the effective server configuration and confirming the local ai-memory CLI supports the documented flags.

## Official sources

- [ai-memory installation](https://github.com/akitaonrails/ai-memory/blob/main/docs/install.md)
- [ai-memory marker file](https://github.com/akitaonrails/ai-memory/blob/main/docs/marker-file.md)
- [ai-memory auto-improvement loop](https://github.com/akitaonrails/ai-memory/blob/main/docs/auto-improvement-loop.md)
- [ai-memory architecture](https://github.com/akitaonrails/ai-memory/blob/main/docs/ARCHITECTURE.md)
- [Understand Anything skill](https://github.com/egonex-ai/understand-anything/blob/main/understand-anything-plugin/skills/understand/SKILL.md)

# Agentic Workflow Manager (AWM)

**Version:** 0.2.0  
**Project:** Harness-agnostic agentic software development workflow

AWM installs one shared Core across four coding harnesses.
It uses native agent formats and permission systems.
It does not replace the harness runtime.

## Supported harnesses

- OpenCode V2
- Claude Code
- Codex CLI
- Oh My Pi

## Quick start

One-line install from GitHub Releases.

macOS and Linux:

```sh
curl -fsSL https://raw.githubusercontent.com/nathann-tran/agentic-workflow/main/installers/install.sh | sh
```

Windows PowerShell:

```powershell
irm https://raw.githubusercontent.com/nathann-tran/agentic-workflow/main/installers/install.ps1 | iex
```

Pin a release with `curl ... | AWM_VERSION=0.2.0 sh`, or
`$env:AWM_VERSION='0.2.0'; irm ... | iex` on Windows.
The installer downloads the archive, verifies `SHA256SUMS.txt`, and places
`awm` in `~/.local/bin`.

From a repository checkout, both installers also work without network access:

```sh
./installers/install.sh
```

```powershell
.\\installers\\install.ps1
```

Then configure your harnesses:

```sh
awm install
awm doctor
```

The wizard detects installed harnesses.
It supports one or multiple harnesses.
It checks skill IDs, bundle versions, declared dependencies, and compatibility metadata.
It offers Yes, No, or Select for missing or outdated bundled skills.
It audits known plugin markers and memory integrations.
It never installs external plugins or unpinned skills without approval.
For OpenCode, select `awm/conductor` as the primary agent after installation.
AWM does not overwrite your existing default-agent setting.

## Agent roles

| Role | Responsibility | Production writes |
|---|---|---|
| Conductor | Workflow orchestration, routing, state, and Human gates | Approved workflow state only |
| Plan | Requirements, acceptance criteria, and SDD | Assigned planning artifacts |
| Scout | Codebase evidence and knowledge-graph freshness | Assigned findings and graph artifacts |
| Architect | Interfaces, design, risks, and rollback | Assigned design artifacts |
| Builder | Approved implementation and tests | Only approved scope |
| Worker | Bounded read-only review or investigation | None |

The default workflow is Plan → Scout → Architect → Builder.
Conductor may select a valid entry point.
Human approval is required before consequential phase transitions.
Specialist handoffs go through Conductor.

## Main commands

```sh
awm install
awm detect
awm doctor
awm status
awm skills audit
awm plugins audit
awm memory doctor
awm workspace init --repo . --workspace product-example
awm builder render --task task.json --harness opencode --human-approved
awm builder rollback --task-id TASK-001
awm rollback
```

See `docs/architecture.md` for the Core and adapter design.
See `docs/` for detailed operations and adapter limitations.
See `core/` for canonical agent definitions and policies.
See `source-agent-specs/` for the six original source prompts supplied for this release.
See `reference/aw2-06-context-memory-reference/` for the historical context and memory work package.
Those two directories are repository-only and are not installed.
See `docs/reference/agentic-workflow-source.png` for the original workflow diagram.

## Safety boundaries

AWM backs up replaced files and records their hashes.
Rollback skips files that changed after installation.
Builder task profiles use separate manifests and rollback records.
AWM rejects symlink targets and path escapes during installation.
OpenCode and Claude Code provide stronger Builder path controls.
Codex and OMP require a separate approved session and isolated worktree.
Shell effects remain a residual risk on every adapter.

AWM audits ai-memory and Understand Anything without changing live settings.
AWM installs only its four bundled skills.
AWM does not install third-party plugins.

## Development checks

```sh
make test
make vet
make build
make release
```

See `docs/security-model.md` before production use.
See `docs/supported-harnesses.md` before selecting a runtime.
See `docs/release-validation.md` for the checks and known test limits in this release.

## Release status

This release includes static, unit, smoke, and cross-build checks.
The user must validate installation and runtime behavior on their machine.
`make release` writes per-platform archives, `RELEASE-MANIFEST.json`, and
`SHA256SUMS.txt` to `dist/`.

## License

MIT. See `LICENSE.txt`.

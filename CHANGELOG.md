# Changelog

## Unreleased

- Replaced the artifact priority order with a gate-artifact contract: 120-line and 1500-word memo, separate appendix, decisions first.
- Every Architect design now carries an Archify diagram; the Architect authors the candidate and Conductor renders the HTML and receipt.
- Added `docs/artifacts/<workflow_id>/design/` as the committed design artifact location.
- Made Archify a required skill for Architect and Conductor, mapped on all four harnesses.
- Widened Architect and Conductor write scope to `docs/artifacts/**`.
- Verified Archify presence in the shared skill root for all four harnesses; runtime load remains a manual check.
- Added `awm update` to re-apply the installed payload, reuse the previous harness selection, and prune stale managed files.
- Added `awm uninstall` to restore backups, strip managed guidance blocks, remove task-scoped Builder profiles, and remove the AWM home; `--remove-cli` also removes the binary.
- Managed state now carries unchanged files forward and records managed guidance block names.

## 0.2.0 — 2026-11-08

- Added harness-neutral Core and six canonical agent roles.
- Added OpenCode V2, Claude Code, Codex CLI, and OMP adapters.
- Added interactive Go CLI and safe home-directory installer.
- Added bundled skill audit and explicit selection.
- Added plugin and memory integration diagnostics.
- Added task-scoped Builder generation for OpenCode and Claude Code.
- Added backup, hash-guarded rollback, and workspace marker setup.
- Added cross-platform build targets and installer scripts.
- Selected the MIT license for open-source distribution.
- Kept `source-agent-specs/`, `core-docs/`, and the AW2-06 reference package repository-only.
- Added an embedded-payload parity test and asset sync target.
- Added release archives, manifest generation, and SHA-256 checksum verification.

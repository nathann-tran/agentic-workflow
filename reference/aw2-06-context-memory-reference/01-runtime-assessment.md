# AW2-06 — Runtime Assessment

## 1. Confirmed ai-memory integration

| Item | Observed result | Interpretation |
|---|---|---|
| MCP server ID | `ai-memory` | Use this exact server ID in the OpenCode adapter. |
| MCP tools | 24 | The user reported all 24 tools available. |
| Current scope | `default/semester-1-paper` | This is the scope of the inspected session, not the future workspace taxonomy. |
| Session handoff | A pending handoff was injected and consumed | Session continuity is active in the inspected session. |
| Raw observations | Not read | Keep raw observation access restricted. |
| Auto-improvement settings | Not established | The earlier config search did not establish the effective server config. |

## 2. Findings from the installed OpenCode plugin inspection

The user inspected `~/.config/opencode/plugins/ai-memory-opencode2.ts`.

| Item | Installed-plugin finding | AW2-06 action |
|---|---|---|
| Capture mode | `CAPTURE_MODE = "denylist"` | Regenerate the plugin with allowlist mode. Do not edit generated constants by hand. |
| Assistant text | `CAPTURE_ASSISTANT = false` | Keep disabled. |
| Allowlist branch | Branch exists in the generated file | Re-run the installer with `--capture-mode allowlist`. |
| Pre-spool filtering | Capture policy runs before spool enqueue | Retain this boundary. |
| Marker resolution | Local helper walks ancestors with a Git boundary | Verify this against the current installer output before using a group-level marker. |
| `project_strategy` | `repo-root` is resolved client-side by the generated TypeScript integration | Keep it in the marker and/or install-time setting. |
| Server profiles | The generated TypeScript integration does not support profile routing | Do not set `server = "..."` in OpenCode markers. |
| Lifecycle hooks | Start, end, pre-compact, prompt, context, and execute hooks are present | Preserve the generated integration. |

The local plugin report is authoritative for the installed file. Current upstream documentation may describe newer behavior. Refresh the generated plugin before relying on newer marker semantics.

## 3. Important reconciliation: group marker lookup

The approved design uses one workspace marker above related repositories.

Current upstream ai-memory documentation says lifecycle hooks search allowed ancestors toward `$HOME`. It stops at the nearest Git checkout root only when the checkout is outside `$HOME`.

The user's installed generated plugin inspection reported a Git-boundary stop in its local marker helper.

These two observations may refer to different plugin versions.

Therefore:

1. Keep the approved Workspace → Project → Repository model.
2. Regenerate the OpenCode plugin from a current ai-memory CLI.
3. Test marker resolution with `--check-capture` before enabling group-level allowlist capture.
4. If the generated plugin still does not see an ancestor workspace marker, put the same minimal marker at each repository root. This is the safe fallback.
5. Do not assume a workspace marker opts in descendant repositories until the local check confirms it.

## 4. Current defaults that require deliberate configuration

Current ai-memory documentation describes these defaults:

- `auto_improve.require_approval = false`.
- `auto_improve.scheduler.enabled = true`.

The first rollout must override both settings.

Use `require_approval = true` and `scheduler.enabled = false` until the memory workflow is stable.

The two settings are independent.

## 5. Configuration location

The ai-memory server reads `<data_dir>/config.toml`.

On macOS, the documented default data directory is:

```text
~/Library/Application Support/ai-memory/
```

The likely default config path is therefore:

```text
~/Library/Application Support/ai-memory/config.toml
```

This path is a default, not proof of the active service configuration.

A LaunchAgent, Docker volume, `AI_MEMORY_DATA_DIR`, or explicit `--config` argument can select another path. Discover the active path before editing.

## 6. Unresolved but bounded items

- Exact installed ai-memory CLI/server version.
- Effective server config path and existing auto-improvement values.
- Whether the current generated plugin supports ancestor workspace markers after refresh.
- Whether the workspace group directory is inside `$HOME`.

No repository list is needed to define the architecture.

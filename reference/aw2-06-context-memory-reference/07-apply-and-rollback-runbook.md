# AW2-06 — Apply and Rollback Runbook

## 0. Scope

This runbook changes three things only:

1. ai-memory server auto-improvement settings.
2. The generated OpenCode V2 ai-memory plugin capture mode.
3. `.ai-memory.toml` markers for opted-in workspace groups or repositories.

Do not migrate existing memory pages during this rollout.

Do not run consolidation, retention sweep, or deletion.

## 1. Discover before applying

Run:

```bash
ai-memory --version
opencode --version
ai-memory install-hooks --help
```

Then run the read-only inspector:

```bash
sh scripts/inspect_runtime.sh
```

Confirm the active ai-memory server and config path.

Do not proceed if the server config path is uncertain.

Confirm the local CLI supports `--capture-mode` and `--project-strategy`.

## 2. Back up the current generated plugin

Keep backups outside the plugin load directory.

```bash
BACKUP_DIR="$HOME/Library/Application Support/agentic-workflow/backups"
mkdir -p "$BACKUP_DIR"
cp "$HOME/.config/opencode/plugins/ai-memory-opencode2.ts" \
  "$BACKUP_DIR/ai-memory-opencode2.ts.$(date +%Y%m%d-%H%M%S).bak"
```

Adjust the source filename only if the active plugin uses another name.

## 3. Configure server approval

Locate the active `<data_dir>/config.toml`.

On a default macOS install, check:

```text
~/Library/Application Support/ai-memory/config.toml
```

Back up the file.

Merge:

```toml
[auto_improve]
require_approval = true

[auto_improve.scheduler]
enabled = false
```

Do not duplicate existing tables.

Restart the existing ai-memory service after editing.

Do not change auth, bind, server URL, data directory, LLM provider, or embedding settings.

## 4. Preview OpenCode capture installation

```bash
ai-memory install-hooks \
  --agent opencode \
  --capture-mode allowlist \
  --project-strategy repo-root
```

Review the generated integration target and reported mode.

Use `--agent opencode2` only if the current CLI does not auto-detect OpenCode V2 and the help confirms the alias.

Do not apply if the preview selects an incompatible integration or loses existing server settings.

## 5. Apply OpenCode capture installation

```bash
ai-memory install-hooks \
  --agent opencode \
  --capture-mode allowlist \
  --project-strategy repo-root \
  --apply
```

Restart OpenCode.

Inspect the generated file:

```bash
rg -n 'CAPTURE_MODE|CAPTURE_ASSISTANT|findMarker|capturePolicy|repoRootProject' \
  "$HOME/.config/opencode/plugins/ai-memory-opencode2.ts"
```

Expected:

```text
CAPTURE_MODE = "allowlist"
CAPTURE_ASSISTANT = false
```

Confirm the capture-policy check runs before local spool and POST.

## 6. Create workspace markers

Use a workspace marker at the product or trust-boundary root after ancestor resolution has been verified.

Example:

```toml
workspace = "product-iris-academy"
project_strategy = "repo-root"
```

The marker opts descendant repositories into allowlist capture when the installed integration finds it.

If the local generated plugin still stops at the Git root, add the same marker at each repository root. This fallback preserves the workspace grouping.

Do not set `project`, `identity`, `server`, or `recall.default_global` in the workspace template.

## 7. Validate scope before normal use

From an opted-in repository:

```bash
ai-memory doctor
ai-memory status
```

Confirm the expected workspace and repository project.

Use the native check-capture mode when the installed CLI exposes it:

```bash
python3 - <<'PY' | ai-memory hook --event user-prompt-submit --agent opencode2 --check-capture
import json, os
print(json.dumps({"cwd": os.getcwd()}))
PY
```

Expected for an opted-in repository: `admits_capture: true` and the expected marker source.

Run the same check from a repository with no marker.

Expected: `admits_capture: false`.

The check-capture path must not spool or contact the server.

If the installed CLI does not expose this command shape, inspect `ai-memory hook --help`. Do not guess flags.

## 8. Understand Anything policy

No additional plugin installation is required.

For each repository:

- Run `/understand` for initial graph generation when discovery is required.
- Use incremental update for later changes.
- Use `--full` only when necessary.
- Do not enable `--auto-update` without Human approval.
- Keep the graph in `.ua/` or the existing legacy `.understand-anything/` directory.

## 9. Rollback

### Restore OpenCode plugin

1. Stop OpenCode.
2. Copy the exact backup file into the plugin directory.
3. Start OpenCode.
4. Confirm the original plugin loads.

### Restore capture mode

The install-wide mode can be returned to denylist:

```bash
ai-memory install-hooks \
  --agent opencode \
  --capture-mode denylist \
  --apply
```

Use this only if capture behavior must be reverted.

### Restore server config

Restore the backed-up effective `config.toml`.

Restart the same ai-memory service.

### Remove a marker

Remove only the exact `.ai-memory.toml` file created for this rollout.

Under allowlist mode, removing the marker stops capture for that repository or inherited group.

Do not delete memory pages or project records during rollback.

## 10. Completion criteria

AW2-06 is operationally complete when:

- The effective server config is identified.
- `require_approval = true` is effective.
- Scheduler is disabled for initial rollout.
- OpenCode plugin reports allowlist capture.
- Assistant-text capture remains disabled.
- An opted-in repository resolves to the intended workspace and project.
- An unmarked repository is rejected before capture.
- Understand Anything remains repository-scoped and Scout-owned.
- No memory pages or repositories were migrated or deleted.

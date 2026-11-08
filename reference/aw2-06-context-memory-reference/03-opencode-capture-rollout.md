# AW2-06 — OpenCode V2 Capture Policy Rollout

## 1. Desired state

| Setting | Desired value |
|---|---|
| Generated integration | OpenCode V2 TypeScript plugin |
| Capture mode | `allowlist` |
| Project strategy | `repo-root` |
| Assistant-text capture | `false` |
| Capture filtering | Before local spool and network delivery |
| Manual generated-file edits | Prohibited |

## 2. Why regenerate the plugin

The installed file currently contains `CAPTURE_MODE = "denylist"`.

The allowlist code path exists, but the installed constant selects denylist behavior.

ai-memory stores capture mode per installation.

Its installer generates the selected mode into the OpenCode integration.

Do not edit `CAPTURE_MODE` by hand. The next installer refresh would overwrite that edit.

## 3. Preflight

Run:

```bash
ai-memory --version
opencode --version
ai-memory install-hooks --help
```

Confirm that the local CLI supports:

- `--agent opencode` or the explicit V2 alias `--agent opencode2`.
- `--capture-mode allowlist`.
- `--project-strategy repo-root`.
- `--apply`.

The current upstream installer uses generic `--agent opencode` to probe the installed OpenCode major and select its V2 integration. `--agent opencode2` forces the V2 compatibility path.

If the installed CLI does not expose these options, stop. Do not edit the generated TypeScript file. Review the installed package manager or wrapper before upgrading.

## 4. Back up the generated plugin

Keep backups outside `~/.config/opencode/plugins/`.

A backup with a `.ts` extension inside the plugin directory may load as a second plugin.

For the known installed path:

```bash
BACKUP_DIR="$HOME/Library/Application Support/agentic-workflow/backups"
mkdir -p "$BACKUP_DIR"
cp "$HOME/.config/opencode/plugins/ai-memory-opencode2.ts" \
  "$BACKUP_DIR/ai-memory-opencode2.ts.$(date +%Y%m%d-%H%M%S).bak"
```

If the installed filename differs, back up the exact active file instead.

## 5. Preview first

Run without `--apply`:

```bash
ai-memory install-hooks \
  --agent opencode \
  --capture-mode allowlist \
  --project-strategy repo-root
```

Review the output.

Confirm it selects OpenCode V2 and reports the requested capture mode.

Confirm that it preserves the existing server endpoint and authentication configuration.

Do not paste tokens into chat.

If the generic `opencode` probe fails, use `--agent opencode2` only after checking the CLI help.

## 6. Apply only after the preview is correct

```bash
ai-memory install-hooks \
  --agent opencode \
  --capture-mode allowlist \
  --project-strategy repo-root \
  --apply
```

Restart OpenCode.

The installer should regenerate the V2 TypeScript plugin.

Do not add duplicate ai-memory lifecycle integrations through Claude Code compatibility hooks inside OpenCode.

## 7. Verify the generated file

Inspect only the relevant constants and resolver function names:

```bash
rg -n 'CAPTURE_MODE|CAPTURE_ASSISTANT|findMarker|findSettingsMarker|capturePolicy|repoRootProject' \
  "$HOME/.config/opencode/plugins/ai-memory-opencode2.ts"
```

Expected values:

```text
CAPTURE_MODE = "allowlist"
CAPTURE_ASSISTANT = false
```

Confirm that marker presence is checked before an event is spooled or posted.

## 8. Capture-mode behavior

Allowlist mode is install-wide, not per-agent.

A repository without an eligible `.ai-memory.toml` marker must emit no lifecycle event.

The local capture gate must reject the event before local spool and network delivery.

A marker opts the repository in. It does not require a special opt-in key.

A later bare `install-hooks --apply` preserves the installed capture mode.

To return to the default denylist mode, pass `--capture-mode denylist` explicitly.

## 9. Important compatibility warning

Current upstream documentation says generated OpenCode TypeScript integrations enforce allowlist capture and marker presence before POST.

The user's installed generated file contains the allowlist branch but currently selects denylist.

The generated file also reported a Git-root marker boundary. The current upstream marker documentation describes a different walk rule for repositories inside `$HOME`.

After regeneration, verify group-marker discovery before relying on it.

If the group marker is not detected, use a repository-root marker for each repository. Keep the same workspace value.

## 10. Do not enable these features in AW2-06

- Assistant-text capture.
- Server profiles selected through the `.ai-memory.toml` `server` key.
- Global recall by default.
- Automatic Understand Anything graph updates.

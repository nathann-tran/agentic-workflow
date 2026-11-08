# AW2-06 — ai-memory Server Configuration

## 1. Required initial settings

Merge this snippet into the **effective server** `<data_dir>/config.toml`.

Do not replace the whole file.

Do not add duplicate TOML tables. Update existing keys when they already exist.

```toml
[auto_improve]
require_approval = true

[auto_improve.scheduler]
enabled = false
```

## 2. Why both settings matter

`require_approval` controls whether validated auto-improvement proposals apply automatically or remain pending.

`scheduler.enabled` controls background review.

They are independent settings.

The initial rollout disables background review and keeps proposals behind approval.

Manual review remains available.

After the workflow is stable, the scheduler may be enabled while `require_approval` remains `true`.

## 3. Discover the effective configuration path

The server uses `<data_dir>/config.toml`.

Do not assume `~/.config/ai-memory/config.toml` is active.

On macOS, the documented default data directory is `~/Library/Application Support/ai-memory`.

Check for an explicit data directory or config path in the active LaunchAgent or container configuration.

Use `scripts/inspect_runtime.sh` for a read-only inspection.

Do not print environment variables, bearer tokens, provider keys, or the full config file.

## 4. Apply the change

1. Back up the effective config file.
2. Open the file with a local editor.
3. Merge the two settings.
4. Preserve every other setting.
5. Restart the service using its existing service manager.
6. Re-run the read-only memory status check.
7. Inspect the auto-improvement settings through the server's supported admin/configuration interface, if available.

If the effective config path is not known, stop.

Do not create a new config at a guessed path.

## 5. Keep assistant-text capture disabled

The current OpenCode plugin reports `CAPTURE_ASSISTANT = false`.

Keep it disabled.

Assistant final text can contain private material. ai-memory treats assistant-text capture as an opt-in feature with client and server gates.

This workflow does not need assistant-text capture.

## 6. Maintenance policy

- Keep automatic lint and retention maintenance separate from auto-improvement scheduling.
- Do not disable all maintenance without a reason.
- Do not enable embedding backfill as part of AW2-06.
- Do not run `memory_forget_sweep` during this rollout.
- Do not run broad consolidation during this rollout.

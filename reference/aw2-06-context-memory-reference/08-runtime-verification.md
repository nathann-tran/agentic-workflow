# AW2-06 — Runtime Verification Checklist

Use this checklist after applying the runbook.

## ai-memory server

- [ ] `ai-memory --version` recorded.
- [ ] Effective data directory identified.
- [ ] Effective config file identified.
- [ ] `[auto_improve] require_approval = true` is active.
- [ ] `[auto_improve.scheduler] enabled = false` is active.
- [ ] Service restarted using its existing service manager.
- [ ] `memory_status` still succeeds.

## OpenCode integration

- [ ] Existing generated plugin backed up outside plugin directory.
- [ ] Installer preview selected OpenCode V2.
- [ ] Installer preview showed allowlist mode.
- [ ] Generated file contains `CAPTURE_MODE = "allowlist"`.
- [ ] Generated file contains `CAPTURE_ASSISTANT = false`.
- [ ] Capture policy runs before spool and network delivery.
- [ ] OpenCode restarted after plugin refresh.

## Workspace scope

- [ ] Workspace name follows the naming rule.
- [ ] Marker contains `workspace`.
- [ ] Marker contains `project_strategy = "repo-root"`.
- [ ] Marker does not pin `project` or `identity`.
- [ ] Marker does not contain server URL or token.
- [ ] Group-level marker lookup is confirmed for the installed plugin.
- [ ] If group lookup fails, repository-root markers are used.
- [ ] `ai-memory doctor` resolves the expected scope.
- [ ] An unmarked repository is rejected by allowlist capture.

## Understand Anything

- [ ] Plugin symlink still resolves to the existing plugin checkout.
- [ ] Internal agent definitions remain present.
- [ ] Graph output uses `.ua/` or existing legacy directory.
- [ ] Scout owns graph refresh.
- [ ] Automatic graph updates remain disabled unless approved.

## Memory safety

- [ ] Assistant-text capture remains disabled.
- [ ] No `memory_forget_sweep` ran.
- [ ] No bulk consolidation ran.
- [ ] No memory pages were deleted.
- [ ] No workspace migration ran.

## Reporting

Record:

- ai-memory version.
- Resolved data/config path (no secrets).
- Generated plugin path.
- Capture mode.
- Resolved workspace/project for one opted-in repository.
- `admits_capture` for one opted-in and one unmarked directory.
- Remaining warnings.

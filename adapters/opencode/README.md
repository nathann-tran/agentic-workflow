# OpenCode V2 adapter

AWM writes native Markdown agents under `~/.config/opencode/agents/awm/`.
The directory creates scoped agent IDs such as `awm/plan`.
Conductor uses `mode: primary`.
Plan uses `mode: all`.
Other roles use `mode: subagent`.

OpenCode V2 uses ordered `permissions` rules.
The last matching rule controls the action.
Read denies common secret path patterns.
Grep asks for approval because V2 matches the grep query, not its path.
The base Builder remains hidden and cannot edit files.
Use `awm builder render` to create a task-scoped Builder.
The task profile allows only listed edit paths.
Shell commands remain approval-gated because shell effects exceed file-path rules.

Set `experimental.subagent_depth` to a safe value in OpenCode settings.
Do not assume this adapter configures that runtime setting.

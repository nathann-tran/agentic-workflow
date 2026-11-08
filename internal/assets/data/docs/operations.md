# Operations

`awm detect` detects supported harnesses.
`awm doctor` runs a read-only environment audit.
`awm status` reports AWM installation state.
`awm skills audit` checks bundled and external skills.
`awm plugins audit` checks plugin markers.
`awm memory doctor` checks memory integration paths.
`awm workspace init` creates a repository marker.
`awm builder render` creates an approved task-scoped profile.
`awm builder rollback` restores a task-scoped profile and its allowlist.
`awm rollback` restores AWM-managed files when their hashes match.

The audit does not change the local environment.
The workspace and Builder commands write only after explicit invocation.

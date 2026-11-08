# Canonical Contract Index

| Contract | Schema or policy |
|---|---|
| Agent task input | `schemas/task-envelope.schema.json` |
| Agent task result | `schemas/agent-result.schema.json` |
| Workflow state | `schemas/workflow-state.schema.json` |
| Human decision | `schemas/decision-record.schema.json` |
| Memory candidate | `schemas/memory-candidate.schema.json` |
| Skill activation | `schemas/skill-activation-record.schema.json` |
| Runtime adapter | `schemas/adapter-manifest.schema.json` |

## Identity rule

`workspace_id`, `project_id`, `repository_id`, `workflow_id`, and `task_id` are Core identifiers.

`harness_id`, `native_session_id`, and `agent_instance_id` are adapter identifiers.

Do not use adapter identifiers as Core identity.

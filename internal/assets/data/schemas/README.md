# Core Schemas

All schemas use JSON Schema Draft 2020-12.

| Schema | Purpose |
|---|---|
| `task-envelope.schema.json` | Bounded agent task input |
| `agent-result.schema.json` | Canonical agent output |
| `workflow-state.schema.json` | Cross-session workflow state |
| `decision-record.schema.json` | Human-approved decisions |
| `memory-candidate.schema.json` | Proposed durable memory |
| `skill-activation-record.schema.json` | Auditable skill activation |
| `adapter-manifest.schema.json` | Runtime capability declaration |

Harness-native session IDs appear only in `adapter_bindings` or adapter manifests.

They do not define workflow identity.

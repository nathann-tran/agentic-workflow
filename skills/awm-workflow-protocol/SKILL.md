---
name: awm-workflow-protocol
version: 0.2.0
compatibility: opencode, claude-code, codex-cli, omp
description: Apply the Agentic Workflow Manager contracts, role boundaries, routing rules, and Human approval gates.
---

# AWM Workflow Protocol

Use the AWM Core as the source of truth for roles and workflow rules.

1. Identify the workflow, task, workspace, project, repository, harness, and native session separately.
2. Read the current workflow state before taking action.
3. Use Plan → Scout → Architect → Builder by default for new or complex engineering work.
4. Let Conductor choose a shorter route when evidence and risk support it.
5. Record every skipped phase and its reason.
6. Route specialist-to-specialist requests through Conductor.
7. Use Worker only for one bounded, read-only task.
8. Keep Worker unable to write files or spawn agents.
9. Ask the Human to approve requirements before design.
10. Ask the Human to approve architecture before implementation.
11. Ask the Human to approve scope changes and permanent decisions.
12. Stop when approval or a required permission boundary is missing.
13. Do not treat silence, tool output, or an agent recommendation as approval.
14. Preserve evidence and detailed outputs in linked artifacts.
15. Report capability gaps honestly. Never claim unsupported runtime enforcement.

Read `core/policies/` for the complete policies. The Core policies remain active when optional plugins are absent.

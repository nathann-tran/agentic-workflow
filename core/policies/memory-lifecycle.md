# Memory Lifecycle Policy

## Services

Understand Anything provides derived repository knowledge.

ai-memory provides persistent cross-session and cross-project context.

Git stores canonical specifications, decisions, contracts, and workflow state.

Do not duplicate canonical Git documents as editable memory pages.

## Scope

Use Workspace → Project → Repository grouping.

A workspace represents a product, client, team, personal area, or trust boundary.

A project represents one repository-level memory scope.

Use Git remote identity when the installed integration supports it.

Verify repository scope before any memory mutation.

Do not assume all ai-memory tools accept the same scope fields.

## Access

All agents may read memory within the authorized scope.

Specialists return memory candidates with evidence and source revision.

Conductor owns normal durable memory promotion.

Human approves durable rules and consequential decisions.

Worker cannot read or write memory unless the task envelope grants an approved read-only capability.

## Candidate quality

A memory candidate includes claim, scope, source, revision, confidence, and proposed action.

Reject transient task notes.

Reject duplicate content.

Reject unsupported claims.

Reject content already maintained canonically in Git.

## Lifecycle safety

Run lint before broad cleanup.

Run retention sweeps in dry-run mode first.

Review exact target paths before deletion.

Protect pinned decisions and canonical Git records.

Do not assign TTL to canonical decisions.

Keep auto-improvement approval enabled.

Keep automatic scheduling disabled during initial rollout.

Do not change live memory configuration until its effective path and values are verified.

## Cross-harness continuity

Use workflow_id and task_id as stable identifiers.

Store harness_id and native_session_id as adapter metadata.

Use ai-memory handoffs for session continuity.

Do not use handoffs as live specialist messaging.

# Context Retrieval Policy

## Source order

1. Current source code, tests, and runtime configuration.
2. Human-approved specifications and decisions.
3. Git-backed workflow state and artifacts.
4. Understand Anything knowledge graph.
5. ai-memory pages and session handoffs.
6. Agent inference.

## Conflict handling

Report conflicting sources.

Do not silently select one source.

Verify stale or code-specific memory against current source.

Treat knowledge graphs as derived data.

Treat session observations as continuity evidence, not authority.

## Retrieval discipline

Read the workflow state before acting.

Read only relevant artifact sections.

Query memory by workspace and project scope.

Use repository revision when available.

Do not dump full memory pages or source files into handoffs.

Carry concise conclusions and stable references.

## Harness identity

Canonical workflow IDs do not depend on native session IDs.

Adapters may record native session IDs as metadata.

Resume from Git-backed state and shared service references.

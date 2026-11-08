# Memory integration

AWM treats Git as the canonical store for specifications and decisions.
Understand Anything supplies repository-derived knowledge.
ai-memory supplies persistent cross-session context.

## Understand Anything

Scout owns graph creation and refresh.
Other agents read the graph.
Check graph freshness against the current Git revision.
Verify important claims against source files.

## ai-memory

Use Workspace → Project → Repository grouping.
A workspace represents one product, client, team, or trust boundary.
Each Git repository receives its own project scope.
Use repository-level `.ai-memory.toml` markers.
Repeat the same workspace ID across repositories in one workspace.
Verify effective scope through the installed memory status tool.

This release checks the user's OpenCode ai-memory integration path.
It does not install or verify native ai-memory adapters for other harnesses.
Keep memory tooling optional until each runtime has a tested integration.

## Safe defaults

Keep assistant-text capture disabled during initial rollout.
Use allowlist capture only after the installed integration supports it.
Set auto-improvement approval to required.
Disable automatic scheduling during initial rollout.
Do not change live settings without reviewing effective configuration.
Do not run destructive cleanup without a dry run and Human review.

Run:

```sh
awm memory doctor
awm workspace init --repo . --workspace product-example
```

The workspace command writes a repository marker.
It does not configure the ai-memory plugin.

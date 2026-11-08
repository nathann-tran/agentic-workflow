# AW2-06 — Understand Anything Operations

## 1. Installation state

The user verified this symlink:

```text
~/.understand-anything-plugin
  -> ~/.understand-anything/repo/understand-anything-plugin
```

The Understand skill resolves inside that plugin checkout.

The user manually verified the internal agent definitions.

No reinstall is required.

## 2. Ownership

Scout owns codebase graph creation and refresh.

Other agents may read or query the graph.

Other agents must verify important claims against source code, tests, and current configuration.

Understand Anything is a derived codebase index.

It is not the authority for current behavior or approved intent.

## 3. Data directory

The skill writes to the repository data directory:

- `.ua/` for a new repository.
- `.understand-anything/` when the legacy directory already exists.

The skill redirects output from a Git worktree to the main repository root by default.

This avoids disposable worktree graphs.

Do not create a second graph for each worktree unless Human explicitly requests it.

## 4. Refresh policy

At task start, Scout checks:

1. Current Git HEAD.
2. Graph source revision, when recorded.
3. Whether the checkout is a worktree.
4. Whether incremental update state is valid.

Use incremental update by default.

Use `--full` only for a first scan or when incremental state is invalid.

Verify load-bearing results against source.

## 5. Auto-update policy

Do not enable `--auto-update` by default.

The option writes `autoUpdate: true` to the repository's `$UA_DIR/config.json`.

It changes repository behavior.

Ask Human before enabling it.

Use the normal manual update workflow until the user chooses automatic refresh.

## 6. Internal Understand agents

The installed plugin contains internal agents for scanning, file analysis, graph assembly, architecture analysis, tours, and graph review.

These agents belong to the Understand skill.

They are not first-party Agentic Workflow specialists.

Scout may invoke them only through the documented Understand workflow.

Their results do not bypass Conductor communication rules for first-party specialists.

## 7. Dependency installation

A fresh plugin checkout may require Node.js and pnpm to build the core package.

Do not run package installation or build commands silently.

The current verified installation already has the agent definitions.

Only rebuild if the skill reports missing build artifacts or a dependency error.

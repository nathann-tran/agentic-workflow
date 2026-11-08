# AW2-06 — Workspace and Repository Onboarding

## 1. Grouping model

Use this hierarchy:

```text
Workspace
  └── Project (one independently maintained repository)
        └── Git checkout / worktree
```

A workspace represents one product, client, team, personal area, or trust boundary.

A project represents one Git repository's memory scope.

A worktree should resolve to the same project as its main Git repository.

## 2. Naming rules

Use lowercase ASCII letters, digits, dots, dashes, and underscores.

Recommended pattern:

```text
^[a-z0-9][a-z0-9._-]*$
```

Examples:

- `personal`
- `product-iris-academy`
- `client-acme`
- `opensource-foo`

Do not use a person name, token, URL, filesystem path, or secret in a workspace name.

## 3. Workspace marker template

Place this marker at the workspace group root when ancestor resolution has been verified:

```toml
workspace = "product-example"
project_strategy = "repo-root"
```

Do not set `project` or `identity` at the workspace level.

This lets each repository retain its own project identity.

The Git remote is preferred when valid.

`repo-root` provides a stable fallback when no valid remote exists.

## 4. Repository marker fallback

Use this same marker at each repository root if the installed OpenCode plugin does not discover the workspace-level ancestor marker:

```toml
workspace = "product-example"
project_strategy = "repo-root"
```

The fallback remains scalable.

Repository onboarding can create this one file from a template.

Do not place a workspace marker beside a repository and assume the hook will inherit it. Verify the actual marker walk first.

## 5. Marker precedence

The nearest marker is authoritative for ordinary marker settings.

Markers do not merge.

A nested repository marker must repeat the required `workspace` value.

Do not add a child marker with only `project_strategy` and expect the workspace to inherit.

The `server` field has separate inheritance rules, but do not use it in OpenCode markers. The current generated TypeScript integration does not support server-profile routing.

## 6. Project identity

Do not pin `project` or `identity` unless a deliberate exception exists.

Prefer the normalized `upstream` or `origin` remote identity.

Use `project_strategy = "repo-root"` for a remote-less repository.

A monorepo is one project by default.

Split a monorepo into multiple memory projects only when each subproject needs independent memory ownership.

## 7. Workspace taxonomy examples

```text
~/Code/
├── Personal/
│   ├── learning-platform/
│   └── personal-tools/
├── Products/
│   └── IrisAcademy/
│       ├── api/
│       ├── web/
│       ├── worker/
│       └── shared/
├── Clients/
│   └── Acme/
│       ├── api/
│       └── dashboard/
└── OpenSource/
    └── Foo/
        ├── core/
        └── cli/
```

This is a taxonomy example.

It does not require these directories or repositories to exist now.

## 8. Capture policy and privacy

Allowlist mode is the repository-level opt-in gate.

Use `[capture] ignore_paths` only as an additional best-effort exclusion for sensitive subpaths.

Example:

```toml
[capture]
ignore_paths = ["**/private/**", "**/secrets/**"]
```

The exact marker is authoritative. Marker sections do not merge.

Current ai-memory documentation describes path-aware file-tool filtering. Shell-command filtering is lexical and has documented gaps. Unknown tools may not be inspected.

Therefore `ignore_paths` is not a complete data-loss-prevention boundary.

Do not send secrets to agents. Keep secrets outside their readable scope.

## 9. Optional subagent capture control

`drop_subagent_captures = "true"` is available as a per-project option in current ai-memory documentation.

It keeps top-level sessions but drops stored subagent-session captures.

Do not enable it by default during the first rollout.

It reduces noisy Worker and specialist observations, but it also removes their session records from persistent memory.

Review actual observation volume before enabling it for a project.

## 10. Cross-project retrieval

Do not enable `[recall] default_global = "true"` by default.

Conductor should request explicit scopes when it needs sibling repository context.

This avoids unrelated project knowledge appearing in every query.

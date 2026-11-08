# Workspace identity policy

A workspace groups repositories with one trust boundary.
A project represents one Git repository.
Use a stable workspace ID across repositories for one product or client.
Do not use a universal `default` workspace for unrelated projects.
Place `.ai-memory.toml` in each repository root.
Repeat the approved workspace ID in each repository marker.
Do not assume that parent-directory marker discovery crosses the Git root.
Confirm the effective workspace and project through `memory_status`.

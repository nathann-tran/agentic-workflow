# Delegation Policy

## Canonical topology

Conductor may dispatch Plan, Scout, Architect, Builder, and Worker.

Plan, Scout, Architect, and Builder may request or spawn Worker when the adapter permits it.

Worker cannot spawn another agent.

Specialists cannot message each other directly.

Worker returns its result to its direct caller.

## Worker constraints

Every Worker task must be bounded.

Every Worker task must define objective, scope, constraints, and done conditions.

Worker is read-only.

Worker makes no decisions.

Worker writes no files or memory.

Worker does not perform builds, installs, migrations, or external mutations.

## Adapter fallback

If nested delegation is unavailable, use Conductor as a delegation broker.

The specialist sends a Worker request to Conductor.

Conductor dispatches Worker and returns the result.

Do not claim direct delegation unless conformance tests prove it.

## Parallel work

Parallelize only independent tasks.

Do not parallelize tasks with shared mutable state or unresolved dependencies.

Conductor controls cross-specialist task routing.

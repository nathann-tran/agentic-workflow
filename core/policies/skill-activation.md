# Skill Activation Policy

## Registry authority

`core/registries/skills.yaml` is the canonical skill registry.

A canonical skill ID identifies a capability, not a runtime command.

Each adapter maps canonical IDs to exact runtime IDs.

Do not assume that IDs match across harnesses.

## Activation levels

- **required:** Load before a matching task.
- **conditional:** Load only when its trigger matches.
- **explicit:** Load by exact runtime ID only after explicit Human request or workflow authorization.
- **blocked:** Do not use in the default workflow.

## Deterministic activation

1. Classify the task.
2. Read the agent's skill policy.
3. Identify matching required skills.
4. Resolve exact runtime IDs through the adapter.
5. Load required skills.
6. Check dependencies and permissions.
7. Stop dependent work when a required skill is missing.
8. Record loaded IDs and load failures.
9. Do not claim a skill ran unless the runtime confirms it.

## Plugin independence

Core policies remain active when optional plugins are absent.

A plugin provides an implementation of a capability.

A plugin does not own the Core policy.

Use native or equivalent skills when a plugin is unavailable.

Record degraded behavior when an equivalent capability is used.

## Explicit-only skills

Plan owns Grill Me and Grill With Docs.

Do not invoke either silently.

Load `grilling` before either wrapper.

Load `domain-modeling` for Grill With Docs when required.

External issue publication requires tracker configuration and Human approval.

## Duplicate procedures

Do not load multiple skills that duplicate the same procedure unless the registry requires both.

If two skills are required, define their distinct contribution.

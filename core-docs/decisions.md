# Core Decisions

## D-001 — Harness-neutral Core

The Core is canonical. Harnesses are adapters.

## D-002 — Human governance

Human approval controls phase transitions and permanent decisions.

## D-003 — Default workflow

Plan → Scout → Architect → Builder is the default path.

Conductor may choose a valid entry point.

## D-004 — Delegation

Specialists may request or spawn Worker within verified adapter limits.

Worker cannot spawn another agent.

Specialists communicate through Conductor.

Worker returns to its direct caller.

## D-005 — Source ownership

Builder is the only production-source writer.

Scout, Architect, and Worker remain source-read-only.

## D-006 — Engineering

Use SDD before implementation.

Use TDD during implementation.

Apply Clean Architecture pragmatically.

Apply Ponytail minimality without removing correctness or safety.

## D-007 — Comments

Builder writes no comments by default.

Permitted comments explain non-obvious reasons.

## D-008 — Memory

Understand Anything provides derived repository knowledge.

ai-memory provides persistent context.

Git remains canonical for specifications, decisions, and workflow state.

## D-009 — Portability

Core skills use canonical IDs.

Adapters map canonical IDs to exact runtime IDs.

Do not assume runtime IDs match.

## D-010 — Reference adapter

OpenCode is the first reference adapter.

This does not make OpenCode a Core dependency.

## D-011 — No custom runtime

Do not build a custom agent runtime in this phase.

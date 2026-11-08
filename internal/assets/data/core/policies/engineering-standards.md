# Engineering Standards

## Development sequence

Use Spec-Driven Development before implementation.

Human approves the requirements and architecture gates.

Use Test-Driven Development during implementation.

Implement one vertical slice at a time.

Use Red → Green → Refactor for each behavior.

## Design

Follow Clean Architecture pragmatically.

Keep dependency direction explicit.

Prefer deep modules and small interfaces.

Create abstractions only for demonstrated needs.

Reuse existing code before adding dependencies.

Avoid speculative flexibility.

## Ponytail principles

Ask whether the change must exist.

Prefer standard library features before custom code.

Prefer native platform features before new dependencies.

Implement the smallest correct solution.

Do not remove validation, security, error handling, or required tests to reduce code.

## Code comments

Default to no comments.

Use names, types, structure, and tests to explain code.

Do not repeat obvious behavior in comments.

Do not use comments to hide unclear design.

Add a comment only when code cannot express the reason clearly.

Valid exceptions include business invariants, security constraints, concurrency assumptions, complex algorithm assumptions, external limitations, compatibility workarounds, licenses, and required tooling directives.

Explain why, not what.

Keep permitted comments short.

Do not add routine docstrings when types express the contract.

Review every changed comment.

## Production quality

Use strict types where supported.

Validate untrusted input at boundaries.

Handle failure paths explicitly.

Avoid swallowed errors and unbounded retries.

Use named constants instead of unexplained values.

Do not leave TODO placeholders or incomplete production paths.

Test correctness, security, edge cases, and error behavior.

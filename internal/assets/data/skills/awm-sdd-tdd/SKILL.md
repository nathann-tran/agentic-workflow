---
name: awm-sdd-tdd
version: 0.2.0
compatibility: opencode, claude-code, codex-cli, omp
description: Apply Specification-Driven Development before code changes and Test-Driven Development during implementation.
---

# AWM Specification and Test Driven Development

## Before implementation

1. Confirm approved requirements and acceptance criteria.
2. Confirm the approved architecture and interfaces.
3. Confirm current source revision and repository instructions.
4. Confirm the exact implementation scope.
5. Confirm required skills and permission boundaries.
6. Stop when a required approval is missing.

## During implementation

1. Write one failing test for one behavior.
2. Run it and confirm the expected failure.
3. Implement the smallest correct change.
4. Run the test and confirm it passes.
5. Refactor without changing behavior.
6. Repeat for each vertical slice.
7. Run relevant integration and regression checks.
8. Review changes and report exact evidence.

Do not claim a test passed unless the command completed successfully. Do not remove required tests to reduce code. Do not silently expand scope.

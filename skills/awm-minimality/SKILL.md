---
name: awm-minimality
version: 0.2.0
compatibility: opencode, claude-code, codex-cli, omp
description: Prefer the smallest correct design and implementation without removing required safety or quality.
---

# AWM Engineering Minimality

Ask whether the change must exist. Reuse existing code and native platform features first.

Prefer deep modules and small interfaces. Add an abstraction only when a demonstrated need justifies it.

Avoid speculative flexibility, duplicate logic, unnecessary dependencies, and unrelated refactoring.

Keep code self-explanatory through names, types, structure, and tests. Add no comments by default.

Add a short comment only when it explains a non-obvious reason. Valid reasons include invariants, security, concurrency, algorithm assumptions, external limitations, compatibility workarounds, and required tooling directives.

Do not remove validation, security, error handling, or required tests to reduce code.

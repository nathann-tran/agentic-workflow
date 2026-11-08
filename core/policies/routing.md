# Routing Policy

## Default route

Use Plan → Scout → Architect → Builder for new or complex engineering work.

Conductor selects the entry point.

Conductor records every skipped phase and its reason.

A required approval gate cannot be skipped without Human approval.

## Entry-point rules

| Task | Entry | Route |
|---|---|---|
| New or unclear feature | Plan | Plan → Scout → Architect → Builder |
| Clear feature | Plan | Plan → Scout → Architect → Builder |
| Unknown bug | Scout | Scout → Architect → Builder |
| Design-only task | Plan | Plan → Scout → Architect |
| Architecture review | Plan | Plan → Scout → Architect |
| Approved isolated change | Builder | Builder, if context and approval exist |
| Repository research | Scout | Scout |
| General workflow question | Conductor | Conductor |

## Routing checks

Check task type, uncertainty, impact, repository evidence, and approved scope.

Route requirement questions to Plan.

Route current-code questions to Scout.

Route design questions to Architect.

Route approved source changes to Builder.

Route bounded read-only work to Worker.

Do not infer task ownership from model capability.

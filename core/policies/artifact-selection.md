# Artifact Selection Policy

## Purpose

Choose the smallest format that preserves the content accurately.

## Gate artifacts

Each phase role writes one memo for the Human gate.

Worker returns an Agent Result only. It writes no artifact.

The memo is canonical.

A generated HTML view helps the Human read the memo. It does not replace the memo.

Use Markdown for the memo and the appendix. Markdown stays diffable, searchable, and readable by Builder.

## Budget

Keep the memo body at or below 120 lines and at or below 1500 words.

The budget excludes frontmatter and the appendix.

Put decisions and risks before background.

Move detail to the appendix instead of deleting it.

## Required order

1. Objective or status.
2. Decisions required.
3. Diagram.
4. Blast radius or risk.
5. Phases or next actions.
6. Rollback or recovery.

The agent definitions add role-specific sections inside this order.

## Design artifacts

The Architect writes one folder per design:

```text
docs/artifacts/<workflow_id>/design/
  memo.md
  appendix.md
  diagram.json
  diagram.html
  *.finalize*.json
```

| File | Owner | Content |
|---|---|---|
| `memo.md` | Architect | Gate memo within the budget |
| `appendix.md` | Architect | Files, interface contracts, testing detail, pattern notes, constraints, open questions |
| `diagram.json` | Architect | Archify candidate |
| `diagram.html` | Conductor | Rendered view |
| `*.finalize*.json` | Conductor | Passing receipt |

The Architect authors the memo, the appendix, and the candidate.

The Architect does not run commands.

Conductor runs `archify finalize` and writes the HTML and receipt beside the candidate.

Plan, Scout, and Builder write their memo and appendix in the folder named by the Task Envelope.

## Diagrams

Every Architect design carries an Archify diagram.

Plan, Scout, and Conductor add a diagram only when the content is genuinely relational. Conductor renders it.

Design artifacts use Archify, not Mermaid.

Do not publish an HTML view without a passing finalize receipt.

If Node, Archify, or Chromium is unavailable, block the gate and name the missing tool. Do not degrade silently.

## Other formats

Use the first format that represents the content:

1. Procedural animation.
2. Interactive HTML micro app.
3. Structured diagram.
4. Constrained text.

This order applies to non-design artifacts. It does not override the design artifact contract.

## Constraints

Never remove requirements, evidence, risks, caveats, decisions, or context to shorten output.

Keep generated HTML self-contained when practical.

Use accessible labels and keyboard interaction.

Do not send secrets or private data to external rendering services.

Treat prototypes as throwaway work.

Do not promote prototype code to production without normal review and tests.

## Technical language

Use Simplified Technical English principles for English technical communication.

Use equivalent clear-language principles for other languages.

Do not claim formal ASD-STE100 compliance without a separate compliance review.

## Context preservation

Every artifact preserves outcome, evidence, constraints, risks, decisions, and next actions.

Use references instead of copying large source content.

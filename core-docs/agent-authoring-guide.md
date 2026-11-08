# Canonical Agent Authoring Guide

## Purpose

Keep one source for each agent role.

Keep runtime configuration outside Core.

## Required metadata

Each canonical agent file starts with Core metadata.

Allowed fields include `agent_id`, `version`, `model_profile`, and `policy_refs`.

Do not add native harness fields.

## Role content

Each agent defines mission, authority, required behavior, boundaries, skills, done conditions, and output.

Keep one primary responsibility per role.

Do not copy full shared policies into every agent.

Reference policy IDs instead.

## Policy assembly

An adapter must include the referenced policy text in the native prompt context.

A file path alone does not prove that the runtime can read the policy.

The adapter may inline policy text or expose a verified read mechanism.

## Prompt changes

Change canonical role instructions first.

Run shared role conformance tests.

Regenerate or update each affected adapter.

Do not edit generated agent files manually.

## Language and length

Use concise technical English.

Use active voice where practical.

Keep sentences direct.

Preserve constraints, evidence, exceptions, and escalation paths.

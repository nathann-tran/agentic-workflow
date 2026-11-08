# Core Release Process

## Change process

1. Identify the affected role, policy, contract, or registry.
2. Update the canonical source.
3. Validate JSON Schemas and YAML registries.
4. Run shared conformance scenarios.
5. Identify affected adapters.
6. Run adapter-specific tests.
7. Record compatibility changes.
8. Obtain Human approval for release.

## Compatibility

A breaking contract change requires a schema version change.

A role policy change requires affected role tests.

A skill mapping change requires exact runtime load tests.

An adapter capability change requires updated evidence.

## Generated files

Generated adapter files must record Core and adapter versions.

Do not hand-edit generated outputs.

# Registry Governance

## Skills

Use canonical capability IDs in Core.

Use exact runtime IDs in adapter manifests.

Record source, version, checksum, owner, invocation level, activation trigger, and dependencies.

Do not assume the same runtime ID across harnesses.

Do not silently update upstream skills.

## Plugins

Record plugin capabilities and limits.

Treat plugins as optional runtime providers.

Keep the Core policy usable without a plugin.

## Models

Use logical model profiles in Core.

Map provider model IDs inside adapters.

Model selection does not change agent authority.

## Harness targets

Separate documented capabilities from local test results.

Record harness version and verification date.

Do not promote an unverified capability to native or adapter-enforced.

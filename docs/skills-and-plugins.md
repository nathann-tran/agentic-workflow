# Skills and plugins

## Skill discovery

AWM scans the selected harness skill roots for `SKILL.md`.
It reads common frontmatter fields for display.
It matches exact skill IDs or folder names.
It reads version, compatibility, and declared dependency metadata.

## Bundled skills

AWM ships four first-party skills:

- `awm-workflow-protocol`
- `awm-clear-communication`
- `awm-sdd-tdd`
- `awm-minimality`

The installer offers Yes, No, or Select for missing or outdated bundles.
It checks the exact bundled version before offering an update.
It asks before replacing an unmanaged skill file.

## External skills

The registry maps external skills to roles, dependencies, and activation conditions.
AWM marks sources without immutable pins as non-installable.
A missing dependency appears in the audit report.
The user installs and updates those skills through their trusted provider.

## Plugins

AWM checks local markers for Superpowers, Ponytail, ai-memory, and Understand Anything.
This check confirms a marker only.
It does not confirm plugin health or active configuration.
The user remains responsible for plugin installation and updates.

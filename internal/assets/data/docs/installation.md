# Installation

## 1. Install the AWM CLI

One-line install from GitHub Releases.

macOS and Linux:

```sh
curl -fsSL https://raw.githubusercontent.com/nathann-tran/agentic-workflow/main/installers/install.sh | sh
```

Windows PowerShell:

```powershell
irm https://raw.githubusercontent.com/nathann-tran/agentic-workflow/main/installers/install.ps1 | iex
```

The installer detects the OS and architecture, downloads
`awm_<version>_<os>_<arch>.tar.gz` or `.zip` from the latest release, verifies
it against `SHA256SUMS.txt`, and places the executable in `~/.local/bin`.
Add that directory to PATH when your shell does not include it.
Pin a version with `AWM_VERSION=0.2.0` in the environment.
Set `AWM_REPO` to use a fork.

From a repository checkout, the same script builds from source when Go is
installed:

```sh
./installers/install.sh
```

From a release archive:

```sh
tar -xzf awm_<version>_<os>_<arch>.tar.gz
shasum -a 256 -c SHA256SUMS.txt   # sha256sum -c on Linux
install -m 755 awm ~/.local/bin/awm
```

Windows archives use `.zip` and contain `awm.exe`.
The executable does not install harnesses or third-party plugins.

## 2. Run the harness wizard

```sh
awm install
```

Select one or more detected harnesses.
Review the Builder permission notes for each selection.
Choose Yes, No, or Select for missing bundled skills.
AWM installs only its four bundled skills.
AWM never downloads a third-party skill without a trusted immutable source.

To preview changes:

```sh
awm install --harness opencode --dry-run
```

To select explicitly:

```sh
awm install --harness opencode,claude-code
```

Use `--allow-undetected` only when you intentionally install an adapter without its runtime.
Use `--force` only after reviewing files that AWM would replace.

## 3. Verify

```sh
awm doctor
awm status
```

Review generated agent files in each harness.
Restart the harness when its agent discovery requires a reload.

## 4. Roll back

```sh
awm rollback
```

AWM checks installed file hashes before rollback.
It skips files that changed after installation.
Use `awm rollback --force` only after reviewing the affected files.

Builder task profiles use a separate manifest.
Restore one task with `awm builder rollback --task-id TASK-001`.
This command also checks hashes and preserves Human edits by default.

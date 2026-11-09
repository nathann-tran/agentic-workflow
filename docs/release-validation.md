# Release validation — 0.3.0

## Automated checks

`make validate` runs:

- `gofmt -l` formatting check
- `go test ./...`, including the embedded-payload parity test
- `python3 -m unittest discover -s tests/python -v`
- `go vet ./...`
- a host build

CI runs `make validate` on Linux and macOS.
CI also runs `make release` and verifies `dist/SHA256SUMS.txt` on Linux.

## Release artifacts

`make release` cross-compiles five binaries and writes to `dist/`:

- `awm_<version>_<os>_<arch>.tar.gz` for macOS and Linux, `.zip` for Windows.
  Each archive contains the binary, `LICENSE.txt`, and `README.md`.
- `awm-<os>-<arch>` and `awm-windows-amd64.exe`, used by the installers.
- `RELEASE-MANIFEST.json` with version, commit, license, and per-artifact hashes.
- `SHA256SUMS.txt`, covering both binaries and archives.

Verify with `cd dist && shasum -a 256 -c SHA256SUMS.txt` on macOS or
`sha256sum -c SHA256SUMS.txt` on Linux.
Both installers verify the raw binary against `SHA256SUMS.txt` when that file
sits next to the binary.

## Manual smoke coverage

Run these in a temporary HOME before publishing:

- Detect harnesses and audit empty skill, plugin, and memory states.
- Install all four harness adapters in one run.
- Install bundled skills in shared and OMP-specific skill roots.
- Run `doctor` and `status` after installation.
- Create a repository-level `.ai-memory.toml` marker in a Git repository.
- Render and roll back a task-scoped OpenCode Builder profile.
- Roll back the full AWM installation.
- Update an installed copy with `awm update`; verify stale files are removed and unchanged files stay in the state.
- Uninstall with `awm uninstall`; verify restored backups, stripped guidance blocks, and removed task-scoped Builder profiles.
- Preview installation with `--dry-run` and verify that no files changed.
- Install the CLI through the macOS/Linux shell installer.
- Interactive multi-harness selection installs only the selected adapters.
- Interactive skill selection installs only the selected bundled skills.
- Builder rollback preserves Human-modified task profiles unless `--force` is set.
- Full rollback preserves Human-modified managed files.

## Not tested in this environment

The Windows PowerShell installer is reviewed but not executed.
The macOS and Windows binaries are cross-compiled but not run on those operating systems.
No native harness is installed or configured by these checks.
Final runtime validation must run on the target machine and harness versions.

## Security scope

The adapters do not provide equal permission guarantees.
Codex and OMP require a separate approved writable Builder session.
OpenCode and Claude Code still have shell-side-effect risks.
The AWM CLI does not change live ai-memory settings or stored memories.

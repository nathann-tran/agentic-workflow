#!/usr/bin/env sh
set -eu

# Source trees that ship inside the awm binary. Repository-only material
# (core-docs, reference, source-agent-specs) is intentionally excluded.
ROOT=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)
DEST="$ROOT/internal/assets/data"
TREES="core docs skills templates schemas registries adapters examples hooks"

rm -rf "$DEST"
mkdir -p "$DEST"
for tree in $TREES; do
  cp -R "$ROOT/$tree" "$DEST/$tree"
done

# go:embed skips dotfiles; keep the mirror free of host clutter.
find "$DEST" -name '.DS_Store' -delete
echo "Synced embedded payload to $DEST"

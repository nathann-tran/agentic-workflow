#!/usr/bin/env sh
set -eu

REPO="${AWM_REPO:-nathann-tran/agentic-workflow}"
VERSION="${AWM_VERSION:-}"
VERSION="${VERSION#v}"
BASE_URL="${AWM_BASE_URL:-}"
TMPDIR=""

cleanup() {
  if [ -n "$TMPDIR" ]; then rm -rf "$TMPDIR"; fi
}
trap cleanup EXIT INT TERM

if [ -f "$0" ] && [ -r "$0" ]; then
  SCRIPT_DIR=$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)
  ROOT=$(dirname -- "$SCRIPT_DIR")
else
  ROOT=""
fi

OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)
case "$ARCH" in
  arm64|aarch64) ARCH=arm64 ;;
  x86_64|amd64) ARCH=amd64 ;;
  *) echo "Unsupported architecture: $ARCH" >&2; exit 2 ;;
esac
case "$OS" in
  darwin|linux) ;;
  *) echo "Unsupported operating system: $OS" >&2; exit 2 ;;
esac

fetch() {
  URL=$1
  OUTPUT=$2
  if command -v curl >/dev/null 2>&1; then
    curl -fsSL "$URL" -o "$OUTPUT"
  elif command -v wget >/dev/null 2>&1; then
    wget -qO "$OUTPUT" "$URL"
  else
    echo "curl or wget is required to download AWM." >&2
    exit 2
  fi
}

sha256_file() {
  if command -v sha256sum >/dev/null 2>&1; then
    sha256sum "$1" | awk '{ print $1 }'
  else
    shasum -a 256 "$1" | awk '{ print $1 }'
  fi
}

BINARY=""
VERIFY_SUMS=""
if [ -n "$ROOT" ] && [ -x "$ROOT/dist/awm-$OS-$ARCH" ]; then
  BINARY="$ROOT/dist/awm-$OS-$ARCH"
  VERIFY_SUMS="$ROOT/dist/SHA256SUMS.txt"
elif [ -n "$ROOT" ] && [ -f "$ROOT/go.mod" ] && command -v go >/dev/null 2>&1; then
  echo "Building AWM from source for $OS/$ARCH..."
  mkdir -p "$ROOT/.build"
  (cd "$ROOT" && CGO_ENABLED=0 GOOS="$OS" GOARCH="$ARCH" go build -trimpath -ldflags "-s -w" -o "$ROOT/.build/awm" ./cmd/awm)
  BINARY="$ROOT/.build/awm"
else
  TMPDIR=$(mktemp -d)
  if [ -z "$BASE_URL" ]; then
    if [ -n "$VERSION" ]; then
      BASE_URL="https://github.com/$REPO/releases/download/v$VERSION"
    else
      BASE_URL="https://github.com/$REPO/releases/latest/download"
    fi
  fi
  if [ -z "$VERSION" ]; then
    fetch "$BASE_URL/RELEASE-MANIFEST.json" "$TMPDIR/RELEASE-MANIFEST.json"
    VERSION=$(sed -n 's/.*"version": *"\([^"]*\)".*/\1/p' "$TMPDIR/RELEASE-MANIFEST.json" | head -n 1)
    if [ -z "$VERSION" ]; then
      echo "Cannot determine the release version from $BASE_URL." >&2
      exit 2
    fi
  fi
  ARCHIVE="awm_${VERSION}_${OS}_${ARCH}.tar.gz"
  fetch "$BASE_URL/SHA256SUMS.txt" "$TMPDIR/SHA256SUMS.txt"
  fetch "$BASE_URL/$ARCHIVE" "$TMPDIR/$ARCHIVE"
  EXPECTED=$(awk -v name="$ARCHIVE" '$2 == name { print $1 }' "$TMPDIR/SHA256SUMS.txt")
  if [ -z "$EXPECTED" ]; then
    echo "No checksum for $ARCHIVE in SHA256SUMS.txt" >&2
    exit 2
  fi
  ACTUAL=$(sha256_file "$TMPDIR/$ARCHIVE")
  if [ "$ACTUAL" != "$EXPECTED" ]; then
    echo "Checksum mismatch for $ARCHIVE" >&2
    exit 2
  fi
  tar -xzf "$TMPDIR/$ARCHIVE" -C "$TMPDIR" awm
  BINARY="$TMPDIR/awm"
fi

if [ -n "$VERIFY_SUMS" ] && [ -f "$VERIFY_SUMS" ]; then
  NAME=$(basename -- "$BINARY")
  EXPECTED=$(awk -v name="$NAME" '$2 == name { print $1 }' "$VERIFY_SUMS")
  if [ -z "$EXPECTED" ]; then
    echo "No checksum for $NAME in $VERIFY_SUMS" >&2
    exit 2
  fi
  ACTUAL=$(sha256_file "$BINARY")
  if [ "$ACTUAL" != "$EXPECTED" ]; then
    echo "Checksum mismatch for $BINARY" >&2
    exit 2
  fi
fi

DEST_DIR="$HOME/.local/bin"
DEST="$DEST_DIR/awm"
mkdir -p "$DEST_DIR"
if [ -e "$DEST" ]; then
  printf 'Replace existing %s? [y/N]: ' "$DEST"
  read -r ANSWER || ANSWER=n
  case "$ANSWER" in
    y|Y|yes|YES)
      BACKUP="$DEST.bak.$(date -u +%Y%m%dT%H%M%SZ)"
      cp -p "$DEST" "$BACKUP"
      echo "Previous binary backed up to $BACKUP"
      ;;
    *) echo "Installation cancelled."; exit 1 ;;
  esac
fi
cp "$BINARY" "$DEST"
chmod 755 "$DEST"
echo "Installed AWM at $DEST"
case ":$PATH:" in
  *":$DEST_DIR:"*) ;;
  *) echo "Add this to your shell profile: export PATH=\"\$HOME/.local/bin:\$PATH\"" ;;
esac
echo "Run: awm install"

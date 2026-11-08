#!/bin/sh
set -eu

section() { printf '\n== %s ==\n' "$1"; }

section "Runtime versions"
if command -v ai-memory >/dev/null 2>&1; then
  printf 'ai-memory path: %s\n' "$(command -v ai-memory)"
  ai-memory --version 2>&1 | sed -n '1,3p'
else
  echo "ai-memory CLI: not found on PATH"
fi
if command -v opencode >/dev/null 2>&1; then
  printf 'opencode path: %s\n' "$(command -v opencode)"
  opencode --version 2>&1 | sed -n '1,3p'
else
  echo "opencode CLI: not found on PATH"
fi

section "Generated OpenCode ai-memory plugin"
PLUGIN="$HOME/.config/opencode/plugins/ai-memory-opencode2.ts"
if [ -f "$PLUGIN" ]; then
  printf 'Plugin file: %s\n' "$PLUGIN"
  grep -nE 'CAPTURE_MODE|CAPTURE_ASSISTANT|findMarker|findSettingsMarker|capturePolicy|repoRootProject' "$PLUGIN" \
    | sed -E 's/(TOKEN|AUTH_TOKEN|SERVER_URL)[^,;]*/\1=<redacted>/g' || true
else
  echo "Known V2 plugin path not found: $PLUGIN"
  echo "Inspect the active OpenCode plugin path manually."
fi

section "ai-memory config candidates (paths only)"
if [ -n "${AI_MEMORY_DATA_DIR:-}" ]; then
  printf 'Current-shell AI_MEMORY_DATA_DIR: %s\n' "$AI_MEMORY_DATA_DIR"
fi
for candidate in \
  "$HOME/Library/Application Support/ai-memory/config.toml" \
  "$HOME/.local/share/ai-memory/config.toml" \
  "$HOME/.config/ai-memory/config.toml" \
  "$HOME/.ai-memory/config.toml"; do
  [ -f "$candidate" ] && printf 'FOUND %s\n' "$candidate"
done

section "macOS LaunchAgent arguments (safe fields only)"
PLIST="$HOME/Library/LaunchAgents/com.github.akitaonrails.ai-memory.plist"
if [ -f "$PLIST" ] && command -v plutil >/dev/null 2>&1; then
  python3 - "$PLIST" <<'PY'
import plistlib, sys
from pathlib import Path
p = Path(sys.argv[1])
try:
    data = plistlib.loads(p.read_bytes())
except Exception as exc:
    print(f"LaunchAgent plist parse failed: {type(exc).__name__}")
    raise SystemExit(0)
args = data.get("ProgramArguments", [])
print(f"LaunchAgent: {p}")
for i, value in enumerate(args):
    if value in ("--data-dir", "--config") and i + 1 < len(args):
        print(f"{value}: {args[i + 1]}")
print("No config argument shown means the service may use the platform default.")
PY
else
  echo "No standard ai-memory LaunchAgent found. Check the active service manager or container."
fi

section "Docker containers (names and images only)"
if command -v docker >/dev/null 2>&1; then
  CONTAINERS="$(docker ps --format '{{.Names}}' 2>/dev/null | grep -i 'ai-memory' || true)"
  if [ -n "$CONTAINERS" ]; then
    printf '%s\n' "$CONTAINERS" | while IFS= read -r name; do
      [ -n "$name" ] || continue
      docker inspect --format '{{.Name}} {{.Config.Image}}' "$name" 2>/dev/null | sed 's#^/##'
      docker inspect --format '{{range .Mounts}}{{println .Source "->" .Destination}}{{end}}' "$name" 2>/dev/null
    done
  else
    echo "No running container name/image matched ai-memory."
  fi
else
  echo "Docker CLI not found."
fi

section "Safety"
echo "This script did not print environment variables, tokens, provider keys, or config file contents."
echo "It did not write files or change services."

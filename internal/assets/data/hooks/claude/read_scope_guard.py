import json
import re
import sys
from pathlib import PurePosixPath

SENSITIVE_PARTS = {".ssh", ".gnupg", ".aws", ".azure", ".kube", ".docker", "secrets", "secret", "credentials", "credential", "private", "certs"}
SENSITIVE_SUFFIXES = {".pem", ".key", ".p12", ".pfx", ".keystore", ".jks"}
SENSITIVE_NAMES = {".env", ".npmrc", ".pypirc", "credentials", "credentials.json", "secrets.json", "id_rsa", "id_ed25519", "id_ecdsa", "service-account.json"}
BROAD_PATTERNS = {".", "./", "*", "**", "**/*", "**/**", "./**/*"}


def fail(message):
    print(message, file=sys.stderr)
    raise SystemExit(2)


def is_sensitive(raw):
    value = raw.replace("\\", "/").lower()
    parts = [part for part in PurePosixPath(value).parts if part not in {".", ""}]
    name = parts[-1] if parts else value
    if name in SENSITIVE_NAMES:
        return True
    if any(part in SENSITIVE_PARTS for part in parts):
        return True
    if any(name.endswith(suffix) for suffix in SENSITIVE_SUFFIXES):
        return True
    if re.search(r"(^|/)\.env(?:\.[^/]*)?$", value):
        return True
    return False


def main():
    try:
        event = json.load(sys.stdin)
        tool_input = event.get("tool_input") or {}
        tool_name = event.get("tool_name") or event.get("toolName") or ""
        if tool_name == "Read":
            targets = [tool_input.get("file_path") or tool_input.get("path")]
        elif tool_name == "Grep":
            target = tool_input.get("path")
            if not target or target in BROAD_PATTERNS:
                fail("Grep requires a narrow explicit path. Broad search is blocked by AWM guard.")
            targets = [target]
        elif tool_name == "Glob":
            target = tool_input.get("pattern")
            if not target or target in BROAD_PATTERNS:
                fail("Glob requires a narrow explicit pattern. Broad listing is blocked by AWM guard.")
            targets = [target]
        else:
            targets = [tool_input.get("file_path") or tool_input.get("path")]
        if not targets or any(not isinstance(target, str) or not target.strip() for target in targets):
            fail("No explicit read target supplied. Read blocked by AWM guard.")
        for value in targets:
            if is_sensitive(value):
                fail(f"Read blocked for a sensitive path: {value}")
    except SystemExit:
        raise
    except Exception as error:
        fail(f"Read guard failed closed: {error}")


if __name__ == "__main__":
    main()

import argparse
import json
import os
import re
import sys
from pathlib import Path, PurePosixPath


def fail(message):
    print(message, file=sys.stderr)
    raise SystemExit(2)


def collect_paths(value):
    paths = []
    if isinstance(value, dict):
        for key, child in value.items():
            if key in {"file_path", "filePath", "path"} and isinstance(child, str):
                paths.append(child)
            elif key in {"edits", "files", "operations"}:
                paths.extend(collect_paths(child))
    elif isinstance(value, list):
        for child in value:
            paths.extend(collect_paths(child))
    return paths


def exact_allowlist(path, cwd):
    if not path:
        fail("Builder allowlist path is missing.")
    try:
        payload = json.loads(Path(path).expanduser().read_text(encoding="utf-8"))
    except Exception as error:
        fail(f"Cannot read Builder allowlist: {error}")
    raw = payload.get("write_allowlist")
    if not isinstance(raw, list) or not raw:
        fail("Builder write allowlist is empty or invalid.")
    allowed = set()
    for item in raw:
        if not isinstance(item, str) or not item.strip():
            fail("Builder allowlist contains an invalid path.")
        normalized = item.strip().replace("\\", "/")
        candidate = PurePosixPath(normalized)
        if (
            candidate.is_absolute()
            or re.match(r"^[A-Za-z]:", normalized)
            or any(part in {"", ".", ".."} for part in normalized.split("/"))
            or any(ch in normalized for ch in "*?[]")
            or ":" in normalized
        ):
            fail(f"Builder path must be exact and repository-relative: {item!r}")
        allowed.add(candidate.as_posix())
    return allowed


def main():
    parser = argparse.ArgumentParser()
    parser.add_argument("--role", required=True)
    parser.add_argument("--allowlist")
    args = parser.parse_args()
    try:
        event = json.load(sys.stdin)
        tool_input = event.get("tool_input") or {}
        raw_paths = collect_paths(tool_input)
        if not raw_paths:
            fail("No target path supplied. Write blocked.")
        cwd = Path(event.get("cwd") or os.getcwd()).expanduser().resolve()
        allowed = None
        if args.role == "builder-task":
            allowed = exact_allowlist(args.allowlist, cwd)
        elif args.role == "builder":
            fail("Base Builder cannot write. Generate a task-scoped Builder after Human approval.")
        else:
            roots = {
                "conductor": [".agentic-workflow/state", ".agentic-workflow/decisions", ".agentic-workflow/manifest.json"],
                "plan": ["memory/active/plans", ".agentic-workflow/active/plans"],
                "scout": ["memory/active/explorations", ".agentic-workflow/active/explorations", ".ua"],
                "architect": ["memory/active/designs", ".agentic-workflow/active/designs"],
            }.get(args.role)
            if roots is None:
                fail(f"No approved output roots exist for role {args.role!r}.")
        for raw in raw_paths:
            normalized_raw = raw.replace("\\", "/")
            if re.match(r"^[A-Za-z]:", normalized_raw):
                fail(f"Write blocked for an absolute drive path: {raw}")
            requested = Path(raw).expanduser()
            target = (requested if requested.is_absolute() else cwd / requested).resolve()
            try:
                relative = target.relative_to(cwd).as_posix()
            except ValueError:
                fail(f"Write blocked outside repository: {target}")
            if allowed is not None:
                if relative not in allowed:
                    fail(f"Builder write blocked. {relative!r} is not in the exact allowlist.")
            elif not any(relative == root or relative.startswith(root.rstrip("/") + "/") for root in roots):
                fail(f"{args.role} write blocked outside approved output roots: {relative}")
    except SystemExit:
        raise
    except Exception as error:
        fail(f"Write guard failed closed: {error}")


if __name__ == "__main__":
    main()

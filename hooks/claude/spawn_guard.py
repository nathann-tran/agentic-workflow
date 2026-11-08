import argparse
import fnmatch
import json
import sys


def main():
    parser = argparse.ArgumentParser()
    parser.add_argument("--allow", required=True)
    args = parser.parse_args()
    allowed = [name.strip() for name in args.allow.split(",") if name.strip()]
    try:
        event = json.load(sys.stdin)
        tool_input = event.get("tool_input") or {}
        target = (tool_input.get("subagent_type") or tool_input.get("agent_type") or tool_input.get("type") or "").strip()
        if not target:
            raise ValueError("Missing child-agent type. Spawn blocked.")
        if not any(fnmatch.fnmatchcase(target, pattern) for pattern in allowed):
            raise ValueError(f"Child agent {target!r} is not allowed. Allowed: {allowed}")
    except Exception as error:
        print(str(error), file=sys.stderr)
        sys.exit(2)


if __name__ == "__main__":
    main()

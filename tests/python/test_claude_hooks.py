import json
import subprocess
import sys
import tempfile
import unittest
from pathlib import Path

ROOT = Path(__file__).resolve().parents[2]
HOOKS = ROOT / "hooks" / "claude"


class ClaudeHookTests(unittest.TestCase):
    def run_hook(self, script, payload, args=(), cwd=None):
        return subprocess.run(
            [sys.executable, str(HOOKS / script), *args],
            input=json.dumps(payload),
            text=True,
            capture_output=True,
            cwd=cwd,
            check=False,
        )

    def test_read_guard_blocks_env_file(self):
        result = self.run_hook(
            "read_scope_guard.py",
            {"tool_name": "Read", "tool_input": {"file_path": "/repo/config/.env.production"}},
        )
        self.assertEqual(result.returncode, 2)
        self.assertIn("sensitive path", result.stderr)

    def test_read_guard_allows_normal_source(self):
        result = self.run_hook(
            "read_scope_guard.py",
            {"tool_name": "Read", "tool_input": {"file_path": "/repo/src/main.go"}},
        )
        self.assertEqual(result.returncode, 0, result.stderr)

    def test_read_guard_fails_closed_without_path(self):
        result = self.run_hook("read_scope_guard.py", {"tool_name": "Read", "tool_input": {}})
        self.assertEqual(result.returncode, 2)

    def test_read_guard_blocks_grep_without_narrow_path(self):
        result = self.run_hook(
            "read_scope_guard.py",
            {"tool_name": "Grep", "tool_input": {"pattern": "API_KEY"}},
        )
        self.assertEqual(result.returncode, 2)
        self.assertIn("narrow explicit path", result.stderr)

    def test_read_guard_blocks_broad_glob(self):
        result = self.run_hook(
            "read_scope_guard.py",
            {"tool_name": "Glob", "tool_input": {"pattern": "**/*"}},
        )
        self.assertEqual(result.returncode, 2)

    def test_builder_guard_allows_exact_path(self):
        with tempfile.TemporaryDirectory() as temp:
            root = Path(temp)
            allowlist = root / "allowlist.json"
            allowlist.write_text(json.dumps({"write_allowlist": ["src/handler.go"]}), encoding="utf-8")
            (root / "src").mkdir()
            payload = {"cwd": str(root), "tool_input": {"file_path": "src/handler.go"}}
            result = self.run_hook(
                "write_scope_guard.py",
                payload,
                args=("--role", "builder-task", "--allowlist", str(allowlist)),
                cwd=root,
            )
            self.assertEqual(result.returncode, 0, result.stderr)

    def test_builder_guard_rejects_unlisted_path(self):
        with tempfile.TemporaryDirectory() as temp:
            root = Path(temp)
            allowlist = root / "allowlist.json"
            allowlist.write_text(json.dumps({"write_allowlist": ["src/handler.go"]}), encoding="utf-8")
            payload = {"cwd": str(root), "tool_input": {"file_path": "src/secret.go"}}
            result = self.run_hook(
                "write_scope_guard.py",
                payload,
                args=("--role", "builder-task", "--allowlist", str(allowlist)),
                cwd=root,
            )
            self.assertEqual(result.returncode, 2)
            self.assertIn("not in the exact allowlist", result.stderr)

    def test_builder_guard_rejects_path_escape(self):
        with tempfile.TemporaryDirectory() as temp:
            root = Path(temp)
            allowlist = root / "allowlist.json"
            allowlist.write_text(json.dumps({"write_allowlist": ["src/handler.go"]}), encoding="utf-8")
            payload = {"cwd": str(root), "tool_input": {"file_path": "../outside.go"}}
            result = self.run_hook(
                "write_scope_guard.py",
                payload,
                args=("--role", "builder-task", "--allowlist", str(allowlist)),
                cwd=root,
            )
            self.assertEqual(result.returncode, 2)
            self.assertIn("outside repository", result.stderr)

    def test_spawn_guard_allows_only_listed_agents(self):
        allowed = self.run_hook(
            "spawn_guard.py",
            {"tool_input": {"subagent_type": "awm-worker"}},
            args=("--allow", "awm-worker,awm-plan"),
        )
        blocked = self.run_hook(
            "spawn_guard.py",
            {"tool_input": {"subagent_type": "arbitrary-agent"}},
            args=("--allow", "awm-worker,awm-plan"),
        )
        self.assertEqual(allowed.returncode, 0, allowed.stderr)
        self.assertEqual(blocked.returncode, 2)


if __name__ == "__main__":
    unittest.main()

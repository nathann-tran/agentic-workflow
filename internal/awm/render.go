package awm

import (
	"bytes"
	"fmt"
	"io/fs"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/agentic-workflow/awm/internal/assets"
)

var roleIDs = []string{"conductor", "plan", "scout", "architect", "builder", "worker"}

func CoreAgent(role string) ([]byte, error) {
	if !contains(roleIDs, role) {
		return nil, fmt.Errorf("unknown role %q", role)
	}
	return fs.ReadFile(assets.Files, filepath.ToSlash(filepath.Join("data", "core", "agents", role+".md")))
}

func CorePolicy(name string) ([]byte, error) {
	return fs.ReadFile(assets.Files, filepath.ToSlash(filepath.Join("data", "core", "policies", name+".md")))
}

func RenderAgent(harnessID, role, home string) ([]byte, error) {
	if !contains(roleIDs, role) {
		return nil, fmt.Errorf("unknown role %q", role)
	}
	canonical, err := CoreAgent(role)
	if err != nil {
		return nil, err
	}
	body := "<!-- Managed by Agentic Workflow Manager -->\n\n" + stripFrontmatter(string(canonical))
	body = strings.TrimSpace(body) + "\n\n" + runtimeContract(home, role, harnessID)
	switch harnessID {
	case "opencode":
		return renderOpenCode(role, body, home), nil
	case "claude-code":
		return renderClaude(role, body, home), nil
	case "codex-cli":
		return renderCodex(role, body), nil
	case "omp":
		return renderOMP(role, body), nil
	default:
		return nil, fmt.Errorf("unsupported harness %q", harnessID)
	}
}

func runtimeContract(home, role, harness string) string {
	var references []string
	coreRoot := filepath.Join(home, ".agentic-workflow", "core")
	references = append(references,
		"`"+filepath.Join(coreRoot, "CONTRACTS.md")+"`",
		"`"+filepath.Join(coreRoot, "policies", "global-principles.md")+"`",
		"`"+filepath.Join(coreRoot, "policies", "communication.md")+"`",
		"`"+filepath.Join(coreRoot, "policies", "human-approval.md")+"`",
	)
	return fmt.Sprintf(`## AWM Runtime Contract

The Harness-Neutral Core is authoritative. Read the following files before substantive work:

- %s

Use the selected harness's documented tools and permissions. Do not assume capability parity across harnesses.

Keep workflow_id, task_id, workspace_id, project_id, repository_id, harness_id, and native_session_id separate.

Use the model selected by the Human or the harness. Do not choose or hard-code a provider-specific model ID.

Load skills by exact runtime ID. Confirm each load. Do not claim a skill or plugin ran without runtime evidence.

Use short, clear sentences. Express one main idea per sentence. Preserve evidence, requirements, risks, caveats, and decisions.

For English, follow Simplified Technical English principles where practical. For Vietnamese, use equivalent clear-language rules. Do not claim formal ASD-STE100 compliance.

Choose diagrams or interactive artifacts only when they improve understanding. Do not replace useful evidence with visual decoration.

This role runs in the %s adapter. The current role is %s.
`, strings.Join(references, ", "), harness, role)
}

func renderOpenCode(role, body, home string) []byte {
	mode := "subagent"
	if role == "conductor" {
		mode = "primary"
	}
	if role == "plan" {
		mode = "all"
	}
	var b strings.Builder
	b.WriteString("---\n")
	b.WriteString("description: ")
	b.WriteString(opencodeDescription(role))
	b.WriteString("\nmode: ")
	b.WriteString(mode)
	b.WriteString("\npermissions:\n")
	for _, rule := range openCodeRules(role, home) {
		fmt.Fprintf(&b, "  - action: %s\n    resource: %q\n    effect: %s\n", rule.Action, rule.Resource, rule.Effect)
	}
	if role == "builder" {
		b.WriteString("hidden: true\n")
	}
	b.WriteString("---\n\n")
	b.WriteString(body)
	if role == "builder" {
		b.WriteString("\n\n## Builder activation\n\nThis base profile cannot edit files. Use a task-scoped Builder profile after Human approval. Use an isolated worktree. Shell commands require Human approval. AWM marks shell side effects as a residual risk.\n")
	}
	return []byte(b.String())
}

type permissionRule struct {
	Action   string
	Resource string
	Effect   string
}

func openCodeRules(role, home string) []permissionRule {
	allowRead := []permissionRule{
		{Action: "read", Resource: "*", Effect: "allow"},
		{Action: "glob", Resource: "*", Effect: "allow"},
		{Action: "grep", Resource: "*", Effect: "ask"},
		{Action: "read", Resource: "*.env", Effect: "deny"},
		{Action: "read", Resource: "*.env.*", Effect: "deny"},
		{Action: "read", Resource: "**/secrets/**", Effect: "deny"},
		{Action: "read", Resource: "**/*.pem", Effect: "deny"},
		{Action: "read", Resource: "**/*.key", Effect: "deny"},
		{Action: "read", Resource: "**/*.p12", Effect: "deny"},
		{Action: "read", Resource: "**/*.pfx", Effect: "deny"},
		{Action: "grep", Resource: "*.env", Effect: "deny"},
		{Action: "grep", Resource: "*.env.*", Effect: "deny"},
		{Action: "grep", Resource: "**/secrets/**", Effect: "deny"},
		{Action: "grep", Resource: "**/*.pem", Effect: "deny"},
		{Action: "grep", Resource: "**/*.key", Effect: "deny"},
		{Action: "grep", Resource: "**/*.p12", Effect: "deny"},
		{Action: "grep", Resource: "**/*.pfx", Effect: "deny"},
		{Action: "external_directory", Resource: "*", Effect: "deny"},
		{Action: "external_directory", Resource: filepath.ToSlash(filepath.Join(home, ".agentic-workflow", "core", "**")), Effect: "allow"},
		{Action: "read", Resource: filepath.ToSlash(filepath.Join(home, ".agentic-workflow", "core", "**")), Effect: "allow"},
		{Action: "skill", Resource: "*", Effect: "ask"},
		{Action: "skill", Resource: "awm-*", Effect: "allow"},
	}
	editAll := permissionRule{Action: "edit", Resource: "*", Effect: "deny"}
	switch role {
	case "conductor":
		return append(allowRead,
			editAll,
			permissionRule{Action: "edit", Resource: ".agentic-workflow/state/**", Effect: "allow"},
			permissionRule{Action: "edit", Resource: ".agentic-workflow/decisions/**", Effect: "allow"},
			permissionRule{Action: "edit", Resource: ".agentic-workflow/manifest.json", Effect: "allow"},
			permissionRule{Action: "subagent", Resource: "*", Effect: "deny"},
			permissionRule{Action: "subagent", Resource: "awm/plan", Effect: "allow"},
			permissionRule{Action: "subagent", Resource: "awm/scout", Effect: "allow"},
			permissionRule{Action: "subagent", Resource: "awm/architect", Effect: "allow"},
			permissionRule{Action: "subagent", Resource: "awm/worker", Effect: "allow"},
			permissionRule{Action: "subagent", Resource: "awm/builder-task-*", Effect: "allow"},
			permissionRule{Action: "shell", Resource: "*", Effect: "ask"},
			permissionRule{Action: "question", Resource: "*", Effect: "allow"},
		)
	case "plan":
		return append(allowRead, editAll,
			permissionRule{Action: "edit", Resource: "memory/active/plans/**", Effect: "allow"},
			permissionRule{Action: "edit", Resource: ".agentic-workflow/active/plans/**", Effect: "allow"},
			permissionRule{Action: "subagent", Resource: "*", Effect: "deny"},
			permissionRule{Action: "subagent", Resource: "awm/worker", Effect: "allow"},
			permissionRule{Action: "shell", Resource: "*", Effect: "deny"},
			permissionRule{Action: "question", Resource: "*", Effect: "allow"},
		)
	case "scout":
		return append(allowRead, editAll,
			permissionRule{Action: "edit", Resource: "memory/active/explorations/**", Effect: "allow"},
			permissionRule{Action: "edit", Resource: ".agentic-workflow/active/explorations/**", Effect: "allow"},
			permissionRule{Action: "edit", Resource: ".ua/**", Effect: "allow"},
			permissionRule{Action: "subagent", Resource: "*", Effect: "deny"},
			permissionRule{Action: "subagent", Resource: "awm/worker", Effect: "allow"},
			permissionRule{Action: "shell", Resource: "*", Effect: "ask"},
		)
	case "architect":
		return append(allowRead, editAll,
			permissionRule{Action: "edit", Resource: "memory/active/designs/**", Effect: "allow"},
			permissionRule{Action: "edit", Resource: ".agentic-workflow/active/designs/**", Effect: "allow"},
			permissionRule{Action: "subagent", Resource: "*", Effect: "deny"},
			permissionRule{Action: "subagent", Resource: "awm/worker", Effect: "allow"},
			permissionRule{Action: "shell", Resource: "*", Effect: "deny"},
		)
	case "builder":
		return append(allowRead, editAll,
			permissionRule{Action: "subagent", Resource: "*", Effect: "deny"},
			permissionRule{Action: "subagent", Resource: "awm/worker", Effect: "allow"},
			permissionRule{Action: "shell", Resource: "*", Effect: "deny"},
		)
	case "worker":
		return append(allowRead, editAll,
			permissionRule{Action: "subagent", Resource: "*", Effect: "deny"},
			permissionRule{Action: "shell", Resource: "*", Effect: "deny"},
			permissionRule{Action: "webfetch", Resource: "*", Effect: "deny"},
			permissionRule{Action: "websearch", Resource: "*", Effect: "deny"},
		)
	}
	return allowRead
}

func opencodeDescription(role string) string {
	items := map[string]string{
		"conductor": "Coordinates AWM workflow phases, routing, state, and Human approval gates.",
		"plan":      "Clarifies requirements and creates testable investigation plans.",
		"scout":     "Investigates repository behavior and reports verified evidence.",
		"architect": "Designs implementation boundaries, interfaces, risks, and rollback.",
		"builder":   "Read-only Builder base. Use a task-scoped profile after approval.",
		"worker":    "Read-only helper for one bounded task.",
	}
	return items[role]
}

func renderClaude(role, body, home string) []byte {
	name := "awm-" + role
	description := opencodeDescription(role)
	tools := map[string]string{
		"conductor": "Agent, Skill, Read, Grep, Glob, Write, Edit, AskUserQuestion",
		"plan":      "Agent, Skill, Read, Grep, Glob, Write, Edit, AskUserQuestion",
		"scout":     "Agent, Skill, Read, Grep, Glob, Write, Edit",
		"architect": "Agent, Skill, Read, Grep, Glob, Write, Edit",
		"builder":   "Read, Grep, Glob",
		"worker":    "Read, Grep, Glob",
	}
	var b strings.Builder
	fmt.Fprintf(&b, "---\nname: %s\ndescription: %q\ntools: %s\n", name, description, tools[role])
	if role == "conductor" || role == "plan" || role == "scout" || role == "architect" || role == "builder" || role == "worker" {
		b.WriteString("hooks:\n  PreToolUse:\n")
		command := "python3 " + shellQuote(filepath.Join(home, ".agentic-workflow", "hooks", "claude", "read_scope_guard.py"))
		fmt.Fprintf(&b, "    - matcher: Read|Grep|Glob\n      hooks:\n        - type: command\n          command: %s\n          timeout: 10\n", strconv.Quote(command))
		if role != "worker" && role != "builder" {
			command = "python3 " + shellQuote(filepath.Join(home, ".agentic-workflow", "hooks", "claude", "write_scope_guard.py")) + " --role " + shellQuote(role)
			fmt.Fprintf(&b, "    - matcher: Write|Edit\n      hooks:\n        - type: command\n          command: %s\n          timeout: 10\n", strconv.Quote(command))
		}
		if role == "conductor" || role == "plan" || role == "scout" || role == "architect" {
			command = "python3 " + shellQuote(filepath.Join(home, ".agentic-workflow", "hooks", "claude", "spawn_guard.py")) + " --allow " + shellQuote(strings.Join(claudeSpawnAllowlist(role), ","))
			fmt.Fprintf(&b, "    - matcher: Agent\n      hooks:\n        - type: command\n          command: %s\n          timeout: 10\n", strconv.Quote(command))
		}
	}
	b.WriteString("---\n\n")
	b.WriteString(body)
	return []byte(b.String())
}

func claudeSpawnAllowlist(role string) []string {
	switch role {
	case "conductor":
		return []string{"awm-plan", "awm-scout", "awm-architect", "awm-worker", "awm-builder-task-*"}
	case "plan", "scout", "architect", "builder":
		return []string{"awm-worker"}
	default:
		return nil
	}
}

func renderCodex(role, body string) []byte {
	name := "awm_" + strings.ReplaceAll(role, "-", "_")
	var b strings.Builder
	b.WriteString("# Managed by Agentic Workflow Manager\n")
	fmt.Fprintf(&b, "name = %q\ndescription = %q\n", name, opencodeDescription(role))
	if role == "worker" {
		body += "\n\nCodex limitation: this custom agent inherits the parent sandbox and approvals. Do not accept tasks that need stronger isolation."
	}
	if role == "builder" {
		body += "\n\nThis base Builder profile is read-only. Do not implement source changes in this subagent. Use a separate Human-approved writable Codex session and an isolated worktree. The adapter cannot enforce exact per-file scope."
	}
	b.WriteString("developer_instructions = \"\"\"\n")
	b.WriteString(strings.ReplaceAll(body, "\"\"\"", "'''") + "\n")
	b.WriteString("\"\"\"\n")
	return []byte(b.String())
}

func renderOMP(role, body string) []byte {
	name := "awm-" + role
	tools := map[string]string{
		"conductor": "read,grep,glob,find,edit,write,task",
		"plan":      "read,grep,glob,find,edit,write,task",
		"scout":     "read,grep,glob,find,edit,task",
		"architect": "read,grep,glob,find,edit,task",
		"builder":   "read,grep,glob,find,task",
		"worker":    "read,grep,glob,find",
	}
	var spawns string
	switch role {
	case "conductor":
		spawns = "awm-plan,awm-scout,awm-architect,awm-worker,awm-builder"
	case "plan", "scout", "architect", "builder":
		spawns = "awm-worker"
	}
	var b strings.Builder
	fmt.Fprintf(&b, "---\nname: %s\ndescription: %q\ntools: %s\n", name, opencodeDescription(role), tools[role])
	if spawns != "" {
		fmt.Fprintf(&b, "spawns: %s\n", spawns)
	}
	if role == "worker" {
		b.WriteString("read-summarize: false\n")
	}
	b.WriteString("---\n\n")
	b.WriteString(body)
	if role == "builder" {
		b.WriteString("\n\nThis base Builder profile is read-only. OMP cannot enforce exact file paths through agent frontmatter. Use a separate Human-approved writable session and an isolated worktree.\n")
	}
	return []byte(b.String())
}

func stripFrontmatter(source string) string {
	trimmed := strings.TrimSpace(source)
	if !strings.HasPrefix(trimmed, "---\n") {
		return trimmed
	}
	parts := strings.SplitN(trimmed, "\n---\n", 2)
	if len(parts) == 2 {
		return strings.TrimSpace(parts[1])
	}
	return trimmed
}

func contains(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}

func RenderGuidance(harnessID, home string) ([]byte, error) {
	guidance := fmt.Sprintf(`# Agentic Workflow Manager (AWM)

<!-- Managed by Agentic Workflow Manager -->

Use the AWM Conductor as the primary workflow coordinator for complex engineering work.

- Use Plan → Scout → Architect → Builder by default.
- Ask the Human to approve requirements and architecture before implementation.
- Route specialist-to-specialist handoffs through Conductor.
- Read role definitions and policies before substantive work.
- Load skills by exact runtime ID. Confirm each skill load.
- Preserve native permissions and documented adapter limitations.
- Do not change memory settings or delete memory without Human approval.

Core contract: %s
Core policies: %s
`, filepath.Join(home, ".agentic-workflow", "core", "CONTRACTS.md"), filepath.Join(home, ".agentic-workflow", "core", "policies"))
	if harnessID != "codex-cli" && harnessID != "omp" && harnessID != "claude-code" {
		return nil, fmt.Errorf("global guidance is not supported for %s", harnessID)
	}
	return []byte(guidance), nil
}

func RenderTaskBuilder(taskID, harnessID, role, writeAllowlist, home string) ([]byte, error) {
	if role != "builder" {
		return nil, fmt.Errorf("task-scoped generation supports Builder only")
	}
	if strings.TrimSpace(taskID) == "" || strings.ContainsAny(taskID, "/\\\n\r") {
		return nil, fmt.Errorf("task ID must be a simple identifier")
	}
	canonical, err := CoreAgent("builder")
	if err != nil {
		return nil, err
	}
	body := "<!-- Managed by Agentic Workflow Manager -->\n\n" + stripFrontmatter(string(canonical))
	body += "\n\n## Approved task write scope\n\n" + writeAllowlist
	body += "\n\nUse an isolated worktree. Do not write outside the exact allowlist. Run only Human-approved commands.\n"
	if harnessID == "opencode" {
		var b bytes.Buffer
		b.WriteString("---\ndescription: Task-scoped AWM Builder. Use only after Human approval.\nmode: subagent\npermissions:\n")
		for _, rule := range openCodeRules("builder", home) {
			if rule.Action == "edit" || rule.Action == "subagent" || rule.Action == "shell" {
				continue
			}
			fmt.Fprintf(&b, "  - action: %s\n    resource: %q\n    effect: %s\n", rule.Action, rule.Resource, rule.Effect)
		}
		b.WriteString("  - action: edit\n    resource: '*'\n    effect: deny\n")
		for _, line := range strings.Split(writeAllowlist, "\n") {
			path := strings.TrimSpace(strings.TrimPrefix(line, "- "))
			if path == "" {
				continue
			}
			fmt.Fprintf(&b, "  - action: edit\n    resource: %q\n    effect: allow\n", path)
		}
		b.WriteString("  - action: shell\n    resource: '*'\n    effect: ask\n")
		b.WriteString("  - action: subagent\n    resource: '*'\n    effect: deny\n")
		b.WriteString("  - action: subagent\n    resource: 'awm/worker'\n    effect: allow\n")
		b.WriteString("---\n\n")
		b.WriteString(body)
		b.WriteString("\n\nShell operations can still have side effects. Human approval and worktree isolation remain required.\n")
		return b.Bytes(), nil
	}
	if harnessID == "claude-code" {
		var b strings.Builder
		name := "awm-builder-task-" + taskID
		fmt.Fprintf(&b, "---\nname: %s\ndescription: Task-scoped Builder. Use only after Human approval.\nisolation: worktree\ntools: Skill, Read, Grep, Glob, Write, Edit, Bash, Agent\nhooks:\n  PreToolUse:\n", name)
		command := "python3 " + shellQuote(filepath.Join(home, ".agentic-workflow", "hooks", "claude", "write_scope_guard.py")) + " --role builder-task --allowlist " + shellQuote(filepath.Join(home, ".agentic-workflow", "tasks", taskID, "write-allowlist.json"))
		fmt.Fprintf(&b, "    - matcher: Write|Edit\n      hooks:\n        - type: command\n          command: %s\n          timeout: 10\n", strconv.Quote(command))
		command = "python3 " + shellQuote(filepath.Join(home, ".agentic-workflow", "hooks", "claude", "read_scope_guard.py"))
		fmt.Fprintf(&b, "    - matcher: Read|Grep|Glob\n      hooks:\n        - type: command\n          command: %s\n          timeout: 10\n", strconv.Quote(command))
		command = "python3 " + shellQuote(filepath.Join(home, ".agentic-workflow", "hooks", "claude", "spawn_guard.py")) + " --allow awm-worker"
		fmt.Fprintf(&b, "    - matcher: Agent\n      hooks:\n        - type: command\n          command: %s\n          timeout: 10\n---\n\n", strconv.Quote(command))
		b.WriteString(body)
		b.WriteString("\n\nBash operations can still have side effects. Human approval and worktree isolation remain required.\n")
		return []byte(b.String()), nil
	}
	return nil, fmt.Errorf("task-scoped Builder rendering is not available for %s; use documented fallback", harnessID)
}

func shellQuote(value string) string {
	return "'" + strings.ReplaceAll(value, "'", "'\\''") + "'"
}

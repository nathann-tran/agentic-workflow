package awm

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

type denyPrompter struct{}

func (denyPrompter) Confirm(string, bool) bool { return false }

func TestValidateWriteAllowlist(t *testing.T) {
	valid, err := ValidateWriteAllowlist([]string{"src/handler.go", "tests/handler_test.go", "src/handler.go"})
	if err != nil {
		t.Fatalf("valid paths rejected: %v", err)
	}
	if len(valid) != 2 {
		t.Fatalf("expected duplicate paths to collapse, got %v", valid)
	}
	invalid := [][]string{
		{"../outside.go"},
		{"/tmp/outside.go"},
		{"src/../outside.go"},
		{"src/*.go"},
		{"src/"},
		{"C:\\temp\\outside.go"},
		{"src//handler.go"},
		{"./handler.go"},
	}
	for _, paths := range invalid {
		t.Run(strings.Join(paths, "_"), func(t *testing.T) {
			if _, err := ValidateWriteAllowlist(paths); err == nil {
				t.Fatalf("expected unsafe path to fail: %q", paths[0])
			}
		})
	}
}

func TestReadApprovedBuilderTaskRequiresHumanApproval(t *testing.T) {
	path := filepath.Join("..", "..", "examples", "task-envelope-builder.json")
	if _, _, err := ReadApprovedBuilderTask(path, false); err == nil {
		t.Fatal("expected Human approval requirement")
	}
	task, allowlist, err := ReadApprovedBuilderTask(path, true)
	if err != nil {
		t.Fatalf("valid example task rejected: %v", err)
	}
	if task.TaskID != "TASK-001" || len(allowlist) != 2 {
		t.Fatalf("unexpected parsed task: task=%s allowlist=%v", task.TaskID, allowlist)
	}
}

func TestReadApprovedBuilderTaskRequiresExplicitSafetyConstraints(t *testing.T) {
	original, err := os.ReadFile(filepath.Join("..", "..", "examples", "task-envelope-builder.json"))
	if err != nil {
		t.Fatal(err)
	}
	for name, replace := range map[string]func(map[string]any){
		"human-approval": func(payload map[string]any) {
			constraints := payload["constraints"].(map[string]any)
			constraints["human_approval_required"] = false
		},
		"external-mutation": func(payload map[string]any) {
			constraints := payload["constraints"].(map[string]any)
			constraints["external_mutation_allowed"] = true
		},
		"delegation-depth": func(payload map[string]any) {
			constraints := payload["constraints"].(map[string]any)
			constraints["max_delegation_depth"] = 2
		},
	} {
		t.Run(name, func(t *testing.T) {
			var payload map[string]any
			if err := json.Unmarshal(original, &payload); err != nil {
				t.Fatal(err)
			}
			replace(payload)
			modified, err := json.Marshal(payload)
			if err != nil {
				t.Fatal(err)
			}
			taskFile := filepath.Join(t.TempDir(), "task.json")
			if err := os.WriteFile(taskFile, modified, 0o600); err != nil {
				t.Fatal(err)
			}
			if _, _, err := ReadApprovedBuilderTask(taskFile, true); err == nil {
				t.Fatal("unsafe Builder task constraints must be rejected")
			}
		})
	}
}

func TestLoadInstallStateRejectsSymlink(t *testing.T) {
	home := t.TempDir()
	stateDir := filepath.Join(home, ".agentic-workflow", "state")
	if err := os.MkdirAll(stateDir, 0o755); err != nil {
		t.Fatal(err)
	}
	external := filepath.Join(t.TempDir(), "install.json")
	if err := os.WriteFile(external, []byte(`{"version":"0.2.0"}`), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(external, filepath.Join(stateDir, "install.json")); err != nil {
		t.Skipf("symlink creation unavailable: %v", err)
	}
	if _, err := LoadInstallState(home); err == nil {
		t.Fatal("expected symlinked install state to be rejected")
	}
	if _, _, err := Rollback(home, false); err == nil {
		t.Fatal("expected rollback to reject symlinked install state")
	}
}

func TestReadApprovedBuilderTaskRejectsUnknownFields(t *testing.T) {
	data, err := os.ReadFile(filepath.Join("..", "..", "examples", "task-envelope-builder.json"))
	if err != nil {
		t.Fatal(err)
	}
	var payload map[string]any
	if err := json.Unmarshal(data, &payload); err != nil {
		t.Fatal(err)
	}
	payload["unexpected"] = true
	modified, err := json.Marshal(payload)
	if err != nil {
		t.Fatal(err)
	}
	temp := filepath.Join(t.TempDir(), "task.json")
	if err := os.WriteFile(temp, modified, 0o600); err != nil {
		t.Fatal(err)
	}
	if _, _, err := ReadApprovedBuilderTask(temp, true); err == nil {
		t.Fatal("expected unknown field to fail strict decoding")
	}
}

func TestSelectedHarnessesDoesNotAddUnselectedRuntimes(t *testing.T) {
	harnesses := []DetectedHarness{
		{Harness: Harness{ID: "opencode"}, Detected: true},
		{Harness: Harness{ID: "claude-code"}, Detected: true},
		{Harness: Harness{ID: "codex-cli"}, Detected: true},
		{Harness: Harness{ID: "omp"}, Detected: true},
	}
	selected, err := SelectedHarnesses(harnesses, []string{"opencode", "codex-cli"}, false)
	if err != nil {
		t.Fatal(err)
	}
	if len(selected) != 2 || selected[0].ID != "opencode" || selected[1].ID != "codex-cli" {
		t.Fatalf("selected unexpected harnesses: %v", selected)
	}
}

func TestSelectedHarnessesSelectAllOnlyWhenNoIDsProvided(t *testing.T) {
	harnesses := []DetectedHarness{
		{Harness: Harness{ID: "opencode"}, Detected: true},
		{Harness: Harness{ID: "claude-code"}, Detected: false},
		{Harness: Harness{ID: "codex-cli"}, Detected: true},
	}
	selected, err := SelectedHarnesses(harnesses, nil, false)
	if err != nil {
		t.Fatal(err)
	}
	if len(selected) != 2 || selected[0].ID != "opencode" || selected[1].ID != "codex-cli" {
		t.Fatalf("select-all returned unexpected harnesses: %v", selected)
	}
}

func TestRenderAgentCapabilities(t *testing.T) {
	home := filepath.Join(t.TempDir(), "home with spaces")
	for _, harness := range []string{"opencode", "claude-code", "codex-cli", "omp"} {
		for _, role := range roleIDs {
			content, err := RenderAgent(harness, role, home)
			if err != nil {
				t.Fatalf("render %s/%s: %v", harness, role, err)
			}
			if !bytes.Contains(content, []byte("Agentic Workflow Manager")) {
				t.Errorf("rendered %s/%s lacks AWM marker", harness, role)
			}
		}
	}

	openCodeBuilder, _ := RenderAgent("opencode", "builder", home)
	if !bytes.Contains(openCodeBuilder, []byte("hidden: true")) || !bytes.Contains(openCodeBuilder, []byte("This base profile cannot edit files")) {
		t.Error("OpenCode base Builder must be hidden and read-only")
	}
	if !strings.Contains(string(openCodeBuilder), "action: grep\n    resource: \"*\"\n    effect: ask") {
		t.Error("OpenCode must ask before Grep because its matcher targets query text")
	}
	openCodeScout, _ := RenderAgent("opencode", "scout", home)
	if !strings.Contains(string(openCodeScout), "action: shell\n    resource: \"*\"\n    effect: ask") {
		t.Error("OpenCode Scout must request approval for graph and repository commands")
	}
	openCodeArchitect, _ := RenderAgent("opencode", "architect", home)
	if !strings.Contains(string(openCodeArchitect), "action: edit\n    resource: \"docs/artifacts/**\"\n    effect: allow") {
		t.Error("OpenCode Architect must be allowed to write the design artifact folder")
	}
	if !strings.Contains(string(openCodeArchitect), "action: shell\n    resource: \"*\"\n    effect: deny") {
		t.Error("OpenCode Architect must not run commands")
	}
	openCodeConductor, _ := RenderAgent("opencode", "conductor", home)
	if !strings.Contains(string(openCodeConductor), "action: edit\n    resource: \"docs/artifacts/**\"\n    effect: allow") {
		t.Error("OpenCode Conductor must be allowed to render design artifacts")
	}
	openCodeWorker, _ := RenderAgent("opencode", "worker", home)
	if !bytes.Contains(openCodeWorker, []byte("effect: deny")) || bytes.Contains(openCodeWorker, []byte("effect: allow\n    resource: \".agentic-workflow/state")) {
		t.Error("OpenCode Worker must remain read-only")
	}

	claudeBuilder, _ := RenderAgent("claude-code", "builder", home)
	if !bytes.Contains(claudeBuilder, []byte("tools: Read, Grep, Glob")) || bytes.Contains(claudeBuilder, []byte("matcher: Write|Edit")) {
		t.Error("Claude base Builder must expose read tools only")
	}
	claudeScout, _ := RenderAgent("claude-code", "scout", home)
	if !bytes.Contains(claudeScout, []byte("write_scope_guard.py")) || !bytes.Contains(claudeScout, []byte("spawn_guard.py")) {
		t.Error("Claude Scout must have scoped writes and Worker spawn guards")
	}

	codexBuilder, _ := RenderAgent("codex-cli", "builder", home)
	if !bytes.Contains(codexBuilder, []byte("read-only")) || !bytes.Contains(codexBuilder, []byte("exact per-file scope")) {
		t.Error("Codex base Builder must document its read-only fallback")
	}

	ompBuilder, _ := RenderAgent("omp", "builder", home)
	if !bytes.Contains(ompBuilder, []byte("tools: read,grep,glob,find,task")) || !bytes.Contains(ompBuilder, []byte("This base Builder profile is read-only")) {
		t.Error("OMP base Builder must expose read tools only")
	}
	ompWorker, _ := RenderAgent("omp", "worker", home)
	if !bytes.Contains(ompWorker, []byte("tools: read,grep,glob,find")) {
		t.Error("OMP Worker must expose read-only tools")
	}
}

func TestRenderTaskBuilderRequiresExactScopeAndAllowsWorker(t *testing.T) {
	home := filepath.Join(t.TempDir(), "home with spaces")
	allowlist := "- src/handler.go\n- tests/handler_test.go"
	content, err := RenderTaskBuilder("TASK-001", "opencode", "builder", allowlist, home)
	if err != nil {
		t.Fatal(err)
	}
	text := string(content)
	if !strings.Contains(text, `resource: "src/handler.go"`) || !strings.Contains(text, `resource: "tests/handler_test.go"`) {
		t.Fatal("OpenCode task Builder must include exact approved paths")
	}
	if !strings.Contains(text, `resource: 'awm/worker'`) {
		t.Fatal("task Builder must be allowed to spawn a bounded Worker")
	}
	claudeTask, err := RenderTaskBuilder("TASK-001", "claude-code", "builder", allowlist, home)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Contains(claudeTask, []byte("isolation: worktree")) {
		t.Fatal("Claude task Builder must use native worktree isolation")
	}
	if strings.Contains(text, `resource: '*'\n    effect: allow`) {
		t.Fatal("task Builder must not receive a wildcard edit grant")
	}
}

func TestSecurePathWithinRejectsSymlinkEscape(t *testing.T) {
	root := t.TempDir()
	outside := t.TempDir()
	link := filepath.Join(root, "escape")
	if err := os.Symlink(outside, link); err != nil {
		t.Skipf("symlink creation unavailable: %v", err)
	}
	if securePathWithin(root, filepath.Join(link, "file.txt")) {
		t.Fatal("expected symlink escape to be rejected")
	}
	if !securePathWithin(root, filepath.Join(root, "safe", "file.txt")) {
		t.Fatal("expected a new path below home to be accepted")
	}
}

func TestRunInstallAndRollback(t *testing.T) {
	home := t.TempDir()
	result, err := RunInstall(InstallOptions{
		HarnessIDs: []string{"opencode"}, AllowUndetected: true,
		SkipSkills: true, Home: home,
	}, denyPrompter{})
	if err != nil {
		t.Fatalf("install failed: %v", err)
	}
	if len(result.State.ManagedFiles) < 20 {
		t.Fatalf("expected core and agent files, got %d", len(result.State.ManagedFiles))
	}
	agentPath := filepath.Join(home, ".config", "opencode", "agents", "awm", "conductor.md")
	if _, err := os.Stat(agentPath); err != nil {
		t.Fatalf("Conductor profile not installed: %v", err)
	}
	if _, err := LoadInstallState(home); err != nil {
		t.Fatalf("install state missing: %v", err)
	}
	restored, skipped, err := Rollback(home, false)
	if err != nil {
		t.Fatalf("rollback failed: %v", err)
	}
	if len(skipped) != 0 {
		t.Fatalf("rollback skipped unchanged files: %v", skipped)
	}
	if len(restored) == 0 {
		t.Fatal("rollback should restore managed files")
	}
	if _, err := os.Stat(agentPath); !os.IsNotExist(err) {
		t.Fatalf("agent profile still exists after rollback: %v", err)
	}
	if _, err := LoadInstallState(home); !os.IsNotExist(err) {
		t.Fatalf("install state remains after rollback: %v", err)
	}
}

func TestRollbackPreservesFileChangedAfterInstall(t *testing.T) {
	home := t.TempDir()
	_, err := RunInstall(InstallOptions{
		HarnessIDs: []string{"opencode"}, AllowUndetected: true,
		SkipSkills: true, Home: home,
	}, denyPrompter{})
	if err != nil {
		t.Fatal(err)
	}
	agentPath := filepath.Join(home, ".config", "opencode", "agents", "awm", "conductor.md")
	if err := os.WriteFile(agentPath, []byte("Human edit; keep this file.\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	_, skipped, err := Rollback(home, false)
	if err != nil {
		t.Fatal(err)
	}
	if len(skipped) == 0 {
		t.Fatal("rollback should skip a file changed after installation")
	}
	content, err := os.ReadFile(agentPath)
	if err != nil {
		t.Fatal(err)
	}
	if string(content) != "Human edit; keep this file.\n" {
		t.Fatal("rollback overwrote a Human change")
	}
}

func TestInstallerDoesNotReplaceUnmanagedFileWithoutApproval(t *testing.T) {
	home := t.TempDir()
	target := filepath.Join(home, ".config", "opencode", "agents", "awm", "conductor.md")
	if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
		t.Fatal(err)
	}
	original := []byte("User-managed profile.\n")
	if err := os.WriteFile(target, original, 0o644); err != nil {
		t.Fatal(err)
	}
	result, err := RunInstall(InstallOptions{
		HarnessIDs: []string{"opencode"}, AllowUndetected: true,
		SkipSkills: true, Home: home,
	}, denyPrompter{})
	if err != nil {
		t.Fatal(err)
	}
	content, err := os.ReadFile(target)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(content, original) {
		t.Fatal("installer changed an unmanaged profile without approval")
	}
	if !contains(result.Skipped, target) {
		t.Fatalf("expected unmanaged file to appear in skipped results: %v", result.Skipped)
	}
	state, err := LoadInstallState(home)
	if err != nil {
		t.Fatalf("load install state: %v", err)
	}
	if !contains(state.SkippedFiles, target) {
		t.Fatalf("install state must persist skipped files: %v", state.SkippedFiles)
	}
}

func TestRunInstallDryRunDoesNotWriteFiles(t *testing.T) {
	home := t.TempDir()
	result, err := RunInstall(InstallOptions{
		HarnessIDs: []string{"opencode"}, AllowUndetected: true,
		SkipSkills: true, DryRun: true, Home: home,
	}, denyPrompter{})
	if err != nil {
		t.Fatalf("dry run failed: %v", err)
	}
	if len(result.Written) == 0 {
		t.Fatal("dry run should report planned writes")
	}
	if _, err := os.Stat(filepath.Join(home, ".agentic-workflow")); !os.IsNotExist(err) {
		t.Fatalf("dry run changed the home directory: %v", err)
	}
	if _, err := os.Stat(filepath.Join(home, ".config", "opencode")); !os.IsNotExist(err) {
		t.Fatalf("dry run created harness configuration: %v", err)
	}
}

type selectingPrompter struct {
	selected []string
}

func (p selectingPrompter) Confirm(string, bool) bool { return true }

func (p selectingPrompter) SelectSkills([]BundledSkill) ([]string, error) {
	return p.selected, nil
}

func TestInstallerOffersOutdatedSkillAndPreservesOriginalMode(t *testing.T) {
	home := t.TempDir()
	skillFile := filepath.Join(home, ".agents", "skills", "awm-workflow-protocol", "SKILL.md")
	if err := os.MkdirAll(filepath.Dir(skillFile), 0o755); err != nil {
		t.Fatal(err)
	}
	old := []byte("---\nname: awm-workflow-protocol\nversion: 0.1.0\n---\n# Previous skill\n")
	if err := os.WriteFile(skillFile, old, 0o600); err != nil {
		t.Fatal(err)
	}
	result, err := RunInstall(InstallOptions{
		HarnessIDs: []string{"opencode"}, AllowUndetected: true,
		Home: home,
	}, selectingPrompter{selected: []string{"awm-workflow-protocol"}})
	if err != nil {
		t.Fatalf("install failed: %v", err)
	}
	updated, err := os.ReadFile(skillFile)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Contains(updated, []byte("version: 0.2.0")) {
		t.Fatal("selected outdated skill was not updated")
	}
	if len(result.Skipped) < 3 {
		t.Fatalf("unselected bundled skills should be reported as skipped: %v", result.Skipped)
	}
	_, skipped, err := Rollback(home, false)
	if err != nil {
		t.Fatal(err)
	}
	if len(skipped) > 0 {
		t.Fatalf("rollback unexpectedly skipped files: %v", skipped)
	}
	restored, err := os.ReadFile(skillFile)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(restored, old) {
		t.Fatal("rollback did not restore the previous skill")
	}
	info, err := os.Stat(skillFile)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0o600 {
		t.Fatalf("rollback did not preserve original mode: %o", info.Mode().Perm())
	}
}

func TestRenderAndRollbackTaskScopedBuilder(t *testing.T) {
	home := t.TempDir()
	taskFile := filepath.Join("..", "..", "examples", "task-envelope-builder.json")
	target, err := RenderApprovedBuilderTask(taskFile, "opencode", home, "", true, false)
	if err != nil {
		t.Fatalf("render Builder task failed: %v", err)
	}
	if _, err := os.Stat(target); err != nil {
		t.Fatalf("Builder profile missing: %v", err)
	}
	allowlist := filepath.Join(home, ".agentic-workflow", "tasks", "TASK-001", "write-allowlist.json")
	if _, err := os.Stat(allowlist); err != nil {
		t.Fatalf("Builder allowlist missing: %v", err)
	}
	if _, err := RenderApprovedBuilderTask(taskFile, "opencode", home, "", true, false); err == nil {
		t.Fatal("rendering the same task twice without rollback must fail")
	}
	restored, skipped, err := RollbackBuilderTask(home, "TASK-001", false)
	if err != nil {
		t.Fatalf("Builder task rollback failed: %v", err)
	}
	if len(skipped) != 0 || len(restored) != 2 {
		t.Fatalf("unexpected task rollback result: restored=%v skipped=%v", restored, skipped)
	}
	if _, err := os.Stat(target); !os.IsNotExist(err) {
		t.Fatalf("Builder profile remains after rollback: %v", err)
	}
	if _, err := os.Stat(allowlist); !os.IsNotExist(err) {
		t.Fatalf("Builder allowlist remains after rollback: %v", err)
	}
}

func TestBuilderTaskRollbackPreservesChangedFiles(t *testing.T) {
	home := t.TempDir()
	taskFile := filepath.Join("..", "..", "examples", "task-envelope-builder.json")
	target, err := RenderApprovedBuilderTask(taskFile, "opencode", home, "", true, false)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(target, []byte("Human edit\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	_, skipped, err := RollbackBuilderTask(home, "TASK-001", false)
	if err != nil {
		t.Fatal(err)
	}
	if len(skipped) != 1 {
		t.Fatalf("expected changed Builder profile to be skipped, got %v", skipped)
	}
	content, err := os.ReadFile(target)
	if err != nil || string(content) != "Human edit\n" {
		t.Fatalf("rollback changed a Human edit: %q, %v", content, err)
	}
	_, skipped, err = RollbackBuilderTask(home, "TASK-001", true)
	if err != nil {
		t.Fatal(err)
	}
	if len(skipped) != 0 {
		t.Fatalf("forced rollback should finish: %v", skipped)
	}
}

func TestBuilderTaskForceOverwriteCanRestoreOriginal(t *testing.T) {
	home := t.TempDir()
	taskFile := filepath.Join("..", "..", "examples", "task-envelope-builder.json")
	target := filepath.Join(home, ".config", "opencode", "agents", "awm", "builder-task-TASK-001.md")
	if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
		t.Fatal(err)
	}
	original := []byte("Existing profile\n")
	if err := os.WriteFile(target, original, 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := RenderApprovedBuilderTask(taskFile, "opencode", home, "", true, false); err == nil {
		t.Fatal("existing profile should require --force")
	}
	if _, err := RenderApprovedBuilderTask(taskFile, "opencode", home, "", true, true); err != nil {
		t.Fatalf("approved force render failed: %v", err)
	}
	_, skipped, err := RollbackBuilderTask(home, "TASK-001", false)
	if err != nil || len(skipped) != 0 {
		t.Fatalf("rollback failed: skipped=%v err=%v", skipped, err)
	}
	content, err := os.ReadFile(target)
	if err != nil || !bytes.Equal(content, original) {
		t.Fatalf("original profile was not restored: %q err=%v", content, err)
	}
	info, err := os.Stat(target)
	if err != nil || info.Mode().Perm() != 0o600 {
		t.Fatalf("original mode was not restored: info=%v err=%v", info, err)
	}
}

package awm

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func testInstall(t *testing.T, home, harness string, prune bool) InstallResult {
	t.Helper()
	result, err := RunInstall(InstallOptions{
		HarnessIDs: []string{harness}, AllowUndetected: true, Yes: true,
		Home: home, PruneStale: prune,
	}, denyPrompter{})
	if err != nil {
		t.Fatalf("install %s: %v", harness, err)
	}
	return result
}

func TestUpdateKeepsUnchangedManagedFilesInState(t *testing.T) {
	home := t.TempDir()
	testInstall(t, home, "opencode", false)
	first, err := LoadInstallState(home)
	if err != nil {
		t.Fatal(err)
	}
	testInstall(t, home, "opencode", true)
	second, err := LoadInstallState(home)
	if err != nil {
		t.Fatal(err)
	}
	if len(second.ManagedFiles) < len(first.ManagedFiles) {
		t.Fatalf("update dropped managed files: %d -> %d", len(first.ManagedFiles), len(second.ManagedFiles))
	}
	target := filepath.Join(home, ".agentic-workflow", "core", "policies", "artifact-selection.md")
	found := false
	for _, item := range second.ManagedFiles {
		if item.Target == target {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("unchanged managed file missing from updated state: %s", target)
	}
}

func TestUpdatePrunesStaleManagedFile(t *testing.T) {
	home := t.TempDir()
	testInstall(t, home, "opencode", false)
	statePath := filepath.Join(home, ".agentic-workflow", "state", "install.json")
	state, err := LoadInstallState(home)
	if err != nil {
		t.Fatal(err)
	}
	stale := filepath.Join(home, "stale.txt")
	if err := os.WriteFile(stale, []byte("stale\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	state.ManagedFiles = append(state.ManagedFiles, ManagedFile{
		Target: stale, Existed: false, InstalledSHA256: sha256Hex([]byte("stale\n")),
	})
	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(statePath, append(data, '\n'), 0o600); err != nil {
		t.Fatal(err)
	}
	result := testInstall(t, home, "opencode", true)
	if _, err := os.Stat(stale); !os.IsNotExist(err) {
		t.Fatalf("stale file not pruned: %v", err)
	}
	if len(result.Removed) == 0 {
		t.Fatal("expected the update to report a removal")
	}
}

func TestUninstallRestoresGuidanceAndRemovesManagedFiles(t *testing.T) {
	home := t.TempDir()
	guidance := filepath.Join(home, ".claude", "CLAUDE.md")
	if err := os.MkdirAll(filepath.Dir(guidance), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(guidance, []byte("# User notes\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	testInstall(t, home, "claude-code", false)
	agent := filepath.Join(home, ".claude", "agents", "awm-architect.md")
	if _, err := os.Stat(agent); err != nil {
		t.Fatalf("installed agent missing: %v", err)
	}
	result, err := Uninstall(UninstallOptions{Home: home})
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Skipped) != 0 {
		t.Fatalf("unexpected skips: %v", result.Skipped)
	}
	if _, err := os.Stat(agent); !os.IsNotExist(err) {
		t.Fatalf("managed agent still exists: %v", err)
	}
	content, err := os.ReadFile(guidance)
	if err != nil {
		t.Fatal(err)
	}
	if string(content) != "# User notes\n" {
		t.Fatalf("guidance not restored: %q", content)
	}
	if _, err := os.Stat(filepath.Join(home, ".agentic-workflow")); !os.IsNotExist(err) {
		t.Fatalf("AWM home still exists: %v", err)
	}
}

func TestUninstallRemovesBuilderTaskProfiles(t *testing.T) {
	home := t.TempDir()
	testInstall(t, home, "opencode", false)
	target, err := RenderApprovedBuilderTask(filepath.Join("..", "..", "examples", "task-envelope-builder.json"), "opencode", home, "", true, false)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(target); err != nil {
		t.Fatalf("task profile missing: %v", err)
	}
	result, err := Uninstall(UninstallOptions{Home: home})
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Skipped) != 0 {
		t.Fatalf("unexpected skips: %v", result.Skipped)
	}
	if _, err := os.Stat(target); !os.IsNotExist(err) {
		t.Fatalf("task profile remains: %v", err)
	}
	if _, err := os.Stat(filepath.Join(home, ".agentic-workflow")); !os.IsNotExist(err) {
		t.Fatalf("AWM home still exists: %v", err)
	}
}

func TestUninstallSkipsChangedFileWithoutForce(t *testing.T) {
	home := t.TempDir()
	testInstall(t, home, "opencode", false)
	agent := filepath.Join(home, ".config", "opencode", "agents", "awm", "architect.md")
	content, err := os.ReadFile(agent)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(agent, append([]byte("user edit\n"), content...), 0o644); err != nil {
		t.Fatal(err)
	}
	result, err := Uninstall(UninstallOptions{Home: home})
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Skipped) == 0 {
		t.Fatal("expected a changed file to be skipped")
	}
	if _, err := os.Stat(agent); err != nil {
		t.Fatalf("changed file must remain: %v", err)
	}
	if _, err := os.Stat(filepath.Join(home, ".agentic-workflow")); err != nil {
		t.Fatalf("state must remain for retry: %v", err)
	}
	result, err = Uninstall(UninstallOptions{Home: home, Force: true})
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Skipped) != 0 {
		t.Fatalf("force uninstall skipped files: %v", result.Skipped)
	}
	if _, err := os.Stat(agent); !os.IsNotExist(err) {
		t.Fatalf("forced uninstall left file: %v", err)
	}
}

func TestUninstallRemovesEditedGuidanceBlock(t *testing.T) {
	home := t.TempDir()
	guidance := filepath.Join(home, ".claude", "CLAUDE.md")
	if err := os.MkdirAll(filepath.Dir(guidance), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(guidance, []byte("# User notes\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	testInstall(t, home, "claude-code", false)
	content, err := os.ReadFile(guidance)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(guidance, append(content, []byte("\nuser addition\n")...), 0o644); err != nil {
		t.Fatal(err)
	}
	result, err := Uninstall(UninstallOptions{Home: home})
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Skipped) != 0 {
		t.Fatalf("unexpected skips: %v", result.Skipped)
	}
	updated, err := os.ReadFile(guidance)
	if err != nil {
		t.Fatal(err)
	}
	text := string(updated)
	if strings.Contains(text, "AWM:AWM:BEGIN") || strings.Contains(text, "Managed by Agentic Workflow Manager") {
		t.Fatalf("AWM block remains: %s", text)
	}
	if !strings.Contains(text, "# User notes") || !strings.Contains(text, "user addition") {
		t.Fatalf("user content lost: %s", text)
	}
}

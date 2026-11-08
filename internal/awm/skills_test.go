package awm

import (
	"os"
	"path/filepath"
	"testing"
)

func TestScanSkillsReadsVersionDependenciesAndCompatibility(t *testing.T) {
	root := t.TempDir()
	skillDir := filepath.Join(root, "sample-skill")
	if err := os.MkdirAll(skillDir, 0o755); err != nil {
		t.Fatal(err)
	}
	manifest := `---
name: sample-skill
version: 1.2.3
dependencies: [tdd, code-review]
compatibility: opencode, claude-code
---
# Sample
`
	if err := os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte(manifest), 0o644); err != nil {
		t.Fatal(err)
	}
	items, err := ScanSkills([]string{root})
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 1 {
		t.Fatalf("expected one skill, got %d", len(items))
	}
	item := items[0]
	if item.Version != "1.2.3" {
		t.Fatalf("unexpected version: %q", item.Version)
	}
	if len(item.Dependencies) != 2 || item.Dependencies[0] != "tdd" {
		t.Fatalf("unexpected dependencies: %v", item.Dependencies)
	}
	if len(item.Compatibility) != 2 || item.Compatibility[1] != "claude-code" {
		t.Fatalf("unexpected compatibility: %v", item.Compatibility)
	}
}

func TestAuditSkillsReportsOutdatedBundleAndMissingDependencies(t *testing.T) {
	home := t.TempDir()
	skillDir := filepath.Join(home, ".agents", "skills", "awm-workflow-protocol")
	if err := os.MkdirAll(skillDir, 0o755); err != nil {
		t.Fatal(err)
	}
	manifest := "---\nname: awm-workflow-protocol\nversion: 0.1.0\n---\n# Old\n"
	if err := os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte(manifest), 0o644); err != nil {
		t.Fatal(err)
	}
	bundled, external, err := AuditSkills(nil, home)
	if err != nil {
		t.Fatal(err)
	}
	var protocol *SkillFinding
	for index := range bundled {
		if bundled[index].ID == "awm-workflow-protocol" {
			protocol = &bundled[index]
		}
	}
	if protocol == nil || !protocol.Installed || protocol.VersionStatus != "mismatch" || protocol.ExpectedVersion != "0.2.0" {
		t.Fatalf("outdated bundled skill was not identified: %+v", protocol)
	}
	var grillWithDocs *SkillFinding
	for index := range external {
		if external[index].ID == "grill-with-docs" {
			grillWithDocs = &external[index]
		}
	}
	if grillWithDocs == nil || len(grillWithDocs.MissingDependencies) != 2 {
		t.Fatalf("missing external dependencies were not reported: %+v", grillWithDocs)
	}
}

func TestSkillFrontmatterCompatibilityList(t *testing.T) {
	got := parseFrontmatterList("[opencode, claude-code, codex-cli]")
	if len(got) != 3 || got[0] != "opencode" || got[2] != "codex-cli" {
		t.Fatalf("unexpected compatibility list: %v", got)
	}
}

func TestCompareSkillCompatibilityMetadata(t *testing.T) {
	if got := compareCompatibility([]string{"opencode", "claude-code"}, []string{"opencode", "claude-code", "codex-cli"}); got != "match" {
		t.Fatalf("expected compatibility match, got %s", got)
	}
	if got := compareCompatibility([]string{"opencode", "claude-code"}, []string{"opencode"}); got != "mismatch" {
		t.Fatalf("expected compatibility mismatch, got %s", got)
	}
	if got := compareCompatibility([]string{"opencode"}, nil); got != "unknown" {
		t.Fatalf("expected unknown compatibility, got %s", got)
	}
}

func TestUnsupportedDetectedHarnesses(t *testing.T) {
	harnesses := []DetectedHarness{
		{Harness: Harness{ID: "opencode"}, Detected: true},
		{Harness: Harness{ID: "codex-cli"}, Detected: true},
		{Harness: Harness{ID: "omp"}, Detected: false},
	}
	got := unsupportedDetectedHarnesses(harnesses, []string{"opencode"})
	if len(got) != 1 || got[0] != "codex-cli" {
		t.Fatalf("unexpected unsupported harnesses: %v", got)
	}
}

func TestScanSkillsFollowsSkillDirectorySymlinks(t *testing.T) {
	root := t.TempDir()
	target := t.TempDir()
	manifest := "---\nname: understand\nversion: 1.0.0\n---\n# Understand\n"
	if err := os.WriteFile(filepath.Join(target, "SKILL.md"), []byte(manifest), 0o644); err != nil {
		t.Fatal(err)
	}
	link := filepath.Join(root, "understand")
	if err := os.Symlink(target, link); err != nil {
		t.Skipf("symlink creation unavailable: %v", err)
	}
	items, err := ScanSkills([]string{root})
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 1 || items[0].ID != "understand" {
		t.Fatalf("expected symlinked Understand skill to be found, got %+v", items)
	}
}

package awm

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func makeGitRepo(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	command := exec.Command("git", "init", root)
	if output, err := command.CombinedOutput(); err != nil {
		t.Fatalf("git init failed: %v: %s", err, output)
	}
	resolved, err := filepath.EvalSymlinks(root)
	if err != nil {
		t.Fatal(err)
	}
	return resolved
}

func TestInitRepositoryMemoryCreatesWorkspaceMarker(t *testing.T) {
	repo := makeGitRepo(t)
	marker, err := InitRepositoryMemory(repo, "product-iris-academy", false)
	if err != nil {
		t.Fatal(err)
	}
	if marker != filepath.Join(repo, ".ai-memory.toml") {
		t.Fatalf("unexpected marker path: %s", marker)
	}
	content, err := os.ReadFile(marker)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(content), `workspace = "product-iris-academy"`) {
		t.Fatalf("workspace ID missing from marker: %s", content)
	}
}

func TestInitRepositoryMemoryForceBacksUpExistingMarker(t *testing.T) {
	repo := makeGitRepo(t)
	marker := filepath.Join(repo, ".ai-memory.toml")
	original := []byte("workspace = \"old-workspace\"\n")
	if err := os.WriteFile(marker, original, 0o600); err != nil {
		t.Fatal(err)
	}
	created, err := InitRepositoryMemory(repo, "new-workspace", true)
	if err != nil {
		t.Fatal(err)
	}
	if created != marker {
		t.Fatalf("unexpected marker path: %s", created)
	}
	entries, err := filepath.Glob(marker + ".awm-backup-*")
	if err != nil || len(entries) != 1 {
		t.Fatalf("expected one backup, found %v (err=%v)", entries, err)
	}
	backup, err := os.ReadFile(entries[0])
	if err != nil {
		t.Fatal(err)
	}
	if string(backup) != string(original) {
		t.Fatal("backup does not match the original marker")
	}
	backupInfo, err := os.Stat(entries[0])
	if err != nil {
		t.Fatal(err)
	}
	if backupInfo.Mode().Perm() != 0o600 {
		t.Fatalf("backup mode must be 0600, got %o", backupInfo.Mode().Perm())
	}
	newInfo, err := os.Stat(marker)
	if err != nil {
		t.Fatal(err)
	}
	if newInfo.Mode().Perm() != 0o600 {
		t.Fatalf("existing marker mode must be preserved, got %o", newInfo.Mode().Perm())
	}
}

func TestInitRepositoryMemoryRejectsSymlink(t *testing.T) {
	repo := makeGitRepo(t)
	outside := filepath.Join(t.TempDir(), "outside.toml")
	if err := os.WriteFile(outside, []byte("secret"), 0o600); err != nil {
		t.Fatal(err)
	}
	marker := filepath.Join(repo, ".ai-memory.toml")
	if err := os.Symlink(outside, marker); err != nil {
		t.Skipf("symlink creation unavailable: %v", err)
	}
	if _, err := InitRepositoryMemory(repo, "workspace", true); err == nil {
		t.Fatal("expected symlink target to be rejected")
	}
	content, err := os.ReadFile(outside)
	if err != nil {
		t.Fatal(err)
	}
	if string(content) != "secret" {
		t.Fatal("outside symlink target was modified")
	}
}

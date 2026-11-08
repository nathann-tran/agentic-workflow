package awm

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

func RepositoryRoot(path string) (string, error) {
	if path == "" {
		path = "."
	}
	abs, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}
	cmd := exec.Command("git", "-C", abs, "rev-parse", "--show-toplevel")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("cannot locate Git repository at %s: %s", abs, strings.TrimSpace(string(output)))
	}
	return filepath.Clean(strings.TrimSpace(string(output))), nil
}

func InitRepositoryMemory(repoPath, workspace string, force bool) (string, error) {
	root, err := RepositoryRoot(repoPath)
	if err != nil {
		return "", err
	}
	workspace = strings.TrimSpace(workspace)
	if workspace == "" {
		return "", errors.New("workspace name is required")
	}
	for _, r := range workspace {
		if !(r == '-' || r == '_' || r == '.' || r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' || r >= '0' && r <= '9') {
			return "", errors.New("workspace name may contain letters, numbers, dot, underscore, and hyphen")
		}
	}
	target := filepath.Join(root, ".ai-memory.toml")
	mode := os.FileMode(0o644)
	if info, statErr := os.Lstat(target); statErr == nil {
		if info.Mode()&os.ModeSymlink != 0 {
			return "", fmt.Errorf("refusing to replace a symlink: %s", target)
		}
		if !force {
			return "", fmt.Errorf("marker already exists: %s; use --force after reviewing it", target)
		}
		previous, readErr := os.ReadFile(target)
		if readErr != nil {
			return "", readErr
		}
		mode = info.Mode().Perm()
		backup := target + ".awm-backup-" + time.Now().UTC().Format("20060102T150405.000000000Z")
		if err := atomicWrite(backup, previous, 0o600); err != nil {
			return "", fmt.Errorf("backup existing memory marker: %w", err)
		}
	} else if !os.IsNotExist(statErr) {
		return "", statErr
	}
	content := fmt.Sprintf("# Managed by Agentic Workflow Manager. Review before committing.\nworkspace = %q\nproject_strategy = \"repo-root\"\n", workspace)
	if err := atomicWrite(target, []byte(content), mode); err != nil {
		return "", err
	}
	return target, nil
}

func MemoryStatus(home string) []string {
	paths := []string{
		filepath.Join(home, ".config", "opencode", "plugins", "ai-memory-opencode2.ts"),
		filepath.Join(home, ".understand-anything-plugin"),
		filepath.Join(home, ".understand-anything", "repo", "understand-anything-plugin"),
	}
	var result []string
	for _, path := range paths {
		if info, err := os.Stat(path); err == nil {
			kind := "file"
			if info.IsDir() {
				kind = "directory"
			}
			result = append(result, fmt.Sprintf("FOUND %s (%s)", path, kind))
		} else {
			result = append(result, "MISSING "+path)
		}
	}
	return result
}

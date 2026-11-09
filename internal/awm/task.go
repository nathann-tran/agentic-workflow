package awm

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

type taskEnvelope struct {
	SchemaVersion string  `json:"schema_version"`
	WorkspaceID   string  `json:"workspace_id,omitempty"`
	ProjectID     string  `json:"project_id,omitempty"`
	RepositoryID  string  `json:"repository_id,omitempty"`
	WorkflowID    string  `json:"workflow_id"`
	TaskID        string  `json:"task_id"`
	ParentTaskID  *string `json:"parent_task_id,omitempty"`
	CallerRole    string  `json:"caller_role"`
	AssigneeRole  string  `json:"assignee_role"`
	Objective     string  `json:"objective"`
	TaskType      string  `json:"task_type,omitempty"`
	Scope         struct {
		Include        []string `json:"include"`
		Exclude        []string `json:"exclude"`
		Repositories   []string `json:"repositories,omitempty"`
		WriteAllowlist []string `json:"write_allowlist,omitempty"`
	} `json:"scope"`
	Constraints struct {
		ReadOnly                *bool `json:"read_only,omitempty"`
		SourceWriteAllowed      *bool `json:"source_write_allowed,omitempty"`
		NetworkAllowed          *bool `json:"network_allowed,omitempty"`
		HumanApprovalRequired   *bool `json:"human_approval_required,omitempty"`
		MaxDelegationDepth      *int  `json:"max_delegation_depth,omitempty"`
		ExternalMutationAllowed *bool `json:"external_mutation_allowed,omitempty"`
	} `json:"constraints"`
	Context struct {
		ArtifactRefs         []string `json:"artifact_refs"`
		ApprovedDecisionRefs []string `json:"approved_decision_refs"`
		MemoryQueries        []string `json:"memory_queries"`
		SourceRevision       *string  `json:"source_revision,omitempty"`
		RequiredSections     []string `json:"required_sections,omitempty"`
	} `json:"context"`
	Skills struct {
		Required    []string `json:"required"`
		Conditional []string `json:"conditional"`
		Explicit    []string `json:"explicit"`
	} `json:"skills"`
	ExpectedResult struct {
		Format          string   `json:"format"`
		Include         []string `json:"include"`
		MaxSummaryWords *int     `json:"max_summary_words,omitempty"`
	} `json:"expected_result"`
	DoneWhen         []string `json:"done_when"`
	HumanInstruction *string  `json:"human_instruction,omitempty"`
}

var safeTaskID = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._-]{0,79}$`)

func ReadApprovedBuilderTask(filename string, humanApproved bool) (taskEnvelope, []string, error) {
	var task taskEnvelope
	if !humanApproved {
		return task, nil, errors.New("Builder activation requires the explicit --human-approved flag")
	}
	data, err := os.ReadFile(filename)
	if err != nil {
		return task, nil, fmt.Errorf("read task envelope: %w", err)
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&task); err != nil {
		return task, nil, fmt.Errorf("parse task envelope: %w", err)
	}
	var trailing any
	if err := decoder.Decode(&trailing); err == nil {
		return task, nil, errors.New("task envelope has trailing JSON content")
	} else if !errors.Is(err, io.EOF) {
		return task, nil, fmt.Errorf("read task envelope trailer: %w", err)
	}
	if task.SchemaVersion != "1.0" {
		return task, nil, errors.New("task envelope schema_version must be 1.0")
	}
	if strings.TrimSpace(task.WorkflowID) == "" || !safeTaskID.MatchString(task.TaskID) || task.TaskID == "." || task.TaskID == ".." {
		return task, nil, errors.New("workflow_id and a safe task_id are required")
	}
	if !contains([]string{"human", "conductor", "plan", "scout", "architect", "builder", "worker"}, task.CallerRole) {
		return task, nil, fmt.Errorf("invalid caller_role %q", task.CallerRole)
	}
	if task.AssigneeRole != "builder" {
		return task, nil, fmt.Errorf("task assignee_role must be builder, got %q", task.AssigneeRole)
	}
	if strings.TrimSpace(task.Objective) == "" || task.Scope.Include == nil || task.Scope.Exclude == nil || task.Context.ArtifactRefs == nil || task.Context.ApprovedDecisionRefs == nil || task.Context.MemoryQueries == nil || task.Skills.Required == nil || task.Skills.Conditional == nil || task.Skills.Explicit == nil || strings.TrimSpace(task.ExpectedResult.Format) == "" || task.ExpectedResult.Include == nil || len(task.DoneWhen) == 0 {
		return task, nil, errors.New("task envelope is missing required schema fields")
	}
	if task.Constraints.SourceWriteAllowed == nil || !*task.Constraints.SourceWriteAllowed {
		return task, nil, errors.New("task constraints must set source_write_allowed=true")
	}
	if task.Constraints.HumanApprovalRequired == nil || !*task.Constraints.HumanApprovalRequired {
		return task, nil, errors.New("Builder task constraints must set human_approval_required=true")
	}
	if task.Constraints.ExternalMutationAllowed == nil || *task.Constraints.ExternalMutationAllowed {
		return task, nil, errors.New("Builder task constraints must set external_mutation_allowed=false")
	}
	if task.Constraints.MaxDelegationDepth != nil && *task.Constraints.MaxDelegationDepth > 1 {
		return task, nil, errors.New("Builder task delegation depth cannot exceed one Worker level")
	}
	if task.Constraints.ReadOnly != nil && *task.Constraints.ReadOnly {
		return task, nil, errors.New("task constraints conflict: read_only=true")
	}
	if len(task.Scope.WriteAllowlist) == 0 {
		return task, nil, errors.New("Builder task must include a non-empty scope.write_allowlist")
	}
	allowlist, err := ValidateWriteAllowlist(task.Scope.WriteAllowlist)
	if err != nil {
		return task, nil, err
	}
	return task, allowlist, nil
}

func ValidateWriteAllowlist(values []string) ([]string, error) {
	seen := map[string]bool{}
	var paths []string
	for _, raw := range values {
		value := strings.TrimSpace(strings.ReplaceAll(raw, "\\", "/"))
		if value == "" || strings.ContainsAny(value, "\x00\r\n:") || strings.HasPrefix(value, "/") || path.IsAbs(value) || filepath.VolumeName(value) != "" {
			return nil, fmt.Errorf("write allowlist path must be repository-relative: %q", raw)
		}
		for _, segment := range strings.Split(value, "/") {
			if segment == ".." || segment == "." || segment == "" {
				return nil, fmt.Errorf("write allowlist path contains an unsafe segment: %q", raw)
			}
		}
		if strings.ContainsAny(value, "*?[]") || strings.HasSuffix(value, "/") {
			return nil, fmt.Errorf("write allowlist requires exact file paths: %q", raw)
		}
		clean := path.Clean(value)
		if clean == "." || clean == ".." || strings.HasPrefix(clean, "../") {
			return nil, fmt.Errorf("write allowlist path is unsafe: %q", raw)
		}
		if !seen[clean] {
			seen[clean] = true
			paths = append(paths, clean)
		}
	}
	if len(paths) == 0 {
		return nil, errors.New("write allowlist is empty")
	}
	return paths, nil
}

type builderTaskFileState struct {
	Target          string `json:"target"`
	Backup          string `json:"backup,omitempty"`
	Existed         bool   `json:"existed"`
	OriginalMode    uint32 `json:"original_mode,omitempty"`
	InstalledSHA256 string `json:"installed_sha256"`
}

type builderTaskState struct {
	SchemaVersion string                 `json:"schema_version"`
	TaskID        string                 `json:"task_id"`
	HarnessID     string                 `json:"harness_id"`
	Files         []builderTaskFileState `json:"files"`
}

func RenderApprovedBuilderTask(taskFile, harnessID, home, output string, humanApproved, force bool) (string, error) {
	task, allowlist, err := ReadApprovedBuilderTask(taskFile, humanApproved)
	if err != nil {
		return "", err
	}
	if home == "" {
		home = HomeDir()
	}
	if !safeTaskID.MatchString(task.TaskID) || task.TaskID == "." || task.TaskID == ".." {
		return "", errors.New("task_id is not safe for a managed task directory")
	}
	allowlistText := make([]string, 0, len(allowlist))
	for _, item := range allowlist {
		allowlistText = append(allowlistText, "- "+item)
	}
	content, err := RenderTaskBuilder(task.TaskID, harnessID, "builder", strings.Join(allowlistText, "\n"), home)
	if err != nil {
		return "", err
	}
	target := output
	if target == "" {
		switch harnessID {
		case "opencode":
			target = filepath.Join(home, ".config", "opencode", "agents", "awm", "builder-task-"+task.TaskID+".md")
		case "claude-code":
			target = filepath.Join(home, ".claude", "agents", "awm-builder-task-"+task.TaskID+".md")
		default:
			return "", fmt.Errorf("task-scoped Builder is not supported for %s; use a separate Human-approved session and isolated worktree", harnessID)
		}
	}
	if !securePathWithin(home, target) {
		return "", fmt.Errorf("refusing Builder profile output outside the home directory: %s", target)
	}
	allowlistPath := filepath.Join(home, ".agentic-workflow", "tasks", task.TaskID, "write-allowlist.json")
	statePath := filepath.Join(home, ".agentic-workflow", "tasks", task.TaskID, "state.json")
	for _, path := range []string{allowlistPath, target, statePath} {
		if !securePathWithin(home, path) {
			return "", fmt.Errorf("Builder task path escapes the home directory: %s", path)
		}
	}
	if info, statErr := os.Lstat(statePath); statErr == nil {
		if info.Mode()&os.ModeSymlink != 0 {
			return "", errors.New("refusing to replace a symlinked Builder task state")
		}
		return "", fmt.Errorf("Builder task %s already has state; run `awm builder rollback --task-id %s` before re-rendering", task.TaskID, task.TaskID)
	} else if !os.IsNotExist(statErr) {
		return "", statErr
	}
	allowlistData, err := json.MarshalIndent(struct {
		TaskID         string   `json:"task_id"`
		WriteAllowlist []string `json:"write_allowlist"`
	}{task.TaskID, allowlist}, "", "  ")
	if err != nil {
		return "", err
	}
	artifacts := []struct {
		path string
		data []byte
	}{
		{path: allowlistPath, data: append(allowlistData, '\n')},
		{path: target, data: content},
	}
	state := builderTaskState{SchemaVersion: "1.0", TaskID: task.TaskID, HarnessID: harnessID}
	for _, artifact := range artifacts {
		fileState, changed, writeErr := writeBuilderTaskArtifact(home, task.TaskID, artifact.path, artifact.data, force)
		if writeErr != nil {
			rollbackBuilderTaskFiles(home, state.Files, true)
			return "", writeErr
		}
		if changed {
			state.Files = append(state.Files, fileState)
		}
	}
	if len(state.Files) == 0 {
		return target, nil
	}
	manifest, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		rollbackBuilderTaskFiles(home, state.Files, true)
		return "", err
	}
	if err := atomicWrite(statePath, append(manifest, '\n'), 0o600); err != nil {
		rollbackBuilderTaskFiles(home, state.Files, true)
		return "", fmt.Errorf("write Builder task state: %w", err)
	}
	return target, nil
}

func writeBuilderTaskArtifact(home, taskID, target string, content []byte, force bool) (builderTaskFileState, bool, error) {
	state := builderTaskFileState{Target: target, InstalledSHA256: sha256Hex(content)}
	if !securePathWithin(home, target) {
		return state, false, fmt.Errorf("unsafe Builder task target: %s", target)
	}
	if info, err := os.Lstat(target); err == nil {
		if info.Mode()&os.ModeSymlink != 0 {
			return state, false, fmt.Errorf("refusing to replace symlink: %s", target)
		}
		old, readErr := os.ReadFile(target)
		if readErr != nil {
			return state, false, readErr
		}
		if bytes.Equal(old, content) {
			return state, false, nil
		}
		if !force {
			return state, false, fmt.Errorf("Builder target already exists with different content: %s; use --force after review", target)
		}
		state.Existed = true
		state.OriginalMode = uint32(info.Mode().Perm())
		rel, relErr := filepath.Rel(home, target)
		if relErr != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
			return state, false, errors.New("Builder backup target is outside home directory")
		}
		stamp := time.Now().UTC().Format("20060102T150405.000000000Z")
		state.Backup = filepath.Join(home, ".agentic-workflow", "backups", "builder-task", taskID, stamp, rel+".bak")
		if !securePathWithin(home, state.Backup) {
			return state, false, errors.New("Builder backup path is unsafe")
		}
		if err := atomicWrite(state.Backup, old, 0o600); err != nil {
			return state, false, fmt.Errorf("backup existing Builder target: %w", err)
		}
	} else if !os.IsNotExist(err) {
		return state, false, err
	}
	if err := atomicWrite(target, content, 0o644); err != nil {
		if state.Backup != "" {
			_ = restoreBuilderTaskFile(state)
		}
		return state, false, err
	}
	return state, true, nil
}

func RollbackBuilderTask(home, taskID string, force bool) ([]string, []string, error) {
	if home == "" {
		home = HomeDir()
	}
	if !safeTaskID.MatchString(taskID) || taskID == "." || taskID == ".." {
		return nil, nil, errors.New("invalid Builder task ID")
	}
	statePath := filepath.Join(home, ".agentic-workflow", "tasks", taskID, "state.json")
	if !securePathWithin(home, statePath) {
		return nil, nil, errors.New("Builder task state path escapes home directory")
	}
	info, err := os.Lstat(statePath)
	if err != nil {
		return nil, nil, fmt.Errorf("read Builder task state: %w", err)
	}
	if info.Mode()&os.ModeSymlink != 0 {
		return nil, nil, errors.New("refusing to read symlinked Builder task state")
	}
	data, err := os.ReadFile(statePath)
	if err != nil {
		return nil, nil, err
	}
	var state builderTaskState
	if err := json.Unmarshal(data, &state); err != nil {
		return nil, nil, fmt.Errorf("parse Builder task state: %w", err)
	}
	if state.SchemaVersion != "1.0" || state.TaskID != taskID {
		return nil, nil, errors.New("Builder task state identity or schema is invalid")
	}
	restored, skipped := rollbackBuilderTaskFiles(home, state.Files, force)
	if len(skipped) == 0 {
		if err := os.Remove(statePath); err != nil && !os.IsNotExist(err) {
			skipped = append(skipped, statePath+" (state remove failed)")
		}
	}
	return restored, skipped, nil
}

func rollbackBuilderTaskFiles(home string, files []builderTaskFileState, force bool) ([]string, []string) {
	var restored, skipped []string
	for index := len(files) - 1; index >= 0; index-- {
		item := files[index]
		if !securePathWithin(home, item.Target) {
			skipped = append(skipped, item.Target+" (outside home; blocked)")
			continue
		}
		current, err := os.ReadFile(item.Target)
		if err == nil && !force && sha256Hex(current) != item.InstalledSHA256 {
			skipped = append(skipped, item.Target+" (changed after task profile creation)")
			continue
		}
		if err != nil && !os.IsNotExist(err) {
			skipped = append(skipped, item.Target+" (cannot read)")
			continue
		}
		if item.Existed {
			if !securePathWithin(home, item.Backup) {
				skipped = append(skipped, item.Target+" (backup path blocked)")
				continue
			}
			if err := restoreBuilderTaskFile(item); err != nil {
				skipped = append(skipped, item.Target+" (restore failed: "+err.Error()+")")
				continue
			}
		} else if err := os.Remove(item.Target); err != nil && !os.IsNotExist(err) {
			skipped = append(skipped, item.Target+" (remove failed: "+err.Error()+")")
			continue
		}
		restored = append(restored, item.Target)
		removeEmptyParents(filepath.Dir(item.Target), home)
	}
	return restored, skipped
}

func restoreBuilderTaskFile(item builderTaskFileState) error {
	backup, err := os.ReadFile(item.Backup)
	if err != nil {
		return err
	}
	mode := os.FileMode(item.OriginalMode)
	if mode == 0 {
		mode = 0o644
	}
	return atomicWrite(item.Target, backup, mode)
}

package awm

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/agentic-workflow/awm/internal/assets"
)

type InstallPrompter interface {
	Confirm(prompt string, defaultYes bool) bool
}

type SkillSelectionPrompter interface {
	SelectSkills(candidates []BundledSkill) ([]string, error)
}

type InstallResult struct {
	State   InstallState
	Written []string
	Removed []string
	Skipped []string
}

type installer struct {
	home          string
	awmHome       string
	backupRoot    string
	force         bool
	dryRun        bool
	prompt        InstallPrompter
	state         InstallState
	previous      InstallState
	hasPrevious   bool
	previousFiles map[string]ManagedFile
	produced      map[string]bool
	skipSkills    bool
	written       []string
	removed       []string
	skipped       []string
}

func RunInstall(options InstallOptions, prompt InstallPrompter) (InstallResult, error) {
	home := options.Home
	if home == "" {
		home = HomeDir()
	}
	awmHome := filepath.Join(home, ".agentic-workflow")
	stamp := time.Now().UTC().Format("20060102T150405.000000000Z")
	backupRoot := filepath.Join(awmHome, "backups", stamp)
	selected, err := selectByIDs(options.HarnessIDs, home, options.AllowUndetected)
	if err != nil {
		return InstallResult{}, err
	}
	previousManaged := map[string]ManagedFile{}
	var previousState InstallState
	hasPrevious := false
	if previous, stateErr := LoadInstallState(home); stateErr == nil {
		previousState = previous
		hasPrevious = true
		for _, item := range previous.ManagedFiles {
			previousManaged[item.Target] = item
		}
	} else if !os.IsNotExist(stateErr) {
		return InstallResult{}, fmt.Errorf("read previous AWM install state: %w", stateErr)
	}
	inst := &installer{
		home: home, awmHome: awmHome, backupRoot: backupRoot,
		force: options.Force, dryRun: options.DryRun, prompt: prompt,
		previous: previousState, hasPrevious: hasPrevious, previousFiles: previousManaged,
		produced: map[string]bool{}, skipSkills: options.SkipSkills,
		state: InstallState{
			Version: Version, InstalledAt: time.Now().UTC(), BackupRoot: backupRoot,
			InstallSource: "embedded-release",
		},
	}
	for _, h := range selected {
		inst.state.Selected = append(inst.state.Selected, h.ID)
	}
	if err := inst.installCore(); err != nil {
		inst.rollbackWrittenFiles(false)
		return InstallResult{}, err
	}
	for _, harness := range selected {
		if err := inst.installHarness(harness); err != nil {
			inst.rollbackWrittenFiles(false)
			return InstallResult{}, err
		}
	}
	if !options.SkipSkills {
		if err := inst.installSelectedBundledSkills(selected, options.SkillIDs, options.Yes); err != nil {
			inst.rollbackWrittenFiles(false)
			return InstallResult{}, err
		}
	}
	if inst.hasPrevious {
		inst.carryForwardUnchanged()
	}
	if options.PruneStale && inst.hasPrevious {
		inst.pruneStale()
	}
	if options.DryRun {
		return InstallResult{State: inst.state, Written: inst.written, Removed: inst.removed, Skipped: inst.skipped}, nil
	}
	inst.state.SkippedFiles = append(inst.state.SkippedFiles, inst.skipped...)
	if err := inst.saveState(); err != nil {
		inst.rollbackWrittenFiles(false)
		return InstallResult{}, err
	}
	return InstallResult{State: inst.state, Written: inst.written, Removed: inst.removed, Skipped: inst.skipped}, nil
}

func selectByIDs(ids []string, home string, allowUndetected bool) ([]DetectedHarness, error) {
	detected, err := DetectHarnesses(home)
	if err != nil {
		return nil, err
	}
	return SelectedHarnesses(detected, ids, allowUndetected)
}

func (i *installer) installCore() error {
	for _, root := range []string{"core", "schemas", "examples", "registries", "skills", "docs", "adapters", "templates"} {
		if err := i.copyEmbeddedTree(filepath.Join("data", root), filepath.Join(i.awmHome, root)); err != nil {
			return err
		}
	}
	return i.installHooks()
}

func (i *installer) installHooks() error {
	for _, name := range []string{"write_scope_guard.py", "spawn_guard.py", "read_scope_guard.py"} {
		data, err := fs.ReadFile(assets.Files, filepath.ToSlash(filepath.Join("data", "hooks", "claude", name)))
		if err != nil {
			return fmt.Errorf("read hook %s: %w", name, err)
		}
		if err := i.writeManagedFile(filepath.Join(i.awmHome, "hooks", "claude", name), data); err != nil {
			return err
		}
	}
	return nil
}

func (i *installer) installHarness(h DetectedHarness) error {
	agentRoot := ExpandHome(h.GlobalAgents, i.home)
	for _, role := range roleIDs {
		content, err := RenderAgent(h.ID, role, i.home)
		if err != nil {
			return err
		}
		var target string
		switch h.ID {
		case "opencode":
			target = filepath.Join(agentRoot, role+".md")
		case "claude-code":
			target = filepath.Join(agentRoot, "awm-"+role+".md")
		case "codex-cli":
			target = filepath.Join(agentRoot, "awm_"+role+".toml")
		case "omp":
			target = filepath.Join(agentRoot, "awm-"+role+".md")
		default:
			return fmt.Errorf("unsupported harness %q", h.ID)
		}
		if err := i.writeManagedFile(target, content); err != nil {
			return err
		}
	}
	if h.GlobalGuidance != "" {
		guidance, err := RenderGuidance(h.ID, i.home)
		if err != nil {
			return err
		}
		target := ExpandHome(h.GlobalGuidance, i.home)
		if err := i.upsertManagedBlock(target, "AWM", guidance); err != nil {
			return err
		}
	}
	return nil
}

func (i *installer) installSelectedBundledSkills(selected []DetectedHarness, requested []string, yes bool) error {
	catalog, err := LoadEmbeddedJSON[SkillCatalog]("registries/skills.json")
	if err != nil {
		return err
	}
	var roots []string
	for _, h := range selected {
		all := SkillRoots(h, i.home)
		if len(all) > 0 {
			roots = append(roots, all[0])
		}
	}
	roots = uniqueStrings(roots)
	installed, err := ScanSkills(roots)
	if err != nil {
		return err
	}
	installedByID := map[string]InstalledSkill{}
	for _, skill := range installed {
		installedByID[strings.ToLower(skill.ID)] = skill
		installedByID[strings.ToLower(filepath.Base(skill.Path))] = skill
	}
	var candidates []BundledSkill
	candidateReasons := map[string]string{}
	for _, item := range catalog.Bundled {
		current, ok := installedByID[strings.ToLower(item.ID)]
		if !ok {
			candidates = append(candidates, item)
			candidateReasons[strings.ToLower(item.ID)] = "missing"
			continue
		}
		if item.Version != "" && current.Version != item.Version {
			candidates = append(candidates, item)
			candidateReasons[strings.ToLower(item.ID)] = fmt.Sprintf("version %q found; %q required", current.Version, item.Version)
			continue
		}
		i.markProducedSkillFiles(item, roots)
	}
	if len(candidates) == 0 {
		return nil
	}
	fmt.Println("\nBundled AWM skills to install or update:")
	for n, item := range candidates {
		fmt.Printf("  %d. %s — %s (%s)\n", n+1, item.ID, item.Description, candidateReasons[strings.ToLower(item.ID)])
	}
	selectedIDs := map[string]bool{}
	availableIDs := map[string]bool{}
	for _, item := range candidates {
		availableIDs[strings.ToLower(item.ID)] = true
	}
	if len(requested) > 0 {
		for _, id := range requested {
			key := strings.ToLower(strings.TrimSpace(id))
			if !availableIDs[key] {
				return fmt.Errorf("skill %q is not missing or outdated in the selected harnesses", id)
			}
			selectedIDs[key] = true
		}
	} else if yes {
		for _, item := range candidates {
			selectedIDs[strings.ToLower(item.ID)] = true
		}
	} else if selector, ok := i.prompt.(SkillSelectionPrompter); ok {
		ids, err := selector.SelectSkills(candidates)
		if err != nil {
			return err
		}
		for _, id := range ids {
			key := strings.ToLower(strings.TrimSpace(id))
			if !availableIDs[key] {
				return fmt.Errorf("unknown or unavailable bundled skill: %s", id)
			}
			selectedIDs[key] = true
		}
	} else if i.prompt != nil && i.prompt.Confirm("Install or update all missing AWM skills?", true) {
		for _, item := range candidates {
			selectedIDs[strings.ToLower(item.ID)] = true
		}
	}
	for _, item := range candidates {
		key := strings.ToLower(item.ID)
		if !selectedIDs[key] {
			i.skipped = append(i.skipped, "skill:"+item.ID)
			continue
		}
		compatible := map[string]bool{}
		for _, harnessID := range item.Compatibility {
			compatible[strings.ToLower(harnessID)] = true
		}
		for _, harness := range selected {
			if !compatible[harness.ID] {
				return fmt.Errorf("bundled skill %s does not declare compatibility with selected harness %s", item.ID, harness.ID)
			}
		}
		for _, root := range roots {
			from := filepath.ToSlash(filepath.Join("data", "skills", item.ID))
			to := filepath.Join(root, item.ID)
			if err := i.copyEmbeddedTree(from, to); err != nil {
				return err
			}
		}
	}
	return nil
}

func (i *installer) markProducedSkillFiles(item BundledSkill, roots []string) {
	source := filepath.ToSlash(filepath.Join("data", "skills", item.ID))
	for _, root := range roots {
		_ = fs.WalkDir(assets.Files, source, func(path string, entry fs.DirEntry, walkErr error) error {
			if walkErr != nil || entry.IsDir() {
				return walkErr
			}
			rel := strings.TrimPrefix(strings.TrimPrefix(path, source), "/")
			i.produced[filepath.Join(root, item.ID, filepath.FromSlash(rel))] = true
			return nil
		})
	}
}

func (i *installer) copyEmbeddedTree(source, target string) error {
	return fs.WalkDir(assets.Files, source, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		rel := strings.TrimPrefix(path, source)
		rel = strings.TrimPrefix(rel, "/")
		destination := target
		if rel != "" {
			destination = filepath.Join(target, filepath.FromSlash(rel))
		}
		if !securePathWithin(i.home, destination) {
			return fmt.Errorf("refusing to create a directory outside the home directory: %s", destination)
		}
		if entry.IsDir() {
			if i.dryRun {
				return nil
			}
			return os.MkdirAll(destination, 0o755)
		}
		data, err := fs.ReadFile(assets.Files, path)
		if err != nil {
			return err
		}
		return i.writeManagedFile(destination, data)
	})
}

func (i *installer) writeManagedFile(target string, content []byte) error {
	target, err := filepath.Abs(target)
	if err != nil {
		return err
	}
	if !securePathWithin(i.home, target) {
		return fmt.Errorf("refusing to write outside the home directory: %s", target)
	}
	var originalMode os.FileMode
	if info, statErr := os.Lstat(target); statErr == nil {
		if info.Mode()&os.ModeSymlink != 0 {
			return fmt.Errorf("refusing to replace symlink: %s", target)
		}
		originalMode = info.Mode().Perm()
		old, readErr := os.ReadFile(target)
		if readErr != nil {
			return readErr
		}
		if string(old) == string(content) {
			i.produced[target] = true
			return nil
		}
		managed := strings.Contains(string(old), "Managed by Agentic Workflow Manager") || strings.Contains(string(old), "AWM:BEGIN -->")
		if previous, ok := i.previousFiles[target]; ok && previous.InstalledSHA256 == sha256Hex(old) {
			managed = true
		}
		if !managed && !i.force {
			allow := i.prompt != nil && i.prompt.Confirm("Replace existing unmanaged file "+target+"?", false)
			if !allow {
				i.skipped = append(i.skipped, target)
				return nil
			}
		}
	} else if !os.IsNotExist(statErr) {
		return statErr
	}
	if i.dryRun {
		i.produced[target] = true
		i.written = append(i.written, target)
		return nil
	}
	if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
		return err
	}
	backup, existed, err := i.backupExisting(target)
	if err != nil {
		return err
	}
	if err := atomicWrite(target, content, 0o644); err != nil {
		return err
	}
	i.record(target, backup, existed, originalMode, content)
	return nil
}

func (i *installer) upsertManagedBlock(target, name string, block []byte) error {
	absolute, err := filepath.Abs(target)
	if err != nil {
		return err
	}
	target = absolute
	if !securePathWithin(i.home, target) {
		return fmt.Errorf("refusing to read or update a path outside the home directory: %s", target)
	}
	begin := "<!-- AWM:" + name + ":BEGIN -->"
	end := "<!-- AWM:" + name + ":END -->"
	content := strings.TrimSpace(string(block))
	merged := begin + "\n" + content + "\n" + end
	old, err := os.ReadFile(target)
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	if err == nil {
		text := string(old)
		start := strings.Index(text, begin)
		finish := strings.Index(text, end)
		if start >= 0 && finish >= start {
			finish += len(end)
			text = strings.TrimSpace(text[:start] + text[finish:])
		}
		if text != "" {
			merged = text + "\n\n" + merged
		}
	}
	if err := i.writeManagedFile(target, []byte(merged+"\n")); err != nil {
		return err
	}
	for index := len(i.state.ManagedFiles) - 1; index >= 0; index-- {
		if i.state.ManagedFiles[index].Target == target {
			i.state.ManagedFiles[index].Block = name
			break
		}
	}
	return nil
}

func (i *installer) backupExisting(target string) (string, bool, error) {
	info, err := os.Lstat(target)
	if os.IsNotExist(err) {
		return "", false, nil
	}
	if err != nil {
		return "", false, err
	}
	if info.Mode()&os.ModeSymlink != 0 {
		return "", false, errors.New("refusing to back up symlink: " + target)
	}
	data, err := os.ReadFile(target)
	if err != nil {
		return "", false, err
	}
	rel, err := filepath.Rel(i.home, target)
	if err != nil || strings.HasPrefix(rel, ".."+string(filepath.Separator)) || rel == ".." {
		return "", false, errors.New("backup target is outside home directory")
	}
	backup := filepath.Join(i.backupRoot, "files", rel)
	if !securePathWithin(i.home, backup) {
		return "", false, errors.New("backup path is outside the home directory")
	}
	if i.dryRun {
		return backup, true, nil
	}
	if err := os.MkdirAll(filepath.Dir(backup), 0o755); err != nil {
		return "", false, err
	}
	if err := atomicWrite(backup, data, 0o600); err != nil {
		return "", false, err
	}
	return backup, true, nil
}

func (i *installer) record(target, backup string, existed bool, originalMode os.FileMode, content []byte) {
	sum := sha256.Sum256(content)
	i.state.ManagedFiles = append(i.state.ManagedFiles, ManagedFile{
		Target: target, Backup: backup, Existed: existed, OriginalMode: uint32(originalMode.Perm()),
		InstalledSHA256: hex.EncodeToString(sum[:]),
	})
	i.produced[target] = true
	i.written = append(i.written, target)
}

func (i *installer) saveState() error {
	statePath := filepath.Join(i.awmHome, "state", "install.json")
	if !securePathWithin(i.home, statePath) {
		return errors.New("install state path escapes the home directory")
	}
	if previous, err := os.ReadFile(statePath); err == nil {
		backup := filepath.Join(i.backupRoot, "state", "install.json")
		if !securePathWithin(i.home, backup) {
			return errors.New("previous state backup path escapes the home directory")
		}
		if err := atomicWrite(backup, previous, 0o600); err != nil {
			return err
		}
		i.state.PreviousStateBackup = backup
	} else if !os.IsNotExist(err) {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(statePath), 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(i.state, "", "  ")
	if err != nil {
		return err
	}
	return atomicWrite(statePath, append(data, '\n'), 0o600)
}

func Rollback(home string, force bool) ([]string, []string, error) {
	if home == "" {
		home = HomeDir()
	}
	path := filepath.Join(home, ".agentic-workflow", "state", "install.json")
	if !securePathWithin(home, path) {
		return nil, nil, fmt.Errorf("install state path is unsafe: %s", path)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, nil, fmt.Errorf("read install state: %w", err)
	}
	var state InstallState
	if err := json.Unmarshal(data, &state); err != nil {
		return nil, nil, fmt.Errorf("parse install state: %w", err)
	}
	var restored, skipped []string
	for index := len(state.ManagedFiles) - 1; index >= 0; index-- {
		item := state.ManagedFiles[index]
		if !securePathWithin(home, item.Target) {
			skipped = append(skipped, item.Target+" (outside home; blocked)")
			continue
		}
		current, readErr := os.ReadFile(item.Target)
		if readErr == nil && !force {
			sum := sha256.Sum256(current)
			if hex.EncodeToString(sum[:]) != item.InstalledSHA256 {
				skipped = append(skipped, item.Target+" (changed after AWM install)")
				continue
			}
		}
		if readErr != nil && !os.IsNotExist(readErr) {
			skipped = append(skipped, item.Target+" (cannot read)")
			continue
		}
		if item.Existed {
			if !securePathWithin(home, item.Backup) {
				skipped = append(skipped, item.Target+" (backup path outside home; blocked)")
				continue
			}
			backup, backupErr := os.ReadFile(item.Backup)
			if backupErr != nil {
				skipped = append(skipped, item.Target+" (backup missing)")
				continue
			}
			mode := os.FileMode(item.OriginalMode)
			if mode == 0 {
				mode = 0o644
			}
			if err := atomicWrite(item.Target, backup, mode); err != nil {
				skipped = append(skipped, item.Target+" (restore failed: "+err.Error()+")")
				continue
			}
		} else {
			if err := os.Remove(item.Target); err != nil && !os.IsNotExist(err) {
				skipped = append(skipped, item.Target+" (remove failed: "+err.Error()+")")
				continue
			}
		}
		restored = append(restored, item.Target)
	}
	if len(skipped) > 0 {
		b, marshalErr := json.MarshalIndent(state, "", "  ")
		if marshalErr == nil {
			_ = atomicWrite(path, append(b, '\n'), 0o600)
		}
		return restored, skipped, nil
	}
	if state.PreviousStateBackup != "" {
		if securePathWithin(home, state.PreviousStateBackup) {
			previous, readErr := os.ReadFile(state.PreviousStateBackup)
			if readErr == nil {
				if writeErr := atomicWrite(path, previous, 0o600); writeErr == nil {
					restored = append(restored, path+" (previous install state)")
				} else {
					skipped = append(skipped, path+" (previous state restore failed)")
				}
			} else {
				skipped = append(skipped, path+" (previous install state backup missing)")
			}
		} else {
			skipped = append(skipped, path+" (previous state backup path blocked)")
		}
	} else {
		if err := os.Remove(path); err == nil {
			restored = append(restored, path+" (removed)")
		} else if !os.IsNotExist(err) {
			skipped = append(skipped, path+" (state remove failed)")
		}
	}
	return restored, skipped, nil
}

func (i *installer) carryForwardUnchanged() {
	current := map[string]bool{}
	for _, item := range i.state.ManagedFiles {
		current[item.Target] = true
	}
	for _, item := range i.previous.ManagedFiles {
		if current[item.Target] {
			continue
		}
		if !i.skipSkills && !i.produced[item.Target] {
			continue
		}
		content, err := os.ReadFile(item.Target)
		if err != nil {
			continue
		}
		if sha256Hex(content) != item.InstalledSHA256 {
			continue
		}
		i.state.ManagedFiles = append(i.state.ManagedFiles, item)
		current[item.Target] = true
	}
}

func (i *installer) pruneStale() {
	current := map[string]bool{}
	for _, item := range i.state.ManagedFiles {
		current[item.Target] = true
	}
	for index := len(i.previous.ManagedFiles) - 1; index >= 0; index-- {
		item := i.previous.ManagedFiles[index]
		if current[item.Target] {
			continue
		}
		i.removeManaged(item)
	}
}

func (i *installer) removeManaged(item ManagedFile) {
	if !securePathWithin(i.home, item.Target) {
		i.skipped = append(i.skipped, item.Target+" (outside home; blocked)")
		return
	}
	current, err := os.ReadFile(item.Target)
	if err != nil {
		if os.IsNotExist(err) {
			return
		}
		i.skipped = append(i.skipped, item.Target+" (cannot read)")
		return
	}
	if item.Block != "" {
		updated, found := stripManagedBlock(current, item.Block)
		if !found {
			return
		}
		if i.dryRun {
			i.removed = append(i.removed, item.Target)
			return
		}
		if len(bytes.TrimSpace(updated)) == 0 {
			err = os.Remove(item.Target)
		} else {
			mode := os.FileMode(item.OriginalMode)
			if mode == 0 {
				mode = 0o644
			}
			err = atomicWrite(item.Target, updated, mode)
		}
		if err != nil {
			i.skipped = append(i.skipped, item.Target+" (remove failed: "+err.Error()+")")
			return
		}
		i.removed = append(i.removed, item.Target)
		removeEmptyParents(filepath.Dir(item.Target), i.home)
		return
	}
	if !i.force && sha256Hex(current) != item.InstalledSHA256 {
		i.skipped = append(i.skipped, item.Target+" (changed after AWM install)")
		return
	}
	if i.dryRun {
		i.removed = append(i.removed, item.Target)
		return
	}
	if item.Existed {
		if !securePathWithin(i.home, item.Backup) {
			i.skipped = append(i.skipped, item.Target+" (backup path outside home; blocked)")
			return
		}
		backup, readErr := os.ReadFile(item.Backup)
		if readErr != nil {
			i.skipped = append(i.skipped, item.Target+" (backup missing)")
			return
		}
		mode := os.FileMode(item.OriginalMode)
		if mode == 0 {
			mode = 0o644
		}
		if err := atomicWrite(item.Target, backup, mode); err != nil {
			i.skipped = append(i.skipped, item.Target+" (restore failed: "+err.Error()+")")
			return
		}
	} else if err := os.Remove(item.Target); err != nil && !os.IsNotExist(err) {
		i.skipped = append(i.skipped, item.Target+" (remove failed: "+err.Error()+")")
		return
	}
	i.removed = append(i.removed, item.Target)
	removeEmptyParents(filepath.Dir(item.Target), i.home)
}

type UninstallOptions struct {
	Home      string
	Force     bool
	DryRun    bool
	RemoveCLI bool
}

type UninstallResult struct {
	Removed []string
	Skipped []string
}

func Uninstall(options UninstallOptions) (UninstallResult, error) {
	home := options.Home
	if home == "" {
		home = HomeDir()
	}
	awmHome := filepath.Join(home, ".agentic-workflow")
	if !securePathWithin(home, awmHome) {
		return UninstallResult{}, fmt.Errorf("AWM home path is unsafe: %s", awmHome)
	}
	var result UninstallResult
	state, err := LoadInstallState(home)
	if err != nil {
		if !os.IsNotExist(err) {
			return UninstallResult{}, fmt.Errorf("read AWM install state: %w", err)
		}
		if _, statErr := os.Stat(awmHome); statErr == nil {
			if options.DryRun {
				result.Removed = append(result.Removed, awmHome)
			} else if removeErr := os.RemoveAll(awmHome); removeErr != nil {
				return UninstallResult{}, removeErr
			} else {
				result.Removed = append(result.Removed, awmHome)
			}
		}
		if options.RemoveCLI {
			removeCLI(home, options.DryRun, &result)
		}
		return result, nil
	}
	for index := len(state.ManagedFiles) - 1; index >= 0; index-- {
		item := state.ManagedFiles[index]
		if !securePathWithin(home, item.Target) {
			result.Skipped = append(result.Skipped, item.Target+" (outside home; blocked)")
			continue
		}
		current, readErr := os.ReadFile(item.Target)
		if readErr != nil {
			if os.IsNotExist(readErr) {
				continue
			}
			result.Skipped = append(result.Skipped, item.Target+" (cannot read)")
			continue
		}
		if item.Block != "" {
			updated, found := stripManagedBlock(current, item.Block)
			if !found {
				continue
			}
			if options.DryRun {
				result.Removed = append(result.Removed, item.Target)
				continue
			}
			if len(bytes.TrimSpace(updated)) == 0 {
				if removeErr := os.Remove(item.Target); removeErr != nil {
					result.Skipped = append(result.Skipped, item.Target+" (remove failed: "+removeErr.Error()+")")
					continue
				}
			} else {
				mode := os.FileMode(item.OriginalMode)
				if mode == 0 {
					mode = 0o644
				}
				if writeErr := atomicWrite(item.Target, updated, mode); writeErr != nil {
					result.Skipped = append(result.Skipped, item.Target+" (write failed: "+writeErr.Error()+")")
					continue
				}
			}
			result.Removed = append(result.Removed, item.Target)
			removeEmptyParents(filepath.Dir(item.Target), home)
			continue
		}
		if !options.Force && sha256Hex(current) != item.InstalledSHA256 {
			result.Skipped = append(result.Skipped, item.Target+" (changed after AWM install)")
			continue
		}
		if options.DryRun {
			result.Removed = append(result.Removed, item.Target)
			continue
		}
		if item.Existed {
			if !securePathWithin(home, item.Backup) {
				result.Skipped = append(result.Skipped, item.Target+" (backup path outside home; blocked)")
				continue
			}
			backup, backupErr := os.ReadFile(item.Backup)
			if backupErr != nil {
				result.Skipped = append(result.Skipped, item.Target+" (backup missing)")
				continue
			}
			mode := os.FileMode(item.OriginalMode)
			if mode == 0 {
				mode = 0o644
			}
			if writeErr := atomicWrite(item.Target, backup, mode); writeErr != nil {
				result.Skipped = append(result.Skipped, item.Target+" (restore failed: "+writeErr.Error()+")")
				continue
			}
		} else if removeErr := os.Remove(item.Target); removeErr != nil && !os.IsNotExist(removeErr) {
			result.Skipped = append(result.Skipped, item.Target+" (remove failed: "+removeErr.Error()+")")
			continue
		}
		result.Removed = append(result.Removed, item.Target)
		removeEmptyParents(filepath.Dir(item.Target), home)
	}
	tasksDir := filepath.Join(awmHome, "tasks")
	if entries, readErr := os.ReadDir(tasksDir); readErr == nil {
		for _, entry := range entries {
			if !entry.IsDir() {
				continue
			}
			taskID := entry.Name()
			statePath := filepath.Join(tasksDir, taskID, "state.json")
			if _, statErr := os.Stat(statePath); statErr != nil {
				continue
			}
			if options.DryRun {
				data, readErr := os.ReadFile(statePath)
				if readErr != nil {
					result.Skipped = append(result.Skipped, statePath+" (cannot read)")
					continue
				}
				var task builderTaskState
				if err := json.Unmarshal(data, &task); err != nil {
					result.Skipped = append(result.Skipped, statePath+" (cannot parse)")
					continue
				}
				for _, file := range task.Files {
					result.Removed = append(result.Removed, file.Target)
				}
				continue
			}
			restored, skipped, taskErr := RollbackBuilderTask(home, taskID, options.Force)
			if taskErr != nil {
				result.Skipped = append(result.Skipped, statePath+" ("+taskErr.Error()+")")
				continue
			}
			result.Removed = append(result.Removed, restored...)
			result.Skipped = append(result.Skipped, skipped...)
		}
	}
	if len(result.Skipped) == 0 {
		if options.DryRun {
			result.Removed = append(result.Removed, awmHome)
		} else if removeErr := os.RemoveAll(awmHome); removeErr != nil {
			result.Skipped = append(result.Skipped, awmHome+" (remove failed: "+removeErr.Error()+")")
		} else {
			result.Removed = append(result.Removed, awmHome)
		}
	}
	if options.RemoveCLI {
		removeCLI(home, options.DryRun, &result)
	}
	return result, nil
}

func removeCLI(home string, dryRun bool, result *UninstallResult) {
	executable, err := os.Executable()
	if err != nil {
		result.Skipped = append(result.Skipped, "awm binary (cannot locate)")
		return
	}
	executable, err = filepath.Abs(executable)
	if err != nil || !securePathWithin(home, executable) {
		result.Skipped = append(result.Skipped, executable+" (outside home; kept)")
		return
	}
	base := filepath.Base(executable)
	if base != "awm" && base != "awm.exe" {
		result.Skipped = append(result.Skipped, executable+" (not an awm binary; kept)")
		return
	}
	if dryRun {
		result.Removed = append(result.Removed, executable)
		return
	}
	if err := os.Remove(executable); err != nil {
		result.Skipped = append(result.Skipped, executable+" (remove after exit: "+err.Error()+")")
		return
	}
	result.Removed = append(result.Removed, executable)
}

func stripManagedBlock(content []byte, name string) ([]byte, bool) {
	begin := "<!-- AWM:" + name + ":BEGIN -->"
	end := "<!-- AWM:" + name + ":END -->"
	text := string(content)
	start := strings.Index(text, begin)
	finish := strings.Index(text, end)
	if start < 0 || finish < start {
		return content, false
	}
	finish += len(end)
	remaining := strings.TrimSpace(text[:start] + text[finish:])
	if remaining == "" {
		return nil, true
	}
	return []byte(remaining + "\n"), true
}

func removeEmptyParents(dir, stop string) {
	stopAbs, err := filepath.Abs(stop)
	if err != nil {
		return
	}
	dirAbs, err := filepath.Abs(dir)
	if err != nil {
		return
	}
	for dirAbs != stopAbs && pathWithin(stopAbs, dirAbs) {
		if err := os.Remove(dirAbs); err != nil {
			return
		}
		dirAbs = filepath.Dir(dirAbs)
	}
}

func atomicWrite(target string, data []byte, mode os.FileMode) error {
	if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
		return err
	}
	temp, err := os.CreateTemp(filepath.Dir(target), ".awm-*.tmp")
	if err != nil {
		return err
	}
	tempName := temp.Name()
	defer os.Remove(tempName)
	if _, err := temp.Write(data); err != nil {
		temp.Close()
		return err
	}
	if err := temp.Chmod(mode); err != nil {
		temp.Close()
		return err
	}
	if err := temp.Sync(); err != nil {
		temp.Close()
		return err
	}
	if err := temp.Close(); err != nil {
		return err
	}
	return os.Rename(tempName, target)
}

func pathWithin(root, target string) bool {
	rootAbs, rootErr := filepath.Abs(root)
	targetAbs, targetErr := filepath.Abs(target)
	if rootErr != nil || targetErr != nil {
		return false
	}
	rel, err := filepath.Rel(rootAbs, targetAbs)
	return err == nil && rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator))
}

func sha256Hex(content []byte) string {
	sum := sha256.Sum256(content)
	return hex.EncodeToString(sum[:])
}

func (i *installer) rollbackWrittenFiles(force bool) {
	for index := len(i.state.ManagedFiles) - 1; index >= 0; index-- {
		item := i.state.ManagedFiles[index]
		if !securePathWithin(i.home, item.Target) {
			continue
		}
		current, err := os.ReadFile(item.Target)
		if err == nil && !force && sha256Hex(current) != item.InstalledSHA256 {
			continue
		}
		if item.Existed && securePathWithin(i.home, item.Backup) {
			if backup, readErr := os.ReadFile(item.Backup); readErr == nil {
				mode := os.FileMode(item.OriginalMode)
				if mode == 0 {
					mode = 0o644
				}
				_ = atomicWrite(item.Target, backup, mode)
			}
		} else if !item.Existed {
			_ = os.Remove(item.Target)
		}
	}
}

func securePathWithin(root, target string) bool {
	rootAbs, err := filepath.Abs(root)
	if err != nil {
		return false
	}
	rootReal, err := filepath.EvalSymlinks(rootAbs)
	if err != nil {
		if !os.IsNotExist(err) {
			return false
		}
		rootReal = rootAbs
	}
	targetAbs, err := filepath.Abs(target)
	if err != nil {
		return false
	}
	if info, statErr := os.Lstat(targetAbs); statErr == nil && info.Mode()&os.ModeSymlink != 0 {
		return false
	}
	parent := filepath.Dir(targetAbs)
	missing := []string{}
	for {
		resolved, resolveErr := filepath.EvalSymlinks(parent)
		if resolveErr == nil {
			for index := len(missing) - 1; index >= 0; index-- {
				resolved = filepath.Join(resolved, missing[index])
			}
			return pathWithin(rootReal, filepath.Join(resolved, filepath.Base(targetAbs)))
		}
		if !os.IsNotExist(resolveErr) {
			return false
		}
		base := filepath.Base(parent)
		if base == "." || base == string(filepath.Separator) {
			return false
		}
		missing = append(missing, base)
		next := filepath.Dir(parent)
		if next == parent {
			return false
		}
		parent = next
	}
}

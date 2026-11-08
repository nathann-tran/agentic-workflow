package awm

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type InstalledSkill struct {
	ID            string
	Name          string
	Description   string
	Version       string
	Dependencies  []string
	Compatibility []string
	Path          string
}

func SkillRoots(h DetectedHarness, home string) []string {
	seen := map[string]bool{}
	var roots []string
	for _, raw := range h.SkillPaths {
		path := ExpandHome(raw, home)
		if path != "" && !seen[path] {
			seen[path] = true
			roots = append(roots, path)
		}
	}
	return roots
}

func ScanSkills(roots []string) ([]InstalledSkill, error) {
	byID := map[string]InstalledSkill{}
	for _, root := range roots {
		entries, err := os.ReadDir(root)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			if os.IsPermission(err) {
				continue
			}
			return nil, fmt.Errorf("read skill directory %s: %w", root, err)
		}
		for _, entry := range entries {
			folder := filepath.Join(root, entry.Name())
			isDirectory := entry.IsDir()
			if !isDirectory && entry.Type()&os.ModeSymlink != 0 {
				info, statErr := os.Stat(folder)
				isDirectory = statErr == nil && info.IsDir()
			}
			if !isDirectory {
				continue
			}
			manifest := findSkillManifest(folder)
			if manifest == "" {
				continue
			}
			metadata := readSkillMetadata(manifest)
			id := metadata.ID
			if id == "" {
				id = entry.Name()
			}
			key := strings.ToLower(strings.TrimSpace(id))
			if key == "" {
				continue
			}
			byID[key] = InstalledSkill{
				ID:            key,
				Name:          firstNonEmpty(metadata.Name, entry.Name()),
				Description:   metadata.Description,
				Version:       metadata.Version,
				Dependencies:  metadata.Dependencies,
				Compatibility: metadata.Compatibility,
				Path:          folder,
			}
		}
	}
	result := make([]InstalledSkill, 0, len(byID))
	for _, item := range byID {
		result = append(result, item)
	}
	sortSkills(result)
	return result, nil
}

func findSkillManifest(folder string) string {
	entries, err := os.ReadDir(folder)
	if err != nil {
		return ""
	}
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		if strings.EqualFold(entry.Name(), "SKILL.md") {
			return filepath.Join(folder, entry.Name())
		}
	}
	return ""
}

type skillMetadata struct {
	ID            string
	Name          string
	Description   string
	Version       string
	Dependencies  []string
	Compatibility []string
}

func readSkillMetadata(path string) skillMetadata {
	file, err := os.Open(path)
	if err != nil {
		return skillMetadata{}
	}
	defer file.Close()
	var metadata skillMetadata
	scanner := bufio.NewScanner(file)
	inFrontmatter := false
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "---" {
			if !inFrontmatter {
				inFrontmatter = true
				continue
			}
			break
		}
		if !inFrontmatter {
			continue
		}
		key, value, ok := strings.Cut(line, ":")
		if !ok {
			continue
		}
		value = strings.Trim(strings.TrimSpace(value), "\"'")
		switch strings.TrimSpace(key) {
		case "name":
			metadata.Name = value
		case "id":
			metadata.ID = value
		case "description":
			metadata.Description = value
		case "version":
			metadata.Version = value
		case "dependencies":
			metadata.Dependencies = parseFrontmatterList(value)
		case "compatibility":
			metadata.Compatibility = parseFrontmatterList(value)
		}
	}
	return metadata
}

func AuditSkills(harnesses []DetectedHarness, home string) ([]SkillFinding, []SkillFinding, error) {
	catalog, err := LoadEmbeddedJSON[SkillCatalog]("registries/skills.json")
	if err != nil {
		return nil, nil, err
	}
	var roots []string
	for _, h := range harnesses {
		roots = append(roots, SkillRoots(h, home)...)
	}
	roots = append(roots, filepath.Join(home, ".agents", "skills"))
	installed, err := ScanSkills(uniqueStrings(roots))
	if err != nil {
		return nil, nil, err
	}
	installedByID := map[string]InstalledSkill{}
	for _, item := range installed {
		installedByID[strings.ToLower(item.ID)] = item
		installedByID[strings.ToLower(filepath.Base(item.Path))] = item
	}
	var bundled, external []SkillFinding
	for _, item := range catalog.Bundled {
		found, ok := installedByID[strings.ToLower(item.ID)]
		status := "missing"
		compatStatus := "unknown"
		var actualCompatibility, unsupported []string
		if ok {
			status = versionStatus(found.Version, item.Version)
			actualCompatibility = found.Compatibility
			compatStatus = compareCompatibility(item.Compatibility, actualCompatibility)
			unsupported = unsupportedDetectedHarnesses(harnesses, actualCompatibility)
		}
		bundled = append(bundled, SkillFinding{
			ID: item.ID, Description: item.Description, Required: item.Required,
			Installed: ok, Path: found.Path, Source: "AWM bundled", Installable: true,
			Version: found.Version, ExpectedVersion: item.Version, VersionStatus: status,
			Dependencies: item.Dependencies, MissingDependencies: missingDependencies(item.Dependencies, installedByID),
			Compatibility: item.Compatibility, ActualCompatibility: actualCompatibility,
			CompatibilityStatus: compatStatus, UnsupportedHarnesses: unsupported,
		})
	}
	for _, item := range catalog.External {
		found, ok := installedByID[strings.ToLower(item.ID)]
		status := "missing"
		compatStatus := "unknown"
		var actualCompatibility, unsupported []string
		if ok {
			status = versionStatus(found.Version, "")
			actualCompatibility = found.Compatibility
			compatStatus = compareCompatibility(item.Compatibility, actualCompatibility)
			unsupported = unsupportedDetectedHarnesses(harnesses, actualCompatibility)
		}
		external = append(external, SkillFinding{
			ID: item.ID, Description: item.Activation + " for " + strings.Join(item.Roles, ", "),
			Required: item.Activation == "required", Installed: ok,
			Path: found.Path, Version: found.Version, VersionStatus: status,
			Source: item.Source, Installable: item.Installable, Reason: item.Reason,
			Dependencies: item.Dependencies, MissingDependencies: missingDependencies(item.Dependencies, installedByID),
			Compatibility: item.Compatibility, ActualCompatibility: actualCompatibility,
			CompatibilityStatus: compatStatus, UnsupportedHarnesses: unsupported,
		})
	}
	sortFindings(bundled)
	sortFindings(external)
	return bundled, external, nil
}

func compareCompatibility(expected, actual []string) string {
	if len(expected) == 0 || len(actual) == 0 {
		return "unknown"
	}
	actualSet := map[string]bool{}
	for _, item := range actual {
		actualSet[strings.ToLower(strings.TrimSpace(item))] = true
	}
	for _, item := range expected {
		if !actualSet[strings.ToLower(strings.TrimSpace(item))] {
			return "mismatch"
		}
	}
	return "match"
}

func unsupportedDetectedHarnesses(harnesses []DetectedHarness, compatibility []string) []string {
	if len(compatibility) == 0 {
		return nil
	}
	compatible := map[string]bool{}
	for _, item := range compatibility {
		compatible[strings.ToLower(strings.TrimSpace(item))] = true
	}
	var unsupported []string
	for _, harness := range harnesses {
		if harness.Detected && !compatible[strings.ToLower(harness.ID)] {
			unsupported = append(unsupported, harness.ID)
		}
	}
	return unsupported
}

func versionStatus(installed, expected string) string {
	if strings.TrimSpace(installed) == "" {
		return "unknown"
	}
	if strings.TrimSpace(expected) == "" {
		return "reported"
	}
	if strings.TrimSpace(installed) == strings.TrimSpace(expected) {
		return "match"
	}
	return "mismatch"
}

func missingDependencies(dependencies []string, installed map[string]InstalledSkill) []string {
	var missing []string
	for _, dependency := range dependencies {
		if _, ok := installed[strings.ToLower(dependency)]; !ok {
			missing = append(missing, dependency)
		}
	}
	return missing
}

func parseFrontmatterList(value string) []string {
	value = strings.TrimSpace(strings.Trim(value, "[]"))
	if value == "" || strings.EqualFold(value, "none") {
		return nil
	}
	var items []string
	for _, item := range strings.Split(value, ",") {
		item = strings.Trim(strings.TrimSpace(item), "\"'")
		if item != "" {
			items = append(items, item)
		}
	}
	return uniqueStrings(items)
}

func SelectSkillIDs(input string, candidates []SkillFinding) []string {
	input = strings.TrimSpace(strings.ToLower(input))
	if input == "" || input == "y" || input == "yes" || input == "a" || input == "all" {
		var ids []string
		for _, item := range candidates {
			if !item.Installed && item.Installable {
				ids = append(ids, item.ID)
			}
		}
		return ids
	}
	if input == "n" || input == "no" || input == "none" {
		return nil
	}
	wanted := map[string]bool{}
	for _, part := range strings.Split(input, ",") {
		wanted[strings.TrimSpace(part)] = true
	}
	var result []string
	for _, candidate := range candidates {
		if wanted[candidate.ID] && !candidate.Installed && candidate.Installable {
			result = append(result, candidate.ID)
		}
	}
	return result
}

func uniqueStrings(values []string) []string {
	seen := map[string]bool{}
	var result []string
	for _, value := range values {
		value = filepath.Clean(value)
		if value != "." && !seen[value] {
			seen[value] = true
			result = append(result, value)
		}
	}
	return result
}

func sortSkills(items []InstalledSkill) {
	for i := 0; i < len(items); i++ {
		for j := i + 1; j < len(items); j++ {
			if items[j].ID < items[i].ID {
				items[i], items[j] = items[j], items[i]
			}
		}
	}
}

func sortFindings(items []SkillFinding) {
	for i := 0; i < len(items); i++ {
		for j := i + 1; j < len(items); j++ {
			if items[j].ID < items[i].ID {
				items[i], items[j] = items[j], items[i]
			}
		}
	}
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}

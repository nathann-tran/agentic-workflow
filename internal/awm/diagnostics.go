package awm

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func LoadInstallState(home string) (InstallState, error) {
	if home == "" {
		home = HomeDir()
	}
	path := filepath.Join(home, ".agentic-workflow", "state", "install.json")
	if !securePathWithin(home, path) {
		return InstallState{}, fmt.Errorf("install state path is unsafe: %s", path)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return InstallState{}, err
	}
	var state InstallState
	if err := json.Unmarshal(data, &state); err != nil {
		return InstallState{}, err
	}
	return state, nil
}

func AuditReport(home string) (string, error) {
	if home == "" {
		home = HomeDir()
	}
	harnesses, err := DetectHarnesses(home)
	if err != nil {
		return "", err
	}
	bundled, external, err := AuditSkills(harnesses, home)
	if err != nil {
		return "", err
	}
	plugins, err := AuditPlugins(home)
	if err != nil {
		return "", err
	}
	var b strings.Builder
	fmt.Fprintf(&b, "AWM Doctor — %s\nPlatform: %s\n\nHarnesses\n", Version, RuntimePlatform())
	for _, h := range harnesses {
		status := "missing"
		if h.Detected {
			status = "detected"
		}
		fmt.Fprintf(&b, "  %-12s %-9s %s", h.ID, status, h.Version)
		if h.Evidence != "" {
			fmt.Fprintf(&b, " (%s)", h.Evidence)
		}
		b.WriteByte('\n')
	}
	b.WriteString("\nBundled skills\n")
	for _, skill := range bundled {
		status := "missing"
		if skill.Installed {
			status = "installed"
		}
		fmt.Fprintf(&b, "  %-28s %s", skill.ID, status)
		if skill.Version != "" {
			fmt.Fprintf(&b, " version=%s", skill.Version)
		}
		if skill.ExpectedVersion != "" && skill.VersionStatus != "match" {
			fmt.Fprintf(&b, " expected=%s (%s)", skill.ExpectedVersion, skill.VersionStatus)
		}
		if len(skill.MissingDependencies) > 0 {
			fmt.Fprintf(&b, " missing-deps=%s", strings.Join(skill.MissingDependencies, ","))
		}
		if skill.CompatibilityStatus != "" {
			fmt.Fprintf(&b, " compatibility=%s", skill.CompatibilityStatus)
		}
		if len(skill.UnsupportedHarnesses) > 0 {
			fmt.Fprintf(&b, " unsupported-detected=%s", strings.Join(skill.UnsupportedHarnesses, ","))
		}
		b.WriteByte('\n')
	}
	b.WriteString("\nExternal skills\n")
	for _, skill := range external {
		status := "missing"
		if skill.Installed {
			status = "installed"
		}
		fmt.Fprintf(&b, "  %-28s %s", skill.ID, status)
		if skill.Version != "" {
			fmt.Fprintf(&b, " version=%s", skill.Version)
		}
		if len(skill.MissingDependencies) > 0 {
			fmt.Fprintf(&b, " missing-deps=%s", strings.Join(skill.MissingDependencies, ","))
		}
		if skill.CompatibilityStatus != "" {
			fmt.Fprintf(&b, " compatibility=%s", skill.CompatibilityStatus)
		}
		if !skill.Installable && !skill.Installed {
			fmt.Fprintf(&b, " — %s", skill.Reason)
		}
		b.WriteByte('\n')
	}
	b.WriteString("\nPlugins and context services\n")
	PrintPluginFindingsTo(&b, plugins)
	for _, item := range MemoryStatus(home) {
		fmt.Fprintf(&b, "  %s\n", item)
	}
	state, stateErr := LoadInstallState(home)
	if stateErr == nil {
		fmt.Fprintf(&b, "\nAWM installation\n  Version: %s\n  Installed: %s\n  Harnesses: %s\n  Managed files: %d\n", state.Version, state.InstalledAt.Format(time.RFC3339), strings.Join(state.Selected, ", "), len(state.ManagedFiles))
	} else if os.IsNotExist(stateErr) {
		b.WriteString("\nAWM installation\n  Not installed\n")
	} else {
		return "", fmt.Errorf("read AWM install state: %w", stateErr)
	}
	b.WriteString("\nNote: this audit is read-only. It does not install or modify plugins, skills, or memory settings.\n")
	return b.String(), nil
}

func PrintPluginFindingsTo(b *strings.Builder, findings []PluginFinding) {
	for _, item := range findings {
		status := "MISSING"
		if item.Present {
			status = "FOUND"
		}
		fmt.Fprintf(b, "  %-20s %-8s %s\n", item.Plugin.ID, status, item.Plugin.Purpose)
	}
}

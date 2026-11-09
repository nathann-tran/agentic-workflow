package awm

import (
	"context"
	"encoding/json"
	"errors"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/agentic-workflow/awm/internal/assets"
)

var Version = "0.2.0"

var ErrCancelled = errors.New("operation cancelled")

type HarnessCatalog struct {
	Version   string    `json:"version"`
	Harnesses []Harness `json:"harnesses"`
}

type Harness struct {
	ID             string            `json:"id"`
	Name           string            `json:"name"`
	Binary         string            `json:"binary"`
	GlobalAgents   string            `json:"global_agents"`
	SkillPaths     []string          `json:"skill_paths"`
	GlobalGuidance string            `json:"global_guidance"`
	Capabilities   map[string]string `json:"capabilities"`
	Docs           string            `json:"docs"`
}

type DetectedHarness struct {
	Harness
	Executable string `json:"executable,omitempty"`
	Version    string `json:"version,omitempty"`
	Detected   bool   `json:"detected"`
	Evidence   string `json:"evidence"`
}

type SkillCatalog struct {
	Version  string          `json:"version"`
	Bundled  []BundledSkill  `json:"bundled"`
	External []ExternalSkill `json:"external"`
	Policy   SkillPolicy     `json:"policy"`
}

type SkillPolicy struct {
	NeverInstallUnpinnedRemoteSources bool `json:"never_install_unpinned_remote_sources"`
	RequireHumanSelection             bool `json:"require_human_selection"`
	CheckSkillMD                      bool `json:"check_skill_md"`
	DoNotOverwriteExisting            bool `json:"do_not_overwrite_existing"`
}

type BundledSkill struct {
	ID            string   `json:"id"`
	Name          string   `json:"name"`
	Description   string   `json:"description"`
	Version       string   `json:"version"`
	Dependencies  []string `json:"dependencies"`
	Compatibility []string `json:"compatibility"`
	Roles         []string `json:"roles"`
	Required      bool     `json:"required"`
	Path          string   `json:"path"`
}

type ExternalSkill struct {
	ID            string   `json:"id"`
	CanonicalID   string   `json:"canonical_id"`
	Roles         []string `json:"roles"`
	Activation    string   `json:"activation"`
	Source        string   `json:"source"`
	Ref           string   `json:"ref"`
	Installable   bool     `json:"installable"`
	Reason        string   `json:"reason"`
	Dependencies  []string `json:"dependencies"`
	Compatibility []string `json:"compatibility"`
}

type PluginCatalog struct {
	Version string   `json:"version"`
	Plugins []Plugin `json:"plugins"`
}

type Plugin struct {
	ID            string   `json:"id"`
	Purpose       string   `json:"purpose"`
	Markers       []string `json:"markers"`
	Required      bool     `json:"required"`
	InstallPolicy string   `json:"install_policy"`
	Note          string   `json:"note"`
}

type SkillFinding struct {
	ID                   string
	Description          string
	Required             bool
	Installed            bool
	Path                 string
	Source               string
	Installable          bool
	Reason               string
	Version              string
	ExpectedVersion      string
	VersionStatus        string
	Dependencies         []string
	MissingDependencies  []string
	Compatibility        []string
	ActualCompatibility  []string
	CompatibilityStatus  string
	UnsupportedHarnesses []string
}

type PluginFinding struct {
	Plugin  Plugin
	Present bool
	Path    string
}

type InstallOptions struct {
	HarnessIDs      []string
	Force           bool
	Yes             bool
	DryRun          bool
	SkipSkills      bool
	SkillIDs        []string
	AllowUndetected bool
	PruneStale      bool
	Home            string
}

type ManagedFile struct {
	Target          string `json:"target"`
	Backup          string `json:"backup,omitempty"`
	Existed         bool   `json:"existed"`
	OriginalMode    uint32 `json:"original_mode,omitempty"`
	InstalledSHA256 string `json:"installed_sha256"`
	Block           string `json:"block,omitempty"`
}

type InstallState struct {
	Version             string        `json:"version"`
	InstalledAt         time.Time     `json:"installed_at"`
	BackupRoot          string        `json:"backup_root"`
	Selected            []string      `json:"selected_harnesses"`
	ManagedFiles        []ManagedFile `json:"managed_files"`
	SkippedFiles        []string      `json:"skipped_files,omitempty"`
	InstallSource       string        `json:"install_source"`
	PreviousStateBackup string        `json:"previous_state_backup,omitempty"`
}

func HomeDir() string {
	if home := os.Getenv("HOME"); home != "" {
		return home
	}
	home, _ := os.UserHomeDir()
	return home
}

func ExpandHome(path, home string) string {
	if strings.HasPrefix(path, "~/") || path == "~" {
		return filepath.Join(home, strings.TrimPrefix(strings.TrimPrefix(path, "~"), string(filepath.Separator)))
	}
	return filepath.Clean(path)
}

func LoadEmbeddedJSON[T any](name string) (T, error) {
	var result T
	data, err := fs.ReadFile(assets.Files, filepath.ToSlash(filepath.Join("data", name)))
	if err != nil {
		return result, err
	}
	err = json.Unmarshal(data, &result)
	return result, err
}

func DetectHarnesses(home string) ([]DetectedHarness, error) {
	catalog, err := LoadEmbeddedJSON[HarnessCatalog]("registries/harnesses.json")
	if err != nil {
		return nil, err
	}
	found := make([]DetectedHarness, 0, len(catalog.Harnesses))
	for _, h := range catalog.Harnesses {
		item := DetectedHarness{Harness: h}
		if path, lookErr := exec.LookPath(h.Binary); lookErr == nil {
			item.Executable = path
			item.Detected = true
			item.Evidence = "binary found in PATH"
			item.Version = detectVersion(path)
		} else {
			agentPath := ExpandHome(h.GlobalAgents, home)
			if info, statErr := os.Stat(agentPath); statErr == nil && info.IsDir() {
				item.Detected = true
				item.Evidence = "global agent directory exists; executable not found in PATH"
			}
		}
		found = append(found, item)
	}
	return found, nil
}

func detectVersion(binary string) string {
	ctx, cancel := contextWithTimeout(1800 * time.Millisecond)
	defer cancel()
	cmd := exec.CommandContext(ctx, binary, "--version")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "version unavailable"
	}
	value := strings.TrimSpace(string(output))
	if len(value) > 100 {
		value = value[:100]
	}
	return value
}

func contextWithTimeout(duration time.Duration) (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), duration)
}

func SelectedHarnesses(all []DetectedHarness, ids []string, allowUndetected bool) ([]DetectedHarness, error) {
	selectAll := len(ids) == 0
	wanted := make(map[string]bool, len(ids))
	for _, id := range ids {
		wanted[strings.TrimSpace(strings.ToLower(id))] = true
	}
	var selected []DetectedHarness
	for _, h := range all {
		if selectAll {
			if h.Detected {
				selected = append(selected, h)
			}
			continue
		}
		if wanted[h.ID] {
			if !h.Detected && !allowUndetected {
				return nil, errors.New("harness " + h.ID + " is not detected; use --allow-undetected to install anyway")
			}
			selected = append(selected, h)
			delete(wanted, h.ID)
		}
	}
	if len(wanted) > 0 {
		var unknown []string
		for id := range wanted {
			unknown = append(unknown, id)
		}
		return nil, errors.New("unknown harness IDs: " + strings.Join(unknown, ", "))
	}
	if len(selected) == 0 {
		return nil, errors.New("no supported harness was detected or selected")
	}
	return selected, nil
}

func RuntimePlatform() string {
	return runtime.GOOS + "/" + runtime.GOARCH
}

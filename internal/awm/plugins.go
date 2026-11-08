package awm

import (
	"fmt"
	"os"
	"path/filepath"
)

func AuditPlugins(home string) ([]PluginFinding, error) {
	catalog, err := LoadEmbeddedJSON[PluginCatalog]("registries/plugins.json")
	if err != nil {
		return nil, err
	}
	findings := make([]PluginFinding, 0, len(catalog.Plugins))
	for _, plugin := range catalog.Plugins {
		finding := PluginFinding{Plugin: plugin}
		for _, marker := range plugin.Markers {
			path := ExpandHome(marker, home)
			if _, statErr := os.Stat(path); statErr == nil {
				finding.Present = true
				finding.Path = path
				break
			}
			if matches, _ := filepath.Glob(path); len(matches) > 0 {
				finding.Present = true
				finding.Path = matches[0]
				break
			}
		}
		findings = append(findings, finding)
	}
	return findings, nil
}

func PrintPluginFindings(findings []PluginFinding) {
	for _, item := range findings {
		status := "MISSING"
		if item.Present {
			status = "FOUND"
		}
		fmt.Printf("%-20s %-8s %s\n", item.Plugin.ID, status, item.Plugin.Purpose)
		if item.Present {
			fmt.Printf("  %s\n", item.Path)
		} else {
			fmt.Printf("  %s\n", item.Plugin.Note)
		}
	}
}

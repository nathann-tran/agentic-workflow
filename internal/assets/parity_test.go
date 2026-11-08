package assets

import (
	"bytes"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

func ignoredName(name string) bool {
	return strings.HasPrefix(name, ".") || strings.HasPrefix(name, "_")
}

func hasIgnoredSegment(rel string) bool {
	for _, segment := range strings.Split(rel, "/") {
		if ignoredName(segment) {
			return true
		}
	}
	return false
}

func TestEmbeddedPayloadMatchesSource(t *testing.T) {
	sourceRoot, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatal(err)
	}
	entries, err := fs.ReadDir(Files, "data")
	if err != nil {
		t.Fatal(err)
	}
	var roots []string
	for _, entry := range entries {
		if entry.IsDir() {
			roots = append(roots, entry.Name())
		}
	}
	sort.Strings(roots)
	if len(roots) == 0 {
		t.Fatal("embedded payload is empty")
	}

	embedded := map[string][]byte{}
	for _, root := range roots {
		err := fs.WalkDir(Files, filepath.ToSlash(filepath.Join("data", root)), func(path string, entry fs.DirEntry, walkErr error) error {
			if walkErr != nil {
				return walkErr
			}
			if entry.IsDir() {
				return nil
			}
			rel := strings.TrimPrefix(path, "data/")
			if hasIgnoredSegment(rel) {
				return nil
			}
			data, err := fs.ReadFile(Files, path)
			if err != nil {
				return err
			}
			embedded[rel] = data
			return nil
		})
		if err != nil {
			t.Fatal(err)
		}
	}

	source := map[string][]byte{}
	for _, root := range roots {
		rootPath := filepath.Join(sourceRoot, root)
		info, statErr := os.Stat(rootPath)
		if statErr != nil {
			t.Fatalf("source tree missing: %v", statErr)
		}
		if !info.IsDir() {
			continue
		}
		err := filepath.WalkDir(rootPath, func(path string, entry fs.DirEntry, walkErr error) error {
			if walkErr != nil {
				return walkErr
			}
			if entry.IsDir() {
				if path != rootPath && ignoredName(entry.Name()) {
					return fs.SkipDir
				}
				return nil
			}
			if ignoredName(entry.Name()) {
				return nil
			}
			rel, err := filepath.Rel(sourceRoot, path)
			if err != nil {
				return err
			}
			data, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			source[filepath.ToSlash(rel)] = data
			return nil
		})
		if err != nil {
			t.Fatal(err)
		}
	}

	for _, rel := range sortedKeys(source) {
		if _, ok := embedded[rel]; !ok {
			t.Errorf("missing from embedded payload: %s (run make sync-assets)", rel)
		}
	}
	for _, rel := range sortedKeys(embedded) {
		if _, ok := source[rel]; !ok {
			t.Errorf("embedded file has no source: %s (run make sync-assets)", rel)
		}
	}
	for _, rel := range sortedKeys(source) {
		if other, ok := embedded[rel]; ok && !bytes.Equal(source[rel], other) {
			t.Errorf("embedded payload differs from source: %s (run make sync-assets)", rel)
		}
	}
}

func sortedKeys(items map[string][]byte) []string {
	keys := make([]string, 0, len(items))
	for key := range items {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

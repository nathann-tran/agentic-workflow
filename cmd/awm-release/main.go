// Command awm-release builds the release binaries and writes the archives,
// manifest, and checksums. It is a maintainer tool, not part of the awm CLI.
package main

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"time"
)

const product = "Agentic Workflow Manager"

var fixedTime = time.Date(1980, 1, 1, 0, 0, 0, 0, time.UTC)

type target struct {
	os     string
	arch   string
	binary string
}

var targets = []target{
	{"darwin", "amd64", "awm-darwin-amd64"},
	{"darwin", "arm64", "awm-darwin-arm64"},
	{"linux", "amd64", "awm-linux-amd64"},
	{"linux", "arm64", "awm-linux-arm64"},
	{"windows", "amd64", "awm-windows-amd64.exe"},
}

type artifact struct {
	File      string `json:"file"`
	Kind      string `json:"kind"`
	OS        string `json:"os,omitempty"`
	Arch      string `json:"arch,omitempty"`
	SizeBytes int64  `json:"size_bytes"`
	SHA256    string `json:"sha256"`
}

type manifest struct {
	Product                        string     `json:"product"`
	Command                        string     `json:"command"`
	Version                        string     `json:"version"`
	ReleaseDate                    string     `json:"release_date"`
	SourceModule                   string     `json:"source_module"`
	GoVersion                      string     `json:"go_version"`
	BuildHost                      string     `json:"build_host"`
	Commit                         string     `json:"commit"`
	License                        string     `json:"license"`
	Artifacts                      []artifact `json:"artifacts"`
	NativeHarnessRuntimeValidation string     `json:"native_harness_runtime_validation"`
}

func main() {
	dist := flag.String("dist", "dist", "output directory")
	version := flag.String("version", "", "release version (defaults to VERSION)")
	flag.Parse()

	if *version == "" {
		content, err := os.ReadFile("VERSION")
		if err != nil {
			fatal(err)
		}
		*version = strings.TrimSpace(string(content))
	}
	module, err := readModule()
	if err != nil {
		fatal(err)
	}
	license, err := os.ReadFile("LICENSE.txt")
	if err != nil {
		fatal(err)
	}
	readme, err := os.ReadFile("README.md")
	if err != nil {
		fatal(err)
	}
	if err := os.MkdirAll(*dist, 0o755); err != nil {
		fatal(err)
	}

	var artifacts []artifact
	for _, t := range targets {
		binaryPath := filepath.Join(*dist, t.binary)
		if err := buildBinary(t, module, *version, binaryPath); err != nil {
			fatal(err)
		}
		binaryArtifact, err := inspect(binaryPath, "binary", t.os, t.arch)
		if err != nil {
			fatal(err)
		}
		artifacts = append(artifacts, binaryArtifact)

		archivePath := filepath.Join(*dist, archiveName(*version, t))
		if t.os == "windows" {
			err = writeZip(archivePath, t.binary, binaryPath, license, readme)
		} else {
			err = writeTarGz(archivePath, "awm", binaryPath, license, readme)
		}
		if err != nil {
			fatal(err)
		}
		archiveArtifact, err := inspect(archivePath, "archive", t.os, t.arch)
		if err != nil {
			fatal(err)
		}
		artifacts = append(artifacts, archiveArtifact)
	}

	sort.Slice(artifacts, func(i, j int) bool { return artifacts[i].File < artifacts[j].File })
	if err := writeChecksums(filepath.Join(*dist, "SHA256SUMS.txt"), artifacts); err != nil {
		fatal(err)
	}
	document := manifest{
		Product:                        product,
		Command:                        "awm",
		Version:                        *version,
		ReleaseDate:                    time.Now().UTC().Format("2006-01-02"),
		SourceModule:                   module,
		GoVersion:                      runtime.Version(),
		BuildHost:                      runtime.GOOS + "/" + runtime.GOARCH,
		Commit:                         gitCommit(),
		License:                        "MIT",
		Artifacts:                      artifacts,
		NativeHarnessRuntimeValidation: "pending on target machine",
	}
	if err := writeJSON(filepath.Join(*dist, "RELEASE-MANIFEST.json"), document); err != nil {
		fatal(err)
	}
	fmt.Printf("AWM %s: %d binaries, %d archives, manifest and checksums written to %s\n",
		*version, len(targets), len(targets), *dist)
}

func readModule() (string, error) {
	content, err := os.ReadFile("go.mod")
	if err != nil {
		return "", err
	}
	for _, line := range strings.Split(string(content), "\n") {
		if strings.HasPrefix(line, "module ") {
			return strings.TrimSpace(strings.TrimPrefix(line, "module ")), nil
		}
	}
	return "", fmt.Errorf("module path missing from go.mod")
}

func buildBinary(t target, module, version, output string) error {
	command := exec.Command("go", "build", "-trimpath",
		"-ldflags", "-s -w -X "+module+"/internal/awm.Version="+version,
		"-o", output, "./cmd/awm")
	command.Env = append(os.Environ(), "CGO_ENABLED=0", "GOOS="+t.os, "GOARCH="+t.arch)
	if combined, err := command.CombinedOutput(); err != nil {
		return fmt.Errorf("build %s/%s: %w\n%s", t.os, t.arch, err, combined)
	}
	return nil
}

func archiveName(version string, t target) string {
	extension := "tar.gz"
	if t.os == "windows" {
		extension = "zip"
	}
	return fmt.Sprintf("awm_%s_%s_%s.%s", version, t.os, t.arch, extension)
}

func inspect(path, kind, goos, arch string) (artifact, error) {
	digest, size, err := hashFile(path)
	if err != nil {
		return artifact{}, err
	}
	return artifact{
		File:      filepath.Base(path),
		Kind:      kind,
		OS:        goos,
		Arch:      arch,
		SizeBytes: size,
		SHA256:    digest,
	}, nil
}

func hashFile(path string) (string, int64, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", 0, err
	}
	defer file.Close()
	digest := sha256.New()
	size, err := io.Copy(digest, file)
	if err != nil {
		return "", 0, err
	}
	return hex.EncodeToString(digest.Sum(nil)), size, nil
}

func writeTarGz(path, binaryName, binaryPath string, license, readme []byte) (err error) {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer func() {
		if closeErr := file.Close(); err == nil {
			err = closeErr
		}
	}()
	compressed := gzip.NewWriter(file)
	archive := tar.NewWriter(compressed)

	if err := addTarFile(archive, binaryName, binaryPath, 0o755); err != nil {
		return err
	}
	if err := addTarBytes(archive, "LICENSE.txt", license, 0o644); err != nil {
		return err
	}
	if err := addTarBytes(archive, "README.md", readme, 0o644); err != nil {
		return err
	}
	if err := archive.Close(); err != nil {
		return err
	}
	return compressed.Close()
}

func addTarFile(archive *tar.Writer, name, path string, mode int64) error {
	info, err := os.Stat(path)
	if err != nil {
		return err
	}
	header := &tar.Header{Name: name, Mode: mode, Size: info.Size(), ModTime: fixedTime}
	if err := archive.WriteHeader(header); err != nil {
		return err
	}
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()
	_, err = io.Copy(archive, file)
	return err
}

func addTarBytes(archive *tar.Writer, name string, content []byte, mode int64) error {
	header := &tar.Header{Name: name, Mode: mode, Size: int64(len(content)), ModTime: fixedTime}
	if err := archive.WriteHeader(header); err != nil {
		return err
	}
	_, err := archive.Write(content)
	return err
}

func writeZip(path, binaryName, binaryPath string, license, readme []byte) (err error) {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer func() {
		if closeErr := file.Close(); err == nil {
			err = closeErr
		}
	}()
	archive := zip.NewWriter(file)

	if err := addZipFile(archive, binaryName, binaryPath, 0o755); err != nil {
		return err
	}
	if err := addZipBytes(archive, "LICENSE.txt", license, 0o644); err != nil {
		return err
	}
	if err := addZipBytes(archive, "README.md", readme, 0o644); err != nil {
		return err
	}
	return archive.Close()
}

func addZipFile(archive *zip.Writer, name, path string, mode os.FileMode) error {
	info, err := os.Stat(path)
	if err != nil {
		return err
	}
	header, err := zip.FileInfoHeader(info)
	if err != nil {
		return err
	}
	header.Name = name
	header.Method = zip.Deflate
	header.Modified = fixedTime
	header.SetMode(mode)
	writer, err := archive.CreateHeader(header)
	if err != nil {
		return err
	}
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()
	_, err = io.Copy(writer, file)
	return err
}

func addZipBytes(archive *zip.Writer, name string, content []byte, mode os.FileMode) error {
	header := &zip.FileHeader{Name: name, Method: zip.Deflate, Modified: fixedTime}
	header.SetMode(mode)
	writer, err := archive.CreateHeader(header)
	if err != nil {
		return err
	}
	_, err = writer.Write(content)
	return err
}

func writeChecksums(path string, artifacts []artifact) error {
	var builder strings.Builder
	for _, item := range artifacts {
		fmt.Fprintf(&builder, "%s  %s\n", item.SHA256, item.File)
	}
	return os.WriteFile(path, []byte(builder.String()), 0o644)
}

func writeJSON(path string, value any) error {
	content, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}
	content = append(content, '\n')
	return os.WriteFile(path, content, 0o644)
}

func gitCommit() string {
	output, err := exec.Command("git", "rev-parse", "HEAD").Output()
	if err != nil {
		return "uncommitted"
	}
	return strings.TrimSpace(string(output))
}

func fatal(err error) {
	fmt.Fprintln(os.Stderr, "awm-release:", err)
	os.Exit(1)
}

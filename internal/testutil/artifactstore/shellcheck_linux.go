//go:build linux

package artifactstore

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
)

const shellCheckKeySchemaID = "cartulary.shellcheck_key.v1"

// shellCheckImplementation is this producer's own source.
const shellCheckImplementation = "internal/testutil/artifactstore/shellcheck_linux.go"

// shellCheckReleases is where ShellCheck publishes its release archives.
const shellCheckReleases = "https://github.com/koalaman/shellcheck/releases/download"

// ShellCheck installs the pinned ShellCheck release from its archive, which
// must have the committed digest. Its key is the archive's name and digest.
type ShellCheck struct {
	// Root is the repository root, where this producer's source is keyed.
	Root string
	// Version is the pinned version; Archive is the release archive's file
	// name and SHA256 its committed digest.
	Version, Archive, SHA256 string
	// Archives is the directory that keeps downloaded archives.
	Archives string
	// Releases is the release download base; empty means ShellCheck's own.
	Releases string
}

// ShellCheckArchive names the release archive of version for this platform.
func ShellCheckArchive(version string) (string, error) {
	machine := map[string]string{"amd64": "x86_64", "arm64": "aarch64"}[runtime.GOARCH]
	if runtime.GOOS != "linux" || machine == "" {
		return "", fmt.Errorf("no ShellCheck release for %s/%s", runtime.GOOS, runtime.GOARCH)
	}
	return fmt.Sprintf("shellcheck-v%s.linux.%s.tar.xz", version, machine), nil
}

// ID names the producer in key records.
func (s ShellCheck) ID() string { return "shellcheck" }

// Key digests the archive identity and this producer.
func (s ShellCheck) Key() (string, error) {
	implementation, err := fileDigest(filepath.Join(s.Root, shellCheckImplementation))
	if err != nil {
		return "", err
	}
	data, err := json.Marshal(map[string]any{
		"schema_id": shellCheckKeySchemaID, "producer": s.ID(), "version": s.Version, "archive": s.Archive,
		"archive_sha256": s.SHA256, "outputs": []string{"shellcheck"},
		"inputs": map[string]string{shellCheckImplementation: implementation},
	})
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:]), nil
}

// Build obtains the verified archive and extracts its executable into
// staging. An archive with another digest is removed and never extracted.
func (s ShellCheck) Build(staging string) error {
	archive, err := s.verifiedArchive()
	if err != nil {
		return err
	}
	extract, err := os.MkdirTemp(s.Archives, ".extract-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(extract)
	command := exec.Command("tar", "-xJf", archive, "-C", extract)
	command.Stdout, command.Stderr = os.Stderr, os.Stderr
	if err := command.Run(); err != nil {
		return fmt.Errorf("extract %s: %w", s.Archive, err)
	}
	executable := filepath.Join(extract, "shellcheck-v"+s.Version, "shellcheck")
	if info, err := os.Lstat(executable); err != nil || !info.Mode().IsRegular() {
		return fmt.Errorf("%s holds no shellcheck-v%s/shellcheck executable", s.Archive, s.Version)
	}
	file, err := os.Open(executable)
	if err != nil {
		return err
	}
	defer file.Close()
	out, err := os.OpenFile(filepath.Join(staging, "shellcheck"), os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o755)
	if err != nil {
		return err
	}
	_, err = io.Copy(out, file)
	if closeErr := out.Close(); err == nil {
		err = closeErr
	}
	return err
}

// verifiedArchive returns the kept archive when it has the committed digest,
// downloading it first when it is missing or altered.
func (s ShellCheck) verifiedArchive() (string, error) {
	releases := s.Releases
	if releases == "" {
		releases = shellCheckReleases
	}
	return verifiedArchive(s.Archives, s.Archive, releases+"/v"+s.Version+"/"+s.Archive, s.SHA256)
}

//go:build linux

package artifactstore

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
)

// LicenseReportName is the license report published beside the SBOM.
const LicenseReportName = "license-report.json"

// ReleaseInventory regenerates the release SBOM and license report at their
// paths each time it is produced. Its inputs include the package graph of
// every Go module and the installed frontend packages, which no file list
// captures, so the store keeps no key for it: its unit's planner cache reuses
// the two outputs.
type ReleaseInventory struct {
	// Root is the repository root; SBOM the SBOM's path, with the license
	// report beside it.
	Root, SBOM string
	// The executables the generator runs, and the scanners it drives.
	Node, PNPM, NodeRuntime, Go, CycloneDX, Syft string
	Env                                          []string
	Wrap                                         func([]string) []string
}

// ID names the producer.
func (r ReleaseInventory) ID() string { return "release-inventory" }

// Generate writes both outputs through the generator, which publishes them as
// a pair.
func (r ReleaseInventory) Generate() error {
	directory := filepath.Dir(r.SBOM)
	if err := os.MkdirAll(directory, 0o755); err != nil {
		return err
	}
	argv := []string{r.Node, "./tools/release-evidence/generate-sbom-license-evidence.mjs"}
	if r.Wrap != nil {
		argv = r.Wrap(argv)
	}
	command := exec.Command(argv[0], argv[1:]...)
	command.Dir = r.Root
	command.Env = append(nodeEnvironment(slices.Clone(r.Env), r.NodeRuntime), "GO="+r.Go, "NODE_BIN="+r.Node, "PNPM="+r.PNPM,
		"CYCLONEDX_GOMOD_BIN="+r.CycloneDX, "SYFT_BIN="+r.Syft, "RELEASE_ARTIFACT_DIR="+directory,
		"SBOM_ARTIFACT="+r.SBOM, "LICENSE_REPORT_ARTIFACT="+filepath.Join(directory, LicenseReportName))
	command.Stdout, command.Stderr = os.Stderr, os.Stderr
	if err := command.Run(); err != nil {
		return fmt.Errorf("generate the release inventory: %w", err)
	}
	return nil
}

const testServiceImagesKeySchemaID = "cartulary.test_service_images_key.v1"

// testServiceImagesImplementation is this producer's own source.
const testServiceImagesImplementation = "internal/testutil/artifactstore/release_linux.go"

// TestServiceImages pulls the pinned images the test services start. The
// test-services binary compiles their references in, so its digest is the
// key. The images live in the Docker store, outside the receipt's directory,
// so reuse also needs every image present; a missing one installs again.
type TestServiceImages struct {
	// Root is the repository root; Receipt the installation receipt; Binary
	// the test-services executable.
	Root, Receipt, Binary string
	Env                   []string
	Wrap                  func([]string) []string
}

// ID names the producer in receipts.
func (i TestServiceImages) ID() string { return "test-service-images" }

func (i TestServiceImages) key() (string, error) {
	inputs := map[string]string{}
	for name, path := range map[string]string{"test_services": i.Binary, testServiceImagesImplementation: filepath.Join(i.Root, testServiceImagesImplementation)} {
		digest, err := fileDigest(path)
		if err != nil {
			return "", err
		}
		inputs[name] = digest
	}
	data, err := json.Marshal(map[string]any{"schema_id": testServiceImagesKeySchemaID, "producer": i.ID(), "inputs": inputs})
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:]), nil
}

// Current reports whether the receipt records the current key and every
// pinned image is present in the Docker store.
func (i TestServiceImages) Current() (bool, error) {
	key, err := i.key()
	if err != nil || !receiptMatches(i.Receipt, i.ID(), key) {
		return false, err
	}
	command := exec.Command(i.Binary, "images")
	command.Dir, command.Env = i.Root, i.Env
	output, err := command.Output()
	if err != nil {
		return false, nil
	}
	images := strings.Fields(string(output))
	if len(images) == 0 {
		return false, nil
	}
	for _, image := range images {
		inspect := exec.Command(executableIn(i.Env, "docker"), "image", "inspect", image)
		inspect.Env = i.Env
		if current, err := probe(inspect); err != nil || !current {
			return false, err
		}
	}
	return true, nil
}

// Install pulls the images and writes the receipt.
func (i TestServiceImages) Install() error {
	key, err := i.key()
	if err != nil {
		return err
	}
	if err := os.Remove(i.Receipt); err != nil && !os.IsNotExist(err) {
		return err
	}
	argv := []string{i.Binary, "warm-images"}
	if i.Wrap != nil {
		argv = i.Wrap(argv)
	}
	command := exec.Command(argv[0], argv[1:]...)
	command.Dir, command.Env = i.Root, i.Env
	command.Stdout, command.Stderr = os.Stderr, os.Stderr
	if err := command.Run(); err != nil {
		return fmt.Errorf("pull the test service images: %w", err)
	}
	return writeReceipt(i.Receipt, i.ID(), key)
}

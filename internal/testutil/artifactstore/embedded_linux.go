//go:build linux

package artifactstore

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
)

const embeddedAssetsKeySchemaID = "cartulary.embedded_assets_key.v1"

// embeddedAssetsImplementation is this producer's own source. The embedder
// records it among its generator sources, because it chooses the embedder's
// inputs.
const embeddedAssetsImplementation = "internal/testutil/artifactstore/embedded_linux.go"

// embedderPackage is the generator that packages a web bundle for Go embed.
const embedderPackage = "./tools/embedwebassets"

// EmbeddedAssetNames are the files the embedder writes, in the order of its
// outputs; the first is the archive.
var EmbeddedAssetNames = []string{
	"web-assets.zip", "client-asset-set-manifest.json", "client-extension-support-registry.json",
	"client-bootstrap-template.json", "extension-registry-integrity.json", "extension-registry-integrity.sha256",
}

// EmbeddedAssets packages a sealed web bundle for the server's Go embed. Its
// key covers the bundle's files, the authored client support source, the
// embedder's import closure, the module files and the effective Go settings.
type EmbeddedAssets struct {
	// Root is the repository root; Source the sealed web bundle directory;
	// ClientSupport the authored client support source, relative to Root.
	Root, Source, ClientSupport string
	Go                          string
	Env                         []string
}

// ID names the producer in key records.
func (e EmbeddedAssets) ID() string { return "embedded-assets" }

// treeDigests digests every regular file under directory by its slash path.
// A link or any other entry is refused, as an archive must not follow one.
func treeDigests(directory string) (map[string]string, error) {
	files := map[string]string{}
	err := filepath.WalkDir(directory, func(name string, entry fs.DirEntry, err error) error {
		if err != nil || entry.IsDir() {
			return err
		}
		if !entry.Type().IsRegular() {
			return fmt.Errorf("web bundle entry %s is not a regular file", name)
		}
		relative, err := filepath.Rel(directory, name)
		if err != nil {
			return err
		}
		files[filepath.ToSlash(relative)], err = fileDigest(name)
		return err
	})
	if err == nil && len(files) == 0 {
		err = fmt.Errorf("web bundle %s holds no file", directory)
	}
	return files, err
}

// Key digests the bundle, the client support source, the embedder and the
// Go settings.
func (e EmbeddedAssets) Key() (string, error) {
	bundle, err := treeDigests(e.Source)
	if err != nil {
		return "", err
	}
	closure, err := goClosure(e.Go, e.Root, embedderPackage, "", e.Env)
	if err != nil {
		return "", err
	}
	inputs := map[string]string{}
	for _, name := range slices.Concat([]string{"go.mod", "go.sum", embeddedAssetsImplementation, e.ClientSupport}, closure) {
		if inputs[name], err = fileDigest(filepath.Join(e.Root, filepath.FromSlash(name))); err != nil {
			return "", err
		}
	}
	settings, err := effectiveGoSettings(e.Go, e.Root, e.Env)
	if err != nil {
		return "", err
	}
	data, err := json.Marshal(map[string]any{
		"schema_id": embeddedAssetsKeySchemaID, "producer": e.ID(), "outputs": EmbeddedAssetNames,
		"bundle": bundle, "inputs": inputs, "settings": settings,
	})
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:]), nil
}

// Build runs the embedder on the bundle and writes its outputs to staging;
// its scratch stays outside staging.
func (e EmbeddedAssets) Build(staging string) error {
	scratch, err := os.MkdirTemp(environmentValue(e.Env, "GOTMPDIR"), "embedded-assets-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(scratch)
	output := func(index int) string { return filepath.Join(staging, EmbeddedAssetNames[index]) }
	command := exec.Command(e.Go, "run", embedderPackage,
		"--source-dir", e.Source, "--source-index", filepath.Join(e.Source, "index.html"),
		"--output", output(0), "--asset-manifest", output(1),
		"--client-support-source", filepath.Join(e.Root, filepath.FromSlash(e.ClientSupport)),
		"--client-support-registry", output(2), "--bootstrap-template", output(3),
		"--registry-integrity", output(4), "--registry-integrity-digest", output(5),
		"--temp-dir", scratch, "--repository-root", e.Root)
	command.Dir, command.Env = e.Root, e.Env
	command.Stdout, command.Stderr = os.Stderr, os.Stderr
	if err := command.Run(); err != nil {
		return fmt.Errorf("embed web assets: %w", err)
	}
	return nil
}

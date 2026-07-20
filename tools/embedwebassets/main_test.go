package main

import (
	"archive/zip"
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestWriteClientBuildContractsBindsExactStaticMapping(t *testing.T) {
	t.Parallel()

	sourceDir := t.TempDir()
	mustWriteFile(t, filepath.Join(sourceDir, "index.html"), "<html></html>")
	mustWriteFile(t, filepath.Join(sourceDir, "assets", "app.js"), "console.log('ok');")
	supportSource := filepath.Join(t.TempDir(), "client-support.json")
	mustWriteFile(t, supportSource, `{"schema_id":"cartulary.extension_client_support_registry_source.v1","client_build_class":"standard","rows":[{"profile_id":"network_flow_activity","contract_major":2,"workspace_keys":["network_analysis"],"capability_ids":[],"public_schema_ids":[],"client_asset_set_id":"network_flow_activity.standard.v2"}]}`)
	manifestPath := filepath.Join(t.TempDir(), "client-asset-set-manifest.json")
	supportPath := filepath.Join(t.TempDir(), "client-extension-support-registry.json")

	if err := writeClientBuildContracts(sourceDir, supportSource, manifestPath, supportPath); err != nil {
		t.Fatalf("write client build contracts: %v", err)
	}
	manifestBytes, err := os.ReadFile(manifestPath)
	if err != nil {
		t.Fatalf("read manifest: %v", err)
	}
	var manifest clientAssetManifest
	if err := json.Unmarshal(manifestBytes, &manifest); err != nil {
		t.Fatalf("decode manifest: %v", err)
	}
	if len(manifest.Assets) != 2 || manifest.Assets[0].LogicalPath != "assets/app.js" || manifest.Assets[1].LogicalPath != "index.html" {
		t.Fatalf("unexpected manifest rows: %#v", manifest.Assets)
	}
	supportBytes, err := os.ReadFile(supportPath)
	if err != nil {
		t.Fatalf("read client support: %v", err)
	}
	var support clientSupportRegistry
	if err := json.Unmarshal(supportBytes, &support); err != nil {
		t.Fatalf("decode client support: %v", err)
	}
	digest := sha256.Sum256(manifestBytes)
	if support.AssetSetSHA256 != hex.EncodeToString(digest[:]) || support.ClientBuildID != "cartulary.web.standard.sha256:"+support.AssetSetSHA256 {
		t.Fatalf("client support is not bound to manifest: %#v", support)
	}
	if len(support.Profiles) != 1 || support.Profiles[0].SupportedContractMajors[0] != 2 || support.Profiles[0].WorkspaceKeys[0] != "network_analysis" {
		t.Fatalf("unexpected support profiles: %#v", support.Profiles)
	}
}

func TestWriteArchiveIsDeterministicAndServesRootEntries(t *testing.T) {
	t.Parallel()

	sourceDir := t.TempDir()
	mustWriteFile(t, filepath.Join(sourceDir, "index.html"), `<script src="/assets/app.js"></script>`)
	mustWriteFile(t, filepath.Join(sourceDir, "assets", "app.js"), "console.log('ok');")

	first := filepath.Join(t.TempDir(), "web-assets.zip")
	second := filepath.Join(t.TempDir(), "web-assets.zip")
	if err := writeArchive(sourceDir, filepath.Join(sourceDir, "index.html"), first, t.TempDir()); err != nil {
		t.Fatalf("write first archive: %v", err)
	}
	if err := writeArchive(sourceDir, filepath.Join(sourceDir, "index.html"), second, t.TempDir()); err != nil {
		t.Fatalf("write second archive: %v", err)
	}

	firstBytes, err := os.ReadFile(first)
	if err != nil {
		t.Fatalf("read first archive: %v", err)
	}
	secondBytes, err := os.ReadFile(second)
	if err != nil {
		t.Fatalf("read second archive: %v", err)
	}
	if !bytes.Equal(firstBytes, secondBytes) {
		t.Fatalf("archive output is not deterministic")
	}

	reader, err := zip.NewReader(bytes.NewReader(firstBytes), int64(len(firstBytes)))
	if err != nil {
		t.Fatalf("open zip: %v", err)
	}
	indexBytes, err := reader.Open("index.html")
	if err != nil {
		t.Fatalf("open index.html from zip: %v", err)
	}
	defer indexBytes.Close()
	assetBytes, err := reader.Open("assets/app.js")
	if err != nil {
		t.Fatalf("open asset from zip: %v", err)
	}
	defer assetBytes.Close()
}

func TestWriteArchiveRejectsSymlinks(t *testing.T) {
	t.Parallel()

	sourceDir := t.TempDir()
	mustWriteFile(t, filepath.Join(sourceDir, "index.html"), "ok")
	if err := os.Symlink(filepath.Join(sourceDir, "index.html"), filepath.Join(sourceDir, "linked.html")); err != nil {
		t.Skipf("symlink unavailable: %v", err)
	}

	err := writeArchive(sourceDir, filepath.Join(sourceDir, "index.html"), filepath.Join(t.TempDir(), "web-assets.zip"), t.TempDir())
	if err == nil {
		t.Fatalf("expected symlink rejection")
	}
	if !strings.Contains(err.Error(), "must not be a symlink") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestWriteArchiveRequiresIndexInsideSource(t *testing.T) {
	t.Parallel()

	sourceDir := t.TempDir()
	otherDir := t.TempDir()
	mustWriteFile(t, filepath.Join(sourceDir, "index.html"), "ok")
	mustWriteFile(t, filepath.Join(otherDir, "index.html"), "wrong")

	err := writeArchive(sourceDir, filepath.Join(otherDir, "index.html"), filepath.Join(t.TempDir(), "web-assets.zip"), t.TempDir())
	if err == nil {
		t.Fatalf("expected external index rejection")
	}
	if !strings.Contains(err.Error(), "source index must be inside source dir") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func mustWriteFile(t *testing.T, path string, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("create parent: %v", err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write file: %v", err)
	}
}

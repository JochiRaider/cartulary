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
	t.Run("final Extensions integrity package", func(t *testing.T) {
		root, err := filepath.Abs(filepath.Join("..", ".."))
		if err != nil {
			t.Fatal(err)
		}
		source := t.TempDir()
		output := t.TempDir()
		mustWriteFile(t, filepath.Join(source, "assets", "app.js"), "console.log(1)")
		mustWriteFile(t, filepath.Join(source, "index.html"), "<html><head></head><body></body></html>")
		args := []string{"--source-dir", source, "--source-index", filepath.Join(source, "index.html"), "--output", filepath.Join(output, "assets.zip"), "--asset-manifest", filepath.Join(output, "manifest.json"), "--client-support-source", filepath.Join(root, "contracts/extensions/build/client-support.json"), "--client-support-registry", filepath.Join(output, "support.json"), "--registry-integrity", filepath.Join(output, "integrity.json"), "--registry-integrity-digest", filepath.Join(output, "integrity.sha256"), "--repository-root", root, "--bootstrap-template", filepath.Join(output, "bootstrap.json")}
		if err := run(args); err != nil {
			t.Fatal(err)
		}
		body, err := os.ReadFile(filepath.Join(output, "integrity.json"))
		if err != nil {
			t.Fatal(err)
		}
		pin, err := os.ReadFile(filepath.Join(output, "integrity.sha256"))
		if err != nil {
			t.Fatal(err)
		}
		sum := sha256.Sum256(body)
		if string(pin) != hex.EncodeToString(sum[:])+"\n" {
			t.Fatal("final root is not independently pinned")
		}
		var integrity struct {
			SchemaID  string `json:"schema_id"`
			Contracts []struct {
				ID  string `json:"artifact_id"`
				SHA string `json:"artifact_sha256"`
			} `json:"supporting_contract_artifact_digests"`
		}
		if err := json.Unmarshal(body, &integrity); err != nil {
			t.Fatal(err)
		}
		if integrity.SchemaID != "cartulary.extension_registry_integrity.v2" {
			t.Fatal("wrong final integrity generation")
		}
		for id, file := range map[string]string{"client-support-registry": "support.json", "client-asset-set-manifest": "manifest.json", "client-bootstrap-template": "bootstrap.json"} {
			raw, err := os.ReadFile(filepath.Join(output, file))
			if err != nil {
				t.Fatal(err)
			}
			sum := sha256.Sum256(raw)
			found := false
			for _, row := range integrity.Contracts {
				if row.ID == id {
					found = row.SHA == hex.EncodeToString(sum[:])
				}
			}
			if !found {
				t.Fatalf("final root does not bind %s", id)
			}
		}

		// A failed replacement must leave the previous complete container intact.
		before, err := os.ReadFile(filepath.Join(output, "assets.zip"))
		if err != nil {
			t.Fatal(err)
		}
		if err := run(args); err != nil {
			t.Fatal(err)
		}
		rebuilt, err := os.ReadFile(filepath.Join(output, "assets.zip"))
		if err != nil || !bytes.Equal(before, rebuilt) {
			t.Fatal("repeated final-package build changed archive bytes")
		}
		badArgs := append([]string(nil), args...)
		for i := range badArgs {
			if badArgs[i] == "--client-support-source" {
				badArgs[i+1] = filepath.Join(output, "missing.json")
			}
		}
		if err := run(badArgs); err == nil {
			t.Fatal("missing support source accepted")
		}
		after, err := os.ReadFile(filepath.Join(output, "assets.zip"))
		if err != nil || !bytes.Equal(before, after) {
			t.Fatal("failed build replaced the admitted archive")
		}
	})
	t.Parallel()

	sourceDir := t.TempDir()
	mustWriteFile(t, filepath.Join(sourceDir, "index.html"), "<html><head></head></html>")
	mustWriteFile(t, filepath.Join(sourceDir, "assets", "app.js"), "console.log('ok');")
	supportSource := filepath.Join(t.TempDir(), "client-support.json")
	mustWriteFile(t, supportSource, `{"schema_id":"cartulary.extension_client_support_registry_source.v1","client_build_class":"standard","rows":[{"profile_id":"network_flow_activity","contract_major":2,"workspace_keys":["network_analysis"],"capability_ids":[],"public_schema_ids":[],"client_asset_set_id":"network_flow_activity.standard.v2"}]}`)
	manifestPath := filepath.Join(t.TempDir(), "client-asset-set-manifest.json")
	supportPath := filepath.Join(t.TempDir(), "client-extension-support-registry.json")

	archivePath := filepath.Join(t.TempDir(), "assets.zip")
	if err := writeArchive(sourceDir, filepath.Join(sourceDir, "index.html"), archivePath, t.TempDir()); err != nil {
		t.Fatal(err)
	}
	bootstrapPath := filepath.Join(t.TempDir(), "bootstrap.json")
	if err := writeBootstrapTemplate(filepath.Join(sourceDir, "index.html"), bootstrapPath); err != nil {
		t.Fatal(err)
	}
	if err := writeClientBuildContracts(archivePath, bootstrapPath, supportSource, manifestPath, supportPath); err != nil {
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
	if len(manifest.Assets) != 1 || manifest.Assets[0].LogicalPath != "assets/app.js" {
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
	if support.AssetSetSHA256 != hex.EncodeToString(digest[:]) || !strings.HasPrefix(support.ClientBuildID, "cartulary.web.standard.sha256:") {
		t.Fatalf("client support is not bound to manifest: %#v", support)
	}
	if len(support.Profiles) != 1 || support.Profiles[0].SupportedContractMajors[0] != 2 || support.Profiles[0].WorkspaceKeys[0] != "network_analysis" {
		t.Fatalf("unexpected support profiles: %#v", support.Profiles)
	}
	firstBuildID := support.ClientBuildID
	mustWriteFile(t, filepath.Join(sourceDir, "index.html"), "<html><head><title>changed</title></head></html>")
	if err := writeBootstrapTemplate(filepath.Join(sourceDir, "index.html"), bootstrapPath); err != nil {
		t.Fatal(err)
	}
	if err := writeClientBuildContracts(archivePath, bootstrapPath, supportSource, manifestPath, supportPath); err != nil {
		t.Fatal(err)
	}
	changed, err := os.ReadFile(supportPath)
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(changed, &support); err != nil {
		t.Fatal(err)
	}
	if support.ClientBuildID == firstBuildID || support.AssetSetSHA256 != hex.EncodeToString(digest[:]) {
		t.Fatal("template-only change did not receive a distinct build identity")
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
	if _, err := reader.Open("index.html"); err == nil {
		t.Fatal("dynamic bootstrap entered static mapping")
	}
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

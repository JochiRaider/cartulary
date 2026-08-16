package webassets

import (
	"archive/zip"
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io/fs"
	"slices"
	"strings"
	"testing"
	"testing/fstest"
)

func TestReadBrowserRootHTMLValidatesAndInjectsBuildBoundSupport(t *testing.T) {
	t.Parallel()

	archive := testArchive(t, map[string]string{
		"index.html":    "<html><head></head><body></body></html>",
		"assets/app.js": "console.log('ok');",
	})
	manifest, support := testClientContracts(t, map[string]string{
		"assets/app.js": "console.log('ok');",
		"index.html":    "<html><head></head><body></body></html>",
	})
	files := fstest.MapFS{
		"fallback/index.html":     {Data: []byte("fallback")},
		distArchivePath:           {Data: archive},
		clientAssetManifestPath:   {Data: manifest},
		clientSupportRegistryPath: {Data: support},
	}

	root, err := readBrowserRootHTML(files)
	if err != nil {
		t.Fatalf("readBrowserRootHTML(): %v", err)
	}
	if !strings.Contains(string(root), `id="cartulary-client-extension-support-registry"`) || !strings.Contains(string(root), `"asset_set_sha256"`) {
		t.Fatalf("browser bootstrap missing support registry: %s", root)
	}
}

func TestReadBrowserRootHTMLRejectsStaticMappingMismatch(t *testing.T) {
	t.Parallel()

	manifest, support := testClientContracts(t, map[string]string{"index.html": "<html><head></head></html>"})
	files := fstest.MapFS{
		"fallback/index.html":     {Data: []byte("fallback")},
		distArchivePath:           {Data: testArchive(t, map[string]string{"index.html": "<html><head></head></html>", "assets/extra.js": "extra"})},
		clientAssetManifestPath:   {Data: manifest},
		clientSupportRegistryPath: {Data: support},
	}
	if _, err := readBrowserRootHTML(files); err == nil || !strings.Contains(err.Error(), "omits") {
		t.Fatalf("expected extra static asset rejection, got %v", err)
	}
}

func TestReadIndexHTMLUsesArchiveWhenPresent(t *testing.T) {
	t.Parallel()

	files := fstest.MapFS{
		"fallback/index.html": {Data: []byte("fallback")},
		distArchivePath:       {Data: testArchive(t, map[string]string{"index.html": "dist"})},
	}

	got, err := readIndexHTML(files)
	if err != nil {
		t.Fatalf("readIndexHTML(): %v", err)
	}
	if string(got) != "dist" {
		t.Fatalf("readIndexHTML()=%q, want dist", string(got))
	}
}

func TestStaticFSUsesArchiveWhenPresent(t *testing.T) {
	t.Parallel()

	files := fstest.MapFS{
		"fallback/index.html": {Data: []byte("fallback")},
		distArchivePath: {
			Data: testArchive(t, map[string]string{
				"index.html":    "dist",
				"assets/app.js": "console.log('ok');",
			}),
		},
	}

	static, err := staticFS(files)
	if err != nil {
		t.Fatalf("staticFS(): %v", err)
	}
	got, err := fs.ReadFile(static, "assets/app.js")
	if err != nil {
		t.Fatalf("read static asset: %v", err)
	}
	if string(got) != "console.log('ok');" {
		t.Fatalf("asset=%q", string(got))
	}
}

func TestReadIndexHTMLFallsBackWithoutArchive(t *testing.T) {
	t.Parallel()

	files := fstest.MapFS{
		"fallback/index.html": {Data: []byte("fallback")},
		"dist/index.html":     {Data: []byte("legacy loose dist is ignored")},
	}

	got, err := readIndexHTML(files)
	if err != nil {
		t.Fatalf("readIndexHTML(): %v", err)
	}
	if string(got) != "fallback" {
		t.Fatalf("readIndexHTML()=%q, want fallback", string(got))
	}
}

func testArchive(t *testing.T, entries map[string]string) []byte {
	t.Helper()

	var buf bytes.Buffer
	writer := zip.NewWriter(&buf)
	for name, content := range entries {
		file, err := writer.Create(name)
		if err != nil {
			t.Fatalf("create zip entry: %v", err)
		}
		if _, err := file.Write([]byte(content)); err != nil {
			t.Fatalf("write zip entry: %v", err)
		}
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("close zip: %v", err)
	}
	return buf.Bytes()
}

func testClientContracts(t *testing.T, assets map[string]string) ([]byte, []byte) {
	t.Helper()
	paths := make([]string, 0, len(assets))
	for assetPath := range assets {
		paths = append(paths, assetPath)
	}
	slices.Sort(paths)
	rows := make([]clientAssetManifestRow, 0, len(paths))
	for _, assetPath := range paths {
		data := []byte(assets[assetPath])
		digest := sha256.Sum256(data)
		rows = append(rows, clientAssetManifestRow{ByteLength: int64(len(data)), LogicalPath: assetPath, SHA256: hex.EncodeToString(digest[:])})
	}
	manifest, err := json.Marshal(clientAssetManifest{Assets: rows, SchemaID: "cartulary.client_asset_set_manifest.v1"})
	if err != nil {
		t.Fatal(err)
	}
	manifest = append(manifest, '\n')
	manifestDigest := sha256.Sum256(manifest)
	digestHex := hex.EncodeToString(manifestDigest[:])
	support, err := json.Marshal(clientSupportRegistry{
		AssetSetSHA256:   digestHex,
		ClientBuildClass: "standard",
		ClientBuildID:    "cartulary.web.standard.sha256:" + digestHex,
		Profiles: []clientSupportRegistryProfile{{
			CapabilityIDs: []string{}, ProfileID: "network_flow_activity", PublicSchemaIDs: []string{}, SupportedContractMajors: []int64{3}, WorkspaceKeys: []string{"network_analysis"},
		}},
		SchemaID: "cartulary.client_extension_support_registry.v1",
	})
	if err != nil {
		t.Fatal(err)
	}
	return manifest, append(support, '\n')
}

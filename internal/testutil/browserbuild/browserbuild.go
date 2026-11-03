// Package browserbuild provides explicit current browser packages for tests
// that exercise backend composition without executing browser JavaScript.
package browserbuild

import (
	"archive/zip"
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	contractsgen "github.com/JochiRaider/cartulary/internal/gen/contractextensions"
	"github.com/JochiRaider/cartulary/internal/platform/canonicaljson"
	"github.com/JochiRaider/cartulary/internal/platform/httpapi/webassets"
	"sort"
	"strings"
	"testing"
	"testing/fstest"
)

func New(t testing.TB) *webassets.Build {
	t.Helper()
	b, err := webassets.Load(Files(t))
	if err != nil {
		t.Fatal(err)
	}
	return b
}
func Files(t testing.TB) fstest.MapFS {
	t.Helper()
	canonical := func(v any) []byte {
		b, err := canonicaljson.MarshalExtension(v)
		if err != nil {
			t.Fatal(err)
		}
		return b
	}
	decode := func(v string) map[string]any {
		b, err := canonicaljson.DecodeExtension([]byte(v))
		if err != nil {
			t.Fatal(err)
		}
		return b.(map[string]any)
	}
	asset := []byte("console.log('fixture');")
	manifest := canonical(map[string]any{"schema_id": "cartulary.client_asset_set_manifest.v1", "assets": []any{map[string]any{"logical_path": "assets/app.js", "byte_length": len(asset), "sha256": SHA(asset)}}})
	template := canonical(map[string]any{"schema_id": "cartulary.client_bootstrap_template.v1", "html_utf8": `<!doctype html><html><head><script src="/assets/app.js"></script></head><body></body></html>`})
	const prefix = "contracts/extensions/generated/"
	support := decode(contractsgen.Index[prefix+"client-support-registry.json"].JSON)
	support["asset_set_sha256"] = SHA(manifest)
	support["client_build_id"] = "cartulary.web.standard.sha256:" + SHA([]byte(SHA(manifest)+"\n"+SHA(template)+"\n"))
	supportBytes := canonical(support)
	root := decode(contractsgen.Index[prefix+"registry-integrity.json"].JSON)
	root["generator_id"] = "cartulary.embedwebassets.extensions.v1"
	rows := root["supporting_contract_artifact_digests"].([]any)
	for _, raw := range rows {
		row := raw.(map[string]any)
		if row["artifact_id"] == "client-support-registry" {
			row["artifact_sha256"] = SHA(supportBytes)
		}
	}
	for id, value := range map[string][]byte{"client-asset-set-manifest": manifest, "client-bootstrap-template": template} {
		schema := "cartulary.client_asset_set_manifest.v1"
		if id == "client-bootstrap-template" {
			schema = "cartulary.client_bootstrap_template.v1"
		}
		rows = append(rows, map[string]any{"artifact_id": id, "schema_id": schema, "artifact_sha256": SHA(value)})
	}
	sort.Slice(rows, func(i, j int) bool {
		return rows[i].(map[string]any)["artifact_id"].(string) < rows[j].(map[string]any)["artifact_id"].(string)
	})
	root["supporting_contract_artifact_digests"] = rows
	rootBytes := canonical(root)
	var archive bytes.Buffer
	writer := zip.NewWriter(&archive)
	f, err := writer.Create("assets/app.js")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := f.Write(asset); err != nil {
		t.Fatal(err)
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	const dir = "assets/generated/"
	files := fstest.MapFS{dir + "web-assets.zip": {Data: archive.Bytes()}, dir + "client-asset-set-manifest.json": {Data: manifest}, dir + "client-extension-support-registry.json": {Data: supportBytes}, dir + "client-bootstrap-template.json": {Data: template}, dir + "extension-registry-integrity.json": {Data: rootBytes}, dir + "extension-registry-integrity.sha256": {Data: []byte(SHA(rootBytes) + "\n")}}
	Seal(t, files)
	return files
}
func SHA(body []byte) string { sum := sha256.Sum256(body); return hex.EncodeToString(sum[:]) }

// Seal joins the explicit fixture sidecars into the single runtime container.
// The sidecars remain mutable only to make negative fixture construction clear.
func Seal(t testing.TB, files fstest.MapFS) {
	t.Helper()
	const dir = "assets/generated/"
	old := files[dir+"web-assets.zip"].Data
	reader, err := zip.NewReader(bytes.NewReader(old), int64(len(old)))
	if err != nil {
		t.Fatal(err)
	}
	var buffer bytes.Buffer
	writer := zip.NewWriter(&buffer)
	for _, file := range reader.File {
		if strings.HasPrefix(file.Name, "_contracts/") {
			continue
		}
		if err := writer.Copy(file); err != nil {
			t.Fatal(err)
		}
	}
	names := []string{"client-asset-set-manifest.json", "client-extension-support-registry.json", "client-bootstrap-template.json", "extension-registry-integrity.json", "extension-registry-integrity.sha256"}
	for _, name := range names {
		body, ok := files[dir+name]
		if !ok {
			continue
		}
		file, err := writer.Create("_contracts/" + name)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := file.Write(body.Data); err != nil {
			t.Fatal(err)
		}
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	files[dir+"web-assets.zip"] = &fstest.MapFile{Data: buffer.Bytes()}
}

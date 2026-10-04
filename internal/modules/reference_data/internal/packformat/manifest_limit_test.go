package packformat

import (
	"archive/zip"
	"bytes"
	"context"
	"fmt"
	"slices"
	"strings"
	"testing"

	"github.com/JochiRaider/cartulary/internal/platform/canonicaljson"
)

func testCompleteManifestAndPathLimits(t *testing.T) {
	var fixture contentFixture
	for _, f := range contentFixtures(t) {
		if f.Key == "framework.attack" {
			fixture = f
			break
		}
	}
	m, err := DecodeManifest([]byte(manifestFixture(t).Canonical), true)
	if err != nil {
		t.Fatal(err)
	}
	p := profiles[fixture.Key]
	m.Key, m.Kind, m.ProfileID, m.ProfileVersion = p.Key, p.Kind, p.ID, p.Version
	m.Repository = nil
	m.Summary = fixtureManifest(fixture).Summary
	m.Files = []File{}
	m.Artifacts = []SourceArtifact{}
	for i := 0; i < 64; i++ {
		m.Artifacts = append(m.Artifacts, SourceArtifact{Ref: fmt.Sprintf("fixture:%02d", i), Version: "1", SHA256: strings.Repeat("0", 64)})
	}
	m.Dependencies, m.Conflicts = []Dependency{}, []Conflict{}
	for i := 0; i < 64; i++ {
		version := fmt.Sprintf("dep%02d", i)
		m.Dependencies = append(m.Dependencies, Dependency{Key: "enrichment.tor", Version: version, SHA256: strings.Repeat("0", 64)})
		version = fmt.Sprintf("conflict%02d", i)
		m.Conflicts = append(m.Conflicts, Conflict{Key: "enrichment.lolbas", Version: &version})
	}
	m.License.Expression = "MIT"
	m.License.Notices, m.License.Bindings = []string{}, []LicenseBinding{}
	members := contentMemory{}
	for path, data := range fixture.Members {
		members[path] = data
		m.Files = append(m.Files, File{Path: path, Role: "payload", MediaType: "application/x-ndjson", Size: int64(len(data)), SHA256: Digest([]byte(data))})
	}
	// The 1024-byte ASCII path contains three segments at the 255-byte bound.
	longest := "notices/" + strings.Repeat("a", 255) + "/" + strings.Repeat("b", 255) + "/" + strings.Repeat("c", 255) + "/" + strings.Repeat("d", 248)
	if len(longest) != 1024 {
		t.Fatal("path recipe length")
	}
	for i := 0; i < 64; i++ {
		path := fmt.Sprintf("notices/%02d.txt", i)
		if i == 63 {
			path = longest
		}
		members[path] = "Synthetic boundary notice.\n"
		m.License.Notices = append(m.License.Notices, path)
		m.Files = append(m.Files, File{Path: path, Role: "notice", MediaType: "text/plain", Size: int64(len(members[path])), SHA256: Digest([]byte(members[path]))})
	}
	slices.Sort(m.License.Notices)
	slices.SortFunc(m.Files, func(a, b File) int { return strings.Compare(a.Path, b.Path) })
	raw, err := canonicaljson.Marshal(m)
	if err != nil {
		t.Fatal(err)
	}
	admitted, err := DecodeManifest(raw, false)
	if err != nil || len(admitted.Files) != 66 || len(admitted.Artifacts) != 64 {
		t.Fatal("complete manifest boundary", err)
	}
	if err := VerifyDependencies(context.Background(), admitted, func(_ context.Context, dependency Dependency) (Manifest, bool, error) {
		// Authenticated retention is an explicit input to this pure boundary.
		return Manifest{Key: dependency.Key, Version: dependency.Version}, true, nil
	}, newDiagnosticTestScratch()); err != nil {
		t.Fatal("at-limit dependency graph", err)
	}
	members["manifest.json"] = string(raw)
	var container bytes.Buffer
	writer := zip.NewWriter(&container)
	for _, path := range sortedTrustKeys(members) {
		member, err := writer.CreateHeader(&zip.FileHeader{Name: path, Method: zip.Store})
		if err != nil {
			t.Fatal(err)
		}
		if _, err := member.Write([]byte(members[path])); err != nil {
			t.Fatal(err)
		}
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	inventory, err := Extract(context.Background(), bytes.NewReader(container.Bytes()), int64(container.Len()), DefaultArchiveLimits(), &countingDestination{})
	if err != nil {
		t.Fatal("complete at-limit archive", err)
	}
	if err := ValidateInventory(admitted, inventory, false); err != nil {
		t.Fatal(err)
	}
	if err := ValidateContent(context.Background(), admitted, members, nil); err != nil {
		t.Fatal(err)
	}
	for _, path := range admitted.License.Notices {
		if err := ValidateNotice([]byte(members[path])); err != nil {
			t.Fatal(err)
		}
	}
	for _, field := range []string{"files", "source_artifacts", "dependencies", "conflicts"} {
		value, err := canonicaljson.DecodeStrict(raw)
		if err != nil {
			t.Fatal(err)
		}
		node := value.(map[string]any)
		values := node[field].([]any)
		node[field] = append(values, values[0])
		over, err := canonicaljson.Marshal(node)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := DecodeManifest(over, false); err == nil {
			t.Fatal("one-over manifest collection admitted", field)
		}
	}
	t.Run("complete compression ratio equality", func(t *testing.T) { testCompleteCompressionRatio(t, admitted, members) })
	for _, path := range []string{longest + "d", "notices/" + strings.Repeat("a", 256)} {
		if err := ValidatePath(path, false); err == nil {
			t.Fatal("one-over path admitted")
		}
	}
}

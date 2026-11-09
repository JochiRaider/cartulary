package reference_data

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/JochiRaider/cartulary/internal/modules/reference_data/internal/packformat"
	"github.com/JochiRaider/cartulary/internal/testutil/referencepackfixture"
	"github.com/JochiRaider/cartulary/internal/testutil/suiteservices"
	"github.com/JochiRaider/cartulary/referencepacks/format/canonicaljson"
	"github.com/JochiRaider/cartulary/referencepacks/format/dataset"
	"github.com/JochiRaider/cartulary/referencepacks/format/publication"
)

// Source-specific ATT&CK endpoint rules live in its producer. This verifies the
// runtime's generic local-reference boundary using an unregistered advisory key.
func TestGenericSignedEndpointSemantics_Unit(t *testing.T) {
	keys, err := referencepackfixture.NewSigningContext("fixture.endpoint.dataset.v1", "advisory.endpoint_fixture")
	if err != nil {
		t.Fatal(err)
	}
	root := "../../.."
	raw, err := os.ReadFile(filepath.Join(root, "fixtures/reference-packs/dataset-v1/catalog.json"))
	if err != nil {
		t.Fatal(err)
	}
	var catalog map[string]struct {
		Root string `json:"logical_root"`
	}
	if err := json.Unmarshal(raw, &catalog); err != nil {
		t.Fatal(err)
	}
	logical := filepath.Join(root, catalog["windows"].Root)
	encode := func(v any) []byte {
		t.Helper()
		b, err := canonicaljson.Marshal(v)
		if err != nil {
			t.Fatal(err)
		}
		return b
	}
	read := func(name string) []byte {
		t.Helper()
		b, err := os.ReadFile(filepath.Join(logical, name))
		if err != nil {
			t.Fatal(err)
		}
		return b
	}
	artifactRoot, err := suiteservices.ResolveResultsRoot(nil)
	if err != nil {
		t.Fatal("generic endpoint fixture requires Harness artifacts", err)
	}
	artifacts := filepath.Join(artifactRoot, suiteservices.ResolveRunID(nil))
	for _, mode := range []string{"valid", "dangling source", "dangling target", "duplicate relationship", "replacement cycle"} {
		t.Run(mode, func(t *testing.T) {
			m, err := packformat.DecodeManifest(read("manifest.json"), true)
			if err != nil {
				t.Fatal(err)
			}
			repository := keys.Repository()
			m.Key, m.Repository = "advisory.endpoint_fixture", &repository
			members := map[string][]byte{}
			for _, f := range m.Files {
				members[f.Path] = read(f.Path)
			}
			lines := bytes.Split(bytes.TrimSuffix(members[dataset.ObjectsPath], []byte{'\n'}), []byte{'\n'})
			var first, second map[string]any
			if err := json.Unmarshal(lines[0], &first); err != nil {
				t.Fatal(err)
			}
			if err := json.Unmarshal(lines[1], &second); err != nil {
				t.Fatal(err)
			}
			from, to := first["object_id"].(string), second["object_id"].(string)
			// Deliberately inconsistent domain content is inert, never resolved or run.
			first["data"] = map[string]any{"source": map[string]any{"type": "unrecognized", "target_ref": "not-an-object"}, "remote_schema": "https://never-fetch.invalid/schema"}
			switch mode {
			case "dangling source":
				from = "absent"
			case "dangling target":
				to = "absent"
			case "replacement cycle":
				first["deprecated"], second["deprecated"] = true, true
				first["replacement_object_id"], second["replacement_object_id"] = to, from
			}
			lines[0], lines[1] = encode(first), encode(second)
			members[dataset.ObjectsPath] = append(bytes.Join(lines, []byte{'\n'}), '\n')
			relationship := map[string]any{"schema_id": dataset.RelationshipSchema, "relationship_id": "edge:1", "relationship_type": "example", "source_object_id": from, "target_object_id": to, "source_refs": first["source_refs"], "data": map[string]any{}}
			members[dataset.RelationshipsPath] = append(encode(relationship), '\n')
			count := int64(1)
			if mode == "duplicate relationship" {
				members[dataset.RelationshipsPath] = append(members[dataset.RelationshipsPath], members[dataset.RelationshipsPath]...)
				count = 2
			}
			m.Summary.Relationships = &count
			m.Files = publication.LogicalMemberInventory(members)
			members["manifest.json"] = encode(m)
			now := time.Now().UTC()
			container, bootstrap, err := referencepackfixture.Envelope(t.Context(), members, keys, now.Add(365*24*time.Hour).Format(time.RFC3339), now.Add(24*time.Hour).Format(time.RFC3339), 1)
			if err != nil {
				t.Fatal(err)
			}
			trust, err := packformat.DecodeBootstrap(bootstrap)
			if err != nil {
				t.Fatal(err)
			}
			if err := publication.PublishTree(artifacts, "generic-endpoints/"+packformat.Digest(container), map[string][]byte{"container.pack": container, "bootstrap.json": bootstrap}); err != nil {
				t.Fatal(err)
			}
			storage := &engineStorage{container: container}
			ref, _ := ParseStagingRef("staged/endpoint.pack")
			attempt := verificationAttempt{Staged: &ref, ContainerSHA256: packformat.Digest(container), Start: now, ClockTrusted: true, Identity: emptyIdentityHistory{}, Limits: packformat.DefaultArchiveLimits(), Repositories: trust}
			indexed := false
			result, err := verifyTestCanonicalContainer(t.Context(), storage, attempt, func(context.Context, packformat.Manifest, string, string) (packformat.ContentSink, error) {
				indexed = true
				return &builtinRows{}, nil
			})
			if mode == "valid" {
				if err != nil || result == nil || !indexed {
					t.Fatal("valid generic local endpoints rejected", err)
				}
				if err := result.Close(); err != nil {
					t.Fatal(err)
				}
			} else {
				var rejection *ContentRejection
				if result != nil || !errors.As(err, &rejection) || rejection.Code != "content_semantic_invalid" || rejection.CheckID != "content_semantics" || indexed {
					t.Fatalf("invalid generic references reached indexing: %v indexed=%v", err, indexed)
				}
			}
			if !storage.closed {
				t.Fatal("verification workspace leaked")
			}
		})
	}
}

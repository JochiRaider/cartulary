package reference_data

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/JochiRaider/cartulary/internal/modules/reference_data/internal/packformat"
	"github.com/JochiRaider/cartulary/internal/modules/reference_data/producer"
	"github.com/JochiRaider/cartulary/internal/testutil/suiteservices"
)

func TestNativeATTACKSignedEndpointSemantics_Unit(t *testing.T) {
	keys, err := producer.NewFixtureSigningContext()
	if err != nil {
		t.Fatal(err)
	}
	artifactRoot, err := suiteservices.ResolveResultsRoot(nil)
	if err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"matrix", "membership", "strategy", "analytic"} {
		for _, verdict := range []string{"valid", "invalid"} {
			t.Run(name+"/"+verdict, func(t *testing.T) {
				root := "../../.."
				path := filepath.Join(root, "contracts/reference-pack-fixtures/fixtures/native_endpoint_"+name+"_"+verdict+".v1.json")
				b, err := os.ReadFile(path)
				if err != nil {
					t.Fatal(err)
				}
				var descriptor struct {
					Manifest string            `json:"manifest"`
					Members  map[string]string `json:"members"`
				}
				if err := json.Unmarshal(b, &descriptor); err != nil {
					t.Fatal(err)
				}
				descriptor.Members["manifest.json"] = descriptor.Manifest
				members := map[string][]byte{}
				for name, path := range descriptor.Members {
					members[name], err = os.ReadFile(filepath.Join(root, path))
					if err != nil {
						t.Fatal(err)
					}
				}
				now := time.Now().UTC()
				container, bootstrap, err := producer.FixtureEnvelope(t.Context(), members, keys, now.Add(365*24*time.Hour).Format(time.RFC3339), now.Add(24*time.Hour).Format(time.RFC3339), 1)
				if err != nil {
					t.Fatal(err)
				}
				trust, err := packformat.DecodeBootstrap(bootstrap)
				if err != nil {
					t.Fatal(err)
				}
				retained := filepath.Join(artifactRoot, suiteservices.ResolveRunID(nil), "reference-pack-endpoint-envelopes", packformat.Digest(container))
				if err := os.MkdirAll(retained, 0700); err != nil {
					t.Fatal(err)
				}
				output, err := os.OpenRoot(retained)
				if err != nil {
					t.Fatal(err)
				}
				defer output.Close()
				for path, bytes := range map[string][]byte{"container.pack": container, "bootstrap.json": bootstrap} {
					if err := producer.Retain(output, path, bytes, 0600); err != nil {
						t.Fatal(err)
					}
				}
				storage := &engineStorage{container: container}
				ref, _ := ParseStagingRef("staged/endpoint.pack")
				attempt := verificationAttempt{Staged: &ref, ContainerSHA256: packformat.Digest(container), Start: time.Now().UTC(), ClockTrusted: true, Identity: emptyIdentityHistory{}, Limits: packformat.DefaultArchiveLimits(), Repositories: trust}
				published := false
				result, err := verifyCanonicalContainer(t.Context(), storage, attempt, func(context.Context, packformat.Manifest, string, string) (packformat.ContentSink, error) {
					published = true
					return &builtinRows{}, nil
				})
				if verdict == "valid" {
					if err != nil || result == nil || !published {
						t.Fatal("valid signed endpoint", err)
					}
					if err := result.Close(); err != nil {
						t.Fatal(err)
					}
				} else {
					var rejection *ContentRejection
					if result != nil || !errors.As(err, &rejection) || rejection.Code != "content_semantic_invalid" || rejection.CheckID != "content_semantics" || published {
						t.Fatalf("semantic boundary bypass: %v published=%v", err, published)
					}
				}
				if !storage.closed {
					t.Fatal("verification workspace leaked")
				}
			})
		}
	}
}

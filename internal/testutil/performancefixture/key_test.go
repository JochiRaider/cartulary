package performancefixture

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/JochiRaider/cartulary/internal/gen/performancefixtureprofile"
)

type keyVectors struct {
	SchemaID string `json:"schema_id"`
	Vectors  []struct {
		Name          string           `json:"name"`
		Input         SnapshotKeyInput `json:"input"`
		CanonicalJSON string           `json:"canonical_json"`
		SnapshotKey   string           `json:"snapshot_key"`
	} `json:"vectors"`
}

func TestSnapshotKeyMatchesCrossLanguageVectors(t *testing.T) {
	t.Parallel()
	body, err := os.ReadFile(filepath.Join("..", "..", "..", "tools", "performance_fixture_snapshot_key_vectors.json"))
	if err != nil {
		t.Fatal(err)
	}
	var fixture keyVectors
	if err := json.Unmarshal(body, &fixture); err != nil {
		t.Fatal(err)
	}
	if fixture.SchemaID != "cartulary.performance_fixture_snapshot_key_vectors.v1" || len(fixture.Vectors) == 0 {
		t.Fatal("snapshot key vector fixture is invalid")
	}
	for _, vector := range fixture.Vectors {
		vector := vector
		t.Run(vector.Name, func(t *testing.T) {
			profile, ok := performancefixtureprofile.Lookup(vector.Name)
			if !ok {
				t.Fatalf("generated fixture profile %q is missing", vector.Name)
			}
			profile.SourceContractDigest = vector.Input.SourceContractDigest
			profile.FixtureVersion = vector.Input.FixtureVersion
			profile.Seed = vector.Input.Seed
			profile.ArtifactPolicy.SnapshotKeySchemaID = vector.Input.SchemaID
			if vector.Input.FixtureProfileID != "" {
				profile.FixtureProfileID = vector.Input.FixtureProfileID
			}
			canonical, canonicalErr := CanonicalSnapshotKeyInput(profile, vector.Input.MigrationDigest)
			if canonicalErr != nil {
				t.Fatal(canonicalErr)
			}
			if string(canonical) != vector.CanonicalJSON {
				t.Fatalf("canonical input mismatch:\n got %s\nwant %s", canonical, vector.CanonicalJSON)
			}
			key, keyErr := SnapshotKey(profile, vector.Input.MigrationDigest)
			if keyErr != nil {
				t.Fatal(keyErr)
			}
			if key != vector.SnapshotKey {
				t.Fatalf("snapshot key = %s, want %s", key, vector.SnapshotKey)
			}
		})
	}
}

func TestSnapshotKeyRejectsNonCanonicalOrUnsupportedInput(t *testing.T) {
	t.Parallel()
	valid := activeTestProfile(t)
	cases := []performancefixtureprofile.Profile{
		func() performancefixtureprofile.Profile { value := valid; value.Status = "inactive"; return value }(),
		func() performancefixtureprofile.Profile {
			value := valid
			value.FixtureProfileID = "invalid"
			return value
		}(),
		func() performancefixtureprofile.Profile {
			value := valid
			value.FixtureVersion = "invalid"
			return value
		}(),
		func() performancefixtureprofile.Profile { value := valid; value.Seed = -1; return value }(),
		func() performancefixtureprofile.Profile {
			value := valid
			value.SourceContractDigest = strings.Repeat("F", 64)
			return value
		}(),
		func() performancefixtureprofile.Profile {
			value := valid
			value.ArtifactPolicy.SnapshotKeySchemaID += ".next"
			return value
		}(),
	}
	for _, profile := range cases {
		if _, err := SnapshotKey(profile, strings.Repeat("e", 64)); err == nil {
			t.Fatalf("SnapshotKey(%+v) succeeded, want error", profile)
		}
	}
	if _, err := SnapshotKey(valid, "sha256:"+strings.Repeat("e", 64)); err == nil {
		t.Fatal("SnapshotKey accepted a prefixed migration digest")
	}
}

func TestSnapshotKeyIsProfileExplicit(t *testing.T) {
	t.Parallel()
	profile := activeTestProfile(t)
	profile.FixtureProfileID = "synthetic_grid_snapshot_v1"
	profile.FixtureVersion = "cartulary.perf.synthetic_grid.v1"
	profile.Seed++
	profile.SourceContractDigest = strings.Repeat("a", 64)
	profile.ArtifactPolicy.SnapshotKeySchemaID = SnapshotKeySchemaID
	canonical, err := CanonicalSnapshotKeyInput(profile, strings.Repeat("e", 64))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(canonical), `"fixture_profile_id":"synthetic_grid_snapshot_v1"`) {
		t.Fatalf("v2 canonical key input is not profile-explicit: %s", canonical)
	}
}

func TestGeneratedProfileLookupReturnsDefensiveCopies(t *testing.T) {
	t.Parallel()
	profile := activeTestProfile(t)
	profileID := profile.FixtureProfileID
	profile.Contributions[0].ExpectedReceiptCounts[0].Exact = -1
	profile.VerificationBindings[0].PredicateID = "mutated"
	profile.RedactionPolicy.ForbiddenFields[0] = "mutated"
	again, ok := performancefixtureprofile.Lookup(profileID)
	if !ok {
		t.Fatal("generated large-grid fixture profile disappeared")
	}
	if again.Contributions[0].ExpectedReceiptCounts[0].Exact < 0 ||
		again.VerificationBindings[0].PredicateID == "mutated" ||
		again.RedactionPolicy.ForbiddenFields[0] == "mutated" {
		t.Fatal("generated fixture profile lookup leaked mutable catalog state")
	}
}

func activeTestProfile(t *testing.T) performancefixtureprofile.Profile {
	t.Helper()
	profiles := performancefixtureprofile.Profiles()
	for _, profile := range profiles {
		if profile.Status == "active" {
			return profile
		}
	}
	t.Fatal("generated active fixture profile is missing")
	return performancefixtureprofile.Profile{}
}

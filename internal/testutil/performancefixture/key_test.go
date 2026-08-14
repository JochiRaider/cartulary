package performancefixture

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
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
			canonical, canonicalErr := CanonicalSnapshotKeyInput(vector.Input)
			if canonicalErr != nil {
				t.Fatal(canonicalErr)
			}
			if string(canonical) != vector.CanonicalJSON {
				t.Fatalf("canonical input mismatch:\n got %s\nwant %s", canonical, vector.CanonicalJSON)
			}
			key, keyErr := SnapshotKey(vector.Input)
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
	valid := SnapshotKeyInput{
		SchemaID:             SnapshotKeySchemaID,
		MigrationDigest:      strings.Repeat("e", 64),
		SourceContractDigest: strings.Repeat("f", 64),
		FixtureVersion:       LargeGridVersion,
		Seed:                 LargeGridSeed,
	}
	cases := []SnapshotKeyInput{
		func() SnapshotKeyInput { value := valid; value.SchemaID += ".next"; return value }(),
		func() SnapshotKeyInput {
			value := valid
			value.MigrationDigest = "sha256:" + value.MigrationDigest
			return value
		}(),
		func() SnapshotKeyInput {
			value := valid
			value.SourceContractDigest = strings.Repeat("F", 64)
			return value
		}(),
		func() SnapshotKeyInput { value := valid; value.FixtureVersion += ".next"; return value }(),
		func() SnapshotKeyInput { value := valid; value.Seed++; return value }(),
	}
	for _, input := range cases {
		if _, err := SnapshotKey(input); err == nil {
			t.Fatalf("SnapshotKey(%+v) succeeded, want error", input)
		}
	}
}

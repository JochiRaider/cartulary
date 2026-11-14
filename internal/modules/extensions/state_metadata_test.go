package extensions

import (
	"fmt"
	"testing"
	"time"
)

// These literal SHA-256 vectors cover the sorted, six-member canonical metadata
// object with a final LF. They are not computed by the encoder under test.
var syntheticMetadataDigests = map[int]string{
	1: "ee98cba8e3fafc792e329ac1fe21949ddfd0f860801f65144a2e7bcacbda2594",
	2: "3ccd112fdd7f7ef1299ca68aa4942d2243a3b2e3b20edc892aa5e5feb06b8ba4",
	3: "fffd5d9f8c28baa3f0a75ef7b8d4d817791ca737f580f0b68fda1a671cf2c2c3",
}

func TestStateMetadataCanonicalIdentity_Unit(t *testing.T) {
	for version, want := range syntheticMetadataDigests {
		t.Run(fmt.Sprint(version), func(t *testing.T) {
			var lastMigration *string
			if version > 1 {
				id := fmt.Sprintf("test_profile.migrate_%d_%d_v1", version-1, version)
				lastMigration = &id
			}
			metadata := StateMetadata{
				ProfileID: "test_profile", MigrationLineageID: "test_profile.state_v1",
				StateVersion: version, MetadataVersion: version, LastMigrationID: lastMigration,
			}
			if got := stateMetadataSHA256(metadata); got != want {
				t.Fatalf("canonical metadata digest = %s, want %s", got, want)
			}
			metadata.CreatedAt = time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)
			metadata.UpdatedAt = metadata.CreatedAt.Add(time.Hour)
			if got := stateMetadataSHA256(metadata); got != want {
				t.Fatalf("persistence timestamps changed canonical identity: %s", got)
			}
		})
	}
}

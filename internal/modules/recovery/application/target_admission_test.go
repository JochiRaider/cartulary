package application

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/platform/objectstore"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
)

func TestRestoreTargetMarkerV2Admission_Unit(t *testing.T) {
	now := time.Date(2026, 7, 29, 17, 30, 0, 0, time.UTC)
	generationID := uuid.MustParse("00000000-0000-0000-0000-000000005001")
	expected := TargetBindingDigests{
		DatabaseSHA256:    strings.Repeat("1", 64),
		ObjectStoreSHA256: strings.Repeat("2", 64),
	}
	validMarker := RestoreTargetMarker{
		SchemaID:           RestoreTargetMarkerSchemaID,
		Purpose:            RestoreVerificationTargetPurpose,
		TargetGenerationID: generationID.String(),
		BindingDigests:     expected,
		IssuedAt:           now.Add(-time.Minute).Format(time.RFC3339Nano),
		ExpiresAt:          now.Add(time.Hour).Format(time.RFC3339Nano),
	}
	validMaterial := markerMaterialForTest(t, validMarker, generationID)
	if err := ValidateRestoreTargetMarker(validMaterial, RestoreVerificationTargetPurpose, expected, now); err != nil {
		t.Fatalf("valid marker rejected: %v", err)
	}

	tests := []struct {
		name     string
		material TargetMarkerMaterial
		purpose  string
		expected TargetBindingDigests
	}{
		{"wrong purpose", validMaterial, RestoreTargetPurpose, expected},
		{"wrong database binding", validMaterial, RestoreVerificationTargetPurpose, TargetBindingDigests{DatabaseSHA256: strings.Repeat("3", 64), ObjectStoreSHA256: expected.ObjectStoreSHA256}},
		{"wrong object binding", validMaterial, RestoreVerificationTargetPurpose, TargetBindingDigests{DatabaseSHA256: expected.DatabaseSHA256, ObjectStoreSHA256: strings.Repeat("3", 64)}},
		{"wrong generation", TargetMarkerMaterial{MarkerBody: validMaterial.MarkerBody, GenerationBody: []byte("00000000-0000-0000-0000-000000005002\n")}, RestoreVerificationTargetPurpose, expected},
		{"missing generation", TargetMarkerMaterial{MarkerBody: validMaterial.MarkerBody}, RestoreVerificationTargetPurpose, expected},
		{"v1 schema", replaceMarkerMember(validMaterial, RestoreTargetMarkerSchemaID, "cartulary.restore_verification_target.v1"), RestoreVerificationTargetPurpose, expected},
		{"duplicate member", TargetMarkerMaterial{MarkerBody: bytes.Replace(validMaterial.MarkerBody, []byte(`"purpose":`), []byte(`"purpose":"restore_verification_target","purpose":`), 1), GenerationBody: validMaterial.GenerationBody}, RestoreVerificationTargetPurpose, expected},
		{"unknown member", TargetMarkerMaterial{MarkerBody: bytes.Replace(validMaterial.MarkerBody, []byte(`{`), []byte(`{"unknown":true,`), 1), GenerationBody: validMaterial.GenerationBody}, RestoreVerificationTargetPurpose, expected},
		{"trailing data", TargetMarkerMaterial{MarkerBody: append(append([]byte(nil), validMaterial.MarkerBody...), []byte(` {}`)...), GenerationBody: validMaterial.GenerationBody}, RestoreVerificationTargetPurpose, expected},
		{"expired", markerMaterialForTest(t, markerWithTimes(validMarker, now.Add(-2*time.Hour), now.Add(-time.Hour)), generationID), RestoreVerificationTargetPurpose, expected},
		{"future issued", markerMaterialForTest(t, markerWithTimes(validMarker, now.Add(time.Minute), now.Add(time.Hour)), generationID), RestoreVerificationTargetPurpose, expected},
		{"lifetime over 24 hours", markerMaterialForTest(t, markerWithTimes(validMarker, now.Add(-time.Minute), now.Add(24*time.Hour)), generationID), RestoreVerificationTargetPurpose, expected},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if err := ValidateRestoreTargetMarker(test.material, test.purpose, test.expected, now); err == nil {
				t.Fatal("invalid restore target marker was admitted")
			}
		})
	}
}

func TestRestoreTargetBindingDigestsExcludeCredentials_Unit(t *testing.T) {
	base := Deployment{
		DatabaseStorage: RootBinding{BindingKind: "managed_service", ServiceRef: "restore_target"},
		ObjectStorage:   RootBinding{BindingKind: "managed_service", ServiceRef: "restore_target"},
		PostgresSettings: postgres.Settings{
			DSN: "postgres://operator:first-secret@db.example/restore",
		},
		ObjectSettings: objectstore.Settings{
			Endpoint:  "objects.example",
			AccessKey: "first-access",
			SecretKey: "first-secret",
			Bucket:    "restore",
		},
	}
	rotatedCredentials := base
	rotatedCredentials.PostgresSettings.DSN = "postgres://operator:second-secret@db.example/restore"
	rotatedCredentials.ObjectSettings.AccessKey = "second-access"
	rotatedCredentials.ObjectSettings.SecretKey = "second-secret"
	if TargetBindingDigestsFor(base) != TargetBindingDigestsFor(rotatedCredentials) {
		t.Fatal("credential rotation changed non-secret restore target binding digests")
	}
	differentTarget := base
	differentTarget.ObjectStorage.ServiceRef = "other_target"
	if TargetBindingDigestsFor(base) == TargetBindingDigestsFor(differentTarget) {
		t.Fatal("different restore target binding produced identical digest pair")
	}
}

func markerMaterialForTest(t testing.TB, marker RestoreTargetMarker, generationID uuid.UUID) TargetMarkerMaterial {
	t.Helper()
	body, err := json.Marshal(marker)
	if err != nil {
		t.Fatalf("encode marker fixture: %v", err)
	}
	return TargetMarkerMaterial{
		MarkerBody:     body,
		GenerationBody: []byte(generationID.String() + "\n"),
	}
}

func markerWithTimes(marker RestoreTargetMarker, issuedAt time.Time, expiresAt time.Time) RestoreTargetMarker {
	marker.IssuedAt = issuedAt.UTC().Format(time.RFC3339Nano)
	marker.ExpiresAt = expiresAt.UTC().Format(time.RFC3339Nano)
	return marker
}

func replaceMarkerMember(material TargetMarkerMaterial, oldValue string, newValue string) TargetMarkerMaterial {
	return TargetMarkerMaterial{
		MarkerBody:     bytes.Replace(material.MarkerBody, []byte(oldValue), []byte(newValue), 1),
		GenerationBody: material.GenerationBody,
	}
}

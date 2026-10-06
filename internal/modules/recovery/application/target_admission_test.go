package application

import (
	"bytes"
	"encoding/json"
	"errors"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	graphrestore "github.com/JochiRaider/cartulary/internal/modules/graphprojection/restore"
	"github.com/JochiRaider/cartulary/internal/modules/recovery"
	"github.com/JochiRaider/cartulary/internal/platform/objectstore"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
)

func TestRestoreTargetMarkerV2Admission_Unit(t *testing.T) {
	t.Run("owner issuance renewal and interrupted identity", func(t *testing.T) {
		now := time.Date(2026, 10, 6, 0, 0, 0, 0, time.UTC)
		bindings := TargetBindingDigestsFor(Deployment{})
		var stored TargetMarkerMaterial
		inspections, writes := 0, 0
		pristine := func() error { inspections++; return nil }
		publish := func(_, next TargetMarkerMaterial) error { writes++; stored = next; return nil }
		generation, err := prepareTargetMarker(stored, os.ErrNotExist, RestoreTargetPurpose, bindings, now, pristine, publish)
		if err != nil || generation == uuid.Nil || inspections != 1 || writes != 1 {
			t.Fatalf("initial issuance: %s %v %d %d", generation, err, inspections, writes)
		}
		original := stored
		replay, err := prepareTargetMarker(stored, nil, RestoreTargetPurpose, bindings, now.Add(time.Hour), pristine, publish)
		if err != nil || replay != generation || writes != 1 || inspections != 1 {
			t.Fatal("valid replay rewrote or inspected proof", err)
		}
		renewed, err := prepareTargetMarker(stored, nil, RestoreTargetPurpose, bindings, now.Add(25*time.Hour), pristine, publish)
		if err != nil || renewed != generation || writes != 2 || inspections != 2 || bytes.Equal(original.MarkerBody, stored.MarkerBody) {
			t.Fatal("renewal lost generation", err)
		}
		partial := TargetMarkerMaterial{GenerationBody: stored.GenerationBody}
		resumed, err := prepareTargetMarker(partial, os.ErrNotExist, RestoreTargetPurpose, bindings, now.Add(26*time.Hour), pristine, publish)
		if err != nil || resumed != generation {
			t.Fatal("interrupted issuance replaced generation", err)
		}
		writesBefore := writes
		for _, material := range []TargetMarkerMaterial{original, {}} {
			readErr := error(nil)
			if len(material.MarkerBody) == 0 {
				readErr = os.ErrNotExist
			}
			_, err := prepareTargetMarker(material, readErr, RestoreTargetPurpose, bindings, now.Add(48*time.Hour), func() error { return errors.New("partial state") }, publish)
			if err == nil || writes != writesBefore {
				t.Fatal("partial target restamped")
			}
		}
		changed := bindings
		changed.DatabaseSHA256 = strings.Repeat("0", 64)
		if _, err := prepareTargetMarker(original, nil, RestoreTargetPurpose, changed, now.Add(48*time.Hour), pristine, publish); err == nil || writes != writesBefore {
			t.Fatal("changed binding renewed")
		}
		if _, err := prepareTargetMarker(TargetMarkerMaterial{MarkerBody: original.MarkerBody}, os.ErrNotExist, RestoreTargetPurpose, bindings, now, pristine, publish); err == nil {
			t.Fatal("missing generation replaced")
		}
	})
	now := time.Date(2026, 7, 29, 17, 30, 0, 0, time.UTC)
	generationID := uuid.MustParse("00000000-0000-0000-0000-000000005001")
	expected := TargetBindingDigests{
		DatabaseSHA256:             strings.Repeat("1", 64),
		ObjectStoreSHA256:          strings.Repeat("2", 64),
		ReferencePackStorageSHA256: strings.Repeat("3", 64), ExportOutputsSHA256: strings.Repeat("4", 64),
	}
	validMarker := RestoreTargetMarker{
		ApplicationCryptoFormat: recovery.ApplicationCryptoFormatID,
		SchemaID:                RestoreTargetMarkerSchemaID,
		Purpose:                 RestoreVerificationTargetPurpose,
		TargetGenerationID:      generationID.String(),
		BindingDigests:          expected,
		IssuedAt:                now.Add(-time.Minute).Format(time.RFC3339Nano),
		ExpiresAt:               now.Add(time.Hour).Format(time.RFC3339Nano),
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
		{"v2 schema", replaceMarkerMember(validMaterial, RestoreTargetMarkerSchemaID, "cartulary.restore_target_marker.v2"), RestoreVerificationTargetPurpose, expected},
		{"v4 schema", replaceMarkerMember(validMaterial, RestoreTargetMarkerSchemaID, "cartulary.restore_target_marker.v4"), RestoreVerificationTargetPurpose, expected},
		{"wrong application format", replaceMarkerMember(validMaterial, recovery.ApplicationCryptoFormatID, "cartulary.application_crypto_format.v0"), RestoreVerificationTargetPurpose, expected},
		{"wrong Reference Pack root", replaceMarkerMember(validMaterial, expected.ReferencePackStorageSHA256, strings.Repeat("4", 64)), RestoreVerificationTargetPurpose, expected},
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

func TestRestoreTargetMarkerAdmissionReturnsValidatedGeneration_Unit(t *testing.T) {
	now := time.Date(2026, 7, 29, 17, 30, 0, 0, time.UTC)
	generationID := uuid.MustParse("00000000-0000-0000-0000-000000005001")
	expected := TargetBindingDigests{DatabaseSHA256: strings.Repeat("1", 64), ObjectStoreSHA256: strings.Repeat("2", 64), ReferencePackStorageSHA256: strings.Repeat("3", 64), ExportOutputsSHA256: strings.Repeat("4", 64)}
	material := markerMaterialForTest(t, RestoreTargetMarker{
		ApplicationCryptoFormat: recovery.ApplicationCryptoFormatID,
		SchemaID:                RestoreTargetMarkerSchemaID, Purpose: RestoreTargetPurpose,
		TargetGenerationID: generationID.String(), BindingDigests: expected,
		IssuedAt:  now.Add(-time.Minute).Format(time.RFC3339Nano),
		ExpiresAt: now.Add(time.Hour).Format(time.RFC3339Nano),
	}, generationID)
	admitted, err := AdmitRestoreTargetMarker(material, RestoreTargetPurpose, expected, now)
	if err != nil {
		t.Fatalf("admit restore target marker: %v", err)
	}
	if admitted != generationID {
		t.Fatalf("admitted target generation got %s want %s", admitted, generationID)
	}
}

func TestRecoveryJournalPayloadV3RetainsGraphCompletionAndV2Decoder_Unit(t *testing.T) {
	t.Run("portable age is independent of source retention", func(t *testing.T) {
		now := time.Now().UTC()
		id := uuid.New()
		backup := recovery.BackupSet{BackupSetID: id, ConsistencyPointAt: now.Add(-24 * time.Hour), RetainedUntil: now.Add(-time.Hour)}
		if err := admitPortableAge(backup, uuid.Nil, now); err != nil {
			t.Fatal("exact 24-hour boundary rejected", err)
		}
		if err := admitPortableAge(backup, uuid.Nil, now.Add(time.Nanosecond)); err == nil {
			t.Fatal("stale backup accepted without acknowledgement")
		}
		if err := admitPortableAge(backup, id, now.Add(100*24*time.Hour)); err != nil {
			t.Fatal("completed export expired with source retention", err)
		}
		if err := admitPortableAge(backup, uuid.New(), now); err == nil {
			t.Fatal("wrong acknowledgement accepted")
		}
		if err := admitPortableAge(backup, id, backup.ConsistencyPointAt.Add(-time.Nanosecond)); err == nil {
			t.Fatal("future consistency point accepted")
		}
	})

	operationID := uuid.MustParse("00000000-0000-0000-0000-000000004285")
	backupSetID := uuid.MustParse("00000000-0000-0000-0000-000000000428")
	targetGenerationID := uuid.MustParse("00000000-0000-0000-0000-000000005001")
	consistencyPoint := time.Date(2026, 7, 29, 17, 0, 0, 0, time.UTC)
	postcondition := strings.Repeat("a", 64)
	participant := graphrestore.RestoreRebuildResult{
		SchemaID:           graphrestore.RestoreRebuildResultSchemaID,
		RestoreOperationID: operationID.String(), TargetGenerationID: targetGenerationID.String(),
		Status: graphrestore.RestoreStatusSucceeded, ReadinessOutcome: graphrestore.RestoreReadinessReady,
		AlgorithmID:                 graphrestore.RestoreAlgorithmID,
		ImplementationBindingSHA256: strings.Repeat("b", 64), SourceRegistrySHA256: strings.Repeat("c", 64),
		ClearedTableIDs:     graphrestore.RestoreGraphTableIDs(),
		RebuiltViews:        []graphrestore.RestoreRebuiltView{},
		PostconditionSHA256: &postcondition, Warnings: []graphrestore.RestoreSafeMessage{}, Errors: []graphrestore.RestoreSafeMessage{},
	}
	completion := &GraphProjectionCompletionEvidence{
		TargetGenerationID: targetGenerationID, RestoreOperationID: operationID, BackupSetID: backupSetID,
		ConsistencyPointAt: consistencyPoint, RecoveryStateCatalogSHA256: strings.Repeat("9", 64),
		SourceRegistrySHA256:        participant.SourceRegistrySHA256,
		ImplementationBindingSHA256: participant.ImplementationBindingSHA256,
		PostconditionSHA256:         postcondition, ParticipantResult: participant,
	}
	current := recoveryJournalCompletionPayloadV6{
		recoveryJournalCompletionFields: recoveryJournalCompletionFields{
			SchemaID: RecoveryJournalPayloadSchemaID, RecordKind: "completion", OperationID: operationID,
			Operation: OperationRestoreLatest, StartedAt: consistencyPoint, CompletedAt: consistencyPoint.Add(time.Minute),
			Result: ResultSucceeded, BackupSetID: &backupSetID, ConsistencyPointAt: &consistencyPoint, ArtifactCounts: []ArtifactCount{},
		}, GraphProjectionCompletion: completion,
		TargetBindings: &TargetBindingDigests{DatabaseSHA256: strings.Repeat("1", 64), ObjectStoreSHA256: strings.Repeat("2", 64), ReferencePackStorageSHA256: strings.Repeat("3", 64), ExportOutputsSHA256: strings.Repeat("4", 64)},
	}
	encoded, err := json.Marshal(current)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := DecodeRecoveryJournalPayload(encoded); err != nil {
		t.Fatal("valid current journal", err)
	}
	// The historical test identity now records deliberate rejection; current
	// releases do not carry journal readers for historical deployments.
	for _, version := range []string{"v2", "v3", "v4"} {
		historical := bytes.Replace(encoded, []byte(RecoveryJournalPayloadSchemaID), []byte("cartulary.operator_recovery_journal_payload."+version), 1)
		if _, err := DecodeRecoveryJournalPayload(historical); err == nil {
			t.Fatalf("historical %s journal admitted", version)
		}
	}
	var root map[string]any
	if err := json.Unmarshal(encoded, &root); err != nil {
		t.Fatal(err)
	}
	var checkObjects func(map[string]any)
	checkObjects = func(object map[string]any) {
		for name, original := range object {
			delete(object, name)
			changed, _ := json.Marshal(root)
			if _, err := DecodeRecoveryJournalPayload(changed); err == nil {
				t.Fatalf("omitted journal member admitted: %s", name)
			}
			object[name] = true
			changed, _ = json.Marshal(root)
			if _, err := DecodeRecoveryJournalPayload(changed); err == nil {
				t.Fatalf("wrong journal member type admitted: %s", name)
			}
			object[name] = original
			if nested, ok := original.(map[string]any); ok {
				checkObjects(nested)
			}
		}
		object["unexpected"] = true
		changed, _ := json.Marshal(root)
		if _, err := DecodeRecoveryJournalPayload(changed); err == nil {
			t.Fatal("unknown journal member admitted")
		}
		delete(object, "unexpected")
	}
	checkObjects(root)
	for _, bad := range [][]byte{
		bytes.Replace(encoded, []byte(`"record_kind":`), []byte(`"record_kind":"completion","record_kind":`), 1),
		bytes.Replace(encoded, []byte(`"attempt_id":null`), []byte(`"attempt_id":"\ud800"`), 1),
		bytes.Replace(encoded, []byte(`"artifact_counts":[]`), []byte(`"artifact_counts":null`), 1),
	} {
		if _, err := DecodeRecoveryJournalPayload(bad); err == nil {
			t.Fatal("lossy or null journal content admitted")
		}
	}

}

func TestRestoreTargetBindingDigestsExcludeCredentials_Unit(t *testing.T) {
	base := Deployment{
		DatabaseStorage: RootBinding{BindingKind: "managed_service", ServiceRef: "restore_target"},
		ObjectStorage:   RootBinding{BindingKind: "managed_service", ServiceRef: "restore_target"},
		PostgresSettings: postgres.Settings{
			DSN:          "postgres://operator:first-secret@db.example/restore",
			Purpose:      postgres.PurposeRecovery,
			ExpectedRole: "cartulary_recovery",
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
	for _, change := range []func(*Deployment){
		func(d *Deployment) { d.PostgresSettings.DSN = "postgres://operator:secret@db.example/other" },
		func(d *Deployment) { d.ObjectSettings.Bucket = "other" },
	} {
		changed := base
		change(&changed)
		if TargetBindingDigestsFor(base) == TargetBindingDigestsFor(changed) {
			t.Fatal("physical namespace absent from target proof")
		}
	}
	differentTarget := base
	differentTarget.ObjectStorage.ServiceRef = "other_target"
	if TargetBindingDigestsFor(base) == TargetBindingDigestsFor(differentTarget) {
		t.Fatal("different restore target binding produced identical digest pair")
	}
	differentTarget = base
	differentTarget.ReferencePackStorage = RootBinding{BindingKind: "filesystem_root", Path: "/different/reference-packs"}
	if TargetBindingDigestsFor(base) == TargetBindingDigestsFor(differentTarget) {
		t.Fatal("Reference Pack root absent from restore target identity")
	}
	differentTarget = base
	differentTarget.ExportOutputs = RootBinding{BindingKind: "filesystem_root", Path: "/different/exports"}
	if TargetBindingDigestsFor(base) == TargetBindingDigestsFor(differentTarget) {
		t.Fatal("export root absent from target identity")
	}
	factory := func() (recovery.ReferencePackStorage, error) { return nil, nil }
	source := Deployment{PostgresSettings: postgres.Settings{DSN: "source"}, ObjectSettings: objectstore.Settings{BindingKind: "filesystem_root", RootPath: "/source/objects"}, OpenReferencePacks: factory, OpenExportOutputs: func() (recovery.RootObjectStorage, error) { return nil, nil },
		ReferencePackStorage: RootBinding{BindingKind: "filesystem_root", Path: "/source/packs"}, ExportOutputs: RootBinding{BindingKind: "filesystem_root", Path: "/source/exports"}, ObjectStorage: RootBinding{BindingKind: "filesystem_root", Path: "/source/objects"}, BackupStorage: RootBinding{BindingKind: "filesystem_root", Path: "/source/backups"}}
	target := Deployment{PostgresSettings: postgres.Settings{DSN: "target"}, ObjectSettings: objectstore.Settings{BindingKind: "filesystem_root", RootPath: "/target/objects"}, OpenReferencePacks: factory, OpenExportOutputs: func() (recovery.RootObjectStorage, error) { return nil, nil },
		ReferencePackStorage: RootBinding{BindingKind: "filesystem_root", Path: "/target/packs"}, ExportOutputs: RootBinding{BindingKind: "filesystem_root", Path: "/target/exports"}, ObjectStorage: RootBinding{BindingKind: "filesystem_root", Path: "/target/objects"}, BackupStorage: RootBinding{BindingKind: "filesystem_root", Path: "/target/backups"}}
	if err := requireDistinctRestoreTarget("/source.toml", "/target.toml", source, target); err != nil {
		t.Fatal(err)
	}
	for _, path := range []string{"/source/packs", "/source", "/source/packs/nested", "/source/objects", "/source/backups/nested", "/target/objects/nested", "/target/backups"} {
		changed := target
		changed.ReferencePackStorage.Path = path
		if requireDistinctRestoreTarget("/source.toml", "/target.toml", source, changed) == nil {
			t.Fatalf("overlapping pack root admitted: %s", path)
		}
	}
	for _, path := range []string{"/source/exports", "/source/packs", "/source", "/source/objects", "/source/backups/nested", "/target/packs/nested", "/target/objects", "/target/backups"} {
		changed := target
		changed.ExportOutputs.Path = path
		if requireDistinctRestoreTarget("/source.toml", "/target.toml", source, changed) == nil {
			t.Fatalf("overlapping export root admitted: %s", path)
		}
	}
	changed := target
	changed.ObjectStorage.Path = "/source/packs"
	if requireDistinctRestoreTarget("/source.toml", "/target.toml", source, changed) == nil {
		t.Fatal("target objects alias source packs")
	}
	changed = target
	changed.ReferencePackStorage.Path = "/source/packs-other"
	if err := requireDistinctRestoreTarget("/source.toml", "/target.toml", source, changed); err != nil {
		t.Fatal("non-overlapping sibling rejected", err)
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

package recovery_test

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/JochiRaider/cartulary/internal/gen/contractrecovery"
	"github.com/JochiRaider/cartulary/internal/modules/recovery"
)

func TestRestoreVerificationArtifactV2BindsWorkbookExecutionAndBasis_Unit(t *testing.T) {
	basis := recovery.RestoreVerificationBasis{
		MechanismID:                recovery.VNextBackupMechanismID,
		DatabaseBindingSHA256:      strings.Repeat("1", 64),
		ObjectStoreBindingSHA256:   strings.Repeat("2", 64),
		BackupStorageBindingSHA256: strings.Repeat("3", 64),
		RecoveryStateCatalogSHA256: strings.Repeat("4", 64),
		CodecRegistrySHA256:        strings.Repeat("5", 64),
	}
	basisSHA256, err := basis.SHA256()
	if err != nil {
		t.Fatalf("verification basis digest: %v", err)
	}
	rowCount := int64(0)
	artifact := recovery.RestoreVerificationArtifact{
		SchemaID:                   recovery.RestoreVerificationArtifactSchemaID,
		VerificationAttemptID:      "recovery-attempt-0001",
		BackupSetID:                "00000000-0000-0000-0000-000000000201",
		ConsistencyPointAt:         time.Date(2026, 7, 29, 11, 0, 0, 0, time.UTC),
		VerificationBasis:          basis,
		VerificationBasisSHA256:    basisSHA256,
		RecoveryStateCatalogSHA256: basis.RecoveryStateCatalogSHA256,
		CodecRegistrySHA256:        basis.CodecRegistrySHA256,
		ManifestSHA256:             strings.Repeat("7", 64),
		RestoredObjectCount:        1,
		SelectedIncidentID:         stringPointer("00000000-0000-0000-0000-000000000401"),
		WorkbookProbe: recovery.RestoreVerificationWorkbookProbeArtifact{
			Status:         "executed",
			RegistrationID: "timeline.base_restore_probe.v1",
			ViewSchemaID:   "cartulary.view.timeline.v2",
			RowCount:       &rowCount,
		},
		Result:      "pass",
		CompletedAt: time.Date(2026, 7, 29, 12, 5, 0, 0, time.UTC),
	}
	body, err := recovery.EncodeRestoreVerificationArtifact(artifact)
	if err != nil {
		t.Fatalf("encode restore verification v2: %v", err)
	}
	decoded, err := recovery.DecodeRestoreVerificationArtifact(body)
	if err != nil {
		t.Fatalf("decode restore verification v2: %v", err)
	}
	if decoded.WorkbookProbe.RegistrationID != "timeline.base_restore_probe.v1" ||
		decoded.WorkbookProbe.RowCount == nil ||
		*decoded.WorkbookProbe.RowCount != 0 ||
		decoded.VerificationBasisSHA256 != basisSHA256 {
		t.Fatalf("decoded restore verification identity got %#v", decoded)
	}
	if !bytes.Contains(body, []byte(`"row_count":0`)) {
		t.Fatalf("zero row count was omitted: %s", body)
	}

	duplicate := bytes.Replace(
		body,
		[]byte(`"schema_id":"cartulary.restore_verification.v2"`),
		[]byte(`"schema_id":"cartulary.restore_verification.v2","schema_id":"cartulary.restore_verification.v2"`),
		1,
	)
	if _, err := recovery.DecodeRestoreVerificationArtifact(duplicate); err == nil {
		t.Fatal("duplicate restore verification member was accepted")
	}

	artifact.WorkbookProbe.RegistrationID = ""
	if _, err := recovery.EncodeRestoreVerificationArtifact(artifact); err == nil {
		t.Fatal("executed workbook probe without registration identity was accepted")
	}

	artifact.SelectedIncidentID = nil
	artifact.WorkbookProbe = recovery.RestoreVerificationWorkbookProbeArtifact{
		Status: "skipped",
		Reason: "verification_failed_before_probe",
	}
	artifact.Result = "fail"
	if _, err := recovery.EncodeRestoreVerificationArtifact(artifact); err != nil {
		t.Fatalf("failed verification before workbook probe was not encodable: %v", err)
	}
	artifact.Result = "pass"
	if _, err := recovery.EncodeRestoreVerificationArtifact(artifact); err == nil {
		t.Fatal("passing verification accepted prior-failure workbook skip")
	}
}

func TestRestoreVerificationArtifactV2CanonicalFixture_Unit(t *testing.T) {
	for _, artifact := range contractrecovery.Artifacts {
		if artifact.Path != "contracts/recovery/fixtures/restore-verification.v2.json" {
			continue
		}
		decoded, err := recovery.DecodeRestoreVerificationArtifact([]byte(artifact.JSON + "\n"))
		if err != nil {
			t.Fatalf("decode generated canonical restore verification fixture: %v", err)
		}
		if decoded.WorkbookProbe.RegistrationID != "timeline.base_restore_probe.v1" {
			t.Fatalf("canonical workbook registration got %#v", decoded.WorkbookProbe)
		}
		return
	}
	t.Fatal("generated canonical restore verification fixture is missing")
}

func stringPointer(value string) *string {
	return &value
}

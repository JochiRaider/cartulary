package recovery

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
)

const (
	RestoreVerificationArtifactSchemaID = "cartulary.restore_verification.v2"
	VNextBackupMechanismID              = "logical_streaming_backup.v2"
)

var recoveryIdentifierPattern = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9_.:-]{0,159}$`)

type RestoreVerificationBasis struct {
	MechanismID                string `json:"mechanism_id"`
	DatabaseBindingSHA256      string `json:"database_binding_sha256"`
	ObjectStoreBindingSHA256   string `json:"object_store_binding_sha256"`
	BackupStorageBindingSHA256 string `json:"backup_storage_binding_sha256"`
	RecoveryStateCatalogSHA256 string `json:"recovery_state_catalog_sha256"`
	CodecRegistrySHA256        string `json:"codec_registry_sha256"`
}

func (basis RestoreVerificationBasis) SHA256() (string, error) {
	if err := basis.Validate(); err != nil {
		return "", err
	}
	body, err := json.Marshal(basis)
	if err != nil {
		return "", fmt.Errorf("encode restore verification basis: %w", err)
	}
	sum := sha256.Sum256(body)
	return hex.EncodeToString(sum[:]), nil
}

func (basis RestoreVerificationBasis) Validate() error {
	if !recoveryIdentifierPattern.MatchString(basis.MechanismID) {
		return fmt.Errorf("%w: mechanism_id is invalid", ErrInvalidVerificationBasis)
	}
	for name, value := range map[string]string{
		"database_binding_sha256":       basis.DatabaseBindingSHA256,
		"object_store_binding_sha256":   basis.ObjectStoreBindingSHA256,
		"backup_storage_binding_sha256": basis.BackupStorageBindingSHA256,
		"recovery_state_catalog_sha256": basis.RecoveryStateCatalogSHA256,
		"codec_registry_sha256":         basis.CodecRegistrySHA256,
	} {
		if !validSHA256Hex(value) {
			return fmt.Errorf("%w: %s is invalid", ErrInvalidVerificationBasis, name)
		}
	}
	return nil
}

type RestoreVerificationWorkbookProbeArtifact struct {
	Status         string `json:"status"`
	Reason         string `json:"reason,omitempty"`
	RegistrationID string `json:"registration_id,omitempty"`
	ViewSchemaID   string `json:"view_schema_id,omitempty"`
	RowCount       *int64 `json:"row_count,omitempty"`
}

type RestoreVerificationArtifact struct {
	SchemaID                   string                                   `json:"schema_id"`
	VerificationAttemptID      string                                   `json:"verification_attempt_id"`
	BackupSetID                string                                   `json:"backup_set_id"`
	ConsistencyPointAt         time.Time                                `json:"consistency_point_at"`
	VerificationBasis          RestoreVerificationBasis                 `json:"verification_basis"`
	VerificationBasisSHA256    string                                   `json:"verification_basis_sha256"`
	RecoveryStateCatalogSHA256 string                                   `json:"recovery_state_catalog_sha256"`
	CodecRegistrySHA256        string                                   `json:"codec_registry_sha256"`
	ManifestSHA256             string                                   `json:"manifest_sha256"`
	RestoredObjectCount        int64                                    `json:"restored_object_count"`
	SelectedIncidentID         *string                                  `json:"selected_incident_id"`
	WorkbookProbe              RestoreVerificationWorkbookProbeArtifact `json:"workbook_probe"`
	Result                     string                                   `json:"result"`
	CompletedAt                time.Time                                `json:"completed_at"`
}

func EncodeRestoreVerificationArtifact(artifact RestoreVerificationArtifact) ([]byte, error) {
	if err := ValidateRestoreVerificationArtifact(artifact); err != nil {
		return nil, err
	}
	return canonicalRestoreVerificationArtifactV2Bytes(artifact), nil
}

func DecodeRestoreVerificationArtifact(body []byte) (RestoreVerificationArtifact, error) {
	if err := rejectDuplicateJSONKeys(body); err != nil {
		return RestoreVerificationArtifact{}, fmt.Errorf("%w: restore verification v2 JSON keys must be unique: %v", ErrInvalidBackupArtifact, err)
	}
	var artifact RestoreVerificationArtifact
	if err := decodeStrictJSON(body, &artifact); err != nil {
		return RestoreVerificationArtifact{}, fmt.Errorf("%w: decode restore verification v2: %v", ErrInvalidBackupArtifact, err)
	}
	if err := ValidateRestoreVerificationArtifact(artifact); err != nil {
		return RestoreVerificationArtifact{}, err
	}
	if !bytes.Equal(body, canonicalRestoreVerificationArtifactV2Bytes(artifact)) {
		return RestoreVerificationArtifact{}, fmt.Errorf("%w: restore verification v2 is not canonical JSON", ErrInvalidBackupArtifact)
	}
	return artifact, nil
}

func ValidateRestoreVerificationArtifact(artifact RestoreVerificationArtifact) error {
	if artifact.SchemaID != RestoreVerificationArtifactSchemaID {
		return fmt.Errorf("%w: unsupported restore verification schema %q", ErrInvalidBackupArtifact, artifact.SchemaID)
	}
	if !recoveryIdentifierPattern.MatchString(artifact.VerificationAttemptID) {
		return fmt.Errorf("%w: verification_attempt_id is invalid", ErrInvalidBackupArtifact)
	}
	if _, err := uuid.Parse(artifact.BackupSetID); err != nil {
		return fmt.Errorf("%w: restore verification backup_set_id is invalid", ErrInvalidBackupArtifact)
	}
	if artifact.ConsistencyPointAt.IsZero() || artifact.CompletedAt.IsZero() {
		return fmt.Errorf("%w: restore verification timestamps are required", ErrInvalidBackupArtifact)
	}
	if err := artifact.VerificationBasis.Validate(); err != nil {
		return fmt.Errorf("%w: %v", ErrInvalidBackupArtifact, err)
	}
	basisSHA256, err := artifact.VerificationBasis.SHA256()
	if err != nil || basisSHA256 != artifact.VerificationBasisSHA256 {
		return fmt.Errorf("%w: verification_basis_sha256 mismatch", ErrInvalidBackupArtifact)
	}
	if artifact.RecoveryStateCatalogSHA256 != artifact.VerificationBasis.RecoveryStateCatalogSHA256 ||
		artifact.CodecRegistrySHA256 != artifact.VerificationBasis.CodecRegistrySHA256 {
		return fmt.Errorf("%w: verification basis catalog or codec digest mismatch", ErrInvalidBackupArtifact)
	}
	if !validSHA256Hex(artifact.ManifestSHA256) {
		return fmt.Errorf("%w: manifest_sha256 is invalid", ErrInvalidBackupArtifact)
	}
	if artifact.RestoredObjectCount < 0 || artifact.RestoredObjectCount > 10_485_760 {
		return fmt.Errorf("%w: restored_object_count is outside bounds", ErrInvalidBackupArtifact)
	}
	if artifact.SelectedIncidentID != nil {
		if _, err := uuid.Parse(*artifact.SelectedIncidentID); err != nil {
			return fmt.Errorf("%w: selected_incident_id is invalid", ErrInvalidBackupArtifact)
		}
	}
	if artifact.WorkbookProbe.Status == "skipped" {
		if artifact.WorkbookProbe.Reason != "no_incidents" &&
			artifact.WorkbookProbe.Reason != "verification_failed_before_probe" {
			return fmt.Errorf("%w: skipped workbook probe reason is invalid", ErrInvalidBackupArtifact)
		}
		if artifact.WorkbookProbe.Reason == "no_incidents" && artifact.SelectedIncidentID != nil {
			return fmt.Errorf("%w: no-incident workbook probe has a selected incident", ErrInvalidBackupArtifact)
		}
		if artifact.WorkbookProbe.Reason == "verification_failed_before_probe" && artifact.Result != "fail" {
			return fmt.Errorf("%w: prior-failure workbook skip requires failed verification", ErrInvalidBackupArtifact)
		}
		if artifact.WorkbookProbe.RegistrationID != "" ||
			artifact.WorkbookProbe.ViewSchemaID != "" ||
			artifact.WorkbookProbe.RowCount != nil {
			return fmt.Errorf("%w: skipped workbook probe has execution fields", ErrInvalidBackupArtifact)
		}
	} else {
		if artifact.SelectedIncidentID == nil {
			return fmt.Errorf("%w: executed workbook probe requires a selected incident", ErrInvalidBackupArtifact)
		}
		if artifact.WorkbookProbe.Status != "executed" ||
			!recoveryIdentifierPattern.MatchString(artifact.WorkbookProbe.RegistrationID) ||
			!recoveryIdentifierPattern.MatchString(artifact.WorkbookProbe.ViewSchemaID) ||
			artifact.WorkbookProbe.RowCount == nil ||
			*artifact.WorkbookProbe.RowCount < 0 ||
			artifact.WorkbookProbe.Reason != "" {
			return fmt.Errorf("%w: executed workbook probe evidence is incomplete", ErrInvalidBackupArtifact)
		}
	}
	if artifact.Result != "pass" && artifact.Result != "fail" {
		return fmt.Errorf("%w: restore verification result is outside the closed vocabulary", ErrInvalidBackupArtifact)
	}
	return nil
}

func canonicalRestoreVerificationArtifactV2Bytes(artifact RestoreVerificationArtifact) []byte {
	workbookProbe := map[string]any{
		"status": artifact.WorkbookProbe.Status,
	}
	if artifact.WorkbookProbe.Status == "skipped" {
		workbookProbe["reason"] = artifact.WorkbookProbe.Reason
	} else {
		workbookProbe["registration_id"] = artifact.WorkbookProbe.RegistrationID
		workbookProbe["view_schema_id"] = artifact.WorkbookProbe.ViewSchemaID
		workbookProbe["row_count"] = *artifact.WorkbookProbe.RowCount
	}
	return marshalCanonical(map[string]any{
		"schema_id":               artifact.SchemaID,
		"verification_attempt_id": artifact.VerificationAttemptID,
		"backup_set_id":           artifact.BackupSetID,
		"consistency_point_at":    artifact.ConsistencyPointAt,
		"verification_basis": map[string]any{
			"mechanism_id":                  artifact.VerificationBasis.MechanismID,
			"database_binding_sha256":       artifact.VerificationBasis.DatabaseBindingSHA256,
			"object_store_binding_sha256":   artifact.VerificationBasis.ObjectStoreBindingSHA256,
			"backup_storage_binding_sha256": artifact.VerificationBasis.BackupStorageBindingSHA256,
			"recovery_state_catalog_sha256": artifact.VerificationBasis.RecoveryStateCatalogSHA256,
			"codec_registry_sha256":         artifact.VerificationBasis.CodecRegistrySHA256,
		},
		"verification_basis_sha256":     artifact.VerificationBasisSHA256,
		"recovery_state_catalog_sha256": artifact.RecoveryStateCatalogSHA256,
		"codec_registry_sha256":         artifact.CodecRegistrySHA256,
		"manifest_sha256":               artifact.ManifestSHA256,
		"restored_object_count":         artifact.RestoredObjectCount,
		"selected_incident_id":          artifact.SelectedIncidentID,
		"workbook_probe":                workbookProbe,
		"result":                        artifact.Result,
		"completed_at":                  artifact.CompletedAt,
	})
}

func SHA256String(value string) string {
	sum := sha256.Sum256([]byte(strings.TrimSpace(value)))
	return hex.EncodeToString(sum[:])
}

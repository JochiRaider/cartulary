package recovery

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

const (
	TargetStateSchemaID  = "cartulary.recovery_target_state.v1"
	TargetStateTableName = "recovery_target_state"
)

type TargetState string

const (
	TargetUnproven          TargetState = "unproven"
	TargetInitialized       TargetState = "initialized"
	TargetVerificationReady TargetState = "verification_ready"
	TargetRestoring         TargetState = "restoring"
	TargetComplete          TargetState = "complete"
	TargetResetting         TargetState = "resetting"
)

var ErrTargetIncomplete = errors.New("recovery: target completion is not established")

// TargetControl is installation-local authority. It is never captured or
// restored as payload, and application credentials cannot mutate it.
type TargetControl struct {
	SchemaID                   string                `json:"schema_id"`
	State                      TargetState           `json:"state"`
	Revision                   int64                 `json:"revision"`
	TargetGenerationID         *string               `json:"target_generation_id"`
	Purpose                    *string               `json:"purpose"`
	Operation                  *string               `json:"operation"`
	OperationID                *string               `json:"operation_id"`
	AttemptID                  *string               `json:"attempt_id"`
	BackupSetID                *string               `json:"backup_set_id"`
	RecoveryStateCatalogSHA256 *string               `json:"recovery_state_catalog_sha256"`
	CodecRegistrySHA256        *string               `json:"codec_registry_sha256"`
	BindingDigests             *TargetBindingDigests `json:"binding_digests"`
	CompletionEvidenceSHA256   *string               `json:"completion_evidence_sha256"`
}

func (control TargetControl) Validate() error {
	if control.SchemaID != TargetStateSchemaID || control.Revision <= 0 {
		return ErrTargetIncomplete
	}
	unbound := control.TargetGenerationID == nil && control.Purpose == nil && control.BindingDigests == nil
	noOperation := control.Operation == nil && control.OperationID == nil && control.AttemptID == nil &&
		control.BackupSetID == nil && control.RecoveryStateCatalogSHA256 == nil && control.CodecRegistrySHA256 == nil && control.CompletionEvidenceSHA256 == nil
	if control.State == TargetInitialized || control.State == TargetUnproven {
		if unbound && noOperation {
			return nil
		}
		return ErrTargetIncomplete
	}
	if !targetControlUUID(control.TargetGenerationID) || control.Purpose == nil || control.BindingDigests == nil || !control.BindingDigests.valid() {
		return ErrTargetIncomplete
	}
	if control.State == TargetVerificationReady || control.State == TargetResetting {
		if *control.Purpose == "restore_verification_target" && noOperation {
			return nil
		}
		return ErrTargetIncomplete
	}
	if control.State != TargetRestoring && control.State != TargetComplete {
		return ErrTargetIncomplete
	}
	if !targetControlUUID(control.OperationID) || !targetControlUUID(control.BackupSetID) || control.Operation == nil ||
		control.RecoveryStateCatalogSHA256 == nil || !validSHA256Hex(*control.RecoveryStateCatalogSHA256) ||
		control.CodecRegistrySHA256 == nil || !validSHA256Hex(*control.CodecRegistrySHA256) {
		return ErrTargetIncomplete
	}
	switch *control.Operation {
	case "restore_latest", "restore_bundle":
		if *control.Purpose != "restore_target" || control.AttemptID != nil {
			return ErrTargetIncomplete
		}
	case "restore_verify_latest":
		if *control.Purpose != "restore_verification_target" || control.AttemptID != nil {
			return ErrTargetIncomplete
		}
	case "restore_verify_due":
		if *control.Purpose != "restore_verification_target" || control.AttemptID == nil || !targetControlAttempt(*control.AttemptID) {
			return ErrTargetIncomplete
		}
	default:
		return ErrTargetIncomplete
	}
	if control.State == TargetRestoring {
		if control.CompletionEvidenceSHA256 != nil {
			return ErrTargetIncomplete
		}
	} else if control.CompletionEvidenceSHA256 == nil || !validSHA256Hex(*control.CompletionEvidenceSHA256) {
		return ErrTargetIncomplete
	}
	return nil
}

// RequireServing is also capture admission. Verification targets cannot become
// application deployments or backup sources by completing a verification.
func (control TargetControl) RequireServing(bindings TargetBindingDigests) error {
	if err := control.Validate(); err != nil {
		return err
	}
	if control.State == TargetInitialized {
		return nil
	}
	if control.State != TargetComplete || *control.Purpose != "restore_target" || !bindings.valid() || *control.BindingDigests != bindings {
		return ErrTargetIncomplete
	}
	registry, err := loadVNextRecoveryGenerationRegistry()
	if err != nil {
		return errors.Join(ErrTargetIncomplete, err)
	}
	identity := registry.current.identity()
	if *control.RecoveryStateCatalogSHA256 != identity.RecoveryStateCatalogSHA256 || *control.CodecRegistrySHA256 != identity.CodecRegistrySHA256 {
		return ErrTargetIncomplete
	}
	return nil
}

func targetControlUUID(value *string) bool {
	if value == nil {
		return false
	}
	id, err := uuid.Parse(*value)
	return err == nil && id != uuid.Nil && id.String() == *value
}

func targetControlAttempt(value string) bool {
	if len(value) == 0 || len(value) > 160 {
		return false
	}
	for i, c := range []byte(value) {
		if c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' || c >= '0' && c <= '9' {
			continue
		}
		if i == 0 || c != '_' && c != '.' && c != '-' {
			return false
		}
	}
	return true
}

type targetStateQuery interface {
	QueryRow(context.Context, string, ...any) pgx.Row
}

// ReadTargetControl accepts a pool or the already-open capture/restore
// transaction so admission observes the caller's actual consistency boundary.
func ReadTargetControl(ctx context.Context, database targetStateQuery) (TargetControl, error) {
	var control TargetControl
	var databaseDigest, objectDigest, referenceDigest, exportDigest *string
	var count int64
	err := database.QueryRow(ctx, `SELECT schema_id,state,revision,target_generation_id::text,purpose,operation,
 operation_id::text,attempt_id,backup_set_id::text,recovery_state_catalog_sha256,codec_registry_sha256,
 database_sha256,object_store_sha256,reference_pack_storage_sha256,export_outputs_sha256,completion_evidence_sha256,
 (SELECT count(*) FROM public.recovery_target_state) FROM public.recovery_target_state WHERE singleton=true`).Scan(
		&control.SchemaID, &control.State, &control.Revision, &control.TargetGenerationID, &control.Purpose, &control.Operation,
		&control.OperationID, &control.AttemptID, &control.BackupSetID, &control.RecoveryStateCatalogSHA256, &control.CodecRegistrySHA256,
		&databaseDigest, &objectDigest, &referenceDigest, &exportDigest, &control.CompletionEvidenceSHA256, &count)
	if err != nil {
		return TargetControl{}, fmt.Errorf("%w: read control: %w", ErrTargetIncomplete, err)
	}
	if count != 1 {
		return TargetControl{}, ErrTargetIncomplete
	}
	if databaseDigest != nil || objectDigest != nil || referenceDigest != nil || exportDigest != nil {
		if databaseDigest == nil || objectDigest == nil || referenceDigest == nil || exportDigest == nil {
			return TargetControl{}, ErrTargetIncomplete
		}
		control.BindingDigests = &TargetBindingDigests{*databaseDigest, *objectDigest, *referenceDigest, *exportDigest}
	}
	if err := control.Validate(); err != nil {
		return TargetControl{}, err
	}
	return control, nil
}

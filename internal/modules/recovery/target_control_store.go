package recovery

import (
	"context"
	"errors"
	"fmt"
	"math"
	"strings"

	"github.com/google/uuid"
)

var ErrTargetControlConflict = errors.New("recovery: target control changed")

type RestoreTargetIdentity struct {
	TargetGenerationID         uuid.UUID
	Purpose                    string
	Operation                  string
	OperationID                uuid.UUID
	AttemptID                  *string
	BackupSetID                uuid.UUID
	RecoveryStateCatalogSHA256 string
	CodecRegistrySHA256        string
	BindingDigests             TargetBindingDigests
}

// TargetCompletionProof is produced only after the owner has authenticated or
// durably written the exact successful terminal payload. It carries its digest,
// not another journal encoding or a generic readiness notification.
type TargetCompletionProof struct {
	Identity       RestoreTargetIdentity
	EvidenceSHA256 string
}

func (identity RestoreTargetIdentity) control() TargetControl {
	bindings := identity.BindingDigests
	var attempt *string
	if identity.AttemptID != nil {
		attempt = targetControlText(*identity.AttemptID)
	}
	return TargetControl{
		SchemaID: TargetStateSchemaID, State: TargetRestoring, Revision: 1,
		TargetGenerationID: targetControlText(identity.TargetGenerationID.String()), Purpose: targetControlText(identity.Purpose),
		Operation: targetControlText(identity.Operation), OperationID: targetControlText(identity.OperationID.String()), AttemptID: attempt,
		BackupSetID: targetControlText(identity.BackupSetID.String()), RecoveryStateCatalogSHA256: targetControlText(identity.RecoveryStateCatalogSHA256),
		CodecRegistrySHA256: targetControlText(identity.CodecRegistrySHA256), BindingDigests: &bindings,
	}
}

func (identity RestoreTargetIdentity) validate() error {
	if err := identity.control().Validate(); err != nil {
		return err
	}
	registry, err := loadVNextRecoveryGenerationRegistry()
	if err != nil {
		return err
	}
	if identity.RecoveryStateCatalogSHA256 != registry.current.stateCatalog.DigestSHA256() || identity.CodecRegistrySHA256 != registry.current.codecRegistrySHA256 {
		return ErrTargetIncomplete
	}
	return nil
}

func targetControlText(value string) *string { return &value }

func sameTargetText(a, b *string) bool {
	if a == nil || b == nil {
		return a == nil && b == nil
	}
	return *a == *b
}

func sameTargetBinding(a, b TargetControl) bool {
	return sameTargetText(a.TargetGenerationID, b.TargetGenerationID) && sameTargetText(a.Purpose, b.Purpose) &&
		a.BindingDigests != nil && b.BindingDigests != nil && *a.BindingDigests == *b.BindingDigests
}

func sameTargetOperation(a, b TargetControl) bool {
	return sameTargetBinding(a, b) && sameTargetText(a.Operation, b.Operation) && sameTargetText(a.OperationID, b.OperationID) &&
		sameTargetText(a.AttemptID, b.AttemptID) && sameTargetText(a.BackupSetID, b.BackupSetID) &&
		sameTargetText(a.RecoveryStateCatalogSHA256, b.RecoveryStateCatalogSHA256) && sameTargetText(a.CodecRegistrySHA256, b.CodecRegistrySHA256)
}

func validTargetTransition(before, after TargetControl) bool {
	switch before.State {
	case TargetInitialized:
		return after.State == TargetRestoring || after.State == TargetResetting
	case TargetVerificationReady:
		if after.State == TargetVerificationReady {
			return sameTargetText(before.Purpose, after.Purpose) && *before.BindingDigests == *after.BindingDigests
		}
		return (after.State == TargetRestoring || after.State == TargetResetting) && sameTargetBinding(before, after)
	case TargetRestoring:
		if after.State == TargetComplete {
			return sameTargetOperation(before, after)
		}
		return after.State == TargetResetting && *before.Purpose == "restore_verification_target" && sameTargetBinding(before, after)
	case TargetComplete:
		return after.State == TargetResetting && *before.Purpose == "restore_verification_target" && sameTargetBinding(before, after)
	case TargetResetting:
		return after.State == TargetVerificationReady && sameTargetBinding(before, after)
	default:
		return false
	}
}

// The complete old record participates in the predicate. A revision alone is
// not permission to overwrite changed operation, target or completion identity.
const targetControlColumns = `schema_id,state,revision,target_generation_id,purpose,operation,operation_id,attempt_id,backup_set_id,recovery_state_catalog_sha256,codec_registry_sha256,database_sha256,object_store_sha256,reference_pack_storage_sha256,export_outputs_sha256,completion_evidence_sha256`

func targetControlValues(control TargetControl) []any {
	var database, objects, references, exports any
	if b := control.BindingDigests; b != nil {
		database, objects, references, exports = b.DatabaseSHA256, b.ObjectStoreSHA256, b.ReferencePackStorageSHA256, b.ExportOutputsSHA256
	}
	return []any{control.SchemaID, string(control.State), control.Revision, control.TargetGenerationID, control.Purpose, control.Operation, control.OperationID, control.AttemptID, control.BackupSetID, control.RecoveryStateCatalogSHA256, control.CodecRegistrySHA256, database, objects, references, exports, control.CompletionEvidenceSHA256}
}

func advanceTargetControl(ctx context.Context, capability TargetControlCapability, expected, next TargetControl) (TargetControl, error) {
	if capability == nil || capability.TargetDatabase() == nil {
		return TargetControl{}, ErrTargetIncomplete
	}
	var advanced TargetControl
	err := capability.WithControlTransaction(ctx, func(work context.Context, tx TargetControlTransaction) error {
		var err error
		advanced, err = advanceTargetControlInTransaction(work, tx, expected, next)
		return err
	})
	return advanced, err
}

func advanceTargetControlInTransaction(ctx context.Context, tx TargetControlTransaction, expected, next TargetControl) (TargetControl, error) {
	if err := expected.Validate(); err != nil {
		return TargetControl{}, err
	}
	if expected.Revision == math.MaxInt64 {
		return TargetControl{}, ErrTargetControlConflict
	}
	next.SchemaID = TargetStateSchemaID
	next.Revision = expected.Revision + 1
	if err := next.Validate(); err != nil {
		return TargetControl{}, err
	}
	if !validTargetTransition(expected, next) {
		return TargetControl{}, ErrTargetIncomplete
	}
	values := targetControlValues(next)
	count := len(values)
	assign, compare := make([]string, count), make([]string, count)
	for i := range count {
		assign[i] = fmt.Sprintf("$%d", i+1)
		compare[i] = fmt.Sprintf("$%d", count+i+1)
	}
	query := fmt.Sprintf(`UPDATE public.recovery_target_state SET (%s)=(%s)
 WHERE singleton AND (%s) IS NOT DISTINCT FROM (%s)
 AND (SELECT count(*) FROM public.recovery_target_state)=1`, targetControlColumns, strings.Join(assign, ","), targetControlColumns, strings.Join(compare, ","))
	values = append(values, targetControlValues(expected)...)
	tag, err := tx.Exec(ctx, query, values...)
	if err != nil {
		return TargetControl{}, err
	}
	if tag.RowsAffected() != 1 {
		return TargetControl{}, ErrTargetControlConflict
	}
	return next, nil
}

// ReconcileTargetCompletion performs only the terminal control transition.
// Restoring or already complete identity must match the authenticated proof;
// reset, initialized and unproven records cannot inherit old source history.
func ReconcileTargetCompletion(ctx context.Context, capability TargetControlCapability, proof TargetCompletionProof) error {
	if capability == nil || capability.TargetDatabase() == nil || !validSHA256Hex(proof.EvidenceSHA256) {
		return ErrTargetIncomplete
	}
	if err := proof.Identity.validate(); err != nil {
		return err
	}
	return capability.WithControlTransaction(ctx, func(work context.Context, tx TargetControlTransaction) error {
		current, err := readMatchingTargetCompletion(work, tx, proof)
		if err != nil {
			return err
		}
		if current.State == TargetComplete {
			return nil
		}
		next := current
		next.State = TargetComplete
		next.CompletionEvidenceSHA256 = targetControlText(proof.EvidenceSHA256)
		_, err = advanceTargetControlInTransaction(work, tx, current, next)
		return err
	})
}

// CheckTargetCompletion admits metadata repair before a portable journal is
// copied into its target database. It never changes target control state.
func CheckTargetCompletion(ctx context.Context, capability TargetControlCapability, proof TargetCompletionProof) error {
	if capability == nil || !validSHA256Hex(proof.EvidenceSHA256) {
		return ErrTargetIncomplete
	}
	if err := proof.Identity.validate(); err != nil {
		return err
	}
	return capability.WithControlTransaction(ctx, func(work context.Context, tx TargetControlTransaction) error {
		_, err := readMatchingTargetCompletion(work, tx, proof)
		return err
	})
}

func readMatchingTargetCompletion(ctx context.Context, tx TargetControlTransaction, proof TargetCompletionProof) (TargetControl, error) {
	current, err := ReadTargetControl(ctx, tx)
	if err != nil {
		return TargetControl{}, err
	}
	if (current.State != TargetRestoring && current.State != TargetComplete) || !sameTargetOperation(current, proof.Identity.control()) {
		return TargetControl{}, ErrTargetIncomplete
	}
	if current.State == TargetComplete && *current.CompletionEvidenceSHA256 != proof.EvidenceSHA256 {
		return TargetControl{}, ErrTargetIncomplete
	}
	return current, nil
}

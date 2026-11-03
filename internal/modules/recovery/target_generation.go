package recovery

import (
	"context"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/platform/recoverystate"
)

// AdmitVerificationTargetGeneration reconciles an owner-issued marker with a
// pristine verification-ready control record. Incomplete states cannot renew.
func AdmitVerificationTargetGeneration(ctx context.Context, capability TargetControlCapability, target RestoreTarget, generationID uuid.UUID, bindings TargetBindingDigests, extensions *ExtensionBackupCatalog, catalog *recoverystate.Catalog) error {
	if capability == nil || capability.TargetDatabase() == nil || generationID == uuid.Nil || !bindings.valid() {
		return ErrTargetIncomplete
	}
	if err := capability.AssertHeld(); err != nil {
		return err
	}
	current, err := ReadTargetControl(ctx, capability.TargetDatabase())
	if err != nil {
		return err
	}
	if current.State != TargetVerificationReady || current.BindingDigests == nil || *current.BindingDigests != bindings {
		return ErrTargetIncomplete
	}
	if *current.TargetGenerationID == generationID.String() {
		return nil
	}
	target.Postgres = capability.TargetDatabase()
	tx, err := beginRestoreTargetTransaction(ctx, target.Postgres, catalog)
	if err != nil {
		return err
	}
	defer func() { _ = rollbackRestoreTransaction(tx) }()
	if err := requireEmptyRestoreTargetWithReader(ctx, target, tx, extensions, catalog); err != nil {
		return err
	}
	next := current
	next.TargetGenerationID = targetControlText(generationID.String())
	_, err = advanceTargetControl(ctx, capability, current, next)
	return err
}

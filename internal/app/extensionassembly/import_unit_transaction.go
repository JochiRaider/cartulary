package extensionassembly

import (
	"context"

	"github.com/JochiRaider/cartulary/internal/modules/crossownertransaction"
	"github.com/JochiRaider/cartulary/internal/modules/imports"
)

func (access crossOwnerPreconditions) ValidateImportUnit(ctx context.Context, unit *imports.AnalyticalUnitCommit) error {
	t := access.transaction
	if t == nil || t.closed || t.analyticalUnit != nil {
		return crossownertransaction.ErrUnavailable
	}
	if err := imports.ValidateAnalyticalUnitTx(ctx, t.tx, unit); err != nil {
		return err
	}
	t.analyticalUnit = unit
	return nil
}

func (access crossOwnerFinalization) PublishImportUnit(ctx context.Context, unit *imports.AnalyticalUnitCommit, values map[string]any) error {
	t := access.transaction
	if t == nil || t.closed || unit == nil || t.analyticalUnit != unit || t.commitFinalized != nil || t.proveFinalizedCommit != nil {
		return crossownertransaction.ErrWrite
	}
	if err := imports.PublishAnalyticalUnitTx(ctx, t.tx, unit, values); err != nil {
		return err
	}
	t.proveFinalizedCommit = func(proofCtx context.Context) bool { return imports.ProveAnalyticalUnitCommit(proofCtx, unit) }
	return nil
}

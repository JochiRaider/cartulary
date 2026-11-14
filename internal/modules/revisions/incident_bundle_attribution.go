package revisions

import (
	"context"

	"github.com/JochiRaider/cartulary/internal/modules/incidentportability"
)

func (p preparedRevisionsImport) validateAttributions(ctx context.Context, recorder incidentportability.AttributionRecorder) error {
	if recorder == nil {
		return revisionsFailure(revisionsActorsInvariant)
	}
	var expectedCount, actualCount uint64
	if err := p.walkChangeSets(ctx, func(row portableChangeSet) error {
		expectedCount++
		present := false
		if err := recorder.VisitImportedAttributions("change_sets", row.ChangeSetID.String(), func(actual incidentportability.ImportedAttribution) error {
			if present || actual.SourceTable != "change_sets" || actual.SourceRowID != row.ChangeSetID.String() ||
				actual.SourceColumn != "actor_user_id" || actual.SourceActorID != row.PortableActorID.String() || actual.LocalUserID != p.runtimeActor {
				return revisionsFailure(revisionsActorsInvariant)
			}
			present = true
			return ctx.Err()
		}); err != nil {
			return err
		}
		if !present {
			return revisionsFailure(revisionsActorsInvariant)
		}
		return nil
	}); err != nil {
		return err
	}
	if err := recorder.VisitImportedAttributions("change_sets", "", func(incidentportability.ImportedAttribution) error {
		actualCount++
		return ctx.Err()
	}); err != nil {
		return err
	}
	if actualCount != expectedCount {
		return revisionsFailure(revisionsActorsInvariant)
	}
	return nil
}

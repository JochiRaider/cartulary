package tasksdecisions

import conflicttokens "github.com/JochiRaider/cartulary/internal/modules/revisions/conflicts"

import "github.com/JochiRaider/cartulary/internal/modules/tasksdecisions/internal/sourcecatalog"

func taskDecisionConflictSnapshotProjector(catalog *sourcecatalog.Catalog, viewSchemaID string) (conflicttokens.RevisionSnapshotProjector, bool) {
	surface, ok := catalog.SurfaceByViewID(viewSchemaID)
	if !ok {
		return conflicttokens.RevisionSnapshotProjector{}, false
	}
	projector, err := conflicttokens.NewRevisionSnapshotProjector(
		surface.RevisionSnapshotSchemaID,
		catalog.ConflictFieldSourceKeys(viewSchemaID),
	)
	if err != nil {
		panic(err)
	}
	return projector, true
}

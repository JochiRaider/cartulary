package artifacts

import (
	"github.com/JochiRaider/cartulary/internal/modules/artifacts/deleterestore"
	"github.com/JochiRaider/cartulary/internal/modules/artifacts/rollbackprovider"
	"github.com/JochiRaider/cartulary/internal/modules/artifacts/surfacecatalog"
	"github.com/JochiRaider/cartulary/internal/modules/revisions"
)

func RevisionProviderContribution() revisions.ProviderContribution {
	routes := make([]revisions.RecordViewRouteContribution, 0, 8)
	for _, surface := range surfacecatalog.All() {
		routes = append(routes, recordViewRoute(surface))
	}
	return revisions.ProviderContribution{
		SourceOwnerModule:     revisions.SourceOwnerArtifacts,
		ConflictFieldProvider: revisions.NewViewSchemaConflictFieldProvider(),
		Records: []revisions.RecordProviderContribution{{
			SourceOwnerModule:      revisions.SourceOwnerArtifacts,
			RecordType:             "artifact",
			DeleteRestoreSource:    deleterestore.NewSource(),
			RowRollbackProvider:    rollbackprovider.NewProvider(),
			LiveRecordChangePolicy: revisions.LiveRecordChangeRequired,
			RecordViewRoutes:       routes,
		}},
	}
}

func recordViewRoute(surface surfacecatalog.Surface) revisions.RecordViewRouteContribution {
	return revisions.RecordViewRouteContribution{
		ContributionID: "artifacts." + viewSchemaKey(surface.ViewSchemaID),
		Variant: &revisions.RecordVariant{
			Kind:  "artifact_type",
			Value: surface.ArtifactType,
		},
		ViewSchemaIDs: []string{surface.ViewSchemaID},
	}
}

func viewSchemaKey(viewSchemaID string) string {
	const (
		prefix = "cartulary.view."
		suffix = ".v1"
	)
	return viewSchemaID[len(prefix) : len(viewSchemaID)-len(suffix)]
}

package artifacts

import "github.com/JochiRaider/cartulary/internal/modules/artifacts/surfacecatalog"

const (
	CommLogViewSchemaID              = surfacecatalog.CommLogViewSchemaID
	FindingsViewSchemaID             = surfacecatalog.FindingsViewSchemaID
	ForensicKeywordsViewSchemaID     = surfacecatalog.ForensicKeywordsViewSchemaID
	HandoffViewSchemaID              = surfacecatalog.HandoffViewSchemaID
	InvestigativeQueriesViewSchemaID = surfacecatalog.InvestigativeQueriesViewSchemaID
	LessonViewSchemaID               = surfacecatalog.LessonViewSchemaID
	NotesViewSchemaID                = surfacecatalog.NotesViewSchemaID
	StatusReviewViewSchemaID         = surfacecatalog.StatusReviewViewSchemaID
)

func ArtifactTypeForView(viewSchemaID string) string {
	surface, ok := surfacecatalog.Lookup(viewSchemaID)
	if !ok {
		return ""
	}
	return surface.ArtifactType
}

func IsArtifactBackedView(viewSchemaID string) bool {
	return ArtifactTypeForView(viewSchemaID) != ""
}

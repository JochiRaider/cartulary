package artifacts

import (
	"github.com/JochiRaider/cartulary/internal/gen/contractartifacts"
	"github.com/JochiRaider/cartulary/internal/modules/artifacts/internal/sourcecatalog"
)

const (
	CommLogViewSchemaID              = contractartifacts.CommLogViewSchemaID
	FindingsViewSchemaID             = contractartifacts.FindingViewSchemaID
	ForensicKeywordsViewSchemaID     = contractartifacts.ForensicKeywordViewSchemaID
	HandoffViewSchemaID              = contractartifacts.HandoffViewSchemaID
	InvestigativeQueriesViewSchemaID = contractartifacts.InvestigativeQueryViewSchemaID
	LessonViewSchemaID               = contractartifacts.LessonViewSchemaID
	NotesViewSchemaID                = contractartifacts.NoteViewSchemaID
	StatusReviewViewSchemaID         = contractartifacts.StatusReviewViewSchemaID
)

func artifactTypeForView(viewSchemaID string) string {
	catalog, err := sourcecatalog.Load()
	if err != nil {
		return ""
	}
	surface, ok := catalog.SurfaceByViewID(viewSchemaID)
	if !ok {
		return ""
	}
	return surface.ArtifactType
}

func isArtifactBackedView(viewSchemaID string) bool {
	return artifactTypeForView(viewSchemaID) != ""
}

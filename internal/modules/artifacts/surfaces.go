package artifacts

import (
	"fmt"

	"github.com/JochiRaider/cartulary/internal/platform/viewschema"
)

const (
	CommLogViewSchemaID              = "cartulary.view.comm_log.v1"
	FindingsViewSchemaID             = "cartulary.view.findings.v1"
	ForensicKeywordsViewSchemaID     = "cartulary.view.forensic_keywords.v1"
	HandoffViewSchemaID              = "cartulary.view.handoff.v1"
	InvestigativeQueriesViewSchemaID = "cartulary.view.investigative_queries.v1"
	LessonViewSchemaID               = "cartulary.view.lesson.v1"
	NotesViewSchemaID                = "cartulary.view.notes.v1"
	StatusReviewViewSchemaID         = "cartulary.view.status_review.v1"
)

func ArtifactTypeForView(viewSchemaID string) string {
	switch viewSchemaID {
	case NotesViewSchemaID:
		return ArtifactTypeForSurface(viewSchemaID, "note")
	case CommLogViewSchemaID, HandoffViewSchemaID, StatusReviewViewSchemaID, LessonViewSchemaID:
		return ArtifactTypeForSurface(viewSchemaID, "")
	case FindingsViewSchemaID:
		return ArtifactTypeForSurface(viewSchemaID, "finding")
	case InvestigativeQueriesViewSchemaID:
		return ArtifactTypeForSurface(viewSchemaID, "investigative_query")
	case ForensicKeywordsViewSchemaID:
		return ArtifactTypeForSurface(viewSchemaID, "forensic_keyword")
	default:
		return ""
	}
}

func ArtifactTypeForSurface(viewSchemaID string, fallbackArtifactType string) string {
	schema, ok := viewschema.Lookup(viewSchemaID)
	if ok {
		if filter, hasFilter := schema.CanonicalSourceFilter(); hasFilter {
			if schema.BaseProjection != "artifact_grid_projection" {
				panic(fmt.Sprintf("artifact surface %s declares base_projection=%q", viewSchemaID, schema.BaseProjection))
			}
			if filter.Kind != "artifact_type" || filter.Field != "artifact_type" || filter.Value == "" {
				panic(fmt.Sprintf("artifact surface %s declares invalid canonical source filter %#v", viewSchemaID, filter))
			}
			if fallbackArtifactType != "" && fallbackArtifactType != filter.Value {
				panic(fmt.Sprintf("artifact surface %s fallback artifact_type=%q contradicts contract value %q", viewSchemaID, fallbackArtifactType, filter.Value))
			}
			return filter.Value
		}
	}
	if fallbackArtifactType == "" {
		panic(fmt.Sprintf("artifact surface %s missing canonical artifact_type filter", viewSchemaID))
	}
	return fallbackArtifactType
}

func IsArtifactBackedView(viewSchemaID string) bool {
	return ArtifactTypeForView(viewSchemaID) != ""
}

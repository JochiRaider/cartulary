package surfacecatalog

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

var registeredViewSchemaIDs = [...]string{
	CommLogViewSchemaID,
	FindingsViewSchemaID,
	ForensicKeywordsViewSchemaID,
	HandoffViewSchemaID,
	InvestigativeQueriesViewSchemaID,
	LessonViewSchemaID,
	NotesViewSchemaID,
	StatusReviewViewSchemaID,
}

type Surface struct {
	ViewSchemaID string
	ArtifactType string
}

// All resolves the explicit artifact-owned surface allowlist against the
// generated view-schema registry. Generated schemas never acquire artifact
// source ownership merely by declaring an artifact-like filter.
func All() []Surface {
	result := make([]Surface, 0, len(registeredViewSchemaIDs))
	for _, viewSchemaID := range registeredViewSchemaIDs {
		surface, ok := Lookup(viewSchemaID)
		if !ok {
			panic(fmt.Sprintf("registered artifact surface %s is unavailable", viewSchemaID))
		}
		result = append(result, surface)
	}
	return result
}

// Lookup admits only the explicit artifact-owned surface set. Its discriminator
// is derived from the generated projection contract and has no fallback.
func Lookup(viewSchemaID string) (Surface, bool) {
	if !isRegistered(viewSchemaID) {
		return Surface{}, false
	}
	schema, ok := viewschema.Lookup(viewSchemaID)
	if !ok {
		panic(fmt.Sprintf("registered artifact surface %s has no view-schema contract", viewSchemaID))
	}
	if schema.BaseProjection != "artifact_grid_projection" {
		panic(fmt.Sprintf(
			"artifact surface %s declares base_projection=%q",
			viewSchemaID,
			schema.BaseProjection,
		))
	}
	filter, ok := schema.CanonicalSourceFilter()
	if !ok || filter.Kind != "artifact_type" || filter.Field != "artifact_type" || filter.Value == "" {
		panic(fmt.Sprintf(
			"artifact surface %s declares invalid canonical source filter %#v",
			viewSchemaID,
			filter,
		))
	}
	return Surface{ViewSchemaID: viewSchemaID, ArtifactType: filter.Value}, true
}

func LookupByArtifactType(artifactType string) (Surface, bool) {
	var found Surface
	for _, viewSchemaID := range registeredViewSchemaIDs {
		surface, ok := Lookup(viewSchemaID)
		if !ok || surface.ArtifactType != artifactType {
			continue
		}
		if found.ViewSchemaID != "" {
			panic(fmt.Sprintf(
				"artifact type %q is registered by both %s and %s",
				artifactType,
				found.ViewSchemaID,
				surface.ViewSchemaID,
			))
		}
		found = surface
	}
	return found, found.ViewSchemaID != ""
}

func isRegistered(viewSchemaID string) bool {
	for _, registered := range registeredViewSchemaIDs {
		if viewSchemaID == registered {
			return true
		}
	}
	return false
}

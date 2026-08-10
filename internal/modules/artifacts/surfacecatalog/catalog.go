package surfacecatalog

import (
	"fmt"

	"github.com/JochiRaider/cartulary/internal/gen/contractartifacts"
	"github.com/JochiRaider/cartulary/internal/platform/viewschema"
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

type Surface struct {
	ViewSchemaID string
	ArtifactType string
}

// All resolves the generated artifact-owned surface catalog against the live
// generated view-schema registry. It has no handwritten allowlist or fallback.
func All() []Surface {
	result := make([]Surface, 0, len(contractartifacts.SourceCatalog))
	for _, generated := range contractartifacts.SourceCatalog {
		surface, ok := Lookup(generated.ViewSchemaID)
		if !ok {
			panic(fmt.Sprintf("generated artifact surface %s is unavailable", generated.ViewSchemaID))
		}
		result = append(result, surface)
	}
	return result
}

func Lookup(viewSchemaID string) (Surface, bool) {
	generated, ok := lookupGeneratedSurface(viewSchemaID)
	if !ok {
		return Surface{}, false
	}
	schema, ok := viewschema.Lookup(viewSchemaID)
	if !ok {
		panic(fmt.Sprintf("generated artifact surface %s has no view-schema contract", viewSchemaID))
	}
	if schema.BaseProjection != "artifact_grid_projection" {
		panic(fmt.Sprintf("artifact surface %s declares base_projection=%q", viewSchemaID, schema.BaseProjection))
	}
	filter, ok := schema.CanonicalSourceFilter()
	if !ok || filter.Kind != "artifact_type" || filter.Field != "artifact_type" || filter.Value != generated.ArtifactType {
		panic(fmt.Sprintf("artifact surface %s declares invalid canonical source filter %#v", viewSchemaID, filter))
	}
	return Surface{ViewSchemaID: generated.ViewSchemaID, ArtifactType: generated.ArtifactType}, true
}

func LookupByArtifactType(artifactType string) (Surface, bool) {
	var found Surface
	for _, generated := range contractartifacts.SourceCatalog {
		if generated.ArtifactType != artifactType {
			continue
		}
		if found.ViewSchemaID != "" {
			panic(fmt.Sprintf("artifact type %q is registered by both %s and %s", artifactType, found.ViewSchemaID, generated.ViewSchemaID))
		}
		found, _ = Lookup(generated.ViewSchemaID)
	}
	return found, found.ViewSchemaID != ""
}

func lookupGeneratedSurface(viewSchemaID string) (contractartifacts.SourceSurface, bool) {
	for _, surface := range contractartifacts.SourceCatalog {
		if surface.ViewSchemaID == viewSchemaID {
			return surface, true
		}
	}
	return contractartifacts.SourceSurface{}, false
}

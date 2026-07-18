package workbook

import (
	"testing"

	"github.com/JochiRaider/cartulary/internal/platform/viewschema"
)

func TestCoordinationArtifactSurfacesUseContractFilters(t *testing.T) {
	tests := []struct {
		viewSchemaID string
		artifactType string
	}{
		{viewSchemaID: CommLogViewSchemaID, artifactType: "comm_log"},
		{viewSchemaID: HandoffViewSchemaID, artifactType: "handoff"},
		{viewSchemaID: StatusReviewViewSchemaID, artifactType: "status_review"},
		{viewSchemaID: LessonViewSchemaID, artifactType: "lesson"},
	}

	for _, tc := range tests {
		t.Run(tc.viewSchemaID, func(t *testing.T) {
			schema, ok := viewschema.Lookup(tc.viewSchemaID)
			if !ok {
				t.Fatalf("missing view schema %s", tc.viewSchemaID)
			}
			filter, ok := schema.CanonicalSourceFilter()
			if !ok {
				t.Fatalf("missing canonical source filter for %s", tc.viewSchemaID)
			}
			if schema.BaseProjection != "artifact_grid_projection" ||
				filter.Kind != "artifact_type" ||
				filter.Field != "artifact_type" ||
				filter.Value != tc.artifactType {
				t.Fatalf("%s artifact filter mismatch: base=%q filter=%#v", tc.viewSchemaID, schema.BaseProjection, filter)
			}
		})
	}
}

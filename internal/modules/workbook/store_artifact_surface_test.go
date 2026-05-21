package workbook

import (
	"strings"
	"testing"
)

func TestSupportPhase9CoordinationArtifactSurfacesUseContractFilters(t *testing.T) {
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
			if got := artifactTypeForSurface(tc.viewSchemaID, ""); got != tc.artifactType {
				t.Fatalf("%s artifact type: got %q want %q", tc.viewSchemaID, got, tc.artifactType)
			}
			surface, ok := genericSurfaces[tc.viewSchemaID]
			if !ok {
				t.Fatalf("missing generic surface %s", tc.viewSchemaID)
			}
			if !strings.Contains(surface.whereSQL, "p.artifact_type = '"+tc.artifactType+"'") {
				t.Fatalf("%s whereSQL does not use contract artifact filter: %q", tc.viewSchemaID, surface.whereSQL)
			}
		})
	}
}

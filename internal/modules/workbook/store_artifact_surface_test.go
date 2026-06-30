package workbook

import (
	"strings"
	"testing"

	"github.com/JochiRaider/cartulary/internal/modules/projections"
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
			if got := projections.ArtifactTypeForSurface(tc.viewSchemaID, ""); got != tc.artifactType {
				t.Fatalf("%s artifact type: got %q want %q", tc.viewSchemaID, got, tc.artifactType)
			}
			whereSQL, ok := projections.SurfaceWhereSQLForTesting(tc.viewSchemaID)
			if !ok {
				t.Fatalf("missing generic surface %s", tc.viewSchemaID)
			}
			if !strings.Contains(whereSQL, "p.artifact_type = '"+tc.artifactType+"'") {
				t.Fatalf("%s whereSQL does not use contract artifact filter: %q", tc.viewSchemaID, whereSQL)
			}
		})
	}
}

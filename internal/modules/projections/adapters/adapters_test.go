package adapters

import (
	"testing"

	"github.com/JochiRaider/cartulary/internal/modules/recovery/restorecontract"
)

func TestRestoreRebuilderReturnsRecoveryContract(t *testing.T) {
	rebuilder := NewRestoreRebuilder(nil)
	var _ restorecontract.ProjectionRebuilder = rebuilder
	if rebuilder == nil {
		t.Fatal("restore rebuilder adapter returned nil")
	}
}

func TestWorkbookRowsSupportProjectionQuerySurfaces(t *testing.T) {
	rows := NewWorkbookRows(nil)
	for _, viewSchemaID := range []string{
		AssessmentsViewSchemaID,
		EvidenceViewSchemaID,
		NotesViewSchemaID,
		PartiesViewSchemaID,
		TaskRequestsViewSchemaID,
		DecisionsViewSchemaID,
		CommLogViewSchemaID,
		HandoffViewSchemaID,
		StatusReviewViewSchemaID,
		LessonViewSchemaID,
		FindingsViewSchemaID,
		InvestigativeQueriesViewSchemaID,
		ForensicKeywordsViewSchemaID,
	} {
		if !rows.Supports(viewSchemaID) {
			t.Fatalf("workbook rows adapter must support %s", viewSchemaID)
		}
	}
}

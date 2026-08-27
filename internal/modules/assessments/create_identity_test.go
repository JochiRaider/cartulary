package assessments

import (
	"encoding/hex"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestAssessmentCreateIdentityCompatibility(t *testing.T) {
	t.Parallel()
	score := 73
	assessor := uuid.MustParse("40000000-0000-4000-8000-000000000004")
	assessedAt := time.Date(2026, 8, 20, 16, 34, 56, 123456789, time.UTC)
	input := CreateInput{
		ClientTxnID:     "txn-assessment-owner",
		SubjectRef:      uuid.MustParse("10000000-0000-4000-8000-000000000001"),
		SubjectType:     "host",
		AssessmentState: "confirmed",
		ConfidenceScore: &score,
		Rationale:       "First\nSecond",
		Assessor:        &assessor,
		AssessedAt:      &assessedAt,
		SupportRefs: []uuid.UUID{
			uuid.MustParse("30000000-0000-4000-8000-000000000003"),
			uuid.MustParse("20000000-0000-4000-8000-000000000002"),
		},
	}
	command := CreateCommand{
		ActorUserID: uuid.MustParse("50000000-0000-4000-8000-000000000005"),
		IncidentID:  uuid.MustParse("60000000-0000-4000-8000-000000000006"),
		Input:       input,
	}

	key := createIdempotencyKey(command)
	if key.RouteKey != "assessments.rows.create" ||
		key.ActorUserID != command.ActorUserID ||
		key.ScopeKey != command.IncidentID.String()+":"+AssessmentsViewSchemaID ||
		key.ClientTxnID != input.ClientTxnID {
		t.Fatalf("Assessment create key = %#v", key)
	}
	const golden = "6406c647a1b4e4adc65a4161ac1b168775e88376860a1e0bcd6d3f9e699055fb"
	if got := hex.EncodeToString(key.RequestHash); got != golden {
		t.Fatalf("Assessment create hash = %s, want %s", got, golden)
	}

	input.SupportRefs[0], input.SupportRefs[1] = input.SupportRefs[1], input.SupportRefs[0]
	if got := hex.EncodeToString(createRequestHash(input)); got != golden {
		t.Fatalf("reordered Assessment create hash = %s, want %s", got, golden)
	}
}

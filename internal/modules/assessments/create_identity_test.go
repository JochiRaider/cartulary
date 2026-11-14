package assessments

import (
	"bytes"
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
	if got := hex.EncodeToString(legacyCreateRequestHash(input)); got != golden {
		t.Fatalf("Assessment create hash = %s, want %s", got, golden)
	}

	input.SupportRefs[0], input.SupportRefs[1] = input.SupportRefs[1], input.SupportRefs[0]
	if got := hex.EncodeToString(legacyCreateRequestHash(input)); got != golden {
		t.Fatalf("reordered Assessment create hash = %s, want %s", got, golden)
	}

	t.Run("normalized fields and original defaults", func(t *testing.T) {
		defaultTime := time.Date(2026, 9, 1, 12, 0, 0, 123456000, time.UTC)
		original := createRequestHash(input, command.ActorUserID, defaultTime)
		// Independently authored normalized field fixture; no route/key members.
		const normalizedGolden = "887783468fe7de17c6d7c804bde90ae95de77eb665c853ebf4ae7b52fd81a16f"
		if got := hex.EncodeToString(original); got != normalizedGolden {
			t.Fatalf("normalized hash = %s, want %s", got, normalizedGolden)
		}
		equivalent := input
		equivalent.ClientTxnID = "different-key"
		equivalent.SupportRefs = append([]uuid.UUID{input.SupportRefs[1]}, input.SupportRefs...)
		canonicalTime := assessedAt.Truncate(time.Microsecond).In(time.FixedZone("offset", 3600))
		equivalent.AssessedAt = &canonicalTime
		if !bytes.Equal(original, createRequestHash(equivalent, command.ActorUserID, defaultTime.Add(time.Hour))) {
			t.Fatal("set, key exclusion or persisted timestamp normalization differs")
		}

		omitted := input
		omitted.Assessor, omitted.AssessedAt, omitted.ConfidenceScore, omitted.SupportRefs = nil, nil, nil, nil
		explicit := omitted
		explicit.Assessor, explicit.AssessedAt = &command.ActorUserID, &defaultTime
		explicit.SupportRefs = []uuid.UUID{}
		if !bytes.Equal(createRequestHash(omitted, command.ActorUserID, defaultTime), createRequestHash(explicit, command.ActorUserID, defaultTime.Add(time.Hour))) {
			t.Fatal("original create defaults did not compare equal")
		}
		if bytes.Equal(createRequestHash(omitted, command.ActorUserID, defaultTime), createRequestHash(omitted, command.ActorUserID, defaultTime.Add(time.Hour))) {
			t.Fatal("omitted time ignored original default context")
		}
		if bytes.Equal(original, createRequestHash(omitted, command.ActorUserID, defaultTime)) {
			t.Fatal("explicit fields were replaced by defaults")
		}

		for name, change := range map[string]func(*CreateInput){
			"subject":   func(v *CreateInput) { v.SubjectRef = command.ActorUserID },
			"type":      func(v *CreateInput) { v.SubjectType = "identity" },
			"state":     func(v *CreateInput) { v.AssessmentState = "suspected" },
			"score":     func(v *CreateInput) { v.ConfidenceScore = nil },
			"rationale": func(v *CreateInput) { v.Rationale = "Different" },
			"assessor":  func(v *CreateInput) { v.Assessor = &command.ActorUserID },
			"time":      func(v *CreateInput) { v.AssessedAt = &defaultTime },
			"support":   func(v *CreateInput) { v.SupportRefs = nil },
		} {
			t.Run(name, func(t *testing.T) {
				changed := input
				change(&changed)
				if bytes.Equal(original, createRequestHash(changed, command.ActorUserID, defaultTime)) {
					t.Fatal("different authoritative field compared equal")
				}
			})
		}
	})
}

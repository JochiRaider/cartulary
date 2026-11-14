package assessments

import (
	"crypto/sha256"
	"encoding/json"
	"slices"
	"time"

	"github.com/google/uuid"
)

const createRouteKey = "assessments.rows.create"

func createIdempotencyKey(command CreateCommand) CreateIdempotencyKey {
	return CreateIdempotencyKey{
		RouteKey:    createRouteKey,
		ActorUserID: command.ActorUserID,
		ScopeKey:    command.IncidentID.String() + ":" + AssessmentsViewSchemaID,
		ClientTxnID: command.Input.ClientTxnID,
	}
}

// legacyCreateRequestHash is only for receipts written before normalized replay.
// Its admitted representation cannot be recovered from a returned/current row.
func legacyCreateRequestHash(input CreateInput) []byte {
	payload := map[string]any{
		"view_schema_id": AssessmentsViewSchemaID,
		"client_txn_id":  input.ClientTxnID,
	}
	if input.SubjectRef != uuid.Nil {
		payload["assessment.subject_ref"] = input.SubjectRef.String()
	}
	if input.SubjectType != "" {
		payload["assessment.subject_type"] = input.SubjectType
	}
	if input.AssessmentState != "" {
		payload["assessment.assessment_state"] = input.AssessmentState
	}
	if input.ConfidenceScore != nil {
		payload["assessment.confidence_score"] = *input.ConfidenceScore
	}
	if input.Rationale != "" {
		payload["assessment.rationale"] = input.Rationale
	}
	if input.Assessor != nil {
		payload["assessment.assessor"] = input.Assessor.String()
	}
	if input.AssessedAt != nil {
		payload["assessment.assessed_at"] = input.AssessedAt.UTC().Format(time.RFC3339Nano)
	}
	if len(input.SupportRefs) > 0 {
		refs := make([]string, 0, len(input.SupportRefs))
		for _, ref := range input.SupportRefs {
			refs = append(refs, ref.String())
		}
		slices.Sort(refs)
		payload["assessment.support_refs"] = refs
	}
	data, _ := json.Marshal(payload)
	sum := sha256.Sum256(data)
	return append([]byte(nil), sum[:]...)
}

func createRequestHash(input CreateInput, actorID uuid.UUID, defaultAssessedAt time.Time) []byte {
	assessor := actorID
	if input.Assessor != nil {
		assessor = *input.Assessor
	}
	assessedAt := defaultAssessedAt
	if input.AssessedAt != nil {
		assessedAt = *input.AssessedAt
	}
	payload := map[string]any{
		"assessment.subject_ref":      input.SubjectRef.String(),
		"assessment.subject_type":     input.SubjectType,
		"assessment.assessment_state": input.AssessmentState,
		"assessment.rationale":        input.Rationale,
		"assessment.assessor":         assessor.String(),
		// PostgreSQL stores timestamptz at microsecond precision.
		"assessment.assessed_at": assessedAt.UTC().Truncate(time.Microsecond).Format(time.RFC3339Nano),
	}
	if input.ConfidenceScore != nil {
		payload["assessment.confidence_score"] = *input.ConfidenceScore
	}
	if len(input.SupportRefs) > 0 {
		refs := make([]string, 0, len(input.SupportRefs))
		for _, ref := range uniqueUUIDs(input.SupportRefs) {
			refs = append(refs, ref.String())
		}
		slices.Sort(refs)
		payload["assessment.support_refs"] = refs
	}
	data, _ := json.Marshal(payload)
	sum := sha256.Sum256(data)
	return append([]byte(nil), sum[:]...)
}

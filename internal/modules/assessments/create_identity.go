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
		RequestHash: createRequestHash(command.Input),
	}
}

func createRequestHash(input CreateInput) []byte {
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

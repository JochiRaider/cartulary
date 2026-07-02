package merge

import (
	"crypto/sha256"
	"encoding/json"
	"io"
	"strings"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/platform/httpapi"
)

const mergeRouteKey = "entities.records.merge"

type MergeRequest struct {
	LoserRecordID          uuid.UUID
	SurvivorBaseRowVersion int64
	LoserBaseRowVersion    int64
	ClientTxnID            string
	Reason                 *string
}

type MergeExactMatchClassSummary struct {
	IdentifierClass    string `json:"identifier_class"`
	PromotedCount      int    `json:"promoted_count"`
	CarriedCount       int    `json:"carried_count"`
	DuplicateNoopCount int    `json:"duplicate_noop_count"`
	BlockedConflict    int    `json:"blocked_conflict_count"`
	ProvenanceOnly     int    `json:"provenance_only_count"`
	SuggestionOnly     int    `json:"suggestion_only_count"`
}

type MergeSummary struct {
	RecordType                    string                        `json:"record_type"`
	RepointedMentionResolutionCnt int                           `json:"repointed_mention_resolution_count"`
	RepointedLinkCount            int                           `json:"repointed_link_count"`
	DedupedLinkCount              int                           `json:"deduped_link_count"`
	RepointedTagCount             int                           `json:"repointed_tag_count"`
	DedupedTagCount               int                           `json:"deduped_tag_count"`
	RepointedAssessmentCount      int                           `json:"repointed_assessment_count"`
	ExactMatchClasses             []MergeExactMatchClassSummary `json:"exact_match_classes"`
}

type MergeResult struct {
	Payload               map[string]any
	StatusCode            int
	Replayed              bool
	IncidentID            uuid.UUID
	RecordType            string
	SurvivorRecordID      uuid.UUID
	SurvivorRowVersion    int64
	LoserRecordID         uuid.UUID
	LoserRowVersion       int64
	ChangeSetID           uuid.UUID
	MergeSummary          MergeSummary
	TimelineInvalidations []MergeTimelineInvalidation
}

type MergeTimelineInvalidation struct {
	RecordID         uuid.UUID
	RowVersion       int64
	ChangedFieldKeys []string
}

func DecodeMergeRequest(reader io.Reader) (MergeRequest, *httpapi.APIError) {
	raw, apiErr := decodeObject(reader)
	if apiErr != nil {
		return MergeRequest{}, apiErr
	}

	allowed := map[string]struct{}{
		"loser_record_id":           {},
		"survivor_base_row_version": {},
		"loser_base_row_version":    {},
		"client_txn_id":             {},
		"reason":                    {},
	}
	for key := range raw {
		if _, ok := allowed[key]; !ok {
			return MergeRequest{}, invalidMutationPayload(key, "unknown_field")
		}
	}

	var request MergeRequest
	if value, ok := raw["loser_record_id"]; !ok {
		return MergeRequest{}, invalidMutationPayload("loser_record_id", "missing_required_field")
	} else {
		var rawID string
		if err := json.Unmarshal(value, &rawID); err != nil {
			return MergeRequest{}, invalidMutationPayload("loser_record_id", "invalid_value")
		}
		recordID, err := uuid.Parse(rawID)
		if err != nil {
			return MergeRequest{}, invalidMutationPayload("loser_record_id", "invalid_value")
		}
		request.LoserRecordID = recordID
	}
	if value, ok := raw["survivor_base_row_version"]; !ok {
		return MergeRequest{}, invalidMutationPayload("survivor_base_row_version", "missing_required_field")
	} else if err := json.Unmarshal(value, &request.SurvivorBaseRowVersion); err != nil || request.SurvivorBaseRowVersion < 1 {
		return MergeRequest{}, invalidMutationPayload("survivor_base_row_version", "invalid_base_row_version")
	}
	if value, ok := raw["loser_base_row_version"]; !ok {
		return MergeRequest{}, invalidMutationPayload("loser_base_row_version", "missing_required_field")
	} else if err := json.Unmarshal(value, &request.LoserBaseRowVersion); err != nil || request.LoserBaseRowVersion < 1 {
		return MergeRequest{}, invalidMutationPayload("loser_base_row_version", "invalid_base_row_version")
	}
	if value, ok := raw["client_txn_id"]; !ok {
		return MergeRequest{}, invalidMutationPayload("client_txn_id", "missing_required_field")
	} else if err := json.Unmarshal(value, &request.ClientTxnID); err != nil || strings.TrimSpace(request.ClientTxnID) == "" {
		return MergeRequest{}, invalidMutationPayload("client_txn_id", "missing_required_field")
	}
	if value, ok := raw["reason"]; ok {
		if string(value) == "null" {
			request.Reason = nil
		} else {
			var rawReason string
			if err := json.Unmarshal(value, &rawReason); err != nil {
				return MergeRequest{}, invalidMutationPayload("reason", "invalid_value")
			}
			request.Reason = authn.NormalizeReasonNote(&rawReason)
		}
	}

	return request, nil
}

func MergeRequestHash(survivorRecordID uuid.UUID, request MergeRequest) []byte {
	payload := map[string]any{
		"survivor_record_id":        survivorRecordID.String(),
		"loser_record_id":           request.LoserRecordID.String(),
		"survivor_base_row_version": request.SurvivorBaseRowVersion,
		"loser_base_row_version":    request.LoserBaseRowVersion,
		"client_txn_id":             request.ClientTxnID,
		"reason":                    derefString(request.Reason),
	}
	data, _ := json.Marshal(payload)
	sum := sha256.Sum256(data)
	hash := make([]byte, len(sum))
	copy(hash, sum[:])
	return hash
}

func BuildMergePayload(result MergeResult) map[string]any {
	exactMatchClasses := make([]map[string]any, 0, len(result.MergeSummary.ExactMatchClasses))
	for _, summary := range result.MergeSummary.ExactMatchClasses {
		exactMatchClasses = append(exactMatchClasses, map[string]any{
			"identifier_class":       summary.IdentifierClass,
			"promoted_count":         summary.PromotedCount,
			"carried_count":          summary.CarriedCount,
			"duplicate_noop_count":   summary.DuplicateNoopCount,
			"blocked_conflict_count": summary.BlockedConflict,
			"provenance_only_count":  summary.ProvenanceOnly,
			"suggestion_only_count":  summary.SuggestionOnly,
		})
	}
	return map[string]any{
		"incident_id":           result.IncidentID.String(),
		"survivor_record_id":    result.SurvivorRecordID.String(),
		"loser_record_id":       result.LoserRecordID.String(),
		"survivor_row_version":  result.SurvivorRowVersion,
		"loser_row_version":     result.LoserRowVersion,
		"change_set_id":         result.ChangeSetID.String(),
		"merged_into_record_id": result.SurvivorRecordID.String(),
		"merge_summary": map[string]any{
			"record_type":                        result.MergeSummary.RecordType,
			"repointed_mention_resolution_count": result.MergeSummary.RepointedMentionResolutionCnt,
			"repointed_link_count":               result.MergeSummary.RepointedLinkCount,
			"deduped_link_count":                 result.MergeSummary.DedupedLinkCount,
			"repointed_tag_count":                result.MergeSummary.RepointedTagCount,
			"deduped_tag_count":                  result.MergeSummary.DedupedTagCount,
			"repointed_assessment_count":         result.MergeSummary.RepointedAssessmentCount,
			"exact_match_classes":                exactMatchClasses,
		},
	}
}

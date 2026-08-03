package revisions

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/google/uuid"
)

type historyRowMaterializer struct{}

func (historyRowMaterializer) Mutation(record RecordHistoryRecord, row mutationHistoryRow) RecordHistoryItem {
	return RecordHistoryItem{
		ActorUserID:              row.ActorUserID,
		CommittedAt:              row.CommittedAt,
		HistoryItemRef:           historyItemRefForMutation(record.RecordID, row.ChangeSetID, row.SequenceNo),
		Operation:                historyOperation(row.Source, row.OperationKind),
		DiffSummary:              mutationDiffSummary(row.TargetKind, row.TargetID, row.OperationKind, row.SequenceNo, row.BeforeValue, row.AfterValue),
		ChangeSetID:              row.ChangeSetID,
		AvailableRollbackActions: nil,
		HistoryEntryRef:          row.HistoryEntryRef,
		RevisionNo:               row.RevisionNo,
		createdAt:                row.CommittedAt,
		changeSetID:              row.ChangeSetID,
		sequenceNo:               row.SequenceNo,
		syntheticRank:            0,
		targetKey:                row.TargetKind + ":" + row.TargetID,
		hasTargetEntry:           singleEntryAddressable(row.TargetKind, row.TargetID, record.RecordID, row.BeforeValue, row.AfterValue),
	}
}

func (historyRowMaterializer) Revisions(record RecordHistoryRecord, rows []revisionHistoryRow, mutationItems []RecordHistoryItem) []RecordHistoryItem {
	changeSetsWithMutation := make(map[uuid.UUID]bool, len(mutationItems))
	for _, item := range mutationItems {
		changeSetsWithMutation[item.ChangeSetID] = true
	}
	items := make([]RecordHistoryItem, 0, len(rows))
	for _, row := range rows {
		if changeSetsWithMutation[row.ChangeSetID] {
			continue
		}
		revisionNo := row.RevisionNo
		items = append(items, RecordHistoryItem{
			ActorUserID:              row.ActorUserID,
			CommittedAt:              row.CommittedAt,
			HistoryItemRef:           historyItemRefForRevision(record.RecordID, row.ChangeSetID, row.RevisionNo),
			Operation:                historyOperation(row.Source, "row_revision"),
			DiffSummary:              revisionDiffSummary(record.RecordID, row.RevisionNo, row.BeforeValue, row.AfterValue),
			ChangeSetID:              row.ChangeSetID,
			AvailableRollbackActions: nil,
			RevisionNo:               &revisionNo,
			createdAt:                row.CommittedAt,
			changeSetID:              row.ChangeSetID,
			sequenceNo:               int(^uint(0) >> 1),
			syntheticRank:            1,
		})
	}
	return items
}

// historyPageAssembler owns the canonical newest-first order and resource
// projection consumed by the HTTP adapter's cursor/limit selection.
type historyPageAssembler struct{}

func (historyPageAssembler) Resources(items []RecordHistoryItem) []map[string]any {
	sort.SliceStable(items, func(i, j int) bool {
		left := items[i]
		right := items[j]
		if !left.createdAt.Equal(right.createdAt) {
			return left.createdAt.After(right.createdAt)
		}
		if left.changeSetID != right.changeSetID {
			return left.changeSetID.String() > right.changeSetID.String()
		}
		if left.syntheticRank != right.syntheticRank {
			return left.syntheticRank < right.syntheticRank
		}
		return left.sequenceNo < right.sequenceNo
	})
	resources := make([]map[string]any, 0, len(items))
	for _, item := range items {
		resources = append(resources, item.Resource())
	}
	return resources
}

func historyItemRefForMutation(recordID uuid.UUID, changeSetID uuid.UUID, sequenceNo int) string {
	return historyItemRef("mutation", recordID.String(), changeSetID.String(), fmt.Sprintf("%d", sequenceNo))
}

func historyItemRefForRevision(recordID uuid.UUID, changeSetID uuid.UUID, revisionNo int64) string {
	return historyItemRef("revision", recordID.String(), changeSetID.String(), fmt.Sprintf("%d", revisionNo))
}

func historyItemRef(parts ...string) string {
	digest := sha256.Sum256([]byte(strings.Join(parts, ":")))
	return "hitem_" + base64.RawURLEncoding.EncodeToString(digest[:])
}

func singleEntryAddressable(targetKind string, targetID string, recordID uuid.UUID, beforeValue []byte, afterValue []byte) bool {
	if targetID != recordID.String() {
		switch targetKind {
		case "record_link":
			return mutationJSONReferencesRecord(beforeValue, recordID, "src_record_id", "dst_record_id") ||
				mutationJSONReferencesRecord(afterValue, recordID, "src_record_id", "dst_record_id")
		case "entity_mention":
			return mutationJSONReferencesRecord(beforeValue, recordID, "source_record_id") ||
				mutationJSONReferencesRecord(afterValue, recordID, "source_record_id")
		case "record_tag":
			return mutationJSONReferencesRecord(beforeValue, recordID, "record_id") ||
				mutationJSONReferencesRecord(afterValue, recordID, "record_id")
		case "indicator_observation":
			return mutationJSONReferencesRecord(beforeValue, recordID, "source_record_id", "resolved_indicator_record_id") ||
				mutationJSONReferencesRecord(afterValue, recordID, "source_record_id", "resolved_indicator_record_id")
		case "indicator_state_interval":
			return mutationJSONReferencesRecord(beforeValue, recordID, "indicator_record_id") ||
				mutationJSONReferencesRecord(afterValue, recordID, "indicator_record_id")
		default:
			return false
		}
	}
	switch targetKind {
	case "record", "timeline_record", "host", "identity", "indicator", "assessment", "evidence":
		return true
	case "record_tag":
		return mutationJSONReferencesRecord(beforeValue, recordID, "record_id") ||
			mutationJSONReferencesRecord(afterValue, recordID, "record_id")
	default:
		return false
	}
}

func mutationJSONReferencesRecord(raw []byte, recordID uuid.UUID, keys ...string) bool {
	if len(raw) == 0 {
		return false
	}
	var value map[string]any
	if err := json.Unmarshal(raw, &value); err != nil {
		return false
	}
	recordIDText := recordID.String()
	for _, key := range keys {
		if text, ok := value[key].(string); ok && text == recordIDText {
			return true
		}
	}
	return false
}

func historyOperation(source string, operationKind string) string {
	if operationKind != "" {
		return operationKind
	}
	if source != "" {
		return source
	}
	return "unknown"
}

func mutationDiffSummary(targetKind string, targetID string, operationKind string, sequenceNo int, beforeValue []byte, afterValue []byte) map[string]any {
	return map[string]any{
		"summary": fmt.Sprintf("%s %s", historyOperation("", operationKind), targetKind),
		"units": []map[string]any{{
			"target_kind":       targetKind,
			"target_id":         targetID,
			"operation":         historyOperation("", operationKind),
			"sequence_no":       sequenceNo,
			"has_before_value":  len(beforeValue) > 0,
			"has_after_value":   len(afterValue) > 0,
			"history_unit_kind": "mutation",
		}},
	}
}

func revisionDiffSummary(recordID uuid.UUID, revisionNo int64, beforeValue []byte, afterValue []byte) map[string]any {
	return map[string]any{
		"summary": fmt.Sprintf("row revision %d", revisionNo),
		"units": []map[string]any{{
			"record_id":         recordID.String(),
			"revision_no":       revisionNo,
			"has_before_value":  len(beforeValue) > 0,
			"has_after_value":   len(afterValue) > 0,
			"history_unit_kind": "row_revision",
		}},
	}
}

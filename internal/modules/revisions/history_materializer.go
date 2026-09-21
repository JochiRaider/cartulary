package revisions

import (
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"sort"
	"strings"

	"github.com/JochiRaider/cartulary/internal/modules/revisions/historycontract"
	"github.com/JochiRaider/cartulary/internal/modules/revisions/rollbackcontract"
	"github.com/google/uuid"
)

type historyRowMaterializer struct{ catalog *TargetSemanticsCatalog }

func (materializer historyRowMaterializer) Mutation(record RecordHistoryRecord, row mutationHistoryRow) (RecordHistoryItem, error) {
	summary, err := materializer.catalog.projectHistory(record, row.TargetKind, row.TargetID, row.OperationKind, row.BeforeValue, row.AfterValue)
	if err != nil {
		return RecordHistoryItem{}, err
	}
	return RecordHistoryItem{
		ActorUserID:              row.ActorUserID,
		CommittedAt:              row.CommittedAt,
		HistoryItemRef:           historyItemRefForMutation(record.RecordID, row.ChangeSetID, row.SequenceNo),
		Operation:                historyOperation(row.Source, row.OperationKind),
		DiffSummary:              summary,
		ChangeSetID:              row.ChangeSetID,
		AvailableRollbackActions: nil,
		HistoryEntryRef:          row.HistoryEntryRef,
		RevisionNo:               row.RevisionNo,
		createdAt:                row.CommittedAt,
		changeSetID:              row.ChangeSetID,
		sequenceNo:               row.SequenceNo,
		syntheticRank:            0,
		targetKey:                row.TargetKind + ":" + row.TargetID,
		hasTargetEntry:           row.HistoryEntryAddressable,
		rowProjected:             materializer.catalog.byTargetKind[row.TargetKind].dispatchClass == rollbackcontract.DispatchRow,
	}, nil
}

func (materializer historyRowMaterializer) Revisions(record RecordHistoryRecord, rows []revisionHistoryRow, mutationItems []RecordHistoryItem) ([]RecordHistoryItem, error) {
	changeSetsWithMutation := make(map[uuid.UUID]int, len(mutationItems))
	changeSetsWithRow := make(map[uuid.UUID]bool, len(mutationItems))
	for i, item := range mutationItems {
		if _, exists := changeSetsWithMutation[item.ChangeSetID]; !exists {
			changeSetsWithMutation[item.ChangeSetID] = i
		}
		if item.rowProjected {
			changeSetsWithRow[item.ChangeSetID] = true
		}
	}
	items := make([]RecordHistoryItem, 0, len(rows))
	for _, row := range rows {
		revisionNo := row.RevisionNo
		summary, err := materializer.catalog.projectHistory(record, "record", record.RecordID.String(), "row_revision", row.BeforeValue, row.AfterValue)
		if err != nil {
			return nil, err
		}
		if index, exists := changeSetsWithMutation[row.ChangeSetID]; exists {
			if !changeSetsWithRow[row.ChangeSetID] {
				units := append([]historycontract.Unit{}, mutationItems[index].DiffSummary.Units...)
				for _, unit := range summary.Units {
					if unit.Kind == "record" && unit.Operation == "update" && len(unit.Changes) == 0 {
						continue
					}
					units = append(units, unit)
				}
				combined, err := historycontract.Summarize(units)
				if err != nil {
					return nil, err
				}
				mutationItems[index].DiffSummary = combined
			}
			continue
		}
		items = append(items, RecordHistoryItem{
			ActorUserID:              row.ActorUserID,
			CommittedAt:              row.CommittedAt,
			HistoryItemRef:           historyItemRefForRevision(record.RecordID, row.ChangeSetID, row.RevisionNo),
			Operation:                historyOperation(row.Source, "row_revision"),
			DiffSummary:              summary,
			ChangeSetID:              row.ChangeSetID,
			AvailableRollbackActions: nil,
			RevisionNo:               &revisionNo,
			createdAt:                row.CommittedAt,
			changeSetID:              row.ChangeSetID,
			sequenceNo:               int(^uint(0) >> 1),
			syntheticRank:            1,
		})
	}
	return items, nil
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

func historyOperation(source string, operationKind string) string {
	if operationKind != "" {
		return operationKind
	}
	if source != "" {
		return source
	}
	return "unknown"
}

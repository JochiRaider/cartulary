package revisions

import (
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"strings"

	"github.com/JochiRaider/cartulary/internal/modules/revisions/historycontract"
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
		sequenceNo:               row.SequenceNo,
		hasTargetEntry:           row.HistoryEntryAddressable,
	}, nil
}

func (materializer historyRowMaterializer) Revision(record RecordHistoryRecord, row revisionHistoryRow) (RecordHistoryItem, error) {
	summary, err := materializer.catalog.projectHistory(record, "record", record.RecordID.String(), "row_revision", row.BeforeValue, row.AfterValue)
	if err != nil {
		return RecordHistoryItem{}, err
	}
	return RecordHistoryItem{
		ActorUserID: row.ActorUserID, CommittedAt: row.CommittedAt,
		HistoryItemRef: historyItemRefForRevision(record.RecordID, row.ChangeSetID, row.RevisionNo),
		Operation:      historyOperation(row.Source, "row_revision"), DiffSummary: summary,
		ChangeSetID: row.ChangeSetID, RevisionNo: &row.RevisionNo,
	}, nil
}

func (materializer historyRowMaterializer) CoalesceRevision(record RecordHistoryRecord, item *RecordHistoryItem, row revisionHistoryRow) error {
	revision, err := materializer.Revision(record, row)
	if err != nil {
		return err
	}
	units := append([]historycontract.Unit{}, item.DiffSummary.Units...)
	for _, unit := range revision.DiffSummary.Units {
		if unit.Kind == "record" && unit.Operation == "update" && len(unit.Changes) == 0 {
			continue
		}
		units = append(units, unit)
	}
	item.DiffSummary, err = historycontract.Summarize(units)
	return err
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

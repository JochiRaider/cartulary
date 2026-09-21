package revisions

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"

	"github.com/JochiRaider/cartulary/internal/modules/revisions/historycontract"
	"github.com/JochiRaider/cartulary/internal/modules/revisions/rollbackcontract"
)

func decodeHistoryValue(data []byte) (map[string]any, error) {
	if len(data) == 0 {
		return nil, nil
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.UseNumber()
	var value map[string]any
	if err := decoder.Decode(&value); err != nil {
		return nil, fmt.Errorf("%w: JSON object", historycontract.ErrInvalidFacts)
	}
	if err := decoder.Decode(new(any)); err != io.EOF {
		return nil, historycontract.ErrInvalidFacts
	}
	return value, nil
}

func (catalog *TargetSemanticsCatalog) projectHistory(record RecordHistoryRecord, targetKind, targetID, operation string, beforeRaw, afterRaw []byte) (historycontract.Summary, error) {
	if catalog == nil {
		return historycontract.Summary{}, ErrMissingProviderContribution
	}
	entry, ok := catalog.byTargetKind[targetKind]
	if !ok {
		return historycontract.Summary{}, ErrMissingTargetSemantics
	}
	before, err := decodeHistoryValue(beforeRaw)
	if err != nil {
		return historycontract.Summary{}, err
	}
	after, err := decodeHistoryValue(afterRaw)
	if err != nil {
		return historycontract.Summary{}, err
	}
	facts := historycontract.Facts{RecordID: record.RecordID.String(), Operation: operation, Before: before, After: after}
	project := entry.historyProjector
	if entry.dispatchClass == rollbackcontract.DispatchRow {
		projection, ok := entry.rowHistory[record.RecordType]
		if !ok || targetID != record.RecordID.String() {
			return historycontract.Summary{}, ErrInvalidRecordSnapshot
		}
		for _, value := range []map[string]any{before, after} {
			if err := validatePersistedSnapshotValue(projection.schemaID, record.RecordID, record.RecordType, value); err != nil {
				return historycontract.Summary{}, err
			}
		}
		project = projection.project
	} else {
		if _, err := catalog.DescribeMutation(StoredMutation{TargetKind: targetKind, TargetID: targetID, OperationKind: operation, BeforeValue: before, AfterValue: after}); err != nil {
			return historycontract.Summary{}, err
		}
	}
	if project == nil {
		return historycontract.Summary{}, ErrMissingProviderContribution
	}
	units, err := project(facts)
	if err != nil {
		return historycontract.Summary{}, err
	}
	return historycontract.Summarize(units)
}

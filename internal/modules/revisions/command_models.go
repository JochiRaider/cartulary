package revisions

import (
	"crypto/sha256"
	"encoding/json"
)

const (
	deleteRouteKey   = "records.delete"
	restoreRouteKey  = "records.restore"
	rollbackRouteKey = "records.rollback"
)

type DeleteRestoreRequest struct {
	BaseRowVersion int64
	ClientTxnID    string
	Reason         *string
}

type RollbackTarget struct {
	Kind                string
	HistoryEntryRef     string
	ChangeSetID         string
	RestoreToRevisionNo int64
}

type RollbackRequest struct {
	BaseRowVersion int64
	ClientTxnID    string
	Reason         *string
	Target         RollbackTarget
}

func DeleteRestoreRequestHash(request DeleteRestoreRequest) []byte {
	return hashCommandPayload(map[string]any{
		"base_row_version": request.BaseRowVersion,
		"client_txn_id":    request.ClientTxnID,
		"reason":           optionalStringValue(request.Reason),
	})
}

func RollbackRequestHash(request RollbackRequest) []byte {
	return hashCommandPayload(map[string]any{
		"base_row_version": request.BaseRowVersion,
		"reason":           optionalStringValue(request.Reason),
		"target":           request.Target.Normalized(),
	})
}

func (target RollbackTarget) Normalized() map[string]any {
	switch target.Kind {
	case "history_entry":
		return map[string]any{"kind": "history_entry", "history_entry_ref": target.HistoryEntryRef}
	case "change_set":
		return map[string]any{"kind": "change_set", "change_set_id": target.ChangeSetID}
	case "row_restore":
		return map[string]any{"kind": "row_restore", "restore_to_revision_no": target.RestoreToRevisionNo}
	default:
		return map[string]any{"kind": target.Kind}
	}
}

func hashCommandPayload(payload map[string]any) []byte {
	data, _ := json.Marshal(payload)
	sum := sha256.Sum256(data)
	hash := make([]byte, len(sum))
	copy(hash, sum[:])
	return hash
}

func optionalStringValue(value *string) any {
	if value == nil {
		return nil
	}
	return *value
}

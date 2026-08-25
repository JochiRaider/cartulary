package stream

import (
	"encoding/json"
	"errors"
	"slices"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/modules/collaboration/protocol"
)

func validateRecordChangeMessage(message protocol.Message) error {
	var payload struct {
		RecordID         string           `json:"record_id"`
		RowVersion       int64            `json:"row_version"`
		ChangeSetID      string           `json:"change_set_id"`
		ClientTxnID      string           `json:"client_txn_id"`
		ActorUserID      string           `json:"actor_user_id"`
		ChangedFieldKeys []string         `json:"changed_field_keys"`
		AffectedViews    []map[string]any `json:"affected_views"`
	}
	if err := json.Unmarshal(message.Payload, &payload); err != nil {
		return err
	}
	if _, err := uuid.Parse(message.IncidentID); err != nil {
		return errors.New("invalid record_changed incident identity")
	}
	for _, identity := range []string{payload.RecordID, payload.ChangeSetID, payload.ActorUserID, message.EventID} {
		if _, err := uuid.Parse(identity); err != nil {
			return errors.New("invalid record_changed identity")
		}
	}
	if _, err := time.Parse(time.RFC3339Nano, message.EmittedAt); err != nil || payload.RowVersion < 1 || len(payload.AffectedViews) == 0 {
		return errors.New("invalid record_changed envelope")
	}
	viewIDs := make([]string, 0, len(payload.AffectedViews))
	for _, affected := range payload.AffectedViews {
		viewID, _ := affected["view_schema_id"].(string)
		kind, _ := affected["change_kind"].(string)
		_, hasPatch := affected["patch_cells"].(map[string]any)
		if strings.TrimSpace(viewID) == "" || (kind != "invalidate" && kind != "patch" && kind != "remove") || (kind == "patch") != hasPatch {
			return errors.New("invalid record_changed affected view")
		}
		viewIDs = append(viewIDs, viewID)
	}
	if !slices.IsSorted(viewIDs) {
		return errors.New("invalid record_changed affected view order")
	}
	for index := 1; index < len(viewIDs); index++ {
		if viewIDs[index-1] == viewIDs[index] {
			return errors.New("invalid record_changed duplicate affected view")
		}
	}
	return nil
}

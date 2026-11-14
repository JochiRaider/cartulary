package evidence

// Shared Evidence mutation validation and conflict helpers.
import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"slices"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	recordenvelope "github.com/JochiRaider/cartulary/internal/modules/records/enveloperead"
	"github.com/JochiRaider/cartulary/internal/modules/revisions"
	conflicttokens "github.com/JochiRaider/cartulary/internal/modules/revisions/conflicts"
)

func (f *mutationFacade) validateLifecyclePatchTx(ctx context.Context, tx pgx.Tx, recordID uuid.UUID, request patchRequest) error {
	changes := make([]lifecyclePatchChange, 0, len(request.Changes))
	for _, change := range request.Changes {
		var text *string
		if change.Value != nil {
			text = change.Value.Text
		}
		changes = append(changes, lifecyclePatchChange{FieldKey: change.FieldKey, Text: text})
	}
	return f.sourceMutations.validateLifecyclePatchTx(ctx, tx, recordID, changes)
}

func (f *mutationFacade) applyPatchTx(ctx context.Context, tx pgx.Tx, recordID uuid.UUID, request patchRequest, now time.Time) (bool, error) {
	changed := false
	for _, change := range request.Changes {
		if change.Value == nil {
			continue
		}
		if err := validateDirectPatchChange(change.FieldKey, *change.Value); err != nil {
			return false, err
		}
		applied, err := f.sourceMutations.applyDirectChangeTx(ctx, tx, recordID, change.FieldKey, *change.Value, now)
		if err != nil {
			return false, err
		}
		changed = changed || applied
	}
	return changed, nil
}

func validateEvidenceReferencesTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID, values map[string]FieldValue) error {
	ids := make([]uuid.UUID, 0, 2)
	for _, field := range []string{"evidence.collector_party_id", "evidence.source_party_id"} {
		if id := values[field].UUID; id != nil {
			ids = append(ids, *id)
		}
	}
	if len(ids) > 0 {
		rows, err := tx.Query(ctx, `SELECT record_id FROM records WHERE incident_id=$1 AND record_type='party' AND record_id=ANY($2::uuid[]) ORDER BY record_id FOR SHARE`, incidentID, ids)
		if err != nil {
			return err
		}
		for rows.Next() {
		}
		rows.Close()
		if err := rows.Err(); err != nil {
			return err
		}
	}
	for fieldKey, value := range values {
		if value.UUID == nil {
			continue
		}
		switch fieldKey {
		case "evidence.owner_user_id":
			if err := validateEvidenceOwnerTx(ctx, tx, incidentID, *value.UUID); err != nil {
				return err
			}
		case "evidence.collector_party_id", "evidence.source_party_id":
			if err := validateEvidenceTargetRecordTx(ctx, tx, incidentID, *value.UUID, "party", fieldKey); err != nil {
				return err
			}
		}
	}
	return nil
}

func validateEvidencePatchReferencesTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID, request patchRequest) error {
	for _, change := range request.Changes {
		if change.Value == nil || change.Value.UUID == nil {
			continue
		}
		switch change.FieldKey {
		case "evidence.owner_user_id":
			if err := validateEvidenceOwnerTx(ctx, tx, incidentID, *change.Value.UUID); err != nil {
				return err
			}
		case "evidence.collector_party_id", "evidence.source_party_id":
			if err := validateEvidenceTargetRecordTx(ctx, tx, incidentID, *change.Value.UUID, "party", change.FieldKey); err != nil {
				return err
			}
		}
	}
	return nil
}

func validateEvidenceOwnerTx(ctx context.Context, tx pgx.Tx, incidentID, ownerID uuid.UUID) error {
	var exists bool
	if err := tx.QueryRow(ctx, `SELECT EXISTS (
 SELECT 1 FROM users u JOIN incident_memberships m ON m.user_id=u.id
 WHERE u.id=$1 AND u.is_active=true AND m.incident_id=$2
)`, ownerID, incidentID).Scan(&exists); err != nil {
		return fmt.Errorf("validate Evidence owner: %w", err)
	}
	if !exists {
		return &ValidationError{Field: "evidence.owner_user_id", ReasonCode: "invalid_value"}
	}
	return nil
}

func validateEvidenceTargetRecordTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID, recordID uuid.UUID, expectedType string, field string) error {
	var exists bool
	if err := tx.QueryRow(ctx, `
SELECT EXISTS (
    SELECT 1
      FROM records r
      JOIN parties p ON p.record_id=r.record_id AND p.incident_id=r.incident_id
     WHERE r.incident_id = $1
       AND r.record_id = $2
       AND r.record_type = $3
       AND r.deleted_at IS NULL
)
`, incidentID, recordID, expectedType).Scan(&exists); err != nil {
		return fmt.Errorf("validate evidence reference target: %w", err)
	}
	if !exists {
		return &ValidationError{Field: field, ReasonCode: "invalid_value"}
	}
	return nil
}

type evidenceRecordMeta struct {
	IncidentID uuid.UUID
	RecordType string
	RowVersion int64
}

func (f *mutationFacade) loadEvidenceRecordMetaForUpdateTx(ctx context.Context, tx pgx.Tx, recordID uuid.UUID) (evidenceRecordMeta, error) {
	envelope, err := f.recordEnvelopes.LoadEnvelopeTx(ctx, tx, recordID, true)
	if errors.Is(err, recordenvelope.ErrEnvelopeNotFound) {
		return evidenceRecordMeta{}, pgx.ErrNoRows
	}
	if err != nil {
		return evidenceRecordMeta{}, err
	}
	if envelope.DeletedAt != nil {
		return evidenceRecordMeta{}, revisions.ErrRecordDeletedUseRestore
	}
	return evidenceRecordMeta{IncidentID: envelope.IncidentID, RecordType: envelope.RecordType, RowVersion: envelope.RowVersion}, nil
}

func (f *mutationFacade) lockPatchReferenceScopeTx(ctx context.Context, tx pgx.Tx, recordID uuid.UUID, request patchRequest) error {
	ids := []uuid.UUID{recordID}
	for _, change := range request.Changes {
		if (change.FieldKey == "evidence.collector_party_id" || change.FieldKey == "evidence.source_party_id") && change.Value != nil && change.Value.UUID != nil {
			ids = append(ids, *change.Value.UUID)
		}
	}
	if len(ids) == 1 {
		return nil
	}
	envelope, err := f.recordEnvelopes.LoadEnvelopeTx(ctx, tx, recordID, false)
	if errors.Is(err, recordenvelope.ErrEnvelopeNotFound) {
		return pgx.ErrNoRows
	}
	if err != nil {
		return err
	}
	if envelope.RecordType != "evidence" {
		return pgx.ErrNoRows
	}
	rows, err := tx.Query(ctx, `SELECT record_id FROM records WHERE incident_id=$1 AND record_id=ANY($2::uuid[]) AND (record_id=$3 OR record_type='party') ORDER BY record_id FOR UPDATE`, envelope.IncidentID, ids, recordID)
	if err != nil {
		return err
	}
	for rows.Next() {
	}
	rows.Close()
	return rows.Err()
}

func changedFieldKeys(before map[string]any, after map[string]any) []string {
	afterCells, _ := after["cells"].(map[string]any)
	beforeCells := map[string]any{}
	if before != nil {
		beforeCells, _ = before["cells"].(map[string]any)
	}
	keys := make([]string, 0)
	for fieldKey, afterValue := range afterCells {
		if beforeValue, ok := beforeCells[fieldKey]; !ok || !reflect.DeepEqual(beforeValue, afterValue) {
			keys = append(keys, fieldKey)
		}
	}
	slices.Sort(keys)
	return keys
}

func workbookVersionID(recordID uuid.UUID, rowVersion int64) string {
	return fmt.Sprintf("record:%s:%d", recordID.String(), rowVersion)
}

func adaptRevisionWindowError(recordID uuid.UUID, baseRowVersion int64, currentRowVersion int64, err error) error {
	if err == nil {
		return nil
	}
	var windowErr *conflicttokens.RevisionWindowError
	if errors.As(err, &windowErr) {
		return &RowVersionConflictError{RecordID: windowErr.RecordID, BaseRowVersion: windowErr.BaseRowVersion, CurrentRowVersion: windowErr.CurrentRowVersion}
	}
	return &RowVersionConflictError{RecordID: recordID, BaseRowVersion: baseRowVersion, CurrentRowVersion: currentRowVersion}
}

type evidenceSameFieldConflictParams struct {
	RouteKey          string
	RecordID          uuid.UUID
	ViewSchemaID      string
	BaseRowVersion    int64
	CurrentRowVersion int64
	RequestHash       []byte
	Window            conflicttokens.PatchConflictWindow
	Change            patchChange
	Changed           conflicttokens.PatchChangedField
	CurrentRow        map[string]any
	FieldDescriptors  conflicttokens.FieldDescriptorSet
	Codec             conflicttokens.ConflictTokenCodec
}

func overlappingEvidencePatchChange(changes []patchChange, changedFields map[string]conflicttokens.PatchChangedField) (patchChange, conflicttokens.PatchChangedField, bool) {
	for _, change := range changes {
		changed, ok := changedFields[change.FieldKey]
		if ok {
			return change, changed, true
		}
	}
	return patchChange{}, conflicttokens.PatchChangedField{}, false
}

func buildEvidenceSameFieldConflict(params evidenceSameFieldConflictParams) (SameFieldConflict, error) {
	baseValue, ok := rowCellValue(params.Window.BaseRow, params.Change.FieldKey)
	if !ok {
		return SameFieldConflict{}, &conflicttokens.RevisionWindowError{RecordID: params.RecordID, BaseRowVersion: params.BaseRowVersion, CurrentRowVersion: params.CurrentRowVersion}
	}
	serverValue, ok := rowCellValue(params.CurrentRow, params.Change.FieldKey)
	if !ok {
		return SameFieldConflict{}, &conflicttokens.RevisionWindowError{RecordID: params.RecordID, BaseRowVersion: params.BaseRowVersion, CurrentRowVersion: params.CurrentRowVersion}
	}
	clientValue := params.Change.CanonicalValue
	conflictClass := params.FieldDescriptors.ConflictResolutionClass(params.Change.FieldKey)
	if conflictClass == "" {
		conflictClass = "atomic_replace"
	}
	conflictToken, err := evidenceConflictToken(params.RouteKey, params.RecordID, params.ViewSchemaID, params.Change.FieldKey, conflictClass, params.BaseRowVersion, params.CurrentRowVersion, params.RequestHash, params.Codec)
	if err != nil {
		return SameFieldConflict{}, err
	}
	conflict := SameFieldConflict{
		ConflictToken:           conflictToken,
		RecordID:                params.RecordID,
		FieldKey:                params.Change.FieldKey,
		ConflictResolutionClass: conflictClass,
		BaseRowVersion:          params.BaseRowVersion,
		CurrentRowVersion:       params.CurrentRowVersion,
		ClientValue:             clientValue,
		ServerValue:             serverValue,
		BaseValue:               OptionalConflictValue{Present: true, Value: baseValue},
		ServerUpdatedBy:         params.Changed.ServerUpdatedBy,
		ServerUpdatedAt:         params.Changed.ServerUpdatedAt.UTC(),
	}
	if conflictClass == "text_compare_merge" {
		if suggested, ok := conflicttokens.SuggestedTextMergeValue(baseValue, serverValue, clientValue); ok {
			conflict.SuggestedMergedValue = OptionalConflictValue{Present: true, Value: suggested}
		}
	}
	return conflict, nil
}

func rowCellValue(row map[string]any, fieldKey string) (any, bool) {
	cells, _ := row["cells"].(map[string]any)
	cell, ok := cells[fieldKey].(map[string]any)
	if !ok {
		return nil, false
	}
	value, ok := cell["value"]
	return value, ok
}

func evidenceConflictToken(routeKey string, recordID uuid.UUID, viewSchemaID string, fieldKey string, conflictClass string, baseRowVersion int64, currentRowVersion int64, requestHash []byte, codec conflicttokens.ConflictTokenCodec) (string, error) {
	return codec.Issue(conflicttokens.ConflictTokenClaims{
		RouteKey:                routeKey,
		RecordID:                recordID.String(),
		ViewSchemaID:            viewSchemaID,
		FieldKey:                fieldKey,
		ConflictResolutionClass: conflictClass,
		BaseRowVersion:          baseRowVersion,
		CurrentRowVersion:       currentRowVersion,
		RequestHash:             conflicttokens.RequestHashTokenValue(requestHash),
	})
}

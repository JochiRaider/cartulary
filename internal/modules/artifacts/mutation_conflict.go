package artifacts

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/links"
	"github.com/JochiRaider/cartulary/internal/modules/revisions/conflicts"
	"github.com/JochiRaider/cartulary/internal/platform/collectionvalue"
	"github.com/JochiRaider/cartulary/internal/platform/viewschema"
)

type ConflictCommand struct {
	ActorUserID uuid.UUID
	Admission   ConflictResolveAdmission
	RequestID   string
	Now         time.Time
}

func (f *MutationFacade) ResolveConflict(
	ctx context.Context,
	command ConflictCommand,
) (MutationResult, error) {
	if !command.Admission.valid() {
		return MutationResult{}, &ValidationError{Field: "conflict_token", ReasonCode: "invalid_value"}
	}
	request := command.Admission.requestValue()
	context := command.Admission.contextValue()
	mechanics := conflicts.Command{
		ActorUserID: command.ActorUserID, RecordID: context.RecordID,
		Claims: conflictTokenClaims(context), ClientTxnID: request.ClientTxnID,
		RequestHash: command.Admission.requestHash(), RequestID: command.RequestID,
		RouteKey: string(OperationConflictResolve),
	}
	if request.ResolutionKind != "keep_saved" {
		return f.patch(ctx, PatchCommand{
			ActorUserID: command.ActorUserID, RecordID: context.RecordID,
			Admission: command.Admission.patchAdmission(), RequestID: command.RequestID, Now: command.Now,
		}, OperationConflictResolve)
	}
	result, err := conflicts.KeepSaved(
		ctx,
		f.pool,
		f.keepSaved,
		mechanics,
		f.loadConflictTarget,
	)
	if err != nil {
		return MutationResult{}, err
	}
	return MutationResult{
		Row:          conflictResultRow(result.Payload),
		Outcome:      conflictMutationOutcome(result.Replayed),
		IncidentID:   result.IncidentID,
		RecordID:     result.RecordID,
		ClientTxnID:  result.ClientTxnID,
		RowVersion:   result.RowVersion,
		ViewSchemaID: result.ViewSchemaID,
	}, nil
}

func conflictMutationOutcome(replayed bool) MutationOutcome {
	if replayed {
		return MutationOutcomeReplayed
	}
	return MutationOutcomeKeptSaved
}

func conflictTokenClaims(context ConflictAdmissionContext) conflicts.ConflictTokenClaims {
	return conflicts.ConflictTokenClaims{
		Version: context.Version, RecordID: context.RecordID.String(), ViewSchemaID: context.ViewSchemaID,
		RouteKey: context.RouteKey, FieldKey: context.FieldKey,
		ConflictResolutionClass: context.ConflictResolutionClass,
		BaseRowVersion:          context.BaseRowVersion, CurrentRowVersion: context.CurrentRowVersion,
		RequestHash: context.OriginalRequestHash, IssuedAt: context.IssuedAt, ExpiresAt: context.ExpiresAt,
	}
}

func (f *MutationFacade) loadConflictTarget(
	ctx context.Context,
	tx pgx.Tx,
	command conflicts.Command,
) (conflicts.Target, error) {
	meta, err := f.loadArtifactRecordMetaForUpdateTx(ctx, tx, command.RecordID)
	if err != nil {
		return conflicts.Target{}, err
	}
	if meta.RecordType != "artifact" {
		return conflicts.Target{}, pgx.ErrNoRows
	}
	if err := validateArtifactViewRecordTx(
		ctx,
		tx,
		command.RecordID,
		command.Claims.ViewSchemaID,
	); err != nil {
		return conflicts.Target{}, err
	}
	field, ok := viewschema.LookupField(command.Claims.ViewSchemaID, command.Claims.FieldKey)
	if !ok || !field.Writable {
		return conflicts.Target{}, pgx.ErrNoRows
	}
	if err := f.incidentAccess.RequireOpenTx(ctx, tx, meta.IncidentID); err != nil {
		return conflicts.Target{}, err
	}
	row, err := f.source.projections.LoadArtifactTx(ctx, tx, command.Claims.ViewSchemaID, command.RecordID)
	if err != nil {
		return conflicts.Target{}, err
	}
	return conflicts.Target{
		IncidentID: meta.IncidentID,
		RecordID:   command.RecordID,
		RowVersion: meta.RowVersion,
		Row:        row,
	}, nil
}

func conflictResultRow(payload map[string]any) map[string]any {
	row, _ := payload["row"].(map[string]any)
	return row
}

// errInvalidCollectionConflictBase reports a reconstructed base value that
// cannot carry the requested collection actions. Like an unavailable window it
// fails closed as row_version_conflict; an action the admission contract
// excludes is an invariant failure instead.
var errInvalidCollectionConflictBase = errors.New("invalid base collection value")

// adaptRevisionWindowError classifies only the typed unavailable-window
// condition, which fails closed as row_version_conflict. Every other failure,
// including a database, field-catalog or conflict-token error, stays
// unclassified and surfaces as internal_error (REQ-01-706).
func adaptRevisionWindowError(recordID uuid.UUID, baseRowVersion int64, currentRowVersion int64, err error) error {
	if err == nil {
		return nil
	}
	var windowErr *conflicts.RevisionWindowError
	if errors.As(err, &windowErr) {
		return &RowVersionConflictError{RecordID: windowErr.RecordID, BaseRowVersion: windowErr.BaseRowVersion, CurrentRowVersion: windowErr.CurrentRowVersion}
	}
	return fmt.Errorf("record %s revision window %d..%d: %w", recordID, baseRowVersion, currentRowVersion, err)
}

type artifactSameFieldConflictParams struct {
	RouteKey          string
	RecordID          uuid.UUID
	ViewSchemaID      string
	BaseRowVersion    int64
	CurrentRowVersion int64
	RequestHash       []byte
	Window            conflicts.PatchConflictWindow
	Change            patchChange
	Changed           conflicts.PatchChangedField
	CurrentRow        map[string]any
	FieldDescriptors  conflicts.FieldDescriptorSet
	Codec             conflicts.ConflictTokenCodec
}

func overlappingArtifactPatchChange(changes []patchChange, changedFields map[string]conflicts.PatchChangedField) (patchChange, conflicts.PatchChangedField, bool) {
	for _, change := range changes {
		changed, ok := changedFields[change.FieldKey]
		if ok {
			return change, changed, true
		}
	}
	return patchChange{}, conflicts.PatchChangedField{}, false
}

func buildArtifactSameFieldConflict(params artifactSameFieldConflictParams) (SameFieldConflict, error) {
	baseValue, ok := rowCellValue(params.Window.BaseRow, params.Change.FieldKey)
	if !ok {
		return SameFieldConflict{}, &conflicts.RevisionWindowError{RecordID: params.RecordID, BaseRowVersion: params.BaseRowVersion, CurrentRowVersion: params.CurrentRowVersion}
	}
	serverValue, ok := rowCellValue(params.CurrentRow, params.Change.FieldKey)
	if !ok {
		return SameFieldConflict{}, &conflicts.RevisionWindowError{RecordID: params.RecordID, BaseRowVersion: params.BaseRowVersion, CurrentRowVersion: params.CurrentRowVersion}
	}
	// Stored conflict facts may predate REQ-01-067 label order; re-serve them.
	baseValue, serverValue = collectionvalue.Canonicalize(baseValue), collectionvalue.Canonicalize(serverValue)
	clientValue, err := artifactPatchClientConflictValue(params.RecordID, params.Change, baseValue, params.RequestHash)
	if errors.Is(err, errInvalidCollectionConflictBase) {
		return SameFieldConflict{}, &conflicts.RevisionWindowError{RecordID: params.RecordID, BaseRowVersion: params.BaseRowVersion, CurrentRowVersion: params.CurrentRowVersion}
	}
	if err != nil {
		return SameFieldConflict{}, err
	}
	conflictClass := params.FieldDescriptors.ConflictResolutionClass(params.Change.FieldKey)
	if conflictClass == "" {
		conflictClass = "atomic_replace"
	}
	conflictToken, err := artifactConflictToken(params.RouteKey, params.RecordID, params.ViewSchemaID, params.Change.FieldKey, conflictClass, params.BaseRowVersion, params.CurrentRowVersion, params.RequestHash, params.Codec)
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
		if suggested, ok := conflicts.SuggestedTextMergeValue(baseValue, serverValue, clientValue); ok {
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

func artifactPatchClientConflictValue(recordID uuid.UUID, change patchChange, baseValue any, requestHash []byte) (any, error) {
	if change.Collection == nil {
		return change.CanonicalValue, nil
	}
	return applyCollectionConflictActions(recordID, change.FieldKey, baseValue, *change.Collection, requestHash)
}

func applyCollectionConflictActions(recordID uuid.UUID, fieldKey string, baseValue any, payload collectionActionPayload, requestHash []byte) (map[string]any, error) {
	ordered, items, ok := cloneCollectionConflictValue(baseValue)
	if !ok {
		return nil, fmt.Errorf("%w for %s", errInvalidCollectionConflictBase, fieldKey)
	}
	for index, action := range payload.Actions {
		switch action.Op {
		case "add_record_ref", "add_party_ref", "add_tag", "add_risk_ref":
			items = upsertCollectionConflictItem(items, newClientCollectionItem(recordID, fieldKey, action, requestHash, index))
		case "remove_record_ref", "remove_party_ref", "remove_tag", "remove_risk_ref":
			items = removeCollectionConflictItem(items, action.ItemRef)
		default:
			return nil, fmt.Errorf("unsupported collection action: %s", action.Op)
		}
	}
	return collectionvalue.New(ordered, items), nil
}

func cloneCollectionConflictValue(value any) (bool, []map[string]any, bool) {
	object, ok := value.(map[string]any)
	if !ok || object["kind"] != "collection_value_v1" {
		return false, nil, false
	}
	ordered, ok := object["ordered"].(bool)
	if !ok {
		return false, nil, false
	}
	items := make([]map[string]any, 0)
	switch rawItems := object["items"].(type) {
	case []any:
		for _, rawItem := range rawItems {
			item, ok := rawItem.(map[string]any)
			if !ok {
				return false, nil, false
			}
			items = append(items, cloneMap(item))
		}
	case []map[string]any:
		for _, item := range rawItems {
			items = append(items, cloneMap(item))
		}
	default:
		return false, nil, false
	}
	return ordered, items, true
}

func newClientCollectionItem(recordID uuid.UUID, fieldKey string, action collectionAction, requestHash []byte, actionIndex int) map[string]any {
	switch action.Op {
	case "add_record_ref":
		// labelArtifactConflictTx replaces the provisional label (REQ-01-333a).
		linkedID := action.LinkedRecordID.String()
		return map[string]any{
			"item_ref":         links.RecordRefItemRef(*action.LinkedRecordID),
			"item_kind":        "record_ref",
			"display_text":     collectionvalue.UnavailableRecordLabel,
			"linked_record_id": linkedID,
		}
	case "add_party_ref":
		partyID := action.PartyID.String()
		return map[string]any{
			"item_ref":     links.PartyRefItemRef(*action.PartyID),
			"item_kind":    "party_ref",
			"display_text": collectionvalue.UnavailableRecordLabel,
			"party_id":     partyID,
		}
	case "add_tag":
		tagID := conflictLocalUUID(recordID, fieldKey, action, requestHash, actionIndex)
		return map[string]any{
			"item_ref":     links.RecordTagItemRef(recordID, tagID),
			"item_kind":    "tag",
			"display_text": action.RawText,
			"tag_id":       tagID.String(),
		}
	case "add_risk_ref":
		riskRefID := conflictLocalUUID(recordID, fieldKey, action, requestHash, actionIndex)
		return map[string]any{
			"item_ref":      riskRefItemRef(riskRefID),
			"item_kind":     "risk_ref",
			"display_text":  action.RiskRefText,
			"risk_ref_id":   riskRefID.String(),
			"risk_ref_text": action.RiskRefText,
		}
	default:
		return map[string]any{}
	}
}

func conflictLocalUUID(recordID uuid.UUID, fieldKey string, action collectionAction, requestHash []byte, actionIndex int) uuid.UUID {
	text := action.NormalizedText
	if action.Op == "add_risk_ref" {
		text = action.RiskRefText
	}
	seed, _ := json.Marshal(map[string]any{
		"record_id":     recordID.String(),
		"field_key":     fieldKey,
		"request_hash":  base64.RawURLEncoding.EncodeToString(requestHash),
		"action_index":  actionIndex,
		"op":            action.Op,
		"risk_ref_text": text,
	})
	return uuid.NewSHA1(uuid.NameSpaceOID, seed)
}

func upsertCollectionConflictItem(items []map[string]any, item map[string]any) []map[string]any {
	itemRef, _ := item["item_ref"].(string)
	if itemRef == "" {
		return items
	}
	for index, existing := range items {
		if existing["item_ref"] == itemRef {
			items[index] = item
			return items
		}
		if item["item_kind"] == "risk_ref" && existing["item_kind"] == "risk_ref" && existing["risk_ref_text"] == item["risk_ref_text"] {
			items[index] = item
			return items
		}
	}
	return append(items, item)
}

func removeCollectionConflictItem(items []map[string]any, itemRef string) []map[string]any {
	filtered := items[:0]
	for _, item := range items {
		if item["item_ref"] != itemRef {
			filtered = append(filtered, item)
		}
	}
	return filtered
}

// recordSummaryLabels is the projections-owned REQ-01-333a label reader.
type recordSummaryLabels interface {
	RecordSummaryLabelsTx(context.Context, pgx.Tx, uuid.UUID, []uuid.UUID) (map[uuid.UUID]string, error)
}

// labelArtifactConflictTx labels the record_ref and party_ref items of every
// conflict value with REQ-01-333a labels read in one lookup inside the conflict
// transaction. Records not live in the incident read UnavailableRecordLabel.
func labelArtifactConflictTx(ctx context.Context, tx pgx.Tx, labels recordSummaryLabels, incidentID uuid.UUID, conflict *SameFieldConflict) error {
	ids := collectionvalue.LinkedRecordIDs(conflict.BaseValue.Value, conflict.ServerValue, conflict.ClientValue)
	if len(ids) == 0 {
		return nil
	}
	byID, err := labels.RecordSummaryLabelsTx(ctx, tx, incidentID, ids)
	if err != nil {
		return err
	}
	conflict.BaseValue.Value = collectionvalue.WithLinkedRecordLabels(conflict.BaseValue.Value, byID)
	conflict.ServerValue = collectionvalue.WithLinkedRecordLabels(conflict.ServerValue, byID)
	conflict.ClientValue = collectionvalue.WithLinkedRecordLabels(conflict.ClientValue, byID)
	return nil
}

func artifactConflictToken(routeKey string, recordID uuid.UUID, viewSchemaID string, fieldKey string, conflictClass string, baseRowVersion int64, currentRowVersion int64, requestHash []byte, codec conflicts.ConflictTokenCodec) (string, error) {
	return codec.Issue(conflicts.ConflictTokenClaims{
		RouteKey:                routeKey,
		RecordID:                recordID.String(),
		ViewSchemaID:            viewSchemaID,
		FieldKey:                fieldKey,
		ConflictResolutionClass: conflictClass,
		BaseRowVersion:          baseRowVersion,
		CurrentRowVersion:       currentRowVersion,
		RequestHash:             conflicts.RequestHashTokenValue(requestHash),
	})
}

func cloneMap(source map[string]any) map[string]any {
	cloned := make(map[string]any, len(source))
	for key, value := range source {
		cloned[key] = value
	}
	return cloned
}

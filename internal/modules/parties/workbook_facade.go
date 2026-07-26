package parties

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"slices"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/incidents"
	partyprojection "github.com/JochiRaider/cartulary/internal/modules/parties/projectionprovider"
	"github.com/JochiRaider/cartulary/internal/modules/projections"
	"github.com/JochiRaider/cartulary/internal/modules/records"
	"github.com/JochiRaider/cartulary/internal/modules/revisions"
	"github.com/JochiRaider/cartulary/internal/modules/revisions/conflictmerge"
	"github.com/JochiRaider/cartulary/internal/modules/revisions/conflicttokens"
	"github.com/JochiRaider/cartulary/internal/modules/revisions/conflictwindows"
	"github.com/JochiRaider/cartulary/internal/modules/revisions/historyquery"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
)

type WorkbookFacade struct {
	pool             postgres.DB
	authStore        *authn.Store
	incidentAccess   incidents.Access
	recordStore      *records.Store
	projectionRows   *projections.PartyRows
	revisionHistory  historyquery.Reader
	revisionAppender revisions.Appender
	store            *Store
	conflictTokens   conflicttokens.ConflictTokenCodec
}

type WorkbookCreateRequest struct {
	ViewSchemaID string
	ClientTxnID  string
	Values       map[string]FieldValue
}

type WorkbookPatchRequest struct {
	ViewSchemaID   string
	BaseRowVersion int64
	ClientTxnID    string
	Changes        []WorkbookPatchChange
}

type WorkbookPatchChange struct {
	FieldKey       string
	Value          *FieldValue
	CanonicalValue any
}

type WorkbookCreateCommand struct {
	Actor       authn.UserRecord
	IncidentID  uuid.UUID
	Request     WorkbookCreateRequest
	RequestHash []byte
	RequestID   string
	RouteKey    string
	Now         time.Time
}

type WorkbookPatchCommand struct {
	Actor            authn.UserRecord
	RecordID         uuid.UUID
	Request          WorkbookPatchRequest
	RequestHash      []byte
	RequestID        string
	RouteKey         string
	ConflictRouteKey string
	Now              time.Time
}

type WorkbookMutationResult struct {
	Payload          map[string]any
	StatusCode       int
	Replayed         bool
	IncidentID       uuid.UUID
	RecordID         uuid.UUID
	ChangeSetID      uuid.UUID
	ClientTxnID      string
	RowVersion       int64
	ViewSchemaID     string
	ChangedFieldKeys []string
}

type RowVersionConflictError struct {
	RecordID          uuid.UUID
	BaseRowVersion    int64
	CurrentRowVersion int64
}

func (e *RowVersionConflictError) Error() string {
	return "parties: row version conflict"
}

type SameFieldConflictError struct {
	Conflict map[string]any
}

func (e *SameFieldConflictError) Error() string {
	return "parties: same field conflict"
}

func NewWorkbookFacade(pool postgres.DB, conflictTokens conflicttokens.ConflictTokenCodec) *WorkbookFacade {
	return &WorkbookFacade{
		pool:             pool,
		authStore:        authn.NewStore(pool),
		incidentAccess:   incidents.NewAccess(pool),
		recordStore:      records.NewStore(),
		projectionRows:   projections.NewPartyRows(pool, partyprojection.QuerySurfaces()...),
		revisionHistory:  historyquery.NewReader(),
		revisionAppender: revisions.NewAppender(),
		store:            NewStore(pool),
		conflictTokens:   conflictTokens,
	}
}

func (f *WorkbookFacade) Create(ctx context.Context, command WorkbookCreateCommand) (WorkbookMutationResult, error) {
	request := command.Request
	idempotencyKey := authn.RouteIdempotencyKey{
		RouteKey:    command.RouteKey,
		ActorUserID: command.Actor.ID,
		ScopeKey:    command.IncidentID.String() + ":" + request.ViewSchemaID,
		ClientTxnID: request.ClientTxnID,
	}
	if existing, err := f.authStore.GetRouteIdempotency(ctx, idempotencyKey); err == nil {
		if !bytes.Equal(existing.RequestHash, command.RequestHash) {
			return WorkbookMutationResult{}, authn.ErrClientTxnConflict
		}
		payload, err := decodeStoredPayload(existing.ResponseJSON)
		if err != nil {
			return WorkbookMutationResult{}, fmt.Errorf("decode replayed party create payload: %w", err)
		}
		recordID, err := extractPayloadUUID(payload, "row", "record_id")
		if err != nil {
			return WorkbookMutationResult{}, err
		}
		return WorkbookMutationResult{Payload: payload, StatusCode: http.StatusOK, Replayed: true, IncidentID: command.IncidentID, RecordID: recordID, ViewSchemaID: request.ViewSchemaID, ClientTxnID: request.ClientTxnID}, nil
	} else if !errors.Is(err, authn.ErrNotFound) {
		return WorkbookMutationResult{}, fmt.Errorf("query party create idempotency: %w", err)
	}
	params := CreateParams{Values: request.Values}
	if err := ValidateCreateParams(params); err != nil {
		return WorkbookMutationResult{}, err
	}

	tx, err := f.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return WorkbookMutationResult{}, fmt.Errorf("begin party create transaction: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	if err := f.incidentAccess.EnsureOpenTx(ctx, tx, command.IncidentID); err != nil {
		return WorkbookMutationResult{}, err
	}
	now := command.Now.UTC()
	if recordID, found, err := f.store.FindReusablePartyTx(ctx, tx, command.IncidentID, params); err != nil || found {
		if err != nil {
			return WorkbookMutationResult{}, err
		}
		result, err := f.reuseCreateTx(ctx, tx, command, idempotencyKey, recordID, now)
		if err != nil {
			return WorkbookMutationResult{}, err
		}
		if err := tx.Commit(ctx); err != nil {
			return WorkbookMutationResult{}, fmt.Errorf("commit party reuse transaction: %w", err)
		}
		return result, nil
	}

	recordID, err := f.recordStore.InsertTx(ctx, tx, records.InsertParams{
		IncidentID:      command.IncidentID,
		RecordType:      "party",
		CreatedByUserID: command.Actor.ID,
		CreatedAt:       now,
		UpdatedByUserID: command.Actor.ID,
		UpdatedAt:       now,
		RowVersion:      1,
	})
	if err != nil {
		return WorkbookMutationResult{}, err
	}
	if err := f.store.InsertPartyTx(ctx, tx, recordID, command.IncidentID, params, now); err != nil {
		return WorkbookMutationResult{}, err
	}
	if err := f.projectionRows.RefreshTx(ctx, tx, recordID); err != nil {
		return WorkbookMutationResult{}, err
	}
	row, err := f.projectionRows.LoadTx(ctx, tx, recordID)
	if err != nil {
		return WorkbookMutationResult{}, err
	}
	changeSetID, err := f.revisionAppender.AppendChangeSetTx(ctx, tx, revisions.AppendChangeSetParams{
		IncidentID:  command.IncidentID,
		ActorUserID: command.Actor.ID,
		Source:      command.RouteKey,
		ClientTxnID: &request.ClientTxnID,
		RequestID:   &command.RequestID,
		CreatedAt:   now,
	})
	if err != nil {
		return WorkbookMutationResult{}, err
	}
	afterVersionID := workbookVersionID(recordID, 1)
	if err := f.revisionAppender.AppendMutationTx(ctx, tx, revisions.AppendMutationParams{
		ChangeSetID:    changeSetID,
		SequenceNo:     1,
		TargetKind:     "record",
		TargetID:       recordID.String(),
		OperationKind:  "create",
		AfterVersionID: &afterVersionID,
		AfterValue:     row,
	}); err != nil {
		return WorkbookMutationResult{}, err
	}
	if err := f.revisionAppender.AppendRecordRevisionTx(ctx, tx, revisions.AppendRecordRevisionParams{
		ChangeSetID: changeSetID,
		RecordID:    recordID,
		RowVersion:  1,
		AfterValue:  row,
	}); err != nil {
		return WorkbookMutationResult{}, err
	}
	payload := buildMutationPayload(request.ViewSchemaID, changeSetID, row)
	if err := authn.InsertRouteIdempotencyPayload(ctx, tx, idempotencyKey, nil, command.RequestHash, http.StatusCreated, payload); err != nil {
		if authn.IsUniqueViolation(err) {
			return WorkbookMutationResult{}, authn.ErrClientTxnConflict
		}
		return WorkbookMutationResult{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return WorkbookMutationResult{}, fmt.Errorf("commit party create transaction: %w", err)
	}
	return WorkbookMutationResult{
		Payload:          payload,
		StatusCode:       http.StatusCreated,
		IncidentID:       command.IncidentID,
		RecordID:         recordID,
		ChangeSetID:      changeSetID,
		ClientTxnID:      request.ClientTxnID,
		RowVersion:       1,
		ViewSchemaID:     request.ViewSchemaID,
		ChangedFieldKeys: changedFieldKeys(nil, row),
	}, nil
}

func (f *WorkbookFacade) reuseCreateTx(ctx context.Context, tx pgx.Tx, command WorkbookCreateCommand, idempotencyKey authn.RouteIdempotencyKey, recordID uuid.UUID, now time.Time) (WorkbookMutationResult, error) {
	request := command.Request
	if err := f.projectionRows.RefreshTx(ctx, tx, recordID); err != nil {
		return WorkbookMutationResult{}, err
	}
	row, err := f.projectionRows.LoadTx(ctx, tx, recordID)
	if err != nil {
		return WorkbookMutationResult{}, err
	}
	changeSetID, err := f.revisionAppender.AppendChangeSetTx(ctx, tx, revisions.AppendChangeSetParams{
		IncidentID:  command.IncidentID,
		ActorUserID: command.Actor.ID,
		Source:      command.RouteKey,
		ClientTxnID: &request.ClientTxnID,
		RequestID:   &command.RequestID,
		CreatedAt:   now,
	})
	if err != nil {
		return WorkbookMutationResult{}, err
	}
	rowVersion, err := rowVersionFromGenericRow(row)
	if err != nil {
		return WorkbookMutationResult{}, err
	}
	afterVersionID := workbookVersionID(recordID, rowVersion)
	if err := f.revisionAppender.AppendMutationTx(ctx, tx, revisions.AppendMutationParams{
		ChangeSetID:    changeSetID,
		SequenceNo:     1,
		TargetKind:     "record",
		TargetID:       recordID.String(),
		OperationKind:  "reuse",
		AfterVersionID: &afterVersionID,
		AfterValue:     row,
	}); err != nil {
		return WorkbookMutationResult{}, err
	}
	payload := buildMutationPayload(request.ViewSchemaID, changeSetID, row)
	if err := authn.InsertRouteIdempotencyPayload(ctx, tx, idempotencyKey, nil, command.RequestHash, http.StatusOK, payload); err != nil {
		if authn.IsUniqueViolation(err) {
			return WorkbookMutationResult{}, authn.ErrClientTxnConflict
		}
		return WorkbookMutationResult{}, err
	}
	return WorkbookMutationResult{
		Payload:          payload,
		StatusCode:       http.StatusOK,
		IncidentID:       command.IncidentID,
		RecordID:         recordID,
		ChangeSetID:      changeSetID,
		ClientTxnID:      request.ClientTxnID,
		RowVersion:       rowVersion,
		ViewSchemaID:     request.ViewSchemaID,
		ChangedFieldKeys: []string{},
	}, nil
}

func (f *WorkbookFacade) Patch(ctx context.Context, command WorkbookPatchCommand) (WorkbookMutationResult, error) {
	request := command.Request
	idempotencyKey := authn.RouteIdempotencyKey{
		RouteKey:    command.RouteKey,
		ActorUserID: command.Actor.ID,
		ScopeKey:    command.RecordID.String(),
		ClientTxnID: request.ClientTxnID,
	}
	if existing, err := f.authStore.GetRouteIdempotency(ctx, idempotencyKey); err == nil {
		if !bytes.Equal(existing.RequestHash, command.RequestHash) {
			return WorkbookMutationResult{}, authn.ErrClientTxnConflict
		}
		payload, err := decodeStoredPayload(existing.ResponseJSON)
		if err != nil {
			return WorkbookMutationResult{}, fmt.Errorf("decode replayed party patch payload: %w", err)
		}
		return WorkbookMutationResult{Payload: payload, StatusCode: http.StatusOK, Replayed: true, RecordID: command.RecordID, ViewSchemaID: request.ViewSchemaID, ClientTxnID: request.ClientTxnID}, nil
	} else if !errors.Is(err, authn.ErrNotFound) {
		return WorkbookMutationResult{}, fmt.Errorf("query party patch idempotency: %w", err)
	}

	tx, err := f.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return WorkbookMutationResult{}, fmt.Errorf("begin party patch transaction: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	meta, err := loadPartyRecordMetaForUpdateTx(ctx, tx, command.RecordID)
	if err != nil {
		return WorkbookMutationResult{}, err
	}
	if meta.RecordType != "party" {
		return WorkbookMutationResult{}, pgx.ErrNoRows
	}
	if err := f.incidentAccess.EnsureOpenTx(ctx, tx, meta.IncidentID); err != nil {
		return WorkbookMutationResult{}, err
	}
	effectiveBeforeVersion := request.BaseRowVersion
	if meta.RowVersion != request.BaseRowVersion {
		if meta.RowVersion < request.BaseRowVersion {
			return WorkbookMutationResult{}, &RowVersionConflictError{RecordID: command.RecordID, BaseRowVersion: request.BaseRowVersion, CurrentRowVersion: meta.RowVersion}
		}
		windowRows, err := f.revisionHistory.LoadRevisionWindowTx(ctx, tx, command.RecordID, request.BaseRowVersion, meta.RowVersion)
		if err != nil {
			return WorkbookMutationResult{}, adaptRevisionWindowError(command.RecordID, request.BaseRowVersion, meta.RowVersion, err)
		}
		fieldDescriptors := partyConflictFieldDescriptors(request.ViewSchemaID)
		window, err := conflictwindows.BuildPatchConflictWindowWithDescriptors(command.RecordID, request.BaseRowVersion, meta.RowVersion, windowRows, fieldDescriptors)
		if err != nil {
			return WorkbookMutationResult{}, adaptRevisionWindowError(command.RecordID, request.BaseRowVersion, meta.RowVersion, err)
		}
		if change, changed, ok := overlappingPartyPatchChange(request.Changes, window.ChangedFields); ok {
			current, err := f.projectionRows.LoadTx(ctx, tx, command.RecordID)
			if err != nil {
				return WorkbookMutationResult{}, err
			}
			conflictPayload, err := buildPartySameFieldConflict(partySameFieldConflictParams{
				RouteKey:          command.ConflictRouteKey,
				RecordID:          command.RecordID,
				ViewSchemaID:      request.ViewSchemaID,
				BaseRowVersion:    request.BaseRowVersion,
				CurrentRowVersion: meta.RowVersion,
				RequestHash:       command.RequestHash,
				Window:            window,
				Change:            change,
				Changed:           changed,
				CurrentRow:        current,
				FieldDescriptors:  fieldDescriptors,
				Codec:             f.conflictTokens,
			})
			if err != nil {
				return WorkbookMutationResult{}, adaptRevisionWindowError(command.RecordID, request.BaseRowVersion, meta.RowVersion, err)
			}
			return WorkbookMutationResult{}, &SameFieldConflictError{Conflict: conflictPayload}
		}
		effectiveBeforeVersion = meta.RowVersion
	}
	beforeRow, err := f.projectionRows.LoadTx(ctx, tx, command.RecordID)
	if err != nil {
		return WorkbookMutationResult{}, err
	}
	changed, err := f.applyPatchTx(ctx, tx, command.RecordID, request, command.Now.UTC())
	if err != nil {
		return WorkbookMutationResult{}, err
	}
	if !changed {
		return WorkbookMutationResult{}, &ValidationError{Field: "changes", ReasonCode: "no_effective_change"}
	}
	rowVersion, err := f.recordStore.AdvanceVersionTx(ctx, tx, command.RecordID, command.Actor.ID, command.Now.UTC())
	if err != nil {
		return WorkbookMutationResult{}, err
	}
	if err := f.store.TouchPartyTx(ctx, tx, command.RecordID, command.Now.UTC()); err != nil {
		return WorkbookMutationResult{}, err
	}
	if err := f.projectionRows.RefreshTx(ctx, tx, command.RecordID); err != nil {
		return WorkbookMutationResult{}, err
	}
	afterRow, err := f.projectionRows.LoadTx(ctx, tx, command.RecordID)
	if err != nil {
		return WorkbookMutationResult{}, err
	}
	changeSetID, err := f.revisionAppender.AppendChangeSetTx(ctx, tx, revisions.AppendChangeSetParams{
		IncidentID:  meta.IncidentID,
		ActorUserID: command.Actor.ID,
		Source:      command.RouteKey,
		ClientTxnID: &request.ClientTxnID,
		RequestID:   &command.RequestID,
		CreatedAt:   command.Now.UTC(),
	})
	if err != nil {
		return WorkbookMutationResult{}, err
	}
	beforeVersionID := workbookVersionID(command.RecordID, request.BaseRowVersion)
	if effectiveBeforeVersion != request.BaseRowVersion {
		beforeVersionID = workbookVersionID(command.RecordID, effectiveBeforeVersion)
	}
	afterVersionID := workbookVersionID(command.RecordID, rowVersion)
	if err := f.revisionAppender.AppendMutationTx(ctx, tx, revisions.AppendMutationParams{
		ChangeSetID:     changeSetID,
		SequenceNo:      1,
		TargetKind:      "record",
		TargetID:        command.RecordID.String(),
		OperationKind:   "patch",
		BeforeVersionID: &beforeVersionID,
		AfterVersionID:  &afterVersionID,
		BeforeValue:     beforeRow,
		AfterValue:      afterRow,
	}); err != nil {
		return WorkbookMutationResult{}, err
	}
	if err := f.revisionAppender.AppendRecordRevisionTx(ctx, tx, revisions.AppendRecordRevisionParams{
		ChangeSetID: changeSetID,
		RecordID:    command.RecordID,
		RowVersion:  rowVersion,
		BeforeValue: beforeRow,
		AfterValue:  afterRow,
	}); err != nil {
		return WorkbookMutationResult{}, err
	}
	payload := buildMutationPayload(request.ViewSchemaID, changeSetID, afterRow)
	if err := authn.InsertRouteIdempotencyPayload(ctx, tx, idempotencyKey, nil, command.RequestHash, http.StatusOK, payload); err != nil {
		if authn.IsUniqueViolation(err) {
			return WorkbookMutationResult{}, authn.ErrClientTxnConflict
		}
		return WorkbookMutationResult{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return WorkbookMutationResult{}, fmt.Errorf("commit party patch transaction: %w", err)
	}
	return WorkbookMutationResult{
		Payload:          payload,
		StatusCode:       http.StatusOK,
		IncidentID:       meta.IncidentID,
		RecordID:         command.RecordID,
		ChangeSetID:      changeSetID,
		ClientTxnID:      request.ClientTxnID,
		RowVersion:       rowVersion,
		ViewSchemaID:     request.ViewSchemaID,
		ChangedFieldKeys: changedFieldKeys(beforeRow, afterRow),
	}, nil
}

func (f *WorkbookFacade) applyPatchTx(ctx context.Context, tx pgx.Tx, recordID uuid.UUID, request WorkbookPatchRequest, now time.Time) (bool, error) {
	changed := false
	for _, change := range request.Changes {
		if change.Value == nil {
			continue
		}
		if err := ValidateDirectPatchChange(change.FieldKey, *change.Value); err != nil {
			return false, err
		}
		applied, err := f.store.ApplyDirectChangeTx(ctx, tx, recordID, change.FieldKey, *change.Value, now)
		if err != nil {
			return false, err
		}
		changed = changed || applied
	}
	return changed, nil
}

type partyRecordMeta struct {
	IncidentID uuid.UUID
	RecordType string
	RowVersion int64
}

func loadPartyRecordMetaForUpdateTx(ctx context.Context, tx pgx.Tx, recordID uuid.UUID) (partyRecordMeta, error) {
	var meta partyRecordMeta
	var deletedAt sql.NullTime
	err := tx.QueryRow(ctx, `
SELECT incident_id, record_type, row_version, deleted_at
  FROM records
 WHERE record_id = $1
 FOR UPDATE
`, recordID).Scan(&meta.IncidentID, &meta.RecordType, &meta.RowVersion, &deletedAt)
	if err != nil {
		return partyRecordMeta{}, err
	}
	if deletedAt.Valid {
		return partyRecordMeta{}, revisions.ErrRecordDeletedUseRestore
	}
	return meta, nil
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

func rowVersionFromGenericRow(row map[string]any) (int64, error) {
	raw, ok := row["row_version"]
	if !ok {
		return 0, fmt.Errorf("generic row missing row_version")
	}
	switch value := raw.(type) {
	case int64:
		return value, nil
	case int:
		return int64(value), nil
	case int32:
		return int64(value), nil
	case float64:
		return int64(value), nil
	default:
		return 0, fmt.Errorf("generic row has unexpected row_version type %T", value)
	}
}

func workbookVersionID(recordID uuid.UUID, rowVersion int64) string {
	return fmt.Sprintf("record:%s:%d", recordID.String(), rowVersion)
}

func buildMutationPayload(viewSchemaID string, changeSetID uuid.UUID, row map[string]any) map[string]any {
	return map[string]any{"view_schema_id": viewSchemaID, "change_set_id": changeSetID.String(), "row": row}
}

func decodeStoredPayload(data []byte) (map[string]any, error) {
	var payload map[string]any
	if err := json.Unmarshal(data, &payload); err != nil {
		return nil, err
	}
	return payload, nil
}

func extractPayloadUUID(payload map[string]any, path ...string) (uuid.UUID, error) {
	current := any(payload)
	for _, segment := range path {
		object, ok := current.(map[string]any)
		if !ok {
			return uuid.UUID{}, fmt.Errorf("decode payload path %q", strings.Join(path, "."))
		}
		current = object[segment]
	}
	text, ok := current.(string)
	if !ok {
		return uuid.UUID{}, fmt.Errorf("decode payload path %q", strings.Join(path, "."))
	}
	parsed, err := uuid.Parse(text)
	if err != nil {
		return uuid.UUID{}, err
	}
	return parsed, nil
}

func adaptRevisionWindowError(recordID uuid.UUID, baseRowVersion int64, currentRowVersion int64, err error) error {
	if err == nil {
		return nil
	}
	var windowErr *conflictwindows.RevisionWindowError
	if errors.As(err, &windowErr) {
		return &RowVersionConflictError{RecordID: windowErr.RecordID, BaseRowVersion: windowErr.BaseRowVersion, CurrentRowVersion: windowErr.CurrentRowVersion}
	}
	return &RowVersionConflictError{RecordID: recordID, BaseRowVersion: baseRowVersion, CurrentRowVersion: currentRowVersion}
}

type partySameFieldConflictParams struct {
	RouteKey          string
	RecordID          uuid.UUID
	ViewSchemaID      string
	BaseRowVersion    int64
	CurrentRowVersion int64
	RequestHash       []byte
	Window            conflictwindows.PatchConflictWindow
	Change            WorkbookPatchChange
	Changed           conflictwindows.PatchChangedField
	CurrentRow        map[string]any
	FieldDescriptors  conflictwindows.FieldDescriptorSet
	Codec             conflicttokens.ConflictTokenCodec
}

func overlappingPartyPatchChange(changes []WorkbookPatchChange, changedFields map[string]conflictwindows.PatchChangedField) (WorkbookPatchChange, conflictwindows.PatchChangedField, bool) {
	for _, change := range changes {
		changed, ok := changedFields[change.FieldKey]
		if ok {
			return change, changed, true
		}
	}
	return WorkbookPatchChange{}, conflictwindows.PatchChangedField{}, false
}

func buildPartySameFieldConflict(params partySameFieldConflictParams) (map[string]any, error) {
	baseValue, ok := rowCellValue(params.Window.BaseRow, params.Change.FieldKey)
	if !ok {
		return nil, &conflictwindows.RevisionWindowError{RecordID: params.RecordID, BaseRowVersion: params.BaseRowVersion, CurrentRowVersion: params.CurrentRowVersion}
	}
	serverValue, ok := rowCellValue(params.CurrentRow, params.Change.FieldKey)
	if !ok {
		return nil, &conflictwindows.RevisionWindowError{RecordID: params.RecordID, BaseRowVersion: params.BaseRowVersion, CurrentRowVersion: params.CurrentRowVersion}
	}
	clientValue := params.Change.CanonicalValue
	conflictClass := params.FieldDescriptors.ConflictResolutionClass(params.Change.FieldKey)
	if conflictClass == "" {
		conflictClass = "atomic_replace"
	}
	conflict := map[string]any{
		"conflict_token":            partyConflictToken(params.RouteKey, params.RecordID, params.ViewSchemaID, params.Change.FieldKey, conflictClass, params.BaseRowVersion, params.CurrentRowVersion, params.RequestHash, params.Codec),
		"record_id":                 params.RecordID.String(),
		"field_key":                 params.Change.FieldKey,
		"conflict_resolution_class": conflictClass,
		"base_row_version":          params.BaseRowVersion,
		"current_row_version":       params.CurrentRowVersion,
		"client_value":              clientValue,
		"server_value":              serverValue,
		"server_updated_by":         params.Changed.ServerUpdatedBy.String(),
		"server_updated_at":         params.Changed.ServerUpdatedAt.UTC().Format(time.RFC3339Nano),
		"base_value":                baseValue,
	}
	if conflictClass == "text_compare_merge" {
		if suggested, ok := conflictmerge.SuggestedTextMergeValue(baseValue, serverValue, clientValue); ok {
			conflict["suggested_merged_value"] = suggested
		}
	}
	return conflict, nil
}

func partyConflictFieldDescriptors(viewSchemaID string) conflictwindows.FieldDescriptorSet {
	return conflictwindows.ViewSchemaFieldDescriptors(viewSchemaID)
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

func partyConflictToken(routeKey string, recordID uuid.UUID, viewSchemaID string, fieldKey string, conflictClass string, baseRowVersion int64, currentRowVersion int64, requestHash []byte, codec conflicttokens.ConflictTokenCodec) string {
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

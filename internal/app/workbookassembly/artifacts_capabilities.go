package workbookassembly

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"net/http"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/artifacts"
	artifactprojection "github.com/JochiRaider/cartulary/internal/modules/artifacts/workbookprojection"
	"github.com/JochiRaider/cartulary/internal/modules/collaboration"
	"github.com/JochiRaider/cartulary/internal/modules/incidents/admission"
	"github.com/JochiRaider/cartulary/internal/modules/links"
	"github.com/JochiRaider/cartulary/internal/modules/records"
	"github.com/JochiRaider/cartulary/internal/modules/revisions"
	conflicttokens "github.com/JochiRaider/cartulary/internal/modules/revisions/conflicts"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
)

type artifactIdempotency struct {
	store   *authn.Store
	records *records.Store
}

func (a artifactIdempotency) Get(
	ctx context.Context,
	key artifacts.IdempotencyKey,
	requestHash []byte,
) (artifacts.StoredMutationResult, bool, error) {
	record, err := a.store.GetRouteIdempotency(ctx, authn.RouteIdempotencyKey{
		RouteKey: string(key.OperationID), ActorUserID: key.ActorUserID,
		ScopeKey: key.ScopeKey, ClientTxnID: key.ClientTxnID,
	})
	if errors.Is(err, authn.ErrNotFound) {
		return artifacts.StoredMutationResult{}, false, nil
	}
	if err != nil {
		return artifacts.StoredMutationResult{}, false, err
	}
	if !bytes.Equal(record.RequestHash, requestHash) {
		return artifacts.StoredMutationResult{}, false, artifacts.ErrClientTxnConflict
	}
	kind, ok := artifactStoredKindForOperation(key.OperationID)
	if !ok {
		return artifacts.StoredMutationResult{}, false, artifacts.ErrStoredMutationKindMismatch
	}
	if record.StatusCode != artifactStoredStatus(kind) {
		return artifacts.StoredMutationResult{}, false, artifacts.ErrStoredMutationKindMismatch
	}
	recordID, err := artifactStoredRecordID(record.ResponseJSON)
	if err != nil {
		return artifacts.StoredMutationResult{}, false, fmt.Errorf("decode Artifacts stored record identity: %w", err)
	}
	envelope, err := a.records.LoadEnvelope(ctx, recordID)
	if err != nil {
		return artifacts.StoredMutationResult{}, false, fmt.Errorf("load Artifacts stored record envelope: %w", err)
	}
	result, err := decodeArtifactStoredResult(kind, record.ResponseJSON, envelope.IncidentID)
	if err != nil {
		return artifacts.StoredMutationResult{}, false, fmt.Errorf("decode Artifacts stored mutation result: %w", err)
	}
	return result, true, nil
}

func (artifactIdempotency) PutTx(
	ctx context.Context,
	tx pgx.Tx,
	key artifacts.IdempotencyKey,
	requestHash []byte,
	result artifacts.StoredMutationResult,
) error {
	expectedKind, ok := artifactStoredKindForOperation(key.OperationID)
	if !ok || result.Kind() != expectedKind {
		return artifacts.ErrStoredMutationKindMismatch
	}
	payload, err := encodeArtifactStoredResult(result)
	if err != nil {
		return err
	}
	status := artifactStoredStatus(result.Kind())
	err = authn.InsertRouteIdempotencyPayload(ctx, tx, authn.RouteIdempotencyKey{
		RouteKey: string(key.OperationID), ActorUserID: key.ActorUserID,
		ScopeKey: key.ScopeKey, ClientTxnID: key.ClientTxnID,
	}, nil, requestHash, status, payload)
	if authn.IsUniqueViolation(err) {
		return artifacts.ErrClientTxnConflict
	}
	return err
}

func artifactStoredStatus(kind artifacts.StoredMutationKind) int {
	if kind == artifacts.StoredMutationCreate || kind == artifacts.StoredMutationLinkedNote {
		return http.StatusCreated
	}
	return http.StatusOK
}

func artifactStoredKindForOperation(operationID artifacts.OperationID) (artifacts.StoredMutationKind, bool) {
	switch operationID {
	case artifacts.OperationCreate:
		return artifacts.StoredMutationCreate, true
	case artifacts.OperationPatch, artifacts.OperationConflictResolve:
		return artifacts.StoredMutationPatch, true
	case artifacts.OperationLinkedNoteCreate:
		return artifacts.StoredMutationLinkedNote, true
	default:
		return "", false
	}
}

func decodeArtifactStoredResult(
	kind artifacts.StoredMutationKind,
	data []byte,
	incidentID uuid.UUID,
) (artifacts.StoredMutationResult, error) {
	var payload map[string]any
	if err := json.Unmarshal(data, &payload); err != nil {
		return artifacts.StoredMutationResult{}, err
	}
	if !artifactStoredPayloadKeysMatch(kind, payload) {
		return artifacts.StoredMutationResult{}, artifacts.ErrStoredMutationKindMismatch
	}
	viewSchemaID, ok := payload["view_schema_id"].(string)
	if !ok {
		return artifacts.StoredMutationResult{}, fmt.Errorf("view_schema_id is missing")
	}
	row, ok := payload["row"].(map[string]any)
	if !ok {
		return artifacts.StoredMutationResult{}, fmt.Errorf("row is missing")
	}
	recordID, err := artifactPayloadUUID(row, "record_id")
	if err != nil {
		return artifacts.StoredMutationResult{}, err
	}
	changeSetID, err := artifactPayloadUUID(payload, "change_set_id")
	if err != nil {
		return artifacts.StoredMutationResult{}, err
	}
	rowVersion, err := artifactPayloadPositiveInt64(row, "row_version")
	if err != nil {
		return artifacts.StoredMutationResult{}, err
	}
	stored := artifacts.StoredMutationPayload{
		ViewSchemaID: viewSchemaID, IncidentID: incidentID, RecordID: recordID,
		RowVersion: rowVersion, ChangeSetID: &changeSetID, Row: row,
	}
	switch kind {
	case artifacts.StoredMutationCreate:
		return artifacts.NewStoredCreateResult(stored), nil
	case artifacts.StoredMutationPatch:
		return artifacts.NewStoredPatchResult(stored), nil
	case artifacts.StoredMutationLinkedNote:
		sourceRecordID, err := artifactPayloadUUID(payload, "source_record_id")
		if err != nil {
			return artifacts.StoredMutationResult{}, err
		}
		linkType, ok := payload["link_type"].(string)
		if !ok || linkType != "references_artifact" {
			return artifacts.StoredMutationResult{}, fmt.Errorf("link_type is invalid")
		}
		stored.ContextualLink = &artifacts.ContextualLink{SourceRecordID: sourceRecordID, LinkType: linkType}
		return artifacts.NewStoredLinkedNoteResult(stored), nil
	default:
		return artifacts.StoredMutationResult{}, artifacts.ErrStoredMutationKindMismatch
	}
}

func artifactStoredRecordID(data []byte) (uuid.UUID, error) {
	var payload struct {
		Row map[string]any `json:"row"`
	}
	if err := json.Unmarshal(data, &payload); err != nil {
		return uuid.Nil, err
	}
	return artifactPayloadUUID(payload.Row, "record_id")
}

func encodeArtifactStoredResult(result artifacts.StoredMutationResult) (map[string]any, error) {
	stored, ok := result.Payload()
	if !ok || stored.IncidentID == uuid.Nil || stored.RecordID == uuid.Nil || stored.RowVersion < 1 ||
		stored.ChangeSetID == nil || *stored.ChangeSetID == uuid.Nil || stored.Row == nil {
		return nil, artifacts.ErrStoredMutationKindMismatch
	}
	payload := map[string]any{
		"view_schema_id": stored.ViewSchemaID,
		"change_set_id":  stored.ChangeSetID.String(),
		"row":            stored.Row,
	}
	if result.Kind() == artifacts.StoredMutationLinkedNote {
		if stored.ContextualLink == nil || stored.ContextualLink.SourceRecordID == uuid.Nil ||
			stored.ContextualLink.LinkType != "references_artifact" {
			return nil, artifacts.ErrStoredMutationKindMismatch
		}
		payload["source_record_id"] = stored.ContextualLink.SourceRecordID.String()
		payload["link_type"] = stored.ContextualLink.LinkType
	} else if stored.ContextualLink != nil {
		return nil, artifacts.ErrStoredMutationKindMismatch
	}
	return payload, nil
}

func artifactStoredPayloadKeysMatch(kind artifacts.StoredMutationKind, payload map[string]any) bool {
	expected := map[string]struct{}{
		"view_schema_id": {}, "change_set_id": {}, "row": {},
	}
	if kind == artifacts.StoredMutationLinkedNote {
		expected["source_record_id"] = struct{}{}
		expected["link_type"] = struct{}{}
	}
	if len(payload) != len(expected) {
		return false
	}
	for key := range payload {
		if _, ok := expected[key]; !ok {
			return false
		}
	}
	return true
}

func artifactPayloadPositiveInt64(payload map[string]any, key string) (int64, error) {
	value, ok := payload[key].(float64)
	if !ok || value < 1 || value > math.MaxInt64 || value != float64(int64(value)) {
		return 0, fmt.Errorf("%s is invalid", key)
	}
	return int64(value), nil
}

func artifactPayloadUUID(payload map[string]any, key string) (uuid.UUID, error) {
	value, ok := payload[key].(string)
	if !ok {
		return uuid.Nil, fmt.Errorf("%s is missing", key)
	}
	result, err := uuid.Parse(value)
	if err != nil {
		return uuid.Nil, fmt.Errorf("%s is invalid: %w", key, err)
	}
	return result, nil
}

type artifactRevisions struct {
	appender *revisions.Appender
	history  conflicttokens.RevisionWindowReader
}

func (a artifactRevisions) CaptureRecordSnapshotTx(ctx context.Context, tx pgx.Tx, recordID uuid.UUID) (revisions.RecordSnapshot, error) {
	return a.appender.CaptureRecordSnapshotTx(ctx, tx, recordID)
}

func (a artifactRevisions) AppendChangeSetTx(ctx context.Context, tx pgx.Tx, params revisions.AppendChangeSetParams) (uuid.UUID, error) {
	return a.appender.AppendChangeSetTx(ctx, tx, params)
}

func (a artifactRevisions) AppendNonRowMutationTx(ctx context.Context, tx pgx.Tx, params revisions.AppendNonRowMutationParams) error {
	return a.appender.AppendNonRowMutationTx(ctx, tx, params)
}

func (a artifactRevisions) AppendRecordMutationTx(ctx context.Context, tx pgx.Tx, params revisions.AppendRecordMutationParams) error {
	return a.appender.AppendRecordMutationTx(ctx, tx, params)
}

func (a artifactRevisions) AppendLiveRevisionTx(ctx context.Context, tx pgx.Tx, input revisions.LiveRevisionInput) error {
	return a.appender.AppendLiveRevisionTx(ctx, tx, input)
}

func (a artifactRevisions) LoadRevisionWindowTx(ctx context.Context, tx pgx.Tx, recordID uuid.UUID, baseVersion int64, currentVersion int64) ([]conflicttokens.RevisionWindowRow, error) {
	return a.history.LoadRevisionWindowTx(ctx, tx, recordID, baseVersion, currentVersion)
}

func NewArtifactMutationContribution(
	pool postgres.DB,
	conflictTokens conflicttokens.ConflictTokenCodec,
	appender *revisions.Appender,
	conflictFields conflicttokens.FieldResolver,
	projectionRows artifactprojection.Rows,
	publications collaboration.RecordChangedAppender,
) (*artifacts.MutationFacade, error) {
	if appender == nil {
		return nil, fmt.Errorf("compose Artifacts mutation contribution: Revisions appender is required")
	}
	authStore := authn.NewStore(pool)
	recordStore := records.NewStore(pool)
	facade, err := artifacts.NewMutationContribution(pool, conflictTokens, artifacts.MutationDependencies{
		IncidentState:        admission.NewChecker(pool),
		MemberReferences:     artifacts.NewMemberReferenceCapability(),
		Idempotency:          artifactIdempotency{store: authStore, records: recordStore},
		RecordEnvelopes:      recordStore,
		Links:                links.NewStore(),
		Projections:          projectionRows,
		Revisions:            artifactRevisions{appender: appender, history: conflicttokens.NewRevisionWindowReader()},
		ConflictFields:       conflictFields,
		KeepSavedIdempotency: NewConflictIdempotencyPort(pool),
		Collaboration:        publications,
	})
	if err != nil {
		return nil, fmt.Errorf("compose Artifacts mutation contribution: %w", err)
	}
	return facade, nil
}

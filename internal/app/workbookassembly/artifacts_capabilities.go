package workbookassembly

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/artifacts"
	artifactprojection "github.com/JochiRaider/cartulary/internal/modules/artifacts/workbookprojection"
	"github.com/JochiRaider/cartulary/internal/modules/incidents"
	"github.com/JochiRaider/cartulary/internal/modules/links"
	"github.com/JochiRaider/cartulary/internal/modules/records"
	"github.com/JochiRaider/cartulary/internal/modules/revisions"
	conflicttokens "github.com/JochiRaider/cartulary/internal/modules/revisions/conflicts"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
)

type artifactIdempotency struct {
	store *authn.Store
}

func (a artifactIdempotency) Get(
	ctx context.Context,
	key artifacts.IdempotencyKey,
	requestHash []byte,
) (artifacts.IdempotencyRecord, error) {
	record, err := a.store.GetRouteIdempotency(ctx, authn.RouteIdempotencyKey{
		RouteKey: string(key.OperationID), ActorUserID: key.ActorUserID,
		ScopeKey: key.ScopeKey, ClientTxnID: key.ClientTxnID,
	})
	if errors.Is(err, authn.ErrNotFound) {
		return artifacts.IdempotencyRecord{}, artifacts.ErrIdempotencyNotFound
	}
	if err != nil {
		return artifacts.IdempotencyRecord{}, err
	}
	if !bytes.Equal(record.RequestHash, requestHash) {
		return artifacts.IdempotencyRecord{RequestHash: record.RequestHash}, nil
	}
	kind, ok := artifactStoredKindForOperation(key.OperationID)
	if !ok {
		return artifacts.IdempotencyRecord{}, artifacts.ErrStoredMutationKindMismatch
	}
	result, err := decodeArtifactStoredResult(kind, record.ResponseJSON)
	if err != nil {
		return artifacts.IdempotencyRecord{}, fmt.Errorf("decode Artifacts stored mutation result: %w", err)
	}
	return artifacts.IdempotencyRecord{RequestHash: record.RequestHash, Result: result}, nil
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
	status := http.StatusOK
	if result.Kind() == artifacts.StoredMutationCreate || result.Kind() == artifacts.StoredMutationLinkedNote {
		status = http.StatusCreated
	}
	err = authn.InsertRouteIdempotencyPayload(ctx, tx, authn.RouteIdempotencyKey{
		RouteKey: string(key.OperationID), ActorUserID: key.ActorUserID,
		ScopeKey: key.ScopeKey, ClientTxnID: key.ClientTxnID,
	}, nil, requestHash, status, payload)
	if authn.IsUniqueViolation(err) {
		return artifacts.ErrClientTxnConflict
	}
	return err
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
) (artifacts.StoredMutationResult, error) {
	var payload map[string]any
	if err := json.Unmarshal(data, &payload); err != nil {
		return artifacts.StoredMutationResult{}, err
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
	stored := artifacts.StoredMutationPayload{
		ViewSchemaID: viewSchemaID, RecordID: recordID, ChangeSetID: changeSetID, Row: row,
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
		stored.SourceRecordID = &sourceRecordID
		stored.LinkType = linkType
		return artifacts.NewStoredLinkedNoteResult(stored), nil
	default:
		return artifacts.StoredMutationResult{}, artifacts.ErrStoredMutationKindMismatch
	}
}

func encodeArtifactStoredResult(result artifacts.StoredMutationResult) (map[string]any, error) {
	stored, ok := result.Payload()
	if !ok {
		return nil, artifacts.ErrStoredMutationKindMismatch
	}
	payload := map[string]any{
		"view_schema_id": stored.ViewSchemaID,
		"change_set_id":  stored.ChangeSetID.String(),
		"row":            stored.Row,
	}
	if result.Kind() == artifacts.StoredMutationLinkedNote {
		if stored.SourceRecordID == nil || stored.LinkType != "references_artifact" {
			return nil, artifacts.ErrStoredMutationKindMismatch
		}
		payload["source_record_id"] = stored.SourceRecordID.String()
		payload["link_type"] = stored.LinkType
	}
	return payload, nil
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

func (a artifactRevisions) AppendRecordRevisionAndIntentTx(ctx context.Context, tx pgx.Tx, params revisions.AppendRecordRevisionParams) error {
	return a.appender.AppendRecordRevisionAndIntentTx(ctx, tx, params)
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
) (*artifacts.MutationFacade, error) {
	if appender == nil {
		return nil, fmt.Errorf("compose Artifacts mutation contribution: Revisions appender is required")
	}
	authStore := authn.NewStore(pool)
	facade, err := artifacts.NewMutationContribution(pool, conflictTokens, artifacts.MutationDependencies{
		IncidentState:        incidents.NewAccess(pool),
		MemberReferences:     artifacts.NewMemberReferenceCapability(),
		Idempotency:          artifactIdempotency{store: authStore},
		RecordEnvelopes:      records.NewStore(),
		Links:                links.NewStore(),
		Projections:          projectionRows,
		Revisions:            artifactRevisions{appender: appender, history: conflicttokens.NewRevisionWindowReader()},
		ConflictFields:       conflictFields,
		KeepSavedIdempotency: NewConflictIdempotencyPort(pool),
	})
	if err != nil {
		return nil, fmt.Errorf("compose Artifacts mutation contribution: %w", err)
	}
	return facade, nil
}

package artifacts

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/base64"
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

	artifactprojection "github.com/JochiRaider/cartulary/internal/modules/artifacts/projectionprovider"
	"github.com/JochiRaider/cartulary/internal/modules/artifacts/riskrefs"
	"github.com/JochiRaider/cartulary/internal/modules/incidents"
	"github.com/JochiRaider/cartulary/internal/modules/links"
	"github.com/JochiRaider/cartulary/internal/modules/projections"
	"github.com/JochiRaider/cartulary/internal/modules/records"
	"github.com/JochiRaider/cartulary/internal/modules/revisions"
	conflicttokens "github.com/JochiRaider/cartulary/internal/modules/revisions/conflicts"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
)

type WorkbookFacade struct {
	pool             postgres.DB
	authStore        *authn.Store
	idempotency      artifactIdempotencyPort
	incidentAccess   artifactIncidentAdmissionPort
	linkStore        artifactLinkPort
	revisionHistory  artifactRevisionHistoryPort
	revisionAppender artifactRevisionPort
	source           artifactSourceKernel
	conflictTokens   conflicttokens.ConflictTokenCodec
	conflictFields   conflicttokens.FieldResolver
	keepSaved        conflicttokens.IdempotencyPort
}

type artifactLinkPort interface {
	ValidatePartyRefCollectionTx(context.Context, pgx.Tx, links.PartyRefCollectionValidation) error
	ValidateRecordRefCollectionTx(context.Context, pgx.Tx, links.RecordRefCollectionValidation) error
	ValidateTagCollectionTx(context.Context, pgx.Tx, links.TagCollectionValidation) error
	ApplyPartyRefCollectionTx(context.Context, pgx.Tx, links.PartyRefCollectionCommand) (bool, error)
	ApplyRecordRefCollectionTx(context.Context, pgx.Tx, links.RecordRefCollectionCommand) (bool, error)
	ApplyTagCollectionTx(context.Context, pgx.Tx, links.TagCollectionCommand) (bool, error)
	InsertLinkedNoteReferenceTx(context.Context, pgx.Tx, uuid.UUID, uuid.UUID, uuid.UUID, uuid.UUID, time.Time) (links.RecordLink, bool, error)
	LoadRecordLinkValueTx(context.Context, pgx.Tx, uuid.UUID) (map[string]any, error)
}

type WorkbookCreateRequest struct {
	ViewSchemaID string
	ClientTxnID  string
	Values       map[string]FieldValue
	Collections  map[string]WorkbookCollectionActionPayload
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
	Collection     *WorkbookCollectionActionPayload
	CanonicalValue any
}

type WorkbookCollectionActionPayload struct {
	Actions []WorkbookCollectionAction
}

type WorkbookCollectionAction struct {
	Op             string
	RawText        string
	LinkedRecordID *uuid.UUID
	PartyID        *uuid.UUID
	ItemRef        string
	RiskRefText    string
	NormalizedText string
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
	return "artifacts: row version conflict"
}

type SameFieldConflictError struct {
	Conflict map[string]any
}

func (e *SameFieldConflictError) Error() string {
	return "artifacts: same field conflict"
}

func NewWorkbookFacade(pool postgres.DB, conflictTokens conflicttokens.ConflictTokenCodec, appender *revisions.Appender, conflictFields conflicttokens.FieldResolver, keepSaved conflicttokens.IdempotencyPort) *WorkbookFacade {
	authStore := authn.NewStore(pool)
	projectionRows := projections.NewArtifactRows(pool, artifactprojection.QuerySurfaces()...)
	return &WorkbookFacade{
		pool:             pool,
		authStore:        authStore,
		idempotency:      artifactIdempotencyAdapter{store: authStore},
		incidentAccess:   incidents.NewAccess(pool),
		linkStore:        links.NewStore(),
		revisionHistory:  conflicttokens.NewRevisionWindowReader(),
		revisionAppender: appender,
		source: artifactSourceKernel{
			records:     records.NewStore(),
			rows:        newSourceStore(appender),
			projections: projectionRows,
		},
		conflictTokens: conflictTokens,
		conflictFields: conflictFields,
		keepSaved:      keepSaved,
	}
}

func (f *WorkbookFacade) Create(ctx context.Context, command WorkbookCreateCommand) (WorkbookMutationResult, error) {
	return f.create(ctx, command, nil)
}

func (f *WorkbookFacade) create(ctx context.Context, command WorkbookCreateCommand, contextualSourceRecordID *uuid.UUID) (WorkbookMutationResult, error) {
	request := command.Request
	scopeKey := command.IncidentID.String() + ":" + request.ViewSchemaID
	if contextualSourceRecordID != nil {
		scopeKey = contextualSourceRecordID.String()
	}
	idempotencyKey := authn.RouteIdempotencyKey{
		RouteKey:    command.RouteKey,
		ActorUserID: command.Actor.ID,
		ScopeKey:    scopeKey,
		ClientTxnID: request.ClientTxnID,
	}
	if existing, err := f.idempotency.Get(ctx, idempotencyKey); err == nil {
		if !bytes.Equal(existing.RequestHash, command.RequestHash) {
			return WorkbookMutationResult{}, authn.ErrClientTxnConflict
		}
		payload, err := decodeStoredResponse(existing.ResponseJSON)
		if err != nil {
			return WorkbookMutationResult{}, fmt.Errorf("decode replayed artifact create payload: %w", err)
		}
		recordID, err := extractPayloadUUID(payload, "row", "record_id")
		if err != nil {
			return WorkbookMutationResult{}, err
		}
		return WorkbookMutationResult{Payload: payload, StatusCode: http.StatusOK, Replayed: true, IncidentID: command.IncidentID, RecordID: recordID, ViewSchemaID: request.ViewSchemaID, ClientTxnID: request.ClientTxnID}, nil
	} else if !errors.Is(err, authn.ErrNotFound) {
		return WorkbookMutationResult{}, fmt.Errorf("query artifact create idempotency: %w", err)
	}
	if err := ValidateCreateParams(CreateParams{ViewSchemaID: request.ViewSchemaID, Values: request.Values}); err != nil {
		return WorkbookMutationResult{}, err
	}

	tx, err := f.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return WorkbookMutationResult{}, fmt.Errorf("begin artifact create transaction: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	mutation, err := f.executeCreateTx(ctx, tx, command, contextualSourceRecordID)
	if err != nil {
		return WorkbookMutationResult{}, err
	}
	if err := f.idempotency.PutTx(ctx, tx, idempotencyKey, command.RequestHash, http.StatusCreated, mutation.payload); err != nil {
		if authn.IsUniqueViolation(err) {
			return WorkbookMutationResult{}, authn.ErrClientTxnConflict
		}
		return WorkbookMutationResult{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return WorkbookMutationResult{}, fmt.Errorf("commit artifact create transaction: %w", err)
	}
	return WorkbookMutationResult{
		Payload:          mutation.payload,
		StatusCode:       http.StatusCreated,
		IncidentID:       mutation.incidentID,
		RecordID:         mutation.recordID,
		ChangeSetID:      mutation.changeSetID,
		ClientTxnID:      request.ClientTxnID,
		RowVersion:       1,
		ViewSchemaID:     request.ViewSchemaID,
		ChangedFieldKeys: changedFieldKeys(nil, mutation.row),
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
	if existing, err := f.idempotency.Get(ctx, idempotencyKey); err == nil {
		if !bytes.Equal(existing.RequestHash, command.RequestHash) {
			return WorkbookMutationResult{}, authn.ErrClientTxnConflict
		}
		payload, err := decodeStoredResponse(existing.ResponseJSON)
		if err != nil {
			return WorkbookMutationResult{}, fmt.Errorf("decode replayed artifact patch payload: %w", err)
		}
		return WorkbookMutationResult{Payload: payload, StatusCode: http.StatusOK, Replayed: true, RecordID: command.RecordID, ViewSchemaID: request.ViewSchemaID, ClientTxnID: request.ClientTxnID}, nil
	} else if !errors.Is(err, authn.ErrNotFound) {
		return WorkbookMutationResult{}, fmt.Errorf("query artifact patch idempotency: %w", err)
	}

	tx, err := f.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return WorkbookMutationResult{}, fmt.Errorf("begin artifact patch transaction: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	meta, err := loadArtifactRecordMetaForUpdateTx(ctx, tx, command.RecordID)
	if err != nil {
		return WorkbookMutationResult{}, err
	}
	if meta.RecordType != "artifact" {
		return WorkbookMutationResult{}, pgx.ErrNoRows
	}
	if err := validateArtifactViewRecordTx(ctx, tx, command.RecordID, request.ViewSchemaID); err != nil {
		return WorkbookMutationResult{}, err
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
		fieldDescriptors, err := f.conflictFields.ResolveViewSchema(request.ViewSchemaID)
		if err != nil {
			return WorkbookMutationResult{}, adaptRevisionWindowError(command.RecordID, request.BaseRowVersion, meta.RowVersion, err)
		}
		window, err := conflicttokens.BuildPatchConflictWindowWithDescriptors(command.RecordID, request.BaseRowVersion, meta.RowVersion, windowRows, fieldDescriptors)
		if err != nil {
			return WorkbookMutationResult{}, adaptRevisionWindowError(command.RecordID, request.BaseRowVersion, meta.RowVersion, err)
		}
		if change, changed, ok := overlappingArtifactPatchChange(request.Changes, window.ChangedFields); ok {
			current, err := f.source.projections.LoadTx(ctx, tx, request.ViewSchemaID, command.RecordID)
			if err != nil {
				return WorkbookMutationResult{}, err
			}
			conflictPayload, err := buildArtifactSameFieldConflict(artifactSameFieldConflictParams{
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
	beforeRow, err := f.source.projections.LoadTx(ctx, tx, request.ViewSchemaID, command.RecordID)
	if err != nil {
		return WorkbookMutationResult{}, err
	}
	if err := validateArtifactPatchReferencesTx(ctx, tx, f.linkStore, meta.IncidentID, request); err != nil {
		return WorkbookMutationResult{}, err
	}
	changed, err := f.applyPatchTx(ctx, tx, meta.IncidentID, command.RecordID, command.Actor.ID, request, command.Now.UTC())
	if err != nil {
		return WorkbookMutationResult{}, err
	}
	if !changed {
		return WorkbookMutationResult{}, &ValidationError{Field: "changes", ReasonCode: "no_effective_change"}
	}
	rowVersion, err := f.source.records.AdvanceVersionTx(ctx, tx, command.RecordID, command.Actor.ID, command.Now.UTC())
	if err != nil {
		return WorkbookMutationResult{}, err
	}
	if err := f.source.rows.TouchRowTx(ctx, tx, command.RecordID, command.Now.UTC()); err != nil {
		return WorkbookMutationResult{}, err
	}
	if err := f.source.projections.RefreshTx(ctx, tx, command.RecordID); err != nil {
		return WorkbookMutationResult{}, err
	}
	afterRow, err := f.source.projections.LoadTx(ctx, tx, request.ViewSchemaID, command.RecordID)
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
	if err := f.idempotency.PutTx(ctx, tx, idempotencyKey, command.RequestHash, http.StatusOK, payload); err != nil {
		if authn.IsUniqueViolation(err) {
			return WorkbookMutationResult{}, authn.ErrClientTxnConflict
		}
		return WorkbookMutationResult{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return WorkbookMutationResult{}, fmt.Errorf("commit artifact patch transaction: %w", err)
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

func (f *WorkbookFacade) applyPatchTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID, recordID uuid.UUID, actorID uuid.UUID, request WorkbookPatchRequest, now time.Time) (bool, error) {
	changed := false
	for _, change := range request.Changes {
		if change.Value != nil {
			policy, ok := lookupArtifactSourceField(change.FieldKey)
			if !ok || policy.viewSchemaID != request.ViewSchemaID ||
				policy.kind != sourceFieldDirect || !policy.writable {
				return false, &ValidationError{Field: change.FieldKey, ReasonCode: "unsupported_field_key"}
			}
			if err := ValidateDirectPatchChange(change.FieldKey, *change.Value); err != nil {
				return false, err
			}
			applied, err := f.source.rows.ApplyDirectChangeTx(ctx, tx, recordID, change.FieldKey, *change.Value, now)
			if err != nil && strings.Contains(err.Error(), "unsupported field key") {
				return false, &ValidationError{Field: change.FieldKey, ReasonCode: "unsupported_field_key"}
			}
			if err != nil {
				return false, err
			}
			changed = changed || applied
			continue
		}
		if change.Collection != nil {
			applied, err := f.applyCollectionTx(ctx, tx, incidentID, recordID, actorID, request.ViewSchemaID, change.FieldKey, *change.Collection, now)
			if err != nil {
				return false, err
			}
			changed = changed || applied
		}
	}
	if request.ViewSchemaID == FindingsViewSchemaID && touchesArtifactField(request.Changes, "finding.state") {
		applied, err := f.source.rows.NormalizeFindingLifecycleTx(ctx, tx, recordID, now)
		if err != nil {
			return false, err
		}
		changed = changed || applied
	}
	return changed, nil
}

func validateArtifactReferencesTx(ctx context.Context, tx pgx.Tx, linkStore artifactLinkPort, incidentID uuid.UUID, viewSchemaID string, values map[string]FieldValue, collections map[string]WorkbookCollectionActionPayload) error {
	for fieldKey, value := range values {
		if value.UUID != nil && strings.HasSuffix(fieldKey, "_user_id") {
			if err := validateIncidentMemberUserTx(ctx, tx, incidentID, *value.UUID, fieldKey); err != nil {
				return err
			}
		}
	}
	for fieldKey, payload := range collections {
		if err := validateArtifactCollectionPayloadTx(ctx, tx, linkStore, incidentID, viewSchemaID, fieldKey, payload); err != nil {
			return err
		}
	}
	return nil
}

func validateArtifactPatchReferencesTx(ctx context.Context, tx pgx.Tx, linkStore artifactLinkPort, incidentID uuid.UUID, request WorkbookPatchRequest) error {
	for _, change := range request.Changes {
		if change.Value != nil && change.Value.UUID != nil && strings.HasSuffix(change.FieldKey, "_user_id") {
			if err := validateIncidentMemberUserTx(ctx, tx, incidentID, *change.Value.UUID, change.FieldKey); err != nil {
				return err
			}
		}
		if change.Collection != nil {
			if err := validateArtifactCollectionPayloadTx(ctx, tx, linkStore, incidentID, request.ViewSchemaID, change.FieldKey, *change.Collection); err != nil {
				return err
			}
		}
	}
	return nil
}

func validateArtifactCollectionPayloadTx(ctx context.Context, tx pgx.Tx, linkStore artifactLinkPort, incidentID uuid.UUID, viewSchemaID string, fieldKey string, payload WorkbookCollectionActionPayload) error {
	sourcePolicy, ok := lookupArtifactSourceField(fieldKey)
	if !ok || sourcePolicy.viewSchemaID != viewSchemaID ||
		sourcePolicy.kind != sourceFieldCollection || !sourcePolicy.writable {
		return collectionValidationError(fieldKey)
	}
	policy := sourcePolicy.collection
	if policy.AllowsRiskRefs() {
		return ValidateHandoffRiskRefPayload(riskRefPayloadFromWorkbook(payload))
	}
	switch {
	case policy.AllowsRecordRefs():
		command, err := artifactRecordRefValidation(incidentID, policy, payload)
		if err != nil {
			return err
		}
		return linkStore.ValidateRecordRefCollectionTx(ctx, tx, command)
	case policy.AllowsPartyRefs():
		command, err := artifactPartyRefValidation(incidentID, policy, payload)
		if err != nil {
			return err
		}
		return linkStore.ValidatePartyRefCollectionTx(ctx, tx, command)
	case policy.AllowsTags():
		command, err := artifactTagValidation(policy, payload)
		if err != nil {
			return err
		}
		return linkStore.ValidateTagCollectionTx(ctx, tx, command)
	default:
		return collectionValidationError(fieldKey)
	}
}

func (f *WorkbookFacade) applyCollectionsTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID, recordID uuid.UUID, actorID uuid.UUID, viewSchemaID string, collections map[string]WorkbookCollectionActionPayload, now time.Time) error {
	for fieldKey, payload := range collections {
		if _, err := f.applyCollectionTx(ctx, tx, incidentID, recordID, actorID, viewSchemaID, fieldKey, payload, now); err != nil {
			return err
		}
	}
	return nil
}

func (f *WorkbookFacade) applyCollectionTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID, recordID uuid.UUID, actorID uuid.UUID, viewSchemaID string, fieldKey string, payload WorkbookCollectionActionPayload, now time.Time) (bool, error) {
	sourcePolicy, ok := lookupArtifactSourceField(fieldKey)
	if !ok || sourcePolicy.viewSchemaID != viewSchemaID ||
		sourcePolicy.kind != sourceFieldCollection || !sourcePolicy.writable {
		return false, collectionValidationError(fieldKey)
	}
	policy := sourcePolicy.collection
	if policy.AllowsRiskRefs() {
		changed, err := f.source.rows.ApplyHandoffRiskRefPayloadTx(ctx, tx, incidentID, recordID, actorID, riskRefPayloadFromWorkbook(payload), now)
		return changed, err
	}
	switch {
	case policy.AllowsRecordRefs():
		command, err := artifactRecordRefCommand(incidentID, recordID, actorID, policy, payload, now)
		if err != nil {
			return false, err
		}
		return f.linkStore.ApplyRecordRefCollectionTx(ctx, tx, command)
	case policy.AllowsPartyRefs():
		command, err := artifactPartyRefCommand(incidentID, recordID, actorID, policy, payload, now)
		if err != nil {
			return false, err
		}
		return f.linkStore.ApplyPartyRefCollectionTx(ctx, tx, command)
	case policy.AllowsTags():
		command, err := artifactTagCommand(incidentID, recordID, actorID, policy, payload, now)
		if err != nil {
			return false, err
		}
		return f.linkStore.ApplyTagCollectionTx(ctx, tx, command)
	default:
		return false, collectionValidationError(fieldKey)
	}
}

func artifactRecordRefValidation(incidentID uuid.UUID, policy CollectionPolicy, payload WorkbookCollectionActionPayload) (links.RecordRefCollectionValidation, error) {
	adds, removes, err := artifactRecordRefActions(policy, payload)
	return links.RecordRefCollectionValidation{IncidentID: incidentID, FieldKey: policy.FieldKey, LinkType: links.LinkType(policy.LinkType), ExpectedTargetType: policy.ExpectedTargetType, AddRecordIDs: adds, RemoveRecordIDs: removes}, err
}

func artifactRecordRefCommand(incidentID uuid.UUID, recordID uuid.UUID, actorID uuid.UUID, policy CollectionPolicy, payload WorkbookCollectionActionPayload, now time.Time) (links.RecordRefCollectionCommand, error) {
	adds, removes, err := artifactRecordRefActions(policy, payload)
	return links.RecordRefCollectionCommand{IncidentID: incidentID, SourceRecordID: recordID, ActorUserID: actorID, FieldKey: policy.FieldKey, LinkType: links.LinkType(policy.LinkType), ExpectedTargetType: policy.ExpectedTargetType, AddRecordIDs: adds, RemoveRecordIDs: removes, Now: now}, err
}

func artifactRecordRefActions(policy CollectionPolicy, payload WorkbookCollectionActionPayload) ([]uuid.UUID, []uuid.UUID, error) {
	adds := make([]uuid.UUID, 0)
	removes := make([]uuid.UUID, 0)
	for _, action := range payload.Actions {
		if !policy.AllowsOp(action.Op) {
			return nil, nil, collectionValidationError(policy.FieldKey)
		}
		switch action.Op {
		case "add_record_ref":
			if action.LinkedRecordID == nil {
				return nil, nil, collectionValidationError(policy.FieldKey)
			}
			adds = append(adds, *action.LinkedRecordID)
		case "remove_record_ref":
			recordID, err := links.ParseRecordRefItemRef(action.ItemRef)
			if err != nil {
				return nil, nil, collectionValidationError(policy.FieldKey)
			}
			removes = append(removes, recordID)
		default:
			return nil, nil, collectionValidationError(policy.FieldKey)
		}
	}
	return adds, removes, nil
}

func artifactPartyRefValidation(incidentID uuid.UUID, policy CollectionPolicy, payload WorkbookCollectionActionPayload) (links.PartyRefCollectionValidation, error) {
	adds, removes, err := artifactPartyRefActions(policy, payload)
	return links.PartyRefCollectionValidation{IncidentID: incidentID, FieldKey: policy.FieldKey, LinkType: links.LinkType(policy.LinkType), ExpectedTargetType: policy.ExpectedTargetType, AddPartyIDs: adds, RemovePartyIDs: removes}, err
}

func artifactPartyRefCommand(incidentID uuid.UUID, recordID uuid.UUID, actorID uuid.UUID, policy CollectionPolicy, payload WorkbookCollectionActionPayload, now time.Time) (links.PartyRefCollectionCommand, error) {
	adds, removes, err := artifactPartyRefActions(policy, payload)
	return links.PartyRefCollectionCommand{IncidentID: incidentID, SourceRecordID: recordID, ActorUserID: actorID, FieldKey: policy.FieldKey, LinkType: links.LinkType(policy.LinkType), ExpectedTargetType: policy.ExpectedTargetType, AddPartyIDs: adds, RemovePartyIDs: removes, Now: now}, err
}

func artifactPartyRefActions(policy CollectionPolicy, payload WorkbookCollectionActionPayload) ([]uuid.UUID, []uuid.UUID, error) {
	adds := make([]uuid.UUID, 0)
	removes := make([]uuid.UUID, 0)
	for _, action := range payload.Actions {
		if !policy.AllowsOp(action.Op) {
			return nil, nil, collectionValidationError(policy.FieldKey)
		}
		switch action.Op {
		case "add_party_ref":
			if action.PartyID == nil {
				return nil, nil, collectionValidationError(policy.FieldKey)
			}
			adds = append(adds, *action.PartyID)
		case "remove_party_ref":
			partyID, err := links.ParsePartyRefItemRef(action.ItemRef)
			if err != nil {
				return nil, nil, collectionValidationError(policy.FieldKey)
			}
			removes = append(removes, partyID)
		default:
			return nil, nil, collectionValidationError(policy.FieldKey)
		}
	}
	return adds, removes, nil
}

func artifactTagValidation(policy CollectionPolicy, payload WorkbookCollectionActionPayload) (links.TagCollectionValidation, error) {
	adds, removes, err := artifactTagActions(policy, payload)
	return links.TagCollectionValidation{FieldKey: policy.FieldKey, AddTags: adds, RemoveTags: removes}, err
}

func artifactTagCommand(incidentID uuid.UUID, recordID uuid.UUID, actorID uuid.UUID, policy CollectionPolicy, payload WorkbookCollectionActionPayload, now time.Time) (links.TagCollectionCommand, error) {
	adds, removes, err := artifactTagActions(policy, payload)
	return links.TagCollectionCommand{IncidentID: incidentID, RecordID: recordID, ActorUserID: actorID, FieldKey: policy.FieldKey, AddTags: adds, RemoveTags: removes, Now: now}, err
}

func artifactTagActions(policy CollectionPolicy, payload WorkbookCollectionActionPayload) ([]links.TagCollectionAdd, []links.RecordTagRef, error) {
	adds := make([]links.TagCollectionAdd, 0)
	removes := make([]links.RecordTagRef, 0)
	for _, action := range payload.Actions {
		if !policy.AllowsOp(action.Op) {
			return nil, nil, collectionValidationError(policy.FieldKey)
		}
		switch action.Op {
		case "add_tag":
			adds = append(adds, links.TagCollectionAdd{RawText: action.RawText, NormalizedText: action.NormalizedText})
		case "remove_tag":
			recordID, tagID, err := links.ParseRecordTagItemRef(action.ItemRef)
			if err != nil {
				return nil, nil, collectionValidationError(policy.FieldKey)
			}
			removes = append(removes, links.RecordTagRef{RecordID: recordID, RecordTagID: tagID})
		default:
			return nil, nil, collectionValidationError(policy.FieldKey)
		}
	}
	return adds, removes, nil
}

func riskRefPayloadFromWorkbook(payload WorkbookCollectionActionPayload) RiskRefActionPayload {
	actions := make([]RiskRefAction, 0, len(payload.Actions))
	for _, action := range payload.Actions {
		actions = append(actions, RiskRefAction{
			Op:             action.Op,
			ItemRef:        action.ItemRef,
			RiskRefText:    action.RiskRefText,
			NormalizedText: action.NormalizedText,
		})
	}
	return RiskRefActionPayload{Actions: actions}
}

func collectionValidationError(fieldKey string) *links.CollectionValidationError {
	return &links.CollectionValidationError{Field: fieldKey, ReasonCode: "invalid_value"}
}

func validateIncidentMemberUserTx(ctx context.Context, tx pgx.Tx, incidentID, userID uuid.UUID, field string) error {
	var exists bool
	if err := tx.QueryRow(ctx, `
SELECT EXISTS (
  SELECT 1
    FROM users u
    JOIN incident_memberships m ON m.user_id = u.id
   WHERE u.id = $1
     AND u.is_active = true
     AND m.incident_id = $2
)`, userID, incidentID).Scan(&exists); err != nil {
		return fmt.Errorf("validate user: %w", err)
	}
	if !exists {
		return &ValidationError{Field: field, ReasonCode: "invalid_value"}
	}
	return nil
}

type artifactRecordMeta struct {
	IncidentID uuid.UUID
	RecordType string
	RowVersion int64
}

func loadArtifactRecordMetaForUpdateTx(ctx context.Context, tx pgx.Tx, recordID uuid.UUID) (artifactRecordMeta, error) {
	var meta artifactRecordMeta
	var deletedAt sql.NullTime
	err := tx.QueryRow(ctx, `
SELECT incident_id, record_type, row_version, deleted_at
  FROM records
 WHERE record_id = $1
 FOR UPDATE
`, recordID).Scan(&meta.IncidentID, &meta.RecordType, &meta.RowVersion, &deletedAt)
	if err != nil {
		return artifactRecordMeta{}, err
	}
	if deletedAt.Valid {
		return artifactRecordMeta{}, revisions.ErrRecordDeletedUseRestore
	}
	return meta, nil
}

func validateArtifactViewRecordTx(ctx context.Context, tx pgx.Tx, recordID uuid.UUID, viewSchemaID string) error {
	var exists bool
	if err := tx.QueryRow(ctx, `
SELECT EXISTS (
    SELECT 1
      FROM artifacts
     WHERE record_id = $1
       AND artifact_type = $2
)
`, recordID, ArtifactTypeForView(viewSchemaID)).Scan(&exists); err != nil {
		return err
	}
	if !exists {
		return pgx.ErrNoRows
	}
	return nil
}

func touchesArtifactField(changes []WorkbookPatchChange, field string) bool {
	for _, change := range changes {
		if change.FieldKey == field {
			return true
		}
	}
	return false
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

func buildMutationPayload(viewSchemaID string, changeSetID uuid.UUID, row map[string]any) map[string]any {
	return map[string]any{
		"view_schema_id": viewSchemaID,
		"change_set_id":  changeSetID.String(),
		"row":            row,
	}
}

func decodeStoredResponse(data []byte) (map[string]any, error) {
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
	var windowErr *conflicttokens.RevisionWindowError
	if errors.As(err, &windowErr) {
		return &RowVersionConflictError{RecordID: windowErr.RecordID, BaseRowVersion: windowErr.BaseRowVersion, CurrentRowVersion: windowErr.CurrentRowVersion}
	}
	return &RowVersionConflictError{RecordID: recordID, BaseRowVersion: baseRowVersion, CurrentRowVersion: currentRowVersion}
}

type artifactSameFieldConflictParams struct {
	RouteKey          string
	RecordID          uuid.UUID
	ViewSchemaID      string
	BaseRowVersion    int64
	CurrentRowVersion int64
	RequestHash       []byte
	Window            conflicttokens.PatchConflictWindow
	Change            WorkbookPatchChange
	Changed           conflicttokens.PatchChangedField
	CurrentRow        map[string]any
	FieldDescriptors  conflicttokens.FieldDescriptorSet
	Codec             conflicttokens.ConflictTokenCodec
}

func overlappingArtifactPatchChange(changes []WorkbookPatchChange, changedFields map[string]conflicttokens.PatchChangedField) (WorkbookPatchChange, conflicttokens.PatchChangedField, bool) {
	for _, change := range changes {
		changed, ok := changedFields[change.FieldKey]
		if ok {
			return change, changed, true
		}
	}
	return WorkbookPatchChange{}, conflicttokens.PatchChangedField{}, false
}

func buildArtifactSameFieldConflict(params artifactSameFieldConflictParams) (map[string]any, error) {
	baseValue, ok := rowCellValue(params.Window.BaseRow, params.Change.FieldKey)
	if !ok {
		return nil, &conflicttokens.RevisionWindowError{RecordID: params.RecordID, BaseRowVersion: params.BaseRowVersion, CurrentRowVersion: params.CurrentRowVersion}
	}
	serverValue, ok := rowCellValue(params.CurrentRow, params.Change.FieldKey)
	if !ok {
		return nil, &conflicttokens.RevisionWindowError{RecordID: params.RecordID, BaseRowVersion: params.BaseRowVersion, CurrentRowVersion: params.CurrentRowVersion}
	}
	clientValue, err := artifactPatchClientConflictValue(params.RecordID, params.Change, baseValue, params.RequestHash)
	if err != nil {
		return nil, &conflicttokens.RevisionWindowError{RecordID: params.RecordID, BaseRowVersion: params.BaseRowVersion, CurrentRowVersion: params.CurrentRowVersion}
	}
	conflictClass := params.FieldDescriptors.ConflictResolutionClass(params.Change.FieldKey)
	if conflictClass == "" {
		conflictClass = "atomic_replace"
	}
	conflictToken, err := artifactConflictToken(params.RouteKey, params.RecordID, params.ViewSchemaID, params.Change.FieldKey, conflictClass, params.BaseRowVersion, params.CurrentRowVersion, params.RequestHash, params.Codec)
	if err != nil {
		return nil, err
	}
	conflict := map[string]any{
		"conflict_token":            conflictToken,
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
		if suggested, ok := conflicttokens.SuggestedTextMergeValue(baseValue, serverValue, clientValue); ok {
			conflict["suggested_merged_value"] = suggested
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

func artifactPatchClientConflictValue(recordID uuid.UUID, change WorkbookPatchChange, baseValue any, requestHash []byte) (any, error) {
	if change.Collection == nil {
		return change.CanonicalValue, nil
	}
	return applyCollectionConflictActions(recordID, change.FieldKey, baseValue, *change.Collection, requestHash)
}

func applyCollectionConflictActions(recordID uuid.UUID, fieldKey string, baseValue any, payload WorkbookCollectionActionPayload, requestHash []byte) (map[string]any, error) {
	ordered, items, ok := cloneCollectionConflictValue(baseValue)
	if !ok {
		return nil, fmt.Errorf("invalid base collection value for %s", fieldKey)
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
	if !ordered {
		slices.SortFunc(items, func(left map[string]any, right map[string]any) int {
			return strings.Compare(collectionSortKey(left), collectionSortKey(right))
		})
	}
	return map[string]any{"kind": "collection_value_v1", "ordered": ordered, "items": items}, nil
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

func newClientCollectionItem(recordID uuid.UUID, fieldKey string, action WorkbookCollectionAction, requestHash []byte, actionIndex int) map[string]any {
	switch action.Op {
	case "add_record_ref":
		linkedID := action.LinkedRecordID.String()
		targetType := expectedCollectionTargetType(fieldKey)
		if targetType == "" {
			targetType = "record"
		}
		return map[string]any{
			"item_ref":         links.RecordRefItemRef(*action.LinkedRecordID),
			"item_kind":        "record_ref",
			"display_text":     targetType + ":" + linkedID,
			"linked_record_id": linkedID,
		}
	case "add_party_ref":
		partyID := action.PartyID.String()
		return map[string]any{
			"item_ref":     links.PartyRefItemRef(*action.PartyID),
			"item_kind":    "party_ref",
			"display_text": "party:" + partyID,
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
			"item_ref":      riskrefs.RiskRefItemRef(riskRefID),
			"item_kind":     "risk_ref",
			"display_text":  action.RiskRefText,
			"risk_ref_id":   riskRefID.String(),
			"risk_ref_text": action.RiskRefText,
		}
	default:
		return map[string]any{}
	}
}

func conflictLocalUUID(recordID uuid.UUID, fieldKey string, action WorkbookCollectionAction, requestHash []byte, actionIndex int) uuid.UUID {
	seed, _ := json.Marshal(map[string]any{
		"record_id":     recordID.String(),
		"field_key":     fieldKey,
		"request_hash":  base64.RawURLEncoding.EncodeToString(requestHash),
		"action_index":  actionIndex,
		"op":            action.Op,
		"risk_ref_text": action.NormalizedText,
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

func collectionSortKey(item map[string]any) string {
	for _, key := range []string{"item_kind", "display_text", "item_ref"} {
		if value, ok := item[key].(string); ok && value != "" {
			return value
		}
	}
	return ""
}

func expectedCollectionTargetType(fieldKey string) string {
	policy, ok := LookupCollectionPolicy(fieldKey)
	if !ok {
		return ""
	}
	return policy.ExpectedTargetType
}

func artifactConflictToken(routeKey string, recordID uuid.UUID, viewSchemaID string, fieldKey string, conflictClass string, baseRowVersion int64, currentRowVersion int64, requestHash []byte, codec conflicttokens.ConflictTokenCodec) (string, error) {
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

func cloneMap(source map[string]any) map[string]any {
	cloned := make(map[string]any, len(source))
	for key, value := range source {
		cloned[key] = value
	}
	return cloned
}

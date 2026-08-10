package artifacts

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"slices"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/gen/contractartifacts"
	"github.com/JochiRaider/cartulary/internal/modules/artifacts/sourcecontract"
	"github.com/JochiRaider/cartulary/internal/modules/links"
	"github.com/JochiRaider/cartulary/internal/modules/records"
	"github.com/JochiRaider/cartulary/internal/modules/revisions"
	conflicttokens "github.com/JochiRaider/cartulary/internal/modules/revisions/conflicts"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
)

type MutationFacade struct {
	pool              postgres.DB
	idempotency       IdempotencyCapability
	incidentAccess    IncidentStateCapability
	memberReferences  MemberReferenceCapability
	linkStore         LinkCapability
	revisions         RevisionCapability
	source            artifactSourceKernel
	conflictTokens    conflicttokens.ConflictTokenCodec
	conflictFields    conflicttokens.FieldResolver
	conflictSnapshots conflicttokens.RevisionSnapshotProjector
	keepSaved         conflicttokens.IdempotencyPort
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
	ActorUserID uuid.UUID
	IncidentID  uuid.UUID
	Request     WorkbookCreateRequest
	RequestHash []byte
	RequestID   string
	OperationID OperationID
	Now         time.Time
}

type WorkbookPatchCommand struct {
	ActorUserID         uuid.UUID
	RecordID            uuid.UUID
	Request             WorkbookPatchRequest
	RequestHash         []byte
	RequestID           string
	OperationID         OperationID
	ConflictOperationID OperationID
	Now                 time.Time
}

type WorkbookMutationResult struct {
	Row              map[string]any
	Created          bool
	Replayed         bool
	IncidentID       uuid.UUID
	RecordID         uuid.UUID
	ChangeSetID      uuid.UUID
	ClientTxnID      string
	RowVersion       int64
	ViewSchemaID     string
	ChangedFieldKeys []string
	ContextualLink   *ContextualLinkFacts
}

type ContextualLinkFacts struct {
	SourceRecordID uuid.UUID
	LinkType       string
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

func NewMutationContribution(
	pool postgres.DB,
	conflictTokens conflicttokens.ConflictTokenCodec,
	dependencies MutationDependencies,
) (*MutationFacade, error) {
	if pool == nil {
		return nil, fmt.Errorf("artifacts mutation composition: Postgres is required")
	}
	if err := dependencies.validate(); err != nil {
		return nil, err
	}
	if _, err := buildArtifactSourcePolicyCatalog(contractartifacts.SourceCatalog); err != nil {
		return nil, fmt.Errorf("compose Artifacts source catalog: %w", err)
	}
	conflictSnapshots, err := conflicttokens.NewRevisionSnapshotProjector(
		"cartulary.revisions.snapshot.artifact.v1",
		sourcecontract.ConflictFieldSourceKeys(),
	)
	if err != nil {
		return nil, fmt.Errorf("compose Artifacts conflict snapshot projector: %w", err)
	}
	return &MutationFacade{
		pool:             pool,
		idempotency:      dependencies.Idempotency,
		incidentAccess:   dependencies.IncidentState,
		memberReferences: dependencies.MemberReferences,
		linkStore:        dependencies.Links,
		revisions:        dependencies.Revisions,
		source: artifactSourceKernel{
			records:     dependencies.RecordEnvelopes,
			rows:        newSourceStore(),
			projections: dependencies.Projections,
		},
		conflictTokens:    conflictTokens,
		conflictFields:    dependencies.ConflictFields,
		conflictSnapshots: conflictSnapshots,
		keepSaved:         dependencies.KeepSavedIdempotency,
	}, nil
}

func (f *MutationFacade) Create(ctx context.Context, command WorkbookCreateCommand) (WorkbookMutationResult, error) {
	return f.create(ctx, command, nil)
}

func (f *MutationFacade) create(ctx context.Context, command WorkbookCreateCommand, contextualSourceRecordID *uuid.UUID) (WorkbookMutationResult, error) {
	request := command.Request
	wantOperation := OperationWorkbookCreate
	wantKind := StoredMutationCreate
	if contextualSourceRecordID != nil {
		wantOperation = OperationLinkedNoteCreate
		wantKind = StoredMutationLinkedNote
	}
	if command.OperationID != wantOperation {
		return WorkbookMutationResult{}, ErrStoredMutationKindMismatch
	}
	scopeKey := command.IncidentID.String() + ":" + request.ViewSchemaID
	if contextualSourceRecordID != nil {
		scopeKey = contextualSourceRecordID.String()
	}
	idempotencyKey := IdempotencyKey{
		OperationID: command.OperationID,
		ActorUserID: command.ActorUserID,
		ScopeKey:    scopeKey,
		ClientTxnID: request.ClientTxnID,
	}
	if existing, err := f.idempotency.Get(ctx, idempotencyKey, command.RequestHash); err == nil {
		if !bytes.Equal(existing.RequestHash, command.RequestHash) {
			return WorkbookMutationResult{}, ErrClientTxnConflict
		}
		if existing.Result.Kind() != wantKind {
			return WorkbookMutationResult{}, ErrStoredMutationKindMismatch
		}
		stored, ok := existing.Result.WorkbookResult()
		if !ok || stored.ViewSchemaID != request.ViewSchemaID {
			return WorkbookMutationResult{}, ErrStoredMutationKindMismatch
		}
		result := WorkbookMutationResult{
			Row: stored.Row, Replayed: true, IncidentID: command.IncidentID,
			RecordID: stored.RecordID, ChangeSetID: stored.ChangeSetID,
			ViewSchemaID: request.ViewSchemaID, ClientTxnID: request.ClientTxnID,
			RowVersion: rowVersionFromCanonicalRow(stored.Row),
		}
		if stored.SourceRecordID != nil {
			result.ContextualLink = &ContextualLinkFacts{SourceRecordID: *stored.SourceRecordID, LinkType: stored.LinkType}
		}
		return result, nil
	} else if !errors.Is(err, ErrIdempotencyNotFound) {
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
	stored := StoredWorkbookResult{
		ViewSchemaID: request.ViewSchemaID, RecordID: mutation.recordID,
		ChangeSetID: mutation.changeSetID, Row: mutation.row,
	}
	var storedResult StoredMutationResult
	if contextualSourceRecordID == nil {
		storedResult = NewStoredCreateResult(stored)
	} else {
		stored.SourceRecordID = contextualSourceRecordID
		stored.LinkType = "references_artifact"
		storedResult = NewStoredLinkedNoteResult(stored)
	}
	if err := f.idempotency.PutTx(ctx, tx, idempotencyKey, command.RequestHash, storedResult); err != nil {
		return WorkbookMutationResult{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return WorkbookMutationResult{}, fmt.Errorf("commit artifact create transaction: %w", err)
	}
	return WorkbookMutationResult{
		Row:              mutation.row,
		Created:          true,
		IncidentID:       mutation.incidentID,
		RecordID:         mutation.recordID,
		ChangeSetID:      mutation.changeSetID,
		ClientTxnID:      request.ClientTxnID,
		RowVersion:       1,
		ViewSchemaID:     request.ViewSchemaID,
		ChangedFieldKeys: changedFieldKeys(nil, mutation.row),
		ContextualLink:   contextualLinkFacts(contextualSourceRecordID),
	}, nil
}

func (f *MutationFacade) Patch(ctx context.Context, command WorkbookPatchCommand) (WorkbookMutationResult, error) {
	request := command.Request
	if command.OperationID != OperationWorkbookPatch && command.OperationID != OperationConflictResolve {
		return WorkbookMutationResult{}, ErrStoredMutationKindMismatch
	}
	if command.ConflictOperationID != OperationConflictResolve {
		return WorkbookMutationResult{}, ErrStoredMutationKindMismatch
	}
	idempotencyKey := IdempotencyKey{
		OperationID: command.OperationID,
		ActorUserID: command.ActorUserID,
		ScopeKey:    command.RecordID.String(),
		ClientTxnID: request.ClientTxnID,
	}
	if existing, err := f.idempotency.Get(ctx, idempotencyKey, command.RequestHash); err == nil {
		if !bytes.Equal(existing.RequestHash, command.RequestHash) {
			return WorkbookMutationResult{}, ErrClientTxnConflict
		}
		if existing.Result.Kind() != StoredMutationPatch {
			return WorkbookMutationResult{}, ErrStoredMutationKindMismatch
		}
		stored, ok := existing.Result.WorkbookResult()
		if !ok || stored.RecordID != command.RecordID || stored.ViewSchemaID != request.ViewSchemaID {
			return WorkbookMutationResult{}, ErrStoredMutationKindMismatch
		}
		return WorkbookMutationResult{
			Row: stored.Row, Replayed: true, RecordID: command.RecordID,
			ChangeSetID: stored.ChangeSetID, ViewSchemaID: request.ViewSchemaID,
			ClientTxnID: request.ClientTxnID, RowVersion: rowVersionFromCanonicalRow(stored.Row),
		}, nil
	} else if !errors.Is(err, ErrIdempotencyNotFound) {
		return WorkbookMutationResult{}, fmt.Errorf("query artifact patch idempotency: %w", err)
	}

	tx, err := f.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return WorkbookMutationResult{}, fmt.Errorf("begin artifact patch transaction: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	meta, err := f.loadArtifactRecordMetaForUpdateTx(ctx, tx, command.RecordID)
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
		windowRows, err := f.revisions.LoadRevisionWindowTx(ctx, tx, command.RecordID, request.BaseRowVersion, meta.RowVersion)
		if err != nil {
			return WorkbookMutationResult{}, adaptRevisionWindowError(command.RecordID, request.BaseRowVersion, meta.RowVersion, err)
		}
		fieldDescriptors, err := f.conflictFields.ResolveViewSchema(request.ViewSchemaID)
		if err != nil {
			return WorkbookMutationResult{}, adaptRevisionWindowError(command.RecordID, request.BaseRowVersion, meta.RowVersion, err)
		}
		window, err := conflicttokens.BuildCanonicalPatchConflictWindow(command.RecordID, request.BaseRowVersion, meta.RowVersion, windowRows, fieldDescriptors, f.conflictSnapshots)
		if err != nil {
			return WorkbookMutationResult{}, adaptRevisionWindowError(command.RecordID, request.BaseRowVersion, meta.RowVersion, err)
		}
		if change, changed, ok := overlappingArtifactPatchChange(request.Changes, window.ChangedFields); ok {
			current, err := f.source.projections.LoadArtifactTx(ctx, tx, request.ViewSchemaID, command.RecordID)
			if err != nil {
				return WorkbookMutationResult{}, err
			}
			conflictPayload, err := buildArtifactSameFieldConflict(artifactSameFieldConflictParams{
				RouteKey:          string(command.ConflictOperationID),
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
	beforeRow, err := f.source.projections.LoadArtifactTx(ctx, tx, request.ViewSchemaID, command.RecordID)
	if err != nil {
		return WorkbookMutationResult{}, err
	}
	beforeSnapshot, err := f.revisions.CaptureRecordSnapshotTx(ctx, tx, command.RecordID)
	if err != nil {
		return WorkbookMutationResult{}, err
	}
	if err := validateArtifactPatchReferencesTx(ctx, tx, f.memberReferences, f.linkStore, meta.IncidentID, request); err != nil {
		return WorkbookMutationResult{}, err
	}
	changed, collectionMutations, err := f.applyPatchTx(ctx, tx, meta.IncidentID, command.RecordID, command.ActorUserID, request, command.Now.UTC())
	if err != nil {
		return WorkbookMutationResult{}, err
	}
	if !changed {
		return WorkbookMutationResult{}, &ValidationError{Field: "changes", ReasonCode: "no_effective_change"}
	}
	rowVersion, err := f.source.records.AdvanceVersionTx(ctx, tx, command.RecordID, command.ActorUserID, command.Now.UTC())
	if err != nil {
		return WorkbookMutationResult{}, err
	}
	if err := f.source.rows.TouchRowTx(ctx, tx, command.RecordID, command.Now.UTC()); err != nil {
		return WorkbookMutationResult{}, err
	}
	if err := f.source.projections.RefreshArtifactTx(ctx, tx, command.RecordID); err != nil {
		return WorkbookMutationResult{}, err
	}
	afterRow, err := f.source.projections.LoadArtifactTx(ctx, tx, request.ViewSchemaID, command.RecordID)
	if err != nil {
		return WorkbookMutationResult{}, err
	}
	afterSnapshot, err := f.revisions.CaptureRecordSnapshotTx(ctx, tx, command.RecordID)
	if err != nil {
		return WorkbookMutationResult{}, err
	}
	changeSetID, err := f.revisions.AppendChangeSetTx(ctx, tx, revisions.AppendChangeSetParams{
		IncidentID:  meta.IncidentID,
		ActorUserID: command.ActorUserID,
		Source:      string(command.OperationID),
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
	if err := f.revisions.AppendRecordMutationTx(ctx, tx, revisions.AppendRecordMutationParams{
		ChangeSetID:     changeSetID,
		SequenceNo:      1,
		TargetKind:      "record",
		RecordID:        command.RecordID,
		OperationKind:   "patch",
		BeforeVersionID: &beforeVersionID,
		AfterVersionID:  &afterVersionID,
		BeforeSnapshot:  &beforeSnapshot,
		AfterSnapshot:   &afterSnapshot,
	}); err != nil {
		return WorkbookMutationResult{}, err
	}
	if _, err := f.appendCollectionMutationsTx(ctx, tx, changeSetID, 2, collectionMutations); err != nil {
		return WorkbookMutationResult{}, err
	}
	if err := f.revisions.AppendRecordRevisionAndIntentTx(ctx, tx, revisions.AppendRecordRevisionParams{
		ChangeSetID:    changeSetID,
		RecordID:       command.RecordID,
		RowVersion:     rowVersion,
		BeforeSnapshot: &beforeSnapshot,
		AfterSnapshot:  &afterSnapshot,
		LiveChange: revisions.LiveRecordChange{
			BeforeValue: beforeRow,
			AfterValue:  afterRow,
		},
	}); err != nil {
		return WorkbookMutationResult{}, err
	}
	storedResult := NewStoredPatchResult(StoredWorkbookResult{
		ViewSchemaID: request.ViewSchemaID, RecordID: command.RecordID,
		ChangeSetID: changeSetID, Row: afterRow,
	})
	if err := f.idempotency.PutTx(ctx, tx, idempotencyKey, command.RequestHash, storedResult); err != nil {
		return WorkbookMutationResult{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return WorkbookMutationResult{}, fmt.Errorf("commit artifact patch transaction: %w", err)
	}
	return WorkbookMutationResult{
		Row:              afterRow,
		IncidentID:       meta.IncidentID,
		RecordID:         command.RecordID,
		ChangeSetID:      changeSetID,
		ClientTxnID:      request.ClientTxnID,
		RowVersion:       rowVersion,
		ViewSchemaID:     request.ViewSchemaID,
		ChangedFieldKeys: changedFieldKeys(beforeRow, afterRow),
	}, nil
}

func (f *MutationFacade) applyPatchTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID, recordID uuid.UUID, actorID uuid.UUID, request WorkbookPatchRequest, now time.Time) (bool, links.CollectionMutationResult, error) {
	changed := false
	mutations := links.CollectionMutationResult{}
	for _, change := range request.Changes {
		if change.Value != nil {
			policy, ok := lookupArtifactSourceField(change.FieldKey)
			if !ok || policy.viewSchemaID != request.ViewSchemaID ||
				policy.kind != sourceFieldDirect || !policy.writable {
				return false, links.CollectionMutationResult{}, &ValidationError{Field: change.FieldKey, ReasonCode: "unsupported_field_key"}
			}
			if err := ValidateDirectPatchChange(change.FieldKey, *change.Value); err != nil {
				return false, links.CollectionMutationResult{}, err
			}
			applied, err := f.source.rows.ApplyDirectChangeTx(ctx, tx, recordID, change.FieldKey, *change.Value, now)
			if err != nil && strings.Contains(err.Error(), "unsupported field key") {
				return false, links.CollectionMutationResult{}, &ValidationError{Field: change.FieldKey, ReasonCode: "unsupported_field_key"}
			}
			if err != nil {
				return false, links.CollectionMutationResult{}, err
			}
			changed = changed || applied
			continue
		}
		if change.Collection != nil {
			applied, collectionResult, err := f.applyCollectionTx(ctx, tx, incidentID, recordID, actorID, request.ViewSchemaID, change.FieldKey, *change.Collection, now)
			if err != nil {
				return false, links.CollectionMutationResult{}, err
			}
			changed = changed || applied
			mutations.RecordLinks = append(mutations.RecordLinks, collectionResult.RecordLinks...)
			mutations.RecordTags = append(mutations.RecordTags, collectionResult.RecordTags...)
		}
	}
	if request.ViewSchemaID == FindingsViewSchemaID && touchesArtifactField(request.Changes, "finding.state") {
		applied, err := f.source.rows.NormalizeFindingLifecycleTx(ctx, tx, recordID, now)
		if err != nil {
			return false, links.CollectionMutationResult{}, err
		}
		changed = changed || applied
	}
	return changed, mutations, nil
}

func validateArtifactReferencesTx(ctx context.Context, tx pgx.Tx, members MemberReferenceCapability, linkStore LinkCapability, incidentID uuid.UUID, viewSchemaID string, values map[string]FieldValue, collections map[string]WorkbookCollectionActionPayload) error {
	for fieldKey, value := range values {
		if value.UUID != nil && strings.HasSuffix(fieldKey, "_user_id") {
			if err := members.ValidateIncidentMemberUserTx(ctx, tx, incidentID, *value.UUID, fieldKey); err != nil {
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

func validateArtifactPatchReferencesTx(ctx context.Context, tx pgx.Tx, members MemberReferenceCapability, linkStore LinkCapability, incidentID uuid.UUID, request WorkbookPatchRequest) error {
	for _, change := range request.Changes {
		if change.Value != nil && change.Value.UUID != nil && strings.HasSuffix(change.FieldKey, "_user_id") {
			if err := members.ValidateIncidentMemberUserTx(ctx, tx, incidentID, *change.Value.UUID, change.FieldKey); err != nil {
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

func validateArtifactCollectionPayloadTx(ctx context.Context, tx pgx.Tx, linkStore LinkCapability, incidentID uuid.UUID, viewSchemaID string, fieldKey string, payload WorkbookCollectionActionPayload) error {
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

func (f *MutationFacade) applyCollectionsTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID, recordID uuid.UUID, actorID uuid.UUID, viewSchemaID string, collections map[string]WorkbookCollectionActionPayload, now time.Time) (links.CollectionMutationResult, error) {
	fieldKeys := make([]string, 0, len(collections))
	for fieldKey := range collections {
		fieldKeys = append(fieldKeys, fieldKey)
	}
	slices.Sort(fieldKeys)
	mutations := links.CollectionMutationResult{}
	for _, fieldKey := range fieldKeys {
		_, collectionResult, err := f.applyCollectionTx(ctx, tx, incidentID, recordID, actorID, viewSchemaID, fieldKey, collections[fieldKey], now)
		if err != nil {
			return links.CollectionMutationResult{}, err
		}
		mutations.RecordLinks = append(mutations.RecordLinks, collectionResult.RecordLinks...)
		mutations.RecordTags = append(mutations.RecordTags, collectionResult.RecordTags...)
	}
	return mutations, nil
}

func (f *MutationFacade) applyCollectionTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID, recordID uuid.UUID, actorID uuid.UUID, viewSchemaID string, fieldKey string, payload WorkbookCollectionActionPayload, now time.Time) (bool, links.CollectionMutationResult, error) {
	sourcePolicy, ok := lookupArtifactSourceField(fieldKey)
	if !ok || sourcePolicy.viewSchemaID != viewSchemaID ||
		sourcePolicy.kind != sourceFieldCollection || !sourcePolicy.writable {
		return false, links.CollectionMutationResult{}, collectionValidationError(fieldKey)
	}
	policy := sourcePolicy.collection
	if policy.AllowsRiskRefs() {
		changed, err := f.source.rows.ApplyHandoffRiskRefPayloadTx(ctx, tx, incidentID, recordID, actorID, riskRefPayloadFromWorkbook(payload), now)
		return changed, links.CollectionMutationResult{}, err
	}
	switch {
	case policy.AllowsRecordRefs():
		command, err := artifactRecordRefCommand(incidentID, recordID, actorID, policy, payload, now)
		if err != nil {
			return false, links.CollectionMutationResult{}, err
		}
		result, err := f.linkStore.ApplyRecordRefCollectionWithMutationValuesTx(ctx, tx, command)
		return len(result.RecordLinks) > 0, result, err
	case policy.AllowsPartyRefs():
		command, err := artifactPartyRefCommand(incidentID, recordID, actorID, policy, payload, now)
		if err != nil {
			return false, links.CollectionMutationResult{}, err
		}
		result, err := f.linkStore.ApplyPartyRefCollectionWithMutationValuesTx(ctx, tx, command)
		return len(result.RecordLinks) > 0, result, err
	case policy.AllowsTags():
		command, err := artifactTagCommand(incidentID, recordID, actorID, policy, payload, now)
		if err != nil {
			return false, links.CollectionMutationResult{}, err
		}
		result, err := f.linkStore.ApplyTagCollectionWithMutationValuesTx(ctx, tx, command)
		return len(result.RecordTags) > 0, result, err
	default:
		return false, links.CollectionMutationResult{}, collectionValidationError(fieldKey)
	}
}

func (f *MutationFacade) appendCollectionMutationsTx(ctx context.Context, tx pgx.Tx, changeSetID uuid.UUID, startSequence int, mutations links.CollectionMutationResult) (int, error) {
	sequence := startSequence
	for _, mutation := range mutations.RecordLinks {
		if err := f.revisions.AppendNonRowMutationTx(ctx, tx, revisions.AppendNonRowMutationParams{
			ChangeSetID: changeSetID, SequenceNo: sequence, TargetKind: "record_link",
			TargetID: mutation.RecordLinkID.String(), OperationKind: mutation.Operation,
			BeforeValue: mutation.BeforeValue, AfterValue: mutation.AfterValue,
		}); err != nil {
			return sequence, err
		}
		sequence++
	}
	for _, mutation := range mutations.RecordTags {
		if err := f.revisions.AppendNonRowMutationTx(ctx, tx, revisions.AppendNonRowMutationParams{
			ChangeSetID: changeSetID, SequenceNo: sequence, TargetKind: "record_tag",
			TargetID: links.RecordTagItemRef(mutation.RecordID, mutation.RecordTagID), OperationKind: mutation.Operation,
			BeforeValue: mutation.BeforeValue, AfterValue: mutation.AfterValue,
		}); err != nil {
			return sequence, err
		}
		sequence++
	}
	return sequence, nil
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

func (f *MutationFacade) loadArtifactRecordMetaForUpdateTx(
	ctx context.Context,
	tx pgx.Tx,
	recordID uuid.UUID,
) (artifactRecordMeta, error) {
	envelope, err := f.source.records.LoadEnvelopeTx(ctx, tx, recordID, true)
	if errors.Is(err, records.ErrEnvelopeNotFound) {
		return artifactRecordMeta{}, pgx.ErrNoRows
	}
	if err != nil {
		return artifactRecordMeta{}, err
	}
	if envelope.DeletedAt != nil {
		return artifactRecordMeta{}, revisions.ErrRecordDeletedUseRestore
	}
	return artifactRecordMeta{
		IncidentID: envelope.IncidentID,
		RecordType: envelope.RecordType,
		RowVersion: envelope.RowVersion,
	}, nil
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

func rowVersionFromCanonicalRow(row map[string]any) int64 {
	switch value := row["row_version"].(type) {
	case int64:
		return value
	case int:
		return int64(value)
	case float64:
		return int64(value)
	default:
		return 0
	}
}

func contextualLinkFacts(sourceRecordID *uuid.UUID) *ContextualLinkFacts {
	if sourceRecordID == nil {
		return nil
	}
	return &ContextualLinkFacts{SourceRecordID: *sourceRecordID, LinkType: "references_artifact"}
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
			"item_ref":      RiskRefItemRef(riskRefID),
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

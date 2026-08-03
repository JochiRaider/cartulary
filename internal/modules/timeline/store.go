package timeline

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"

	conflicttokens "github.com/JochiRaider/cartulary/internal/modules/revisions/conflicts"
	"github.com/JochiRaider/cartulary/internal/modules/timeline/sourcerepository"
	"github.com/JochiRaider/cartulary/internal/modules/timeline/workbookprojection"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
	"github.com/JochiRaider/cartulary/internal/platform/viewschema"
)

var (
	ErrRecordNotFound         = errors.New("timeline: record not found")
	ErrRowVersionConflict     = errors.New("timeline: row version conflict")
	ErrIllegalTransition      = errors.New("timeline: illegal transition")
	ErrNoEffectiveChange      = errors.New("timeline: no effective change")
	ErrIncidentClosed         = errors.New("timeline: incident closed")
	ErrRecordDeleted          = errors.New("timeline: record deleted use restore")
	ErrResolvedRecordNotFound = errors.New("timeline: resolved record not found")
)

type RowVersionConflictError struct {
	RecordID          uuid.UUID
	BaseRowVersion    int64
	CurrentRowVersion int64
}

func (e *RowVersionConflictError) Error() string {
	return ErrRowVersionConflict.Error()
}

func (e *RowVersionConflictError) Unwrap() error {
	return ErrRowVersionConflict
}

func (e *RowVersionConflictError) Details() map[string]any {
	if e == nil {
		return map[string]any{}
	}
	return map[string]any{
		"record_id":           e.RecordID.String(),
		"base_row_version":    e.BaseRowVersion,
		"current_row_version": e.CurrentRowVersion,
	}
}

type SameFieldConflictError struct {
	Conflict map[string]any
}

func (e *SameFieldConflictError) Error() string {
	return "timeline: same field conflict"
}

type patchConflictWindow struct {
	BaseRow       map[string]any
	ChangedFields map[string]patchChangedField
}

type patchChangedField struct {
	FieldKey        string
	ServerUpdatedBy uuid.UUID
	ServerUpdatedAt time.Time
}

type store struct {
	pool             postgres.DB
	idempotencyStore IdempotencyPort
	incidentAccess   IncidentPort
	recordStore      RecordPort
	revisionsStore   RevisionPort
	projectionStore  ProjectionPort
	linkStore        LinkPort
	mentionStore     MentionPort
	entityStore      EntityPort
	evidenceStore    EvidencePort
	collectionReader CollectionFactPort
	collaboration    CollaborationPort
	sourceRepository *sourcerepository.Repository
	conflictTokens   conflicttokens.ConflictTokenCodec
}

type attachedEvidenceMutation struct {
	RecordLinkID uuid.UUID
	Operation    string
	BeforeValue  map[string]any
	AfterValue   map[string]any
}

type recordTagMutation struct {
	RecordTagID uuid.UUID
	RecordID    uuid.UUID
	Operation   string
	BeforeValue map[string]any
	AfterValue  map[string]any
}

type TimeConversionProfile struct {
	IncidentID         uuid.UUID
	Enabled            bool
	LocalOffsetMinutes *int
	LocalLabel         *string
	ProfileVersion     int64
	UpdatedAt          time.Time
	UpdatedByUserID    *uuid.UUID
}

type createRowOptions struct {
	allowInteractiveAutoResolution bool
}

func newStore(pool postgres.DB, collaborators Collaborators, conflictTokens conflicttokens.ConflictTokenCodec) *store {
	return &store{
		pool:             pool,
		idempotencyStore: collaborators.Core.Idempotency,
		incidentAccess:   collaborators.Core.Incidents,
		recordStore:      collaborators.Core.Records,
		revisionsStore:   collaborators.Core.Revisions,
		projectionStore:  collaborators.Commit.Projection,
		linkStore:        collaborators.Collections.Links,
		mentionStore:     collaborators.Collections.Mentions,
		entityStore:      collaborators.Collections.Entities,
		evidenceStore:    collaborators.Collections.Evidence,
		collectionReader: collaborators.Collections.Facts,
		collaboration:    collaborators.Commit.Collaboration,
		sourceRepository: sourcerepository.New(collaborators.Core.Records),
		conflictTokens:   conflictTokens,
	}
}

func (s *store) GetRecordIncident(ctx context.Context, recordID uuid.UUID) (uuid.UUID, error) {
	incidentID, err := s.recordStore.ResolveIncident(ctx, recordID)
	if err != nil {
		return uuid.UUID{}, ErrRecordNotFound
	}
	return incidentID, nil
}

func (s *store) CreateRow(ctx context.Context, actor authn.UserRecord, incidentID uuid.UUID, request CreateRequest, requestHash []byte, requestID string, now time.Time) (MutationResult, error) {
	return s.createRow(ctx, actor, incidentID, request, requestHash, requestID, now, createRowOptions{
		allowInteractiveAutoResolution: true,
	})
}

func (s *store) CreateImportedRow(ctx context.Context, actor authn.UserRecord, incidentID uuid.UUID, request CreateRequest, requestHash []byte, requestID string, now time.Time) (MutationResult, error) {
	return s.createRow(ctx, actor, incidentID, request, requestHash, requestID, now, createRowOptions{})
}

func (s *store) createRow(ctx context.Context, actor authn.UserRecord, incidentID uuid.UUID, request CreateRequest, requestHash []byte, requestID string, now time.Time, options createRowOptions) (MutationResult, error) {
	scopeKey := incidentID.String() + ":" + TimelineViewSchemaID
	idempotencyKey := authn.RouteIdempotencyKey{
		RouteKey:    createRouteKey,
		ActorUserID: actor.ID,
		ScopeKey:    scopeKey,
		ClientTxnID: request.ClientTxnID,
	}
	if existing, err := s.idempotencyStore.GetRouteIdempotency(ctx, idempotencyKey); err == nil {
		if !hashesEqual(existing.RequestHash, requestHash) {
			return MutationResult{}, authn.ErrClientTxnConflict
		}
		payload, err := decodeStoredResponse(existing.ResponseJSON)
		if err != nil {
			return MutationResult{}, fmt.Errorf("decode replayed timeline create payload: %w", err)
		}
		recordID, err := extractUUIDFromPayload(payload, "row", "record_id")
		if err != nil {
			return MutationResult{}, err
		}
		return MutationResult{
			Payload:    payload,
			StatusCode: http.StatusOK,
			Replayed:   true,
			IncidentID: incidentID,
			RecordID:   recordID,
		}, nil
	} else if !errors.Is(err, authn.ErrNotFound) {
		return MutationResult{}, fmt.Errorf("query timeline create idempotency: %w", err)
	}
	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return MutationResult{}, fmt.Errorf("begin timeline create transaction: %w", err)
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()
	if err := s.incidentAccess.EnsureOpenTx(ctx, tx, incidentID); err != nil {
		return MutationResult{}, err
	}
	changeSetID := uuid.New()
	if _, err := s.revisionsStore.AppendChangeSetTx(ctx, tx, ChangeSetParams{
		ChangeSetID: &changeSetID,
		IncidentID:  incidentID,
		ActorUserID: actor.ID,
		Source:      createRouteKey,
		ClientTxnID: &request.ClientTxnID,
		RequestID:   &requestID,
		CreatedAt:   now.UTC(),
	}); err != nil {
		return MutationResult{}, err
	}
	result, _, err := s.createRowTx(
		ctx,
		tx,
		actor.ID,
		incidentID,
		request,
		changeSetID,
		func(int) (int, error) { return 1, nil },
		now.UTC(),
		options,
	)
	if err != nil {
		return MutationResult{}, err
	}
	if err := s.idempotencyStore.InsertRouteIdempotencyPayload(ctx, tx, idempotencyKey, nil, requestHash, http.StatusCreated, result.Payload); err != nil {
		if authn.IsUniqueViolation(err) {
			return MutationResult{}, authn.ErrClientTxnConflict
		}
		return MutationResult{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return MutationResult{}, fmt.Errorf("commit timeline create transaction: %w", err)
	}
	return result, nil
}

func (s *store) createRowTx(
	ctx context.Context,
	tx pgx.Tx,
	actorUserID uuid.UUID,
	incidentID uuid.UUID,
	request CreateRequest,
	changeSetID uuid.UUID,
	allocateMutationSequence func(int) (int, error),
	now time.Time,
	options createRowOptions,
) (MutationResult, int, error) {
	recordID := uuid.New()
	current := sourcerepository.Snapshot{
		RecordID:              recordID,
		IncidentID:            incidentID,
		DateEnteredText:       request.DateEnteredText,
		AnalystText:           request.AnalystText,
		MitreStageText:        request.MitreStageText,
		DeviceObjectText:      request.DeviceObjectText,
		IPAddressText:         request.IPAddressText,
		ActivityUTCText:       request.ActivityUTCText,
		ActivityLocalText:     request.ActivityLocalText,
		RawActivityText:       request.RawActivityText,
		ActivitySynopsisText:  request.ActivitySynopsisText,
		DataSourceText:        request.DataSourceText,
		ActivityTimePairState: "disabled",
		CaptureState:          InitialCaptureState(),
		RowVersion:            1,
		RecordedAt:            now.UTC(),
		EditedAt:              now.UTC(),
		CreatedByUserID:       actorUserID,
		UpdatedByUserID:       actorUserID,
	}
	profile, err := getTimeConversionProfileTx(ctx, tx, incidentID, now.UTC())
	if err != nil {
		return MutationResult{}, 0, err
	}
	applyTimelineTimeConversion(&current, profile)
	if _, err := s.recordStore.InsertTx(ctx, tx, RecordCreateParams{
		RecordID:        &current.RecordID,
		IncidentID:      incidentID,
		RecordType:      "timeline_event",
		CreatedByUserID: actorUserID,
		CreatedAt:       now.UTC(),
		UpdatedByUserID: actorUserID,
		UpdatedAt:       now.UTC(),
		RowVersion:      1,
	}); err != nil {
		return MutationResult{}, 0, fmt.Errorf("insert timeline record envelope: %w", err)
	}
	if _, err := tx.Exec(ctx, `
INSERT INTO timeline_events (
    record_id, incident_id, date_entered_text, analyst_text, mitre_stage_text,
    device_object_text, ip_address_text, activity_utc_text, activity_local_text,
    raw_activity_text, activity_synopsis_text, data_source_text,
    activity_utc_generated, activity_local_generated, activity_time_pair_state,
    capture_state, row_version, recorded_at, edited_at,
    created_by_user_id, updated_by_user_id
)
VALUES ($1, $2, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, 'rough', 1, $4, $4, $3, $3)
`, current.RecordID, incidentID, actorUserID, now.UTC(), current.DateEnteredText, current.AnalystText, current.MitreStageText, current.DeviceObjectText, current.IPAddressText, current.ActivityUTCText, current.ActivityLocalText, current.RawActivityText, current.ActivitySynopsisText, current.DataSourceText, current.ActivityUTCGenerated, current.ActivityLocalGenerated, current.ActivityTimePairState); err != nil {
		return MutationResult{}, 0, fmt.Errorf("insert timeline source row: %w", err)
	}
	if err := insertSourceProvenanceTx(ctx, tx, current.RecordID, request.RawCaptureColumns); err != nil {
		return MutationResult{}, 0, err
	}

	mentionProjectionRefresh, err := s.applyCreateMentionActionsTx(ctx, tx, actorUserID, current.IncidentID, current.RecordID, request.HostRefs, request.IdentityRefs, options, now.UTC())
	if err != nil {
		return MutationResult{}, 0, err
	}
	if err := s.refreshMentionEntityProjectionsTx(ctx, tx, mentionProjectionRefresh); err != nil {
		return MutationResult{}, 0, err
	}
	tagMutations, err := s.applyCreateTagActionsTx(ctx, tx, actorUserID, current.IncidentID, current.RecordID, request.Tags, now.UTC())
	if err != nil {
		return MutationResult{}, 0, err
	}
	attachedEvidenceMutations, err := s.applyAttachedEvidenceActionsTx(ctx, tx, actorUserID, current.IncidentID, current.RecordID, request.AttachedEvidence, now.UTC())
	if err != nil {
		return MutationResult{}, 0, err
	}

	projected := projectRecord(current, nil)
	if createRequestHasCollectionActions(request) {
		if err := s.hydrateProjectedCollections(ctx, tx, &projected); err != nil {
			return MutationResult{}, 0, err
		}
	}
	mutationSequence, err := allocateMutationSequence(
		1 + len(attachedEvidenceMutations) + len(tagMutations),
	)
	if err != nil {
		return MutationResult{}, 0, err
	}

	afterRow := buildRow(projected)
	afterVersion := versionID(current.RecordID, projected.RowVersion)
	if err := s.revisionsStore.AppendMutationTx(ctx, tx, MutationParams{
		ChangeSetID:    changeSetID,
		SequenceNo:     mutationSequence,
		TargetKind:     "timeline_record",
		TargetID:       current.RecordID.String(),
		OperationKind:  "create",
		AfterVersionID: &afterVersion,
		AfterValue:     afterRow,
	}); err != nil {
		return MutationResult{}, 0, err
	}
	if err := s.insertAttachedEvidenceMutationEntriesTx(ctx, tx, changeSetID, mutationSequence+1, attachedEvidenceMutations); err != nil {
		return MutationResult{}, 0, err
	}
	if err := s.insertRecordTagMutationEntriesTx(ctx, tx, changeSetID, mutationSequence+1+len(attachedEvidenceMutations), tagMutations); err != nil {
		return MutationResult{}, 0, err
	}
	if err := s.revisionsStore.AppendRecordRevisionTx(ctx, tx, RecordRevisionParams{
		ChangeSetID: changeSetID,
		RecordID:    current.RecordID,
		RowVersion:  projected.RowVersion,
		AfterValue:  afterRow,
	}); err != nil {
		return MutationResult{}, 0, err
	}
	if err := s.upsertProjectionTx(ctx, tx, projected); err != nil {
		return MutationResult{}, 0, err
	}

	payload := BuildMutationPayload(projected, changeSetID)
	if err := s.appendRecordChangeIntentTx(
		ctx,
		tx,
		incidentID,
		current.RecordID,
		projected.RowVersion,
		changeSetID,
		request.ClientTxnID,
		actorUserID,
		ComputeChangedFieldKeys(nil, projected),
		afterRow,
		0,
		now,
	); err != nil {
		return MutationResult{}, 0, err
	}
	return MutationResult{
		Payload:          payload,
		StatusCode:       http.StatusCreated,
		IncidentID:       incidentID,
		RecordID:         current.RecordID,
		ChangeSetID:      changeSetID,
		ClientTxnID:      request.ClientTxnID,
		RowVersion:       projected.RowVersion,
		ChangedFieldKeys: ComputeChangedFieldKeys(nil, projected),
		Row:              projected,
	}, mutationSequence, nil
}

func createRequestHasCollectionActions(request CreateRequest) bool {
	return (request.HostRefs != nil && len(request.HostRefs.Actions) > 0) ||
		(request.IdentityRefs != nil && len(request.IdentityRefs.Actions) > 0) ||
		(request.Tags != nil && len(request.Tags.Actions) > 0) ||
		(request.AttachedEvidence != nil && len(request.AttachedEvidence.Actions) > 0)
}

func (s *store) PatchRow(ctx context.Context, actor authn.UserRecord, recordID uuid.UUID, request PatchRequest, requestHash []byte, requestID string, now time.Time) (MutationResult, error) {
	return s.applyPatch(ctx, actor, recordID, request, requestHash, requestID, now, patchRouteKey)
}

func (s *store) ResolveConflict(ctx context.Context, actor authn.UserRecord, recordID uuid.UUID, claims conflicttokens.ConflictTokenClaims, request ConflictResolveRequest, requestHash []byte, requestID string, now time.Time) (MutationResult, error) {
	if request.ResolutionKind == "keep_saved" {
		return s.clearConflict(ctx, actor, recordID, claims, request, requestHash)
	}
	if request.ResolvedChange == nil {
		return MutationResult{}, ErrNoEffectiveChange
	}
	patch := PatchRequest{
		ViewSchemaID:    TimelineViewSchemaID,
		BaseRowVersion:  claims.CurrentRowVersion,
		ClientTxnID:     request.ClientTxnID,
		CanonicalChange: []PatchChange{*request.ResolvedChange},
	}
	return s.applyPatch(ctx, actor, recordID, patch, requestHash, requestID, now, conflictResolveRouteKey)
}

func (s *store) clearConflict(ctx context.Context, actor authn.UserRecord, recordID uuid.UUID, claims conflicttokens.ConflictTokenClaims, request ConflictResolveRequest, requestHash []byte) (MutationResult, error) {
	if claims.ViewSchemaID != TimelineViewSchemaID {
		return MutationResult{}, ErrRecordNotFound
	}
	idempotencyKey := authn.RouteIdempotencyKey{
		RouteKey:    conflictResolveRouteKey,
		ActorUserID: actor.ID,
		ScopeKey:    recordID.String(),
		ClientTxnID: request.ClientTxnID,
	}
	if existing, err := s.idempotencyStore.GetRouteIdempotency(ctx, idempotencyKey); err == nil {
		if !hashesEqual(existing.RequestHash, requestHash) {
			return MutationResult{}, authn.ErrClientTxnConflict
		}
		payload, err := decodeStoredResponse(existing.ResponseJSON)
		if err != nil {
			return MutationResult{}, fmt.Errorf("decode replayed timeline conflict clear payload: %w", err)
		}
		return MutationResult{
			Payload:    payload,
			StatusCode: http.StatusOK,
			Replayed:   true,
			RecordID:   recordID,
		}, nil
	} else if !errors.Is(err, authn.ErrNotFound) {
		return MutationResult{}, fmt.Errorf("query timeline conflict clear idempotency: %w", err)
	}

	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return MutationResult{}, fmt.Errorf("begin timeline conflict clear transaction: %w", err)
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	current, err := s.loadSourceRecordTx(ctx, tx, recordID)
	if err != nil {
		return MutationResult{}, err
	}
	projected := projectRecord(current, nil)
	if err := s.hydrateProjectedCollections(ctx, tx, &projected); err != nil {
		return MutationResult{}, err
	}
	payload := map[string]any{
		"view_schema_id": TimelineViewSchemaID,
		"row":            buildRow(projected),
	}
	if err := s.idempotencyStore.InsertRouteIdempotencyPayload(ctx, tx, idempotencyKey, nil, requestHash, http.StatusOK, payload); err != nil {
		if authn.IsUniqueViolation(err) {
			return MutationResult{}, authn.ErrClientTxnConflict
		}
		return MutationResult{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return MutationResult{}, fmt.Errorf("commit timeline conflict clear transaction: %w", err)
	}
	return MutationResult{
		Payload:     payload,
		StatusCode:  http.StatusOK,
		IncidentID:  current.IncidentID,
		RecordID:    recordID,
		ClientTxnID: request.ClientTxnID,
		RowVersion:  current.RowVersion,
		Row:         projected,
	}, nil
}

func (s *store) applyPatch(ctx context.Context, actor authn.UserRecord, recordID uuid.UUID, request PatchRequest, requestHash []byte, requestID string, now time.Time, routeKey string) (MutationResult, error) {
	idempotencyKey := authn.RouteIdempotencyKey{
		RouteKey:    routeKey,
		ActorUserID: actor.ID,
		ScopeKey:    recordID.String(),
		ClientTxnID: request.ClientTxnID,
	}
	if existing, err := s.idempotencyStore.GetRouteIdempotency(ctx, idempotencyKey); err == nil {
		if !hashesEqual(existing.RequestHash, requestHash) {
			return MutationResult{}, authn.ErrClientTxnConflict
		}
		payload, err := decodeStoredResponse(existing.ResponseJSON)
		if err != nil {
			return MutationResult{}, fmt.Errorf("decode replayed timeline patch payload: %w", err)
		}
		return MutationResult{
			Payload:    payload,
			StatusCode: http.StatusOK,
			Replayed:   true,
			RecordID:   recordID,
		}, nil
	} else if !errors.Is(err, authn.ErrNotFound) {
		return MutationResult{}, fmt.Errorf("query timeline patch idempotency: %w", err)
	}

	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return MutationResult{}, fmt.Errorf("begin timeline patch transaction: %w", err)
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	current, err := s.loadSourceRecordTx(ctx, tx, recordID)
	if err != nil {
		return MutationResult{}, err
	}
	if err := s.incidentAccess.EnsureOpenTx(ctx, tx, current.IncidentID); err != nil {
		return MutationResult{}, err
	}
	if current.RowVersion < request.BaseRowVersion {
		return MutationResult{}, &RowVersionConflictError{
			RecordID:          recordID,
			BaseRowVersion:    request.BaseRowVersion,
			CurrentRowVersion: current.RowVersion,
		}
	}
	if current.RowVersion > request.BaseRowVersion {
		window, err := s.loadPatchConflictWindowTx(ctx, tx, recordID, request.BaseRowVersion, current.RowVersion)
		if err != nil {
			return MutationResult{}, err
		}
		if change, changed, ok := overlappingPatchChange(request.CanonicalChange, window.ChangedFields); ok {
			currentProjected := projectRecord(current, nil)
			if err := s.hydrateProjectedCollections(ctx, tx, &currentProjected); err != nil {
				return MutationResult{}, err
			}
			conflict, err := s.buildSameFieldConflict(recordID, currentProjected, request.BaseRowVersion, requestHash, window, change, changed)
			if err != nil {
				return MutationResult{}, err
			}
			return MutationResult{}, conflict
		}
	}
	if current.CaptureState == "superseded" {
		return MutationResult{}, newIllegalTransitionError("superseded_terminal", current.CaptureState, captureStateEnriched)
	}

	next := current
	mentionChanged := false
	tagChanged := false
	evidenceChanged := false
	for _, change := range request.CanonicalChange {
		switch {
		case change.FieldKey == "timeline.date_entered_text":
			next.DateEnteredText = change.TextValue
		case change.FieldKey == "timeline.analyst_text":
			next.AnalystText = change.TextValue
		case change.FieldKey == "timeline.mitre_stage_text":
			next.MitreStageText = change.TextValue
		case change.FieldKey == "timeline.device_object_text":
			next.DeviceObjectText = change.TextValue
		case change.FieldKey == "timeline.ip_address_text":
			next.IPAddressText = change.TextValue
		case change.FieldKey == "timeline.activity_utc_text":
			next.ActivityUTCText = change.TextValue
			next.ActivityUTCGenerated = false
		case change.FieldKey == "timeline.activity_local_text":
			next.ActivityLocalText = change.TextValue
			next.ActivityLocalGenerated = false
		case change.FieldKey == "timeline.raw_activity_text":
			next.RawActivityText = change.TextValue
		case change.FieldKey == "timeline.activity_synopsis_text":
			next.ActivitySynopsisText = change.TextValue
		case change.FieldKey == "timeline.data_source_text":
			next.DataSourceText = change.TextValue
		case isTimelineMentionCollection(change.FieldKey):
			if change.ActionPayload != nil && len(change.ActionPayload.Actions) > 0 {
				mentionChanged = true
			}
		case isTimelineTagCollection(change.FieldKey):
			if change.ActionPayload != nil && len(change.ActionPayload.Actions) > 0 {
				tagChanged = true
			}
		case isTimelineAttachedEvidenceCollection(change.FieldKey):
			if change.ActionPayload != nil && len(change.ActionPayload.Actions) > 0 {
				evidenceChanged = true
			}
		}
	}
	profile, err := getTimeConversionProfileTx(ctx, tx, current.IncidentID, now.UTC())
	if err != nil {
		return MutationResult{}, err
	}
	applyTimelineTimeConversion(&next, profile)

	beforeProjected := projectRecord(current, nil)
	if err := s.hydrateProjectedCollections(ctx, tx, &beforeProjected); err != nil {
		return MutationResult{}, err
	}
	materialChanged := hasMaterialChange(current, next)
	if mentionChanged {
		mentionProjectionRefresh, err := s.applyPatchMentionActionsTx(ctx, tx, actor, current.IncidentID, recordID, request.CanonicalChange, now.UTC())
		if err != nil {
			return MutationResult{}, err
		}
		if err := s.refreshMentionEntityProjectionsTx(ctx, tx, mentionProjectionRefresh); err != nil {
			return MutationResult{}, err
		}
	}
	var tagMutations []recordTagMutation
	if tagChanged {
		var err error
		tagMutations, err = s.applyPatchTagActionsTx(ctx, tx, actor.ID, current.IncidentID, recordID, request.CanonicalChange, now.UTC())
		if err != nil {
			return MutationResult{}, err
		}
		tagChanged = len(tagMutations) > 0
	}
	var attachedEvidenceMutations []attachedEvidenceMutation
	if evidenceChanged {
		var err error
		attachedEvidenceMutations, err = s.applyPatchAttachedEvidenceActionsTx(ctx, tx, actor.ID, current.IncidentID, recordID, request.CanonicalChange, now.UTC())
		if err != nil {
			return MutationResult{}, err
		}
	}
	if !materialChanged && !mentionChanged && !tagChanged && !evidenceChanged {
		return MutationResult{}, ErrNoEffectiveChange
	}
	stateMaterialChanged := materialChanged || mentionChanged || evidenceChanged
	if stateMaterialChanged {
		nextState, err := CaptureStateAfterMaterialPatch(current.CaptureState)
		if err != nil {
			return MutationResult{}, err
		}
		next.CaptureState = nextState
	} else {
		next.CaptureState = current.CaptureState
	}
	next.RowVersion, err = s.recordStore.AdvanceVersionTx(ctx, tx, current.RecordID, actor.ID, now.UTC())
	if err != nil {
		return MutationResult{}, err
	}
	next.EditedAt = now.UTC()
	next.UpdatedByUserID = actor.ID
	if stateMaterialChanged && current.CaptureState == captureStateReviewed {
		next.ReviewedAt = nil
		next.ReviewedByUserID = nil
	}

	if err := tx.QueryRow(ctx, `
UPDATE timeline_events
   SET date_entered_text = $2,
       analyst_text = $3,
       mitre_stage_text = $4,
       device_object_text = $5,
       ip_address_text = $6,
       activity_utc_text = $7,
       activity_local_text = $8,
       raw_activity_text = $9,
       activity_synopsis_text = $10,
       data_source_text = $11,
       activity_utc_generated = $12,
       activity_local_generated = $13,
       activity_time_pair_state = $14,
       capture_state = $15,
       row_version = $16,
       edited_at = $17,
       updated_by_user_id = $18,
       reviewed_at = $19,
       reviewed_by_user_id = $20
 WHERE record_id = $1
RETURNING recorded_at
`, recordID, next.DateEnteredText, next.AnalystText, next.MitreStageText, next.DeviceObjectText, next.IPAddressText, next.ActivityUTCText, next.ActivityLocalText, next.RawActivityText, next.ActivitySynopsisText, next.DataSourceText, next.ActivityUTCGenerated, next.ActivityLocalGenerated, next.ActivityTimePairState, next.CaptureState, next.RowVersion, next.EditedAt, actor.ID, next.ReviewedAt, next.ReviewedByUserID).Scan(&next.RecordedAt); err != nil {
		return MutationResult{}, fmt.Errorf("update timeline record: %w", err)
	}

	afterProjected := projectRecord(next, nil)
	if err := s.hydrateProjectedCollections(ctx, tx, &afterProjected); err != nil {
		return MutationResult{}, err
	}
	changeSetID, err := s.revisionsStore.AppendChangeSetTx(ctx, tx, ChangeSetParams{
		IncidentID:  current.IncidentID,
		ActorUserID: actor.ID,
		Source:      routeKey,
		ClientTxnID: &request.ClientTxnID,
		RequestID:   &requestID,
		CreatedAt:   now.UTC(),
	})
	if err != nil {
		return MutationResult{}, err
	}

	beforeRow := buildRow(beforeProjected)
	afterRow := buildRow(afterProjected)
	beforeVersion := versionID(current.RecordID, current.RowVersion)
	afterVersion := versionID(next.RecordID, next.RowVersion)
	if err := s.revisionsStore.AppendMutationTx(ctx, tx, MutationParams{
		ChangeSetID:     changeSetID,
		SequenceNo:      1,
		TargetKind:      "timeline_record",
		TargetID:        current.RecordID.String(),
		OperationKind:   "patch",
		BeforeVersionID: &beforeVersion,
		AfterVersionID:  &afterVersion,
		BeforeValue:     beforeRow,
		AfterValue:      afterRow,
	}); err != nil {
		return MutationResult{}, err
	}
	if err := s.insertAttachedEvidenceMutationEntriesTx(ctx, tx, changeSetID, 2, attachedEvidenceMutations); err != nil {
		return MutationResult{}, err
	}
	if err := s.insertRecordTagMutationEntriesTx(ctx, tx, changeSetID, 2+len(attachedEvidenceMutations), tagMutations); err != nil {
		return MutationResult{}, err
	}
	if err := s.revisionsStore.AppendRecordRevisionTx(ctx, tx, RecordRevisionParams{
		ChangeSetID: changeSetID,
		RecordID:    current.RecordID,
		RowVersion:  next.RowVersion,
		BeforeValue: beforeRow,
		AfterValue:  afterRow,
	}); err != nil {
		return MutationResult{}, err
	}
	if err := s.upsertProjectionTx(ctx, tx, afterProjected); err != nil {
		return MutationResult{}, err
	}

	payload := BuildMutationPayload(afterProjected, changeSetID)
	if err := s.appendRecordChangeIntentTx(
		ctx,
		tx,
		current.IncidentID,
		current.RecordID,
		afterProjected.RowVersion,
		changeSetID,
		request.ClientTxnID,
		actor.ID,
		ComputeChangedFieldKeys(&beforeProjected, afterProjected),
		afterRow,
		0,
		now,
	); err != nil {
		return MutationResult{}, err
	}
	if err := s.idempotencyStore.InsertRouteIdempotencyPayload(ctx, tx, idempotencyKey, nil, requestHash, http.StatusOK, payload); err != nil {
		if authn.IsUniqueViolation(err) {
			return MutationResult{}, authn.ErrClientTxnConflict
		}
		return MutationResult{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return MutationResult{}, fmt.Errorf("commit timeline patch transaction: %w", err)
	}
	return MutationResult{
		Payload:          payload,
		StatusCode:       http.StatusOK,
		IncidentID:       current.IncidentID,
		RecordID:         recordID,
		ChangeSetID:      changeSetID,
		ClientTxnID:      request.ClientTxnID,
		RowVersion:       afterProjected.RowVersion,
		ChangedFieldKeys: ComputeChangedFieldKeys(&beforeProjected, afterProjected),
		Row:              afterProjected,
	}, nil
}

func (s *store) loadPatchConflictWindowTx(ctx context.Context, tx pgx.Tx, recordID uuid.UUID, baseRowVersion int64, currentRowVersion int64) (patchConflictWindow, error) {
	rows, err := s.revisionsStore.ListRecordRevisionWindowTx(ctx, tx, recordID, baseRowVersion, currentRowVersion)
	if err != nil {
		return patchConflictWindow{}, fmt.Errorf("query timeline patch conflict window: %w", err)
	}

	window := patchConflictWindow{
		ChangedFields: make(map[string]patchChangedField),
	}
	for _, entry := range rows {
		if entry.RowVersion == baseRowVersion {
			baseRow, ok := decodeRevisionRow(entry.AfterJSON)
			if !ok {
				return patchConflictWindow{}, newRowVersionConflict(recordID, baseRowVersion, currentRowVersion)
			}
			window.BaseRow = baseRow
			continue
		}

		beforeRow, beforeOK := decodeRevisionRow(entry.BeforeJSON)
		afterRow, afterOK := decodeRevisionRow(entry.AfterJSON)
		if !beforeOK || !afterOK {
			return patchConflictWindow{}, newRowVersionConflict(recordID, baseRowVersion, currentRowVersion)
		}
		for _, fieldKey := range changedRevisionWritableFieldKeys(beforeRow, afterRow) {
			window.ChangedFields[fieldKey] = patchChangedField{
				FieldKey:        fieldKey,
				ServerUpdatedBy: entry.ActorUserID,
				ServerUpdatedAt: entry.CreatedAt.UTC(),
			}
		}
	}
	if window.BaseRow == nil {
		return patchConflictWindow{}, newRowVersionConflict(recordID, baseRowVersion, currentRowVersion)
	}
	return window, nil
}

func newRowVersionConflict(recordID uuid.UUID, baseRowVersion int64, currentRowVersion int64) *RowVersionConflictError {
	return &RowVersionConflictError{
		RecordID:          recordID,
		BaseRowVersion:    baseRowVersion,
		CurrentRowVersion: currentRowVersion,
	}
}

func decodeRevisionRow(data []byte) (map[string]any, bool) {
	if len(data) == 0 {
		return nil, false
	}
	var row map[string]any
	if err := json.Unmarshal(data, &row); err != nil {
		return nil, false
	}
	if _, ok := row["cells"].(map[string]any); !ok {
		return nil, false
	}
	return row, true
}

func changedRevisionWritableFieldKeys(beforeRow map[string]any, afterRow map[string]any) []string {
	beforeCells, _ := beforeRow["cells"].(map[string]any)
	afterCells, _ := afterRow["cells"].(map[string]any)
	changed := make([]string, 0)
	for fieldKey, afterCell := range afterCells {
		field, ok := viewschema.LookupField(TimelineViewSchemaID, fieldKey)
		if !ok || !field.Writable {
			continue
		}
		if !reflect.DeepEqual(beforeCells[fieldKey], afterCell) {
			changed = append(changed, fieldKey)
		}
	}
	sort.Strings(changed)
	return changed
}

func overlappingPatchChange(changes []PatchChange, changedFields map[string]patchChangedField) (PatchChange, patchChangedField, bool) {
	for _, change := range changes {
		changed, ok := changedFields[change.FieldKey]
		if ok {
			return change, changed, true
		}
	}
	return PatchChange{}, patchChangedField{}, false
}

func (s *store) buildSameFieldConflict(recordID uuid.UUID, current workbookprojection.DerivedRecord, baseRowVersion int64, requestHash []byte, window patchConflictWindow, change PatchChange, changed patchChangedField) (*SameFieldConflictError, error) {
	baseValue, ok := rowCellValue(window.BaseRow, change.FieldKey)
	if !ok {
		return nil, newRowVersionConflict(recordID, baseRowVersion, current.RowVersion)
	}
	serverValue, ok := rowCellValue(buildRow(current), change.FieldKey)
	if !ok {
		return nil, newRowVersionConflict(recordID, baseRowVersion, current.RowVersion)
	}
	clientValue, err := patchClientConflictValue(recordID, change, baseValue, requestHash)
	if err != nil {
		return nil, newRowVersionConflict(recordID, baseRowVersion, current.RowVersion)
	}

	field, _ := viewschema.LookupField(TimelineViewSchemaID, change.FieldKey)
	conflictClass := field.ConflictResolutionClass
	if conflictClass == "" {
		conflictClass = "atomic_replace"
	}
	token, err := s.conflictToken(recordID, change.FieldKey, baseRowVersion, current.RowVersion, requestHash)
	if err != nil {
		return nil, err
	}
	return &SameFieldConflictError{
		Conflict: map[string]any{
			"conflict_token":            token,
			"record_id":                 recordID.String(),
			"field_key":                 change.FieldKey,
			"conflict_resolution_class": conflictClass,
			"base_row_version":          baseRowVersion,
			"current_row_version":       current.RowVersion,
			"client_value":              clientValue,
			"server_value":              serverValue,
			"server_updated_by":         changed.ServerUpdatedBy.String(),
			"server_updated_at":         formatTimestamp(changed.ServerUpdatedAt),
			"base_value":                baseValue,
		},
	}, nil
}

func (s *store) conflictToken(recordID uuid.UUID, fieldKey string, baseRowVersion int64, currentRowVersion int64, requestHash []byte) (string, error) {
	field, _ := viewschema.LookupField(TimelineViewSchemaID, fieldKey)
	conflictClass := field.ConflictResolutionClass
	if conflictClass == "" {
		conflictClass = "atomic_replace"
	}
	return s.conflictTokens.Issue(conflicttokens.ConflictTokenClaims{
		RouteKey:                conflictResolveRouteKey,
		RecordID:                recordID.String(),
		ViewSchemaID:            TimelineViewSchemaID,
		FieldKey:                fieldKey,
		ConflictResolutionClass: conflictClass,
		BaseRowVersion:          baseRowVersion,
		CurrentRowVersion:       currentRowVersion,
		RequestHash:             conflicttokens.RequestHashTokenValue(requestHash),
	})
}

func (s *store) parseConflictToken(token string) (conflicttokens.ConflictTokenClaims, bool) {
	return parseTimelineConflictTokenWithCodec(s.conflictTokens, token)
}

func parseTimelineConflictTokenWithCodec(codec conflicttokens.ConflictTokenCodec, token string) (conflicttokens.ConflictTokenClaims, bool) {
	claims, ok := codec.Parse(token)
	if !ok {
		return conflicttokens.ConflictTokenClaims{}, false
	}
	if claims.RouteKey != conflictResolveRouteKey || claims.ViewSchemaID != TimelineViewSchemaID {
		return conflicttokens.ConflictTokenClaims{}, false
	}
	return claims, true
}

func rowCellValue(row map[string]any, fieldKey string) (any, bool) {
	cells, ok := row["cells"].(map[string]any)
	if !ok {
		return nil, false
	}
	cell, ok := cells[fieldKey].(map[string]any)
	if !ok {
		return nil, false
	}
	value, ok := cell["value"]
	return value, ok
}

func patchClientConflictValue(recordID uuid.UUID, change PatchChange, baseValue any, requestHash []byte) (any, error) {
	if change.ActionPayload == nil {
		return canonicalChangeValue(change), nil
	}
	return applyCollectionConflictActions(recordID, change.FieldKey, baseValue, change.ActionPayload, requestHash)
}

func applyCollectionConflictActions(recordID uuid.UUID, fieldKey string, baseValue any, payload *CollectionActionPayload, requestHash []byte) (map[string]any, error) {
	ordered, items, ok := cloneCollectionConflictValue(baseValue)
	if !ok {
		return nil, fmt.Errorf("invalid base collection value for %s", fieldKey)
	}
	for index, action := range payload.Actions {
		switch action.Op {
		case "add_token", "add_tag":
			items = append(items, newClientCollectionItem(recordID, fieldKey, action, requestHash, index, false))
		case "add_resolved_ref":
			items = append(items, newClientCollectionItem(recordID, fieldKey, action, requestHash, index, true))
		case "add_record_ref":
			items = append(items, newClientCollectionItem(recordID, fieldKey, action, requestHash, index, true))
		case "resolve_item":
			if item := findCollectionItem(items, action.ItemRef); item != nil {
				item["item_kind"] = "resolved_ref"
				if action.ResolvedRecord != nil {
					item["resolved_record_id"] = action.ResolvedRecord.String()
				}
				removeResolutionMetadata(item, false)
			}
		case "dismiss_item":
			items = removeCollectionItem(items, action.ItemRef)
		case "revert_to_unresolved":
			if item := findCollectionItem(items, action.ItemRef); item != nil {
				item["item_kind"] = "unresolved_mention"
				removeResolutionMetadata(item, true)
			}
		case "remove_record_ref", "remove_tag":
			items = removeCollectionItem(items, action.ItemRef)
		default:
			return nil, fmt.Errorf("unsupported collection action: %s", action.Op)
		}
	}
	if !ordered {
		sort.SliceStable(items, func(left int, right int) bool {
			return collectionSortKey(items[left]) < collectionSortKey(items[right])
		})
	}
	return collectionValue(ordered, items), nil
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

func cloneMap(source map[string]any) map[string]any {
	cloned := make(map[string]any, len(source))
	for key, value := range source {
		cloned[key] = value
	}
	return cloned
}

func newClientCollectionItem(recordID uuid.UUID, fieldKey string, action CollectionAction, requestHash []byte, actionIndex int, resolved bool) map[string]any {
	rawText := action.RawText
	displayText := action.RawText
	if isTimelineTagCollection(fieldKey) {
		rawText = ""
		displayText = action.RawText
	}
	item := map[string]any{
		"item_ref":     clientCollectionItemRef(fieldKey, action, requestHash, actionIndex),
		"display_text": displayText,
		"raw_text":     rawText,
	}
	if isTimelineTagCollection(fieldKey) {
		item["item_kind"] = "tag"
		tagID := clientCollectionLocalUUID(recordID, fieldKey, action, requestHash, actionIndex)
		item["item_ref"] = linkRecordTagItemRef(recordID, tagID)
		item["tag_id"] = tagID.String()
		delete(item, "raw_text")
		return item
	}
	if isTimelineAttachedEvidenceCollection(fieldKey) {
		item["item_kind"] = "record_ref"
		if action.LinkedRecordID != nil {
			item["item_ref"] = linkRecordRefItemRef(*action.LinkedRecordID)
			item["linked_record_id"] = action.LinkedRecordID.String()
			item["display_text"] = action.LinkedRecordID.String()
		}
		return item
	}

	item["entity_type"] = collectionEntityType(fieldKey)
	if resolved {
		item["item_kind"] = "resolved_ref"
		if action.ResolvedRecord != nil {
			item["resolved_record_id"] = action.ResolvedRecord.String()
		}
		return item
	}
	item["item_kind"] = "unresolved_mention"
	return item
}

func clientCollectionLocalUUID(recordID uuid.UUID, fieldKey string, action CollectionAction, requestHash []byte, actionIndex int) uuid.UUID {
	sum := hashCanonicalValue(map[string]any{
		"request_hash": base64.RawURLEncoding.EncodeToString(requestHash),
		"record_id":    recordID.String(),
		"field_key":    fieldKey,
		"action_index": actionIndex,
		"op":           action.Op,
		"text":         action.NormalizedText,
	})
	return uuid.NewSHA1(uuid.NameSpaceOID, sum)
}

func clientCollectionItemRef(fieldKey string, action CollectionAction, requestHash []byte, actionIndex int) string {
	sum := hashCanonicalValue(map[string]any{
		"request_hash":     base64.RawURLEncoding.EncodeToString(requestHash),
		"field_key":        fieldKey,
		"action_index":     actionIndex,
		"op":               action.Op,
		"raw_text":         action.NormalizedText,
		"item_ref":         action.ItemRef,
		"linked_record_id": formatUUIDPointer(action.LinkedRecordID),
	})
	token := base64.RawURLEncoding.EncodeToString(sum)
	if len(token) > 18 {
		token = token[:18]
	}
	return "client:" + token
}

func collectionEntityType(fieldKey string) string {
	if policy, ok := timelineCollectionPolicy(fieldKey); ok && policy.ExpectedTargetType != "" {
		return policy.ExpectedTargetType
	}
	return "host"
}

func findCollectionItem(items []map[string]any, itemRef string) map[string]any {
	for _, item := range items {
		if item["item_ref"] == itemRef {
			return item
		}
	}
	return nil
}

func removeCollectionItem(items []map[string]any, itemRef string) []map[string]any {
	for index, item := range items {
		if item["item_ref"] == itemRef {
			return append(items[:index], items[index+1:]...)
		}
	}
	return items
}

func removeResolutionMetadata(item map[string]any, removeResolvedID bool) {
	if removeResolvedID {
		delete(item, "resolved_record_id")
	}
	delete(item, "resolution_method")
	delete(item, "auto_resolved")
	delete(item, "provenance")
	delete(item, "confidence")
	delete(item, "matched_alias_text")
}

func collectionSortKey(item map[string]any) string {
	for _, key := range []string{"display_text", "raw_text", "item_ref"} {
		if value, ok := item[key].(string); ok {
			return value
		}
	}
	return ""
}

func versionID(recordID uuid.UUID, rowVersion int64) string {
	return fmt.Sprintf("timeline:%s:%d", recordID.String(), rowVersion)
}

func hasMaterialChange(current sourcerepository.Snapshot, next sourcerepository.Snapshot) bool {
	return !stringPointersEqual(current.DateEnteredText, next.DateEnteredText) ||
		!stringPointersEqual(current.AnalystText, next.AnalystText) ||
		!stringPointersEqual(current.MitreStageText, next.MitreStageText) ||
		!stringPointersEqual(current.DeviceObjectText, next.DeviceObjectText) ||
		!stringPointersEqual(current.IPAddressText, next.IPAddressText) ||
		!stringPointersEqual(current.ActivityUTCText, next.ActivityUTCText) ||
		!stringPointersEqual(current.ActivityLocalText, next.ActivityLocalText) ||
		!stringPointersEqual(current.RawActivityText, next.RawActivityText) ||
		!stringPointersEqual(current.ActivitySynopsisText, next.ActivitySynopsisText) ||
		!stringPointersEqual(current.DataSourceText, next.DataSourceText) ||
		current.ActivityUTCGenerated != next.ActivityUTCGenerated ||
		current.ActivityLocalGenerated != next.ActivityLocalGenerated ||
		current.ActivityTimePairState != next.ActivityTimePairState
}

func isRecordLinkConflict(err error) bool {
	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) {
		return false
	}
	return pgErr.Code == "23505"
}

func extractUUIDFromPayload(payload map[string]any, path ...string) (uuid.UUID, error) {
	current := any(payload)
	for _, segment := range path {
		object, ok := current.(map[string]any)
		if !ok {
			return uuid.UUID{}, fmt.Errorf("decode payload path %q", strings.Join(path, "."))
		}
		current = object[segment]
	}
	text, ok := current.(string)
	if !ok || text == "" {
		return uuid.UUID{}, fmt.Errorf("decode payload uuid path %q", strings.Join(path, "."))
	}
	parsed, err := uuid.Parse(text)
	if err != nil {
		return uuid.UUID{}, err
	}
	return parsed, nil
}

func stringPointersEqual(left *string, right *string) bool {
	switch {
	case left == nil && right == nil:
		return true
	case left == nil || right == nil:
		return false
	default:
		return *left == *right
	}
}

func optionalUUIDFromPG(value pgtype.UUID) *uuid.UUID {
	if !value.Valid {
		return nil
	}
	parsed := uuid.Must(uuid.FromBytes(value.Bytes[:]))
	return &parsed
}

func optionalTextFromPG(value pgtype.Text) *string {
	if !value.Valid {
		return nil
	}
	text := value.String
	return &text
}

func optionalIntFromPG(value pgtype.Int4) *int {
	if !value.Valid {
		return nil
	}
	parsed := int(value.Int32)
	return &parsed
}

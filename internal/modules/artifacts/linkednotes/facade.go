package linkednotes

import (
	"bytes"
	"context"
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

	"github.com/JochiRaider/cartulary/internal/modules/artifacts"
	artifactprojection "github.com/JochiRaider/cartulary/internal/modules/artifacts/projectionprovider"
	"github.com/JochiRaider/cartulary/internal/modules/incidents"
	"github.com/JochiRaider/cartulary/internal/modules/links"
	"github.com/JochiRaider/cartulary/internal/modules/projections"
	"github.com/JochiRaider/cartulary/internal/modules/records"
	"github.com/JochiRaider/cartulary/internal/modules/revisions"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
)

type Facade struct {
	pool           postgres.DB
	authStore      *authn.Store
	incidentAccess incidents.Access
	artifactStore  *artifacts.Store
	linkStore      linkedNoteLinkPort
	projectionRows *projections.ArtifactRows
	recordStore    *records.Store
	revisionStore  revisionAppendPort
}

type linkedNoteLinkPort interface {
	ApplyPartyRefCollectionTx(context.Context, pgx.Tx, links.PartyRefCollectionCommand) (bool, error)
	ApplyRecordRefCollectionTx(context.Context, pgx.Tx, links.RecordRefCollectionCommand) (bool, error)
	ApplyTagCollectionTx(context.Context, pgx.Tx, links.TagCollectionCommand) (bool, error)
	InsertLinkedNoteReferenceTx(context.Context, pgx.Tx, uuid.UUID, uuid.UUID, uuid.UUID, uuid.UUID, time.Time) (links.RecordLink, bool, error)
	LoadRecordLinkValueTx(context.Context, pgx.Tx, uuid.UUID) (map[string]any, error)
	ValidatePartyRefCollectionTx(context.Context, pgx.Tx, links.PartyRefCollectionValidation) error
	ValidateRecordRefCollectionTx(context.Context, pgx.Tx, links.RecordRefCollectionValidation) error
	ValidateTagCollectionTx(context.Context, pgx.Tx, links.TagCollectionValidation) error
}

type CreateRequest struct {
	ClientTxnID string
	Values      map[string]artifacts.FieldValue
	Collections map[string]CollectionActionPayload
}

type CollectionActionPayload struct {
	Actions []CollectionAction
}

type CollectionAction struct {
	Op             string
	RawText        string
	LinkedRecordID *uuid.UUID
	PartyID        *uuid.UUID
	ItemRef        string
	NormalizedText string
}

type CreateCommand struct {
	Actor          authn.UserRecord
	SourceRecordID uuid.UUID
	Request        CreateRequest
	RequestHash    []byte
	RequestID      string
	RouteKey       string
	Now            time.Time
}

type MutationResult struct {
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

type MutationValidationError struct {
	Field      string
	ReasonCode string
}

func (e *MutationValidationError) Error() string {
	return "linkednotes: invalid mutation request"
}

func NewFacade(pool postgres.DB, appender *revisions.Appender) *Facade {
	return &Facade{
		pool:           pool,
		authStore:      authn.NewStore(pool),
		incidentAccess: incidents.NewAccess(pool),
		artifactStore:  artifacts.NewStore(appender),
		linkStore:      links.NewStore(),
		projectionRows: projections.NewArtifactRows(pool, artifactprojection.QuerySurfaces()...),
		recordStore:    records.NewStore(),
		revisionStore:  newRevisionAppendAdapter(appender),
	}
}

func (f *Facade) SourceIncident(ctx context.Context, sourceRecordID uuid.UUID) (uuid.UUID, error) {
	tx, err := f.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return uuid.UUID{}, err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	incidentID, err := sourceIncidentTx(ctx, tx, sourceRecordID)
	if err != nil {
		return uuid.UUID{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return uuid.UUID{}, err
	}
	return incidentID, nil
}

func (f *Facade) Create(ctx context.Context, command CreateCommand) (MutationResult, error) {
	request := command.Request
	idempotencyKey := authn.RouteIdempotencyKey{
		RouteKey:    command.RouteKey,
		ActorUserID: command.Actor.ID,
		ScopeKey:    command.SourceRecordID.String(),
		ClientTxnID: request.ClientTxnID,
	}
	if existing, err := f.authStore.GetRouteIdempotency(ctx, idempotencyKey); err == nil {
		if !bytes.Equal(existing.RequestHash, command.RequestHash) {
			return MutationResult{}, authn.ErrClientTxnConflict
		}
		payload, err := decodeStoredResponse(existing.ResponseJSON)
		if err != nil {
			return MutationResult{}, fmt.Errorf("decode replayed linked note payload: %w", err)
		}
		recordID, err := extractPayloadUUID(payload, "row", "record_id")
		if err != nil {
			return MutationResult{}, err
		}
		return MutationResult{Payload: payload, StatusCode: http.StatusOK, Replayed: true, RecordID: recordID, ViewSchemaID: artifacts.NotesViewSchemaID, ClientTxnID: request.ClientTxnID}, nil
	} else if !errors.Is(err, authn.ErrNotFound) {
		return MutationResult{}, fmt.Errorf("query linked note idempotency: %w", err)
	}
	if err := validateCreateRequest(request); err != nil {
		return MutationResult{}, err
	}

	tx, err := f.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return MutationResult{}, fmt.Errorf("begin linked note transaction: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	incidentID, err := sourceIncidentTx(ctx, tx, command.SourceRecordID)
	if err != nil {
		return MutationResult{}, err
	}
	if err := f.incidentAccess.EnsureOpenTx(ctx, tx, incidentID); err != nil {
		return MutationResult{}, err
	}
	if err := validateReferencesTx(ctx, tx, f.linkStore, incidentID, request); err != nil {
		return MutationResult{}, err
	}
	now := command.Now.UTC()
	recordID, err := f.recordStore.InsertTx(ctx, tx, records.InsertParams{
		IncidentID:      incidentID,
		RecordType:      "artifact",
		CreatedByUserID: command.Actor.ID,
		CreatedAt:       now,
		UpdatedByUserID: command.Actor.ID,
		UpdatedAt:       now,
		RowVersion:      1,
	})
	if err != nil {
		return MutationResult{}, err
	}
	if err := f.artifactStore.InsertRowTx(ctx, tx, recordID, incidentID, command.Actor.ID, artifacts.CreateParams{ViewSchemaID: artifacts.NotesViewSchemaID, Values: request.Values}, now); err != nil {
		return MutationResult{}, err
	}
	if _, err := applyLinkedNoteCollectionsTx(ctx, tx, f.linkStore, incidentID, recordID, command.Actor.ID, request.Collections, now); err != nil {
		return MutationResult{}, adaptCollectionValidationError(err)
	}
	linkRecord, linkInserted, err := f.linkStore.InsertLinkedNoteReferenceTx(ctx, tx, incidentID, command.SourceRecordID, recordID, command.Actor.ID, now)
	if err != nil {
		return MutationResult{}, err
	}
	if err := f.projectionRows.RefreshTx(ctx, tx, recordID); err != nil {
		return MutationResult{}, err
	}
	row, err := f.projectionRows.LoadTx(ctx, tx, artifacts.NotesViewSchemaID, recordID)
	if err != nil {
		return MutationResult{}, err
	}
	changeSetID, err := f.revisionStore.AppendChangeSetTx(ctx, tx, revisions.AppendChangeSetParams{
		IncidentID:  incidentID,
		ActorUserID: command.Actor.ID,
		Source:      command.RouteKey,
		ClientTxnID: &request.ClientTxnID,
		RequestID:   &command.RequestID,
		CreatedAt:   now,
	})
	if err != nil {
		return MutationResult{}, err
	}
	afterVersionID := versionID(recordID, 1)
	if err := f.revisionStore.AppendMutationTx(ctx, tx, revisions.AppendMutationParams{
		ChangeSetID:    changeSetID,
		SequenceNo:     1,
		TargetKind:     "record",
		TargetID:       recordID.String(),
		OperationKind:  "create",
		AfterVersionID: &afterVersionID,
		AfterValue:     row,
	}); err != nil {
		return MutationResult{}, err
	}
	if linkInserted {
		linkAfter, err := f.linkStore.LoadRecordLinkValueTx(ctx, tx, linkRecord.RecordLinkID)
		if err != nil {
			return MutationResult{}, err
		}
		if err := f.revisionStore.AppendMutationTx(ctx, tx, revisions.AppendMutationParams{
			ChangeSetID:   changeSetID,
			SequenceNo:    2,
			TargetKind:    "record_link",
			TargetID:      linkRecord.RecordLinkID.String(),
			OperationKind: "create",
			AfterValue:    linkAfter,
		}); err != nil {
			return MutationResult{}, err
		}
	}
	if err := f.revisionStore.AppendRecordRevisionTx(ctx, tx, revisions.AppendRecordRevisionParams{
		ChangeSetID: changeSetID,
		RecordID:    recordID,
		RowVersion:  1,
		AfterValue:  row,
	}); err != nil {
		return MutationResult{}, err
	}
	payload := buildMutationPayload(artifacts.NotesViewSchemaID, changeSetID, row)
	payload["source_record_id"] = command.SourceRecordID.String()
	payload["link_type"] = "references_artifact"
	if err := authn.InsertRouteIdempotencyPayload(ctx, tx, idempotencyKey, nil, command.RequestHash, http.StatusCreated, payload); err != nil {
		if authn.IsUniqueViolation(err) {
			return MutationResult{}, authn.ErrClientTxnConflict
		}
		return MutationResult{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return MutationResult{}, fmt.Errorf("commit linked note transaction: %w", err)
	}
	return MutationResult{
		Payload:          payload,
		StatusCode:       http.StatusCreated,
		IncidentID:       incidentID,
		RecordID:         recordID,
		ChangeSetID:      changeSetID,
		ClientTxnID:      request.ClientTxnID,
		RowVersion:       1,
		ViewSchemaID:     artifacts.NotesViewSchemaID,
		ChangedFieldKeys: changedFieldKeys(nil, row),
	}, nil
}

func validateCreateRequest(request CreateRequest) error {
	if !hasTextValue(request.Values, "note.title") && !hasTextValue(request.Values, "note.body") {
		return &MutationValidationError{Field: "payload", ReasonCode: "missing_minimum_create_signal"}
	}
	return nil
}

func adaptCollectionValidationError(err error) error {
	if err == nil {
		return nil
	}
	var validation *links.CollectionValidationError
	if errors.As(err, &validation) {
		return &MutationValidationError{Field: validation.Field, ReasonCode: validation.ReasonCode}
	}
	return err
}

func validateReferencesTx(ctx context.Context, tx pgx.Tx, linkStore linkedNoteLinkPort, incidentID uuid.UUID, request CreateRequest) error {
	for fieldKey, value := range request.Values {
		if value.UUID != nil && strings.HasSuffix(fieldKey, "_user_id") {
			if err := validateActiveUserTx(ctx, tx, *value.UUID, fieldKey); err != nil {
				return err
			}
		}
	}
	for fieldKey, payload := range request.Collections {
		if err := validateLinkedNoteCollectionPayloadTx(ctx, tx, linkStore, incidentID, fieldKey, payload); err != nil {
			return adaptCollectionValidationError(err)
		}
	}
	return nil
}

func validateLinkedNoteCollectionPayloadTx(ctx context.Context, tx pgx.Tx, linkStore linkedNoteLinkPort, incidentID uuid.UUID, fieldKey string, payload CollectionActionPayload) error {
	policy, ok := artifacts.LookupCollectionPolicy(fieldKey)
	if !ok || !policy.AllowsLinksCollectionMutation() {
		return collectionValidationError(fieldKey)
	}
	switch {
	case policy.AllowsRecordRefs():
		command, err := linkedNoteRecordRefValidation(incidentID, policy, payload)
		if err != nil {
			return err
		}
		return linkStore.ValidateRecordRefCollectionTx(ctx, tx, command)
	case policy.AllowsPartyRefs():
		command, err := linkedNotePartyRefValidation(incidentID, policy, payload)
		if err != nil {
			return err
		}
		return linkStore.ValidatePartyRefCollectionTx(ctx, tx, command)
	case policy.AllowsTags():
		command, err := linkedNoteTagValidation(policy, payload)
		if err != nil {
			return err
		}
		return linkStore.ValidateTagCollectionTx(ctx, tx, command)
	default:
		return collectionValidationError(fieldKey)
	}
}

func applyLinkedNoteCollectionsTx(ctx context.Context, tx pgx.Tx, linkStore linkedNoteLinkPort, incidentID uuid.UUID, recordID uuid.UUID, actorID uuid.UUID, collections map[string]CollectionActionPayload, now time.Time) (bool, error) {
	changed := false
	for fieldKey, payload := range collections {
		policy, ok := artifacts.LookupCollectionPolicy(fieldKey)
		if !ok || !policy.AllowsLinksCollectionMutation() {
			return false, collectionValidationError(fieldKey)
		}
		var (
			applied bool
			err     error
		)
		switch {
		case policy.AllowsRecordRefs():
			command, convertErr := linkedNoteRecordRefCommand(incidentID, recordID, actorID, policy, payload, now)
			if convertErr != nil {
				return false, convertErr
			}
			applied, err = linkStore.ApplyRecordRefCollectionTx(ctx, tx, command)
		case policy.AllowsPartyRefs():
			command, convertErr := linkedNotePartyRefCommand(incidentID, recordID, actorID, policy, payload, now)
			if convertErr != nil {
				return false, convertErr
			}
			applied, err = linkStore.ApplyPartyRefCollectionTx(ctx, tx, command)
		case policy.AllowsTags():
			command, convertErr := linkedNoteTagCommand(incidentID, recordID, actorID, policy, payload, now)
			if convertErr != nil {
				return false, convertErr
			}
			applied, err = linkStore.ApplyTagCollectionTx(ctx, tx, command)
		default:
			return false, collectionValidationError(fieldKey)
		}
		if err != nil {
			return false, err
		}
		changed = changed || applied
	}
	return changed, nil
}

func linkedNoteRecordRefValidation(incidentID uuid.UUID, policy artifacts.CollectionPolicy, payload CollectionActionPayload) (links.RecordRefCollectionValidation, error) {
	adds, removes, err := linkedNoteRecordRefActions(policy, payload)
	return links.RecordRefCollectionValidation{IncidentID: incidentID, FieldKey: policy.FieldKey, LinkType: links.LinkType(policy.LinkType), ExpectedTargetType: policy.ExpectedTargetType, AddRecordIDs: adds, RemoveRecordIDs: removes}, err
}

func linkedNoteRecordRefCommand(incidentID uuid.UUID, recordID uuid.UUID, actorID uuid.UUID, policy artifacts.CollectionPolicy, payload CollectionActionPayload, now time.Time) (links.RecordRefCollectionCommand, error) {
	adds, removes, err := linkedNoteRecordRefActions(policy, payload)
	return links.RecordRefCollectionCommand{IncidentID: incidentID, SourceRecordID: recordID, ActorUserID: actorID, FieldKey: policy.FieldKey, LinkType: links.LinkType(policy.LinkType), ExpectedTargetType: policy.ExpectedTargetType, AddRecordIDs: adds, RemoveRecordIDs: removes, Now: now}, err
}

func linkedNoteRecordRefActions(policy artifacts.CollectionPolicy, payload CollectionActionPayload) ([]uuid.UUID, []uuid.UUID, error) {
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

func linkedNotePartyRefValidation(incidentID uuid.UUID, policy artifacts.CollectionPolicy, payload CollectionActionPayload) (links.PartyRefCollectionValidation, error) {
	adds, removes, err := linkedNotePartyRefActions(policy, payload)
	return links.PartyRefCollectionValidation{IncidentID: incidentID, FieldKey: policy.FieldKey, LinkType: links.LinkType(policy.LinkType), ExpectedTargetType: policy.ExpectedTargetType, AddPartyIDs: adds, RemovePartyIDs: removes}, err
}

func linkedNotePartyRefCommand(incidentID uuid.UUID, recordID uuid.UUID, actorID uuid.UUID, policy artifacts.CollectionPolicy, payload CollectionActionPayload, now time.Time) (links.PartyRefCollectionCommand, error) {
	adds, removes, err := linkedNotePartyRefActions(policy, payload)
	return links.PartyRefCollectionCommand{IncidentID: incidentID, SourceRecordID: recordID, ActorUserID: actorID, FieldKey: policy.FieldKey, LinkType: links.LinkType(policy.LinkType), ExpectedTargetType: policy.ExpectedTargetType, AddPartyIDs: adds, RemovePartyIDs: removes, Now: now}, err
}

func linkedNotePartyRefActions(policy artifacts.CollectionPolicy, payload CollectionActionPayload) ([]uuid.UUID, []uuid.UUID, error) {
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

func linkedNoteTagValidation(policy artifacts.CollectionPolicy, payload CollectionActionPayload) (links.TagCollectionValidation, error) {
	adds, removes, err := linkedNoteTagActions(policy, payload)
	return links.TagCollectionValidation{FieldKey: policy.FieldKey, AddTags: adds, RemoveTags: removes}, err
}

func linkedNoteTagCommand(incidentID uuid.UUID, recordID uuid.UUID, actorID uuid.UUID, policy artifacts.CollectionPolicy, payload CollectionActionPayload, now time.Time) (links.TagCollectionCommand, error) {
	adds, removes, err := linkedNoteTagActions(policy, payload)
	return links.TagCollectionCommand{IncidentID: incidentID, RecordID: recordID, ActorUserID: actorID, FieldKey: policy.FieldKey, AddTags: adds, RemoveTags: removes, Now: now}, err
}

func linkedNoteTagActions(policy artifacts.CollectionPolicy, payload CollectionActionPayload) ([]links.TagCollectionAdd, []links.RecordTagRef, error) {
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

func collectionValidationError(fieldKey string) *links.CollectionValidationError {
	return &links.CollectionValidationError{Field: fieldKey, ReasonCode: "invalid_value"}
}

func validateActiveUserTx(ctx context.Context, tx pgx.Tx, userID uuid.UUID, field string) error {
	var exists bool
	if err := tx.QueryRow(ctx, `SELECT EXISTS (SELECT 1 FROM users WHERE id = $1 AND is_active = true)`, userID).Scan(&exists); err != nil {
		return fmt.Errorf("validate linked note user: %w", err)
	}
	if !exists {
		return &MutationValidationError{Field: field, ReasonCode: "invalid_value"}
	}
	return nil
}

func sourceIncidentTx(ctx context.Context, tx pgx.Tx, sourceRecordID uuid.UUID) (uuid.UUID, error) {
	var incidentID uuid.UUID
	err := tx.QueryRow(ctx, `
SELECT incident_id
  FROM records
 WHERE record_id = $1
   AND record_type IN ('timeline_event', 'host', 'identity', 'evidence')
   AND deleted_at IS NULL
`, sourceRecordID).Scan(&incidentID)
	return incidentID, err
}

func hasTextValue(values map[string]artifacts.FieldValue, field string) bool {
	value, ok := values[field]
	return ok && value.Text != nil && strings.TrimSpace(*value.Text) != ""
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

func versionID(recordID uuid.UUID, rowVersion int64) string {
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
		return uuid.UUID{}, fmt.Errorf("decode payload uuid path %q", strings.Join(path, "."))
	}
	return uuid.Parse(text)
}

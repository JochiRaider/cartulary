package collaboration

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	privatestream "github.com/JochiRaider/cartulary/internal/modules/collaboration/internal/stream"
)

var ErrIntentKeyCollision = errors.New("collaboration intent key collision")

type PublicationContribution struct {
	ContributionID string
	SourceOwnerID  string
	RecordTypes    []string
	AffectedViews  []ViewPublicationContribution
}

type ViewPublicationContribution struct {
	ViewSchemaID    string
	PublicFieldKeys []string
	PatchFieldKeys  []string
}

type publicationView struct {
	ownerID     string
	recordTypes map[string]struct{}
	publicKeys  map[string]struct{}
	patchKeys   map[string]struct{}
}

// PublicationCatalog is an immutable, composition-scoped disclosure policy.
// It validates owner-declared public effects and contains no projection or
// source-row callbacks.
type PublicationCatalog struct {
	views map[string]publicationView
}

func NewPublicationCatalog(contributions []PublicationContribution) (*PublicationCatalog, error) {
	if len(contributions) == 0 {
		return nil, errors.New("collaboration publication catalog is empty")
	}
	catalog := &PublicationCatalog{views: make(map[string]publicationView)}
	contributionIDs := make(map[string]struct{}, len(contributions))
	for _, contribution := range contributions {
		contributionID := strings.TrimSpace(contribution.ContributionID)
		ownerID := strings.TrimSpace(contribution.SourceOwnerID)
		if contributionID == "" || ownerID == "" || len(contribution.RecordTypes) == 0 || len(contribution.AffectedViews) == 0 {
			return nil, fmt.Errorf("collaboration publication contribution %q is incomplete", contributionID)
		}
		if _, duplicate := contributionIDs[contributionID]; duplicate {
			return nil, fmt.Errorf("collaboration publication contribution %q is duplicated", contributionID)
		}
		contributionIDs[contributionID] = struct{}{}
		recordTypes, err := uniqueStrings(contribution.RecordTypes, "record type")
		if err != nil {
			return nil, fmt.Errorf("collaboration publication contribution %q: %w", contributionID, err)
		}
		for _, view := range contribution.AffectedViews {
			viewSchemaID := strings.TrimSpace(view.ViewSchemaID)
			if viewSchemaID == "" {
				return nil, fmt.Errorf("collaboration publication contribution %q has an empty view schema", contributionID)
			}
			if _, duplicate := catalog.views[viewSchemaID]; duplicate {
				return nil, fmt.Errorf("collaboration publication view %q is duplicated", viewSchemaID)
			}
			publicKeys, err := uniqueStrings(view.PublicFieldKeys, "public field key")
			if err != nil {
				return nil, fmt.Errorf("collaboration publication view %q: %w", viewSchemaID, err)
			}
			patchKeys, err := uniqueStrings(view.PatchFieldKeys, "patch field key")
			if err != nil {
				return nil, fmt.Errorf("collaboration publication view %q: %w", viewSchemaID, err)
			}
			for key := range patchKeys {
				if _, public := publicKeys[key]; !public {
					return nil, fmt.Errorf("collaboration publication view %q patch field %q is not public", viewSchemaID, key)
				}
			}
			catalog.views[viewSchemaID] = publicationView{
				ownerID: ownerID, recordTypes: recordTypes, publicKeys: publicKeys, patchKeys: patchKeys,
			}
		}
	}
	return catalog, nil
}

func uniqueStrings(values []string, label string) (map[string]struct{}, error) {
	if len(values) == 0 {
		return nil, fmt.Errorf("%s set is empty", label)
	}
	result := make(map[string]struct{}, len(values))
	for _, value := range values {
		trimmed := strings.TrimSpace(value)
		if trimmed == "" || trimmed != value {
			return nil, fmt.Errorf("%s %q is invalid", label, value)
		}
		if _, duplicate := result[value]; duplicate {
			return nil, fmt.Errorf("%s %q is duplicated", label, value)
		}
		result[value] = struct{}{}
	}
	return result, nil
}

type RecordChangeIntentInput struct {
	IncidentID      uuid.UUID
	RecordID        uuid.UUID
	ChangeSetID     uuid.UUID
	ActorUserID     uuid.UUID
	RowVersion      int64
	ClientTxnID     string
	MutationOrdinal int
	CreatedAt       time.Time
	PublicFieldKeys []string
	AffectedViews   []AffectedViewChange
}

type AffectedViewChange struct {
	ViewSchemaID string
	RecordID     uuid.UUID
	RowVersion   int64
	ChangeKind   string
	PatchCells   map[string]any
}

func validAffectedViewChangeKind(changeKind string) bool {
	switch changeKind {
	case "invalidate", "patch", "remove":
		return true
	default:
		return false
	}
}

type JobProgressIntentInput struct {
	IntentKey        string
	IncidentID       uuid.UUID
	CanonicalPayload json.RawMessage
	SourceIdentity   string
	CreatedAt        time.Time
}

type ExtensionResourceChangeIntentInput struct {
	IntentKey        string
	IncidentID       uuid.UUID
	CanonicalPayload json.RawMessage
	SourceIdentity   string
	CreatedAt        time.Time
}

type RecordChangedAppender interface {
	AppendRecordChangedTx(context.Context, pgx.Tx, RecordChangeIntentInput) error
}

type PublicationAppender interface {
	RecordChangedAppender
	AppendJobProgressTx(context.Context, pgx.Tx, JobProgressIntentInput) error
	AppendExtensionResourceChangedTx(context.Context, pgx.Tx, ExtensionResourceChangeIntentInput) error
}

type publicationAppender struct {
	catalog  *PublicationCatalog
	delegate *privatestream.PostgresStream
}

func NewPublicationAppender(catalog *PublicationCatalog) (PublicationAppender, error) {
	if catalog == nil {
		return nil, errors.New("collaboration publication catalog is required")
	}
	return &publicationAppender{catalog: catalog, delegate: privatestream.NewPostgresStream(nil, nil)}, nil
}

func (appender *publicationAppender) AppendRecordChangedTx(ctx context.Context, tx pgx.Tx, input RecordChangeIntentInput) error {
	intent, err := appender.recordChangedIntent(input)
	if err != nil {
		return err
	}
	return appender.appendTx(ctx, tx, intent)
}

func (appender *publicationAppender) AppendJobProgressTx(ctx context.Context, tx pgx.Tx, input JobProgressIntentInput) error {
	intent, err := privatestream.NewEventIntent(input.IntentKey, input.IncidentID, privatestream.EventFamilyJobProgress, input.CanonicalPayload, input.SourceIdentity, 0, input.CreatedAt)
	if err != nil {
		return err
	}
	return appender.appendTx(ctx, tx, intent)
}

func (appender *publicationAppender) AppendExtensionResourceChangedTx(ctx context.Context, tx pgx.Tx, input ExtensionResourceChangeIntentInput) error {
	intent, err := privatestream.NewEventIntent(input.IntentKey, input.IncidentID, privatestream.EventFamilyExtensionResourceChange, input.CanonicalPayload, input.SourceIdentity, 0, input.CreatedAt)
	if err != nil {
		return err
	}
	return appender.appendTx(ctx, tx, intent)
}

func (appender *publicationAppender) appendTx(ctx context.Context, tx pgx.Tx, intent privatestream.EventIntent) error {
	if appender == nil || appender.delegate == nil || appender.catalog == nil {
		return errors.New("collaboration publication appender is not configured")
	}
	err := appender.delegate.AppendIntentTx(ctx, tx, intent)
	if errors.Is(err, privatestream.ErrIntentKeyCollision) {
		return fmt.Errorf("%w: %v", ErrIntentKeyCollision, err)
	}
	return err
}

func (appender *publicationAppender) recordChangedIntent(input RecordChangeIntentInput) (privatestream.EventIntent, error) {
	if appender == nil || appender.catalog == nil || input.RecordID == uuid.Nil || input.ChangeSetID == uuid.Nil ||
		input.IncidentID == uuid.Nil || input.ActorUserID == uuid.Nil || input.RowVersion < 1 || len(input.AffectedViews) == 0 {
		return privatestream.EventIntent{}, errors.New("record_change_intent_v1 identity is incomplete")
	}
	changedKeys := append([]string(nil), input.PublicFieldKeys...)
	slices.Sort(changedKeys)
	changedKeys = slices.Compact(changedKeys)
	for _, key := range changedKeys {
		if strings.TrimSpace(key) == "" || key != strings.TrimSpace(key) {
			return privatestream.EventIntent{}, fmt.Errorf("record_change_intent_v1 field key %q is invalid", key)
		}
	}
	views := cloneAffectedViewChanges(input.AffectedViews)
	slices.SortFunc(views, func(left AffectedViewChange, right AffectedViewChange) int {
		return strings.Compare(left.ViewSchemaID, right.ViewSchemaID)
	})
	admittedKeys := make(map[string]struct{})
	for index, view := range views {
		policy, known := appender.catalog.views[view.ViewSchemaID]
		if !known || view.RecordID != input.RecordID || view.RowVersion != input.RowVersion ||
			!validAffectedViewChangeKind(view.ChangeKind) || (view.ChangeKind == "patch") != (view.PatchCells != nil) {
			return privatestream.EventIntent{}, fmt.Errorf("record_change_intent_v1 affected view %q is invalid", view.ViewSchemaID)
		}
		if index > 0 && views[index-1].ViewSchemaID == view.ViewSchemaID {
			return privatestream.EventIntent{}, errors.New("record_change_intent_v1 affected views must be unique")
		}
		for key := range policy.publicKeys {
			admittedKeys[key] = struct{}{}
		}
		if view.PatchCells != nil {
			if err := validatePatch(view, policy, input.RecordID, input.RowVersion, changedKeys); err != nil {
				return privatestream.EventIntent{}, err
			}
		}
	}
	for _, key := range changedKeys {
		if _, admitted := admittedKeys[key]; !admitted {
			return privatestream.EventIntent{}, fmt.Errorf("record_change_intent_v1 field key %q is not public for an affected view", key)
		}
	}
	payload := recordChangePayload(input, views, changedKeys)
	intent, err := privatestream.NewEventIntent(
		fmt.Sprintf("record_changed:%s:%s:%d", input.ChangeSetID, input.RecordID, input.RowVersion),
		input.IncidentID,
		privatestream.EventFamilyRecordChanged,
		payload,
		input.ChangeSetID.String()+":"+input.RecordID.String(),
		input.MutationOrdinal,
		input.CreatedAt,
	)
	if err != nil {
		return privatestream.EventIntent{}, err
	}
	intent.SourceChangeSetID = &input.ChangeSetID
	intent.SourceRecordID = &input.RecordID
	intent.SourceRowVersion = &input.RowVersion
	return intent, nil
}

func validatePatch(view AffectedViewChange, policy publicationView, recordID uuid.UUID, rowVersion int64, changedKeys []string) error {
	if len(view.PatchCells) == 0 {
		return fmt.Errorf("record_change_intent_v1 affected view %q has an empty patch", view.ViewSchemaID)
	}
	for key := range view.PatchCells {
		switch key {
		case "record_id", "row_version", "cells", "group_values":
		default:
			return fmt.Errorf("record_change_intent_v1 patch member %q is not admitted", key)
		}
	}
	patchRecordID, ok := view.PatchCells["record_id"].(string)
	if !ok || patchRecordID != recordID.String() {
		return fmt.Errorf("record_change_intent_v1 patch record identity is invalid")
	}
	if !sameRowVersion(view.PatchCells["row_version"], rowVersion) {
		return fmt.Errorf("record_change_intent_v1 patch row version is invalid")
	}
	cells, ok := view.PatchCells["cells"].(map[string]any)
	if !ok || len(cells) == 0 {
		return fmt.Errorf("record_change_intent_v1 patch cells are invalid")
	}
	for key := range cells {
		if _, admitted := policy.patchKeys[key]; !admitted {
			return fmt.Errorf("record_change_intent_v1 patch field %q is not admitted for view %q", key, view.ViewSchemaID)
		}
		if !slices.Contains(changedKeys, key) {
			return fmt.Errorf("record_change_intent_v1 patch field %q was not declared changed", key)
		}
	}
	if groups, present := view.PatchCells["group_values"]; present {
		groupValues, ok := groups.(map[string]any)
		if !ok {
			return fmt.Errorf("record_change_intent_v1 patch group values are invalid")
		}
		for key := range groupValues {
			if _, admitted := policy.patchKeys[key]; !admitted || !slices.Contains(changedKeys, key) {
				return fmt.Errorf("record_change_intent_v1 patch group field %q is not admitted", key)
			}
		}
	}
	return nil
}

func sameRowVersion(value any, expected int64) bool {
	switch typed := value.(type) {
	case int:
		return int64(typed) == expected
	case int32:
		return int64(typed) == expected
	case int64:
		return typed == expected
	case float64:
		return int64(typed) == expected && typed == float64(expected)
	default:
		return false
	}
}

func cloneAffectedViewChanges(input []AffectedViewChange) []AffectedViewChange {
	result := make([]AffectedViewChange, len(input))
	for index, view := range input {
		result[index] = view
		if view.PatchCells != nil {
			result[index].PatchCells = make(map[string]any, len(view.PatchCells))
			for key, value := range view.PatchCells {
				result[index].PatchCells[key] = value
			}
		}
	}
	return result
}

func recordChangePayload(input RecordChangeIntentInput, views []AffectedViewChange, changedKeys []string) map[string]any {
	affectedViews := make([]map[string]any, 0, len(views))
	for _, affected := range views {
		view := map[string]any{"view_schema_id": affected.ViewSchemaID, "change_kind": affected.ChangeKind}
		if affected.PatchCells != nil {
			view["patch_cells"] = affected.PatchCells
		}
		affectedViews = append(affectedViews, view)
	}
	return map[string]any{
		"record_id": input.RecordID.String(), "row_version": input.RowVersion,
		"change_set_id": input.ChangeSetID.String(), "client_txn_id": input.ClientTxnID,
		"actor_user_id": input.ActorUserID.String(), "changed_field_keys": changedKeys,
		"affected_views": affectedViews,
	}
}

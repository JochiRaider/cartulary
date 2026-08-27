package workbookroutetest

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/modules/workbook"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/platform/querypage"
	"github.com/JochiRaider/cartulary/internal/platform/viewschema"
)

// ValueChange is a route-shaped value used by focused catalog tests. Source
// semantics stay in the owner decoders invoked by the catalog.
type ValueChange struct {
	Kind      string
	Text      *string
	Timestamp *time.Time
	UUID      *uuid.UUID
	Number    *int64
	Bool      *bool
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
	RiskRefText    string
	NormalizedText string
}

type CreateInputValue struct {
	UUID uuid.UUID
}

type CreateRequest struct {
	ViewSchemaID string
	ClientTxnID  string
	Values       map[string]ValueChange
	Collections  map[string]CollectionActionPayload
	Inputs       map[string]CreateInputValue
}

type PatchChange struct {
	FieldKey   string
	Value      *ValueChange
	Collection *CollectionActionPayload
}

type PatchRequest struct {
	ViewSchemaID   string
	BaseRowVersion int64
	ClientTxnID    string
	Changes        []PatchChange
}

// QueryRows invokes the exact query-provider boundary used by Workbook routes
// with the effectively unbounded window needed by focused integration tests.
func QueryRows(
	catalog *workbook.WorkbookContributionCatalog,
	ctx context.Context,
	incidentID uuid.UUID,
	viewSchemaID string,
	query viewschema.QueryMeta,
) ([]map[string]any, error) {
	if catalog == nil {
		return nil, fmt.Errorf("workbook contribution catalog is required")
	}
	provider, ok := catalog.QueryFor(viewSchemaID)
	if !ok {
		return nil, fmt.Errorf("workbook query surface %q is not registered", viewSchemaID)
	}
	page, err := provider.QueryRowsPage(ctx, workbook.QueryCommand{
		IncidentID:   incidentID,
		ViewSchemaID: viewSchemaID,
		Query:        query,
		Window:       querypage.Window{Limit: int(^uint(0)>>1) - 1},
	})
	return page.Rows, err
}

// MutationFailureError preserves the catalog's closed failure value without
// recreating Workbook's deleted legacy error hierarchy.
type MutationFailureError struct {
	Failure *workbook.MutationFailure
}

func (err *MutationFailureError) Error() string {
	if err == nil || err.Failure == nil {
		return "workbook catalog mutation failed"
	}
	return fmt.Sprintf("workbook catalog mutation failed: %s", err.Failure.Kind())
}

// CreateWorkbookRow invokes the exact neutral create-provider boundary used by
// Workbook routes. It is intentionally stateless and performs no dispatch of
// its own beyond the immutable catalog lookup.
func CreateWorkbookRow(
	catalog *workbook.WorkbookContributionCatalog,
	ctx context.Context,
	actor authn.UserRecord,
	incidentID uuid.UUID,
	request CreateRequest,
	requestID string,
	now time.Time,
) (workbook.MutationResult, error) {
	if catalog == nil {
		return workbook.MutationResult{}, fmt.Errorf("workbook contribution catalog is required")
	}
	provider, ok := catalog.CreateFor(request.ViewSchemaID)
	if !ok {
		return workbook.MutationResult{}, &MutationFailureError{Failure: workbook.InvalidPayloadFailure("view_schema_id", "unknown_view_schema")}
	}
	payload, err := json.Marshal(createPayload(request))
	if err != nil {
		return workbook.MutationResult{}, err
	}
	operation, failure, err := provider.DecodeCreate(bytes.NewReader(payload))
	if err != nil {
		return workbook.MutationResult{}, err
	}
	if failure != nil {
		return workbook.MutationResult{}, &MutationFailureError{Failure: failure}
	}
	outcome, err := operation.Execute(ctx, workbook.CreateCommand{
		Actor: actor, IncidentID: incidentID, ViewSchemaID: request.ViewSchemaID,
		RequestID: requestID, Now: now,
	})
	return mutationResult(outcome, err)
}

// PatchWorkbookRow invokes the exact neutral patch-provider boundary used by
// Workbook routes after resolving the authoritative record type.
func PatchWorkbookRow(
	catalog *workbook.WorkbookContributionCatalog,
	ctx context.Context,
	actor authn.UserRecord,
	recordID uuid.UUID,
	request PatchRequest,
	requestID string,
	now time.Time,
) (workbook.MutationResult, error) {
	if catalog == nil {
		return workbook.MutationResult{}, fmt.Errorf("workbook contribution catalog is required")
	}
	resource, ok := viewschema.LookupPublicResource(request.ViewSchemaID)
	if !ok || len(resource.SourceRecordTypes) != 1 {
		return workbook.MutationResult{}, &MutationFailureError{Failure: workbook.InvalidPayloadFailure("view_schema_id", "unknown_view_schema")}
	}
	authoritativeRecordType := resource.SourceRecordTypes[0]
	provider, ok := catalog.PatchFor(authoritativeRecordType)
	if !ok {
		return workbook.MutationResult{}, &MutationFailureError{Failure: workbook.InvalidPayloadFailure("view_schema_id", "unknown_view_schema")}
	}
	payload, err := json.Marshal(patchPayload(request))
	if err != nil {
		return workbook.MutationResult{}, err
	}
	operation, failure, err := provider.DecodePatch(bytes.NewReader(payload))
	if err != nil {
		return workbook.MutationResult{}, err
	}
	if failure != nil {
		return workbook.MutationResult{}, &MutationFailureError{Failure: failure}
	}
	outcome, err := operation.Execute(ctx, workbook.PatchCommand{
		Actor: actor, RecordID: recordID, AuthoritativeRecordType: authoritativeRecordType,
		RequestID: requestID, Now: now,
	})
	return mutationResult(outcome, err)
}

func mutationResult(outcome workbook.MutationOutcome, err error) (workbook.MutationResult, error) {
	if err != nil {
		return workbook.MutationResult{}, err
	}
	if validationErr := outcome.Validate(); validationErr != nil {
		return workbook.MutationResult{}, validationErr
	}
	if result, ok := outcome.Result(); ok {
		return result, nil
	}
	failure, _ := outcome.Failure()
	return workbook.MutationResult{}, &MutationFailureError{Failure: failure}
}

func createPayload(request CreateRequest) map[string]any {
	payload := map[string]any{"client_txn_id": request.ClientTxnID}
	for fieldKey, value := range request.Values {
		payload[fieldKey] = canonicalValue(value)
	}
	for fieldKey, collection := range request.Collections {
		payload[fieldKey] = collectionPayload(collection)
	}
	for inputKey, value := range request.Inputs {
		payload[inputKey] = value.UUID.String()
	}
	return payload
}

func patchPayload(request PatchRequest) map[string]any {
	changes := make([]map[string]any, 0, len(request.Changes))
	for _, change := range request.Changes {
		entry := map[string]any{"field_key": change.FieldKey}
		if change.Collection != nil {
			entry["action_payload"] = collectionPayload(*change.Collection)
		} else if change.Value != nil {
			entry["value"] = canonicalValue(*change.Value)
		}
		changes = append(changes, entry)
	}
	return map[string]any{
		"view_schema_id": request.ViewSchemaID, "base_row_version": request.BaseRowVersion,
		"client_txn_id": request.ClientTxnID, "changes": changes,
	}
}

func collectionPayload(payload CollectionActionPayload) map[string]any {
	actions := make([]map[string]any, 0, len(payload.Actions))
	for _, action := range payload.Actions {
		entry := map[string]any{"op": action.Op}
		switch action.Op {
		case "add_token":
			entry["raw_text"] = action.RawText
		case "add_tag":
			entry["tag_name"] = action.RawText
		case "add_record_ref":
			if action.LinkedRecordID != nil {
				entry["linked_record_id"] = action.LinkedRecordID.String()
			}
		case "add_party_ref":
			if action.PartyID != nil {
				entry["party_id"] = action.PartyID.String()
			}
		case "add_risk_ref":
			entry["risk_ref_text"] = action.RiskRefText
		default:
			entry["item_ref"] = action.ItemRef
		}
		actions = append(actions, entry)
	}
	return map[string]any{"kind": "collection_actions_v1", "actions": actions}
}

func canonicalValue(value ValueChange) any {
	switch value.Kind {
	case "null":
		return nil
	case "text":
		if value.Text != nil {
			return *value.Text
		}
	case "timestamp":
		if value.Timestamp != nil {
			return value.Timestamp.UTC().Format(time.RFC3339Nano)
		}
	case "uuid":
		if value.UUID != nil {
			return value.UUID.String()
		}
	case "number":
		if value.Number != nil {
			return *value.Number
		}
	case "bool":
		if value.Bool != nil {
			return *value.Bool
		}
	}
	return nil
}

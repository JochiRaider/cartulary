package parties_test

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/modules/workbook"
	phase4storetest "github.com/JochiRaider/cartulary/internal/modules/workbook/testsupport/phase4storetest"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
)

func TestPhase9_U_9_11_DirectPartyReferencesAcceptOnlyExactStableIDs(t *testing.T) {
	stablePartyID := "11111111-2222-3333-4444-555555555555"
	valid := `{"view_schema_id":"cartulary.view.evidence.v1","base_row_version":1,"client_txn_id":"txn","changes":[{"field_key":"evidence.collector_party_id","value":"` + stablePartyID + `"}]}`
	request, apiErr := workbook.DecodePatchRequest(strings.NewReader(valid))
	if apiErr != nil {
		t.Fatalf("expected exact stable party id to decode: %#v", apiErr)
	}
	if got := request.Changes[0].Value.UUID.String(); got != stablePartyID {
		t.Fatalf("unexpected decoded party id: got %s want %s", got, stablePartyID)
	}

	clear := `{"view_schema_id":"cartulary.view.evidence.v1","base_row_version":1,"client_txn_id":"txn","changes":[{"field_key":"evidence.collector_party_id","value":null}]}`
	request, apiErr = workbook.DecodePatchRequest(strings.NewReader(clear))
	if apiErr != nil {
		t.Fatalf("expected direct null clear to decode: %#v", apiErr)
	}
	if request.Changes[0].Value == nil || request.Changes[0].Value.Kind != "null" {
		t.Fatalf("expected direct null clear value, got %#v", request.Changes[0].Value)
	}

	invalidValues := []string{
		`" 11111111-2222-3333-4444-555555555555"`,
		`"11111111-2222-3333-4444-555555555555 "`,
		`"collector@example.test"`,
		`"Incident Commander"`,
		`"party:11111111-2222-3333-4444-555555555555"`,
		`""`,
		`[]`,
		`{}`,
		`true`,
		`42`,
	}
	for _, rawValue := range invalidValues {
		body := `{"view_schema_id":"cartulary.view.evidence.v1","base_row_version":1,"client_txn_id":"txn","changes":[{"field_key":"evidence.collector_party_id","value":` + rawValue + `}]}`
		if _, apiErr := workbook.DecodePatchRequest(strings.NewReader(body)); apiErr == nil {
			t.Fatalf("expected invalid direct party ref %s to fail", rawValue)
		}
	}

	actionPayloadClear := `{"view_schema_id":"cartulary.view.evidence.v1","base_row_version":1,"client_txn_id":"txn","changes":[{"field_key":"evidence.collector_party_id","action_payload":{"kind":"collection_actions_v1","actions":[{"op":"remove_party_ref","item_ref":"party_ref:` + stablePartyID + `"}]}}]}`
	if _, apiErr := workbook.DecodePatchRequest(strings.NewReader(actionPayloadClear)); apiErr == nil {
		t.Fatalf("expected non-direct clear shape to fail")
	}
}

func TestPhase9_U_9_11_DirectDecisionReferenceDecoderAcceptsOnlyExactStableIDs(t *testing.T) {
	stableDecisionID := "11111111-2222-3333-4444-555555555555"
	valid := `{"view_schema_id":"cartulary.view.task_requests.v1","base_row_version":1,"client_txn_id":"txn","changes":[{"field_key":"task.decision_record_id","value":"` + stableDecisionID + `"}]}`
	request, apiErr := workbook.DecodePatchRequest(strings.NewReader(valid))
	if apiErr != nil {
		t.Fatalf("expected exact stable decision id to decode: %#v", apiErr)
	}
	if got := request.Changes[0].Value.UUID.String(); got != stableDecisionID {
		t.Fatalf("unexpected decoded decision id: got %s want %s", got, stableDecisionID)
	}

	clear := `{"view_schema_id":"cartulary.view.task_requests.v1","base_row_version":1,"client_txn_id":"txn","changes":[{"field_key":"task.decision_record_id","value":null}]}`
	request, apiErr = workbook.DecodePatchRequest(strings.NewReader(clear))
	if apiErr != nil {
		t.Fatalf("expected direct null clear to decode: %#v", apiErr)
	}
	if request.Changes[0].Value == nil || request.Changes[0].Value.Kind != "null" {
		t.Fatalf("expected direct null clear value, got %#v", request.Changes[0].Value)
	}

	invalidValues := []string{
		`" 11111111-2222-3333-4444-555555555555"`,
		`"11111111-2222-3333-4444-555555555555 "`,
		`"decision@example.test"`,
		`"Contain endpoint"`,
		`"decision:11111111-2222-3333-4444-555555555555"`,
		`""`,
		`[]`,
		`{}`,
		`true`,
		`42`,
	}
	for _, rawValue := range invalidValues {
		body := `{"view_schema_id":"cartulary.view.task_requests.v1","base_row_version":1,"client_txn_id":"txn","changes":[{"field_key":"task.decision_record_id","value":` + rawValue + `}]}`
		if _, apiErr := workbook.DecodePatchRequest(strings.NewReader(body)); apiErr == nil {
			t.Fatalf("expected invalid direct decision ref %s to fail", rawValue)
		}
	}

	actionPayloadClear := `{"view_schema_id":"cartulary.view.task_requests.v1","base_row_version":1,"client_txn_id":"txn","changes":[{"field_key":"task.decision_record_id","action_payload":{"kind":"collection_actions_v1","actions":[{"op":"remove_record_ref","item_ref":"record_ref:` + stableDecisionID + `"}]}}]}`
	if _, apiErr := workbook.DecodePatchRequest(strings.NewReader(actionPayloadClear)); apiErr == nil {
		t.Fatalf("expected non-direct clear shape to fail")
	}
}

func TestPhase9_U_9_11_DirectPartyReferencesRequireSameIncidentActiveParties(t *testing.T) {
	ctx := context.Background()
	harness := phase4storetest.StartStore(t, "phase9-u-9-11-party-refs")
	store := workbook.NewStore(harness.DB)
	actor := phase4storetest.SeedLocalUserFlags(t, harness.DB, "u911@example.test", "U911 Party Refs", "U911PartyRefs1!", false, false, true)
	incident := phase4storetest.CreateIncidentInStore(t, harness.DB, actor, "txn-phase9-u-9-11-incident", "IR-U911", "Phase 9 U-9-11")
	otherIncident := phase4storetest.CreateIncidentInStore(t, harness.DB, actor, "txn-phase9-u-9-11-other-incident", "IR-U911B", "Phase 9 U-9-11 Other")

	activeParty := mustCreatePartyForU911(t, store, actor, incident.ID, "txn-phase9-u-9-11-active-party", "Active Party")
	foreignParty := mustCreatePartyForU911(t, store, actor, otherIncident.ID, "txn-phase9-u-9-11-foreign-party", "Foreign Party")
	deletedParty := mustCreatePartyForU911(t, store, actor, incident.ID, "txn-phase9-u-9-11-deleted-party", "Deleted Party")
	if _, err := harness.DB.Exec(ctx, `UPDATE records SET deleted_at = $2, deleted_by_user_id = $3 WHERE record_id = $1`, deletedParty, time.Date(2026, 5, 18, 13, 0, 0, 0, time.UTC), actor.ID); err != nil {
		t.Fatalf("soft-delete party target: %v", err)
	}
	wrongType := mustCreateEvidenceForU911(t, store, actor, incident.ID, "txn-phase9-u-9-11-wrong-type-evidence", "Wrong type evidence")

	for _, tc := range []struct {
		name         string
		viewSchemaID string
		recordID     uuid.UUID
		baseVersion  int64
		fieldKey     string
	}{
		{
			name:         "collector",
			viewSchemaID: workbook.EvidenceViewSchemaID,
			recordID:     mustCreateEvidenceForU911(t, store, actor, incident.ID, "txn-phase9-u-9-11-collector-evidence", "Collector party ref"),
			baseVersion:  1,
			fieldKey:     "evidence.collector_party_id",
		},
		{
			name:         "source",
			viewSchemaID: workbook.EvidenceViewSchemaID,
			recordID:     mustCreateEvidenceForU911(t, store, actor, incident.ID, "txn-phase9-u-9-11-source-evidence", "Source party ref"),
			baseVersion:  1,
			fieldKey:     "evidence.source_party_id",
		},
		{
			name:         "requester",
			viewSchemaID: workbook.TaskRequestsViewSchemaID,
			recordID:     mustCreateTaskForU911(t, store, actor, incident.ID, "txn-phase9-u-9-11-requester-task", "Requester party ref"),
			baseVersion:  1,
			fieldKey:     "task.requester_party_id",
		},
	} {
		linked := mustPatchPartyRefForU911(t, store, actor, tc.recordID, tc.viewSchemaID, tc.baseVersion, tc.fieldKey, &activeParty, "txn-phase9-u-9-11-"+tc.name+"-link-active")
		requireU911CellValue(t, linked, tc.fieldKey, activeParty.String())

		linkedVersion := mustU911RowVersion(t, linked)
		cleared := mustPatchPartyRefForU911(t, store, actor, tc.recordID, tc.viewSchemaID, linkedVersion, tc.fieldKey, nil, "txn-phase9-u-9-11-"+tc.name+"-clear")
		requireU911CellValue(t, cleared, tc.fieldKey, nil)
		clearVersion := mustU911RowVersion(t, cleared)

		for _, invalid := range []struct {
			name string
			id   uuid.UUID
		}{
			{name: "foreign", id: foreignParty},
			{name: "deleted", id: deletedParty},
			{name: "wrong-type", id: wrongType},
			{name: "deployment-user", id: actor.ID},
		} {
			err := patchPartyRefForU911(store, actor, tc.recordID, tc.viewSchemaID, clearVersion, tc.fieldKey, &invalid.id, "txn-phase9-u-9-11-"+tc.name+"-"+invalid.name)
			var validationErr *workbook.MutationValidationError
			if !errors.As(err, &validationErr) {
				t.Fatalf("%s %s: expected mutation validation error, got %v", tc.name, invalid.name, err)
			}
			if validationErr.Field != tc.fieldKey || validationErr.ReasonCode != "invalid_value" {
				t.Fatalf("%s %s: unexpected validation details: %#v", tc.name, invalid.name, validationErr)
			}
			value, version := mustLoadU911PartyRefState(t, harness.DB, tc.recordID, tc.fieldKey)
			if value != nil {
				t.Fatalf("%s %s invalid write changed %s: got %#v want nil", tc.name, invalid.name, tc.fieldKey, value)
			}
			if got := version; got != clearVersion {
				t.Fatalf("%s %s invalid write changed row version: got %d want %d", tc.name, invalid.name, got, clearVersion)
			}
		}
	}
}

func mustCreatePartyForU911(t testing.TB, store *workbook.Store, actor authn.UserRecord, incidentID uuid.UUID, clientTxnID string, displayName string) uuid.UUID {
	t.Helper()
	result, err := store.CreateWorkbookRow(context.Background(), actor, incidentID, workbook.CreateRequest{
		ViewSchemaID: workbook.PartiesViewSchemaID,
		ClientTxnID:  clientTxnID,
		Values: map[string]workbook.ValueChange{
			"party.display_name": {Kind: "text", Text: &displayName},
			"party.party_kind":   {Kind: "text", Text: stringPtrU911("organization")},
		},
	}, []byte(clientTxnID), "req-"+clientTxnID, time.Date(2026, 5, 18, 12, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("create party %s: %v", clientTxnID, err)
	}
	return result.RecordID
}

func mustCreateEvidenceForU911(t testing.TB, store *workbook.Store, actor authn.UserRecord, incidentID uuid.UUID, clientTxnID string, title string) uuid.UUID {
	t.Helper()
	result, err := store.CreateWorkbookRow(context.Background(), actor, incidentID, workbook.CreateRequest{
		ViewSchemaID: workbook.EvidenceViewSchemaID,
		ClientTxnID:  clientTxnID,
		Values: map[string]workbook.ValueChange{
			"evidence.title": {Kind: "text", Text: &title},
		},
	}, []byte(clientTxnID), "req-"+clientTxnID, time.Date(2026, 5, 18, 12, 1, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("create evidence %s: %v", clientTxnID, err)
	}
	return result.RecordID
}

func mustCreateTaskForU911(t testing.TB, store *workbook.Store, actor authn.UserRecord, incidentID uuid.UUID, clientTxnID string, title string) uuid.UUID {
	t.Helper()
	result, err := store.CreateWorkbookRow(context.Background(), actor, incidentID, workbook.CreateRequest{
		ViewSchemaID: workbook.TaskRequestsViewSchemaID,
		ClientTxnID:  clientTxnID,
		Values: map[string]workbook.ValueChange{
			"task.title":     {Kind: "text", Text: &title},
			"task.task_kind": {Kind: "text", Text: stringPtrU911("request")},
		},
	}, []byte(clientTxnID), "req-"+clientTxnID, time.Date(2026, 5, 18, 12, 2, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("create task %s: %v", clientTxnID, err)
	}
	return result.RecordID
}

func mustPatchPartyRefForU911(t testing.TB, store *workbook.Store, actor authn.UserRecord, recordID uuid.UUID, viewSchemaID string, baseRowVersion int64, fieldKey string, partyID *uuid.UUID, clientTxnID string) map[string]any {
	t.Helper()
	result, err := patchPartyRefResultForU911(store, actor, recordID, viewSchemaID, baseRowVersion, fieldKey, partyID, clientTxnID)
	if err != nil {
		t.Fatalf("patch %s: %v", clientTxnID, err)
	}
	return result.Payload["row"].(map[string]any)
}

func patchPartyRefForU911(store *workbook.Store, actor authn.UserRecord, recordID uuid.UUID, viewSchemaID string, baseRowVersion int64, fieldKey string, partyID *uuid.UUID, clientTxnID string) error {
	_, err := patchPartyRefResultForU911(store, actor, recordID, viewSchemaID, baseRowVersion, fieldKey, partyID, clientTxnID)
	return err
}

func patchPartyRefResultForU911(store *workbook.Store, actor authn.UserRecord, recordID uuid.UUID, viewSchemaID string, baseRowVersion int64, fieldKey string, partyID *uuid.UUID, clientTxnID string) (workbook.MutationResult, error) {
	value := workbook.ValueChange{Kind: "null"}
	if partyID != nil {
		value = workbook.ValueChange{Kind: "uuid", UUID: partyID}
	}
	return store.PatchWorkbookRow(context.Background(), actor, recordID, workbook.PatchRequest{
		ViewSchemaID:   viewSchemaID,
		BaseRowVersion: baseRowVersion,
		ClientTxnID:    clientTxnID,
		Changes: []workbook.PatchChange{
			{FieldKey: fieldKey, Value: &value},
		},
	}, []byte(clientTxnID), "req-"+clientTxnID, time.Date(2026, 5, 18, 12, 3, 0, 0, time.UTC))
}

func mustLoadU911PartyRefState(t testing.TB, db postgres.DB, recordID uuid.UUID, fieldKey string) (any, int64) {
	t.Helper()
	table, column := u911TableColumn(t, fieldKey)
	var (
		value      sql.NullString
		rowVersion int64
	)
	query := fmt.Sprintf(`
SELECT %s::text, r.row_version
  FROM %s s
  JOIN records r ON r.record_id = s.record_id
 WHERE s.record_id = $1
`, column, table)
	if err := db.QueryRow(context.Background(), query, recordID).Scan(&value, &rowVersion); err != nil {
		t.Fatalf("load %s state: %v", fieldKey, err)
	}
	if !value.Valid {
		return nil, rowVersion
	}
	return value.String, rowVersion
}

func u911TableColumn(t testing.TB, fieldKey string) (string, string) {
	t.Helper()
	switch fieldKey {
	case "evidence.collector_party_id":
		return "evidence", "collector_party_id"
	case "evidence.source_party_id":
		return "evidence", "source_party_id"
	case "task.requester_party_id":
		return "task_requests", "requester_party_id"
	default:
		t.Fatalf("unsupported field key %s", fieldKey)
		return "", ""
	}
}

func requireU911CellValue(t testing.TB, row map[string]any, fieldKey string, want any) {
	t.Helper()
	got := row["cells"].(map[string]any)[fieldKey].(map[string]any)["value"]
	if got != want {
		t.Fatalf("unexpected %s value: got %#v want %#v", fieldKey, got, want)
	}
}

func mustU911RowVersion(t testing.TB, row map[string]any) int64 {
	t.Helper()
	switch value := row["row_version"].(type) {
	case int64:
		return value
	case int:
		return int64(value)
	case int32:
		return int64(value)
	case float64:
		return int64(value)
	default:
		t.Fatalf("unexpected row_version type %T", value)
		return 0
	}
}

func stringPtrU911(value string) *string {
	return &value
}

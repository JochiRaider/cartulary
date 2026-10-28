package workbook_test

import (
	"context"
	"slices"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/modules/artifacts"
	authstoretest "github.com/JochiRaider/cartulary/internal/modules/auth/testsupport/storetest"
	"github.com/JochiRaider/cartulary/internal/modules/workbook"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
	"github.com/JochiRaider/cartulary/internal/platform/querypage"
	"github.com/JochiRaider/cartulary/internal/platform/viewschema"
	"github.com/JochiRaider/cartulary/internal/testutil/appsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/httptestx"
	"github.com/JochiRaider/cartulary/internal/testutil/workbookroutetest"
)

func TestOptionalStandardizedSurfacesCatalogBehavior_Unit(t *testing.T) {
	requireOptionalSurfaceResources(t)

	ctx := context.Background()
	harness := appsupport.StartStore(t, "workbook_interaction-optional-surfaces-store")
	store := appsupport.NewWorkbookCatalog(harness.DB, workbookTestConflictTokens())
	actor := authstoretest.SeedActor(t, harness.DB, authstoretest.ActorAttributes{Email: "optional-surfaces@example.test", DisplayName: "Optional Surfaces"})
	incident := appsupport.CreateIncidentInStore(t, harness.DB, actor, "txn-workbook_interaction-optional-incident", "IR-OPTIONAL", "Workbook inspector optional surfaces")

	before := countOptionalSurfaceDurableState(t, harness.DB, incident.ID)
	createOutcome1, err := workbookroutetest.CreateWorkbookRow(store, ctx, workbook.CreateCommand{Actor: actor, IncidentID: incident.ID, ViewSchemaID: artifacts.FindingsViewSchemaID, RequestID: "req-workbook_interaction-optional-finding-reject", Now: Time(0)}, httptestx.JSONReader(t, map[string]any{"client_txn_id": "txn-workbook_interaction-optional-finding-reject",
		"finding.kind": "finding"}))
	requireMutationValidation(t, createOutcome1, err, "finding.statement", "missing_required_field")
	requireOptionalSurfaceDurableState(t, harness.DB, incident.ID, before, "rejected finding create")

	support := mustCreateRow(t, store, actor, incident.ID, artifacts.NotesViewSchemaID, "txn-workbook_interaction-optional-support-note", map[string]any{"note.title": "Supporting note"}, nil, Time(0))
	contradiction := mustCreateRow(t, store, actor, incident.ID, artifacts.NotesViewSchemaID, "txn-workbook_interaction-optional-contradiction-note", map[string]any{"note.title": "Contradictory note"}, nil, Time(0))

	finding := mustCreateRow(t, store, actor, incident.ID, artifacts.FindingsViewSchemaID, "txn-workbook_interaction-optional-finding", map[string]any{"finding.statement": "EDR shows suspicious process execution",
		"finding.kind":             "hypothesis",
		"finding.confidence_score": 72}, nil, Time(0))
	findingRow := finding.Payload["row"].(map[string]any)
	requireArtifactType(t, harness, finding.RecordID, "finding")
	requireCoordinationCellValue(t, findingRow, "finding.statement", "EDR shows suspicious process execution")
	requireCoordinationCellValue(t, findingRow, "finding.kind", "hypothesis")
	requireCoordinationCellValue(t, findingRow, "finding.state", "open")
	requireCoordinationCellValue(t, findingRow, "finding.owner_user_id", actor.ID.String())
	requireCellNumericValue(t, findingRow, "finding.confidence_score", 72)
	requireCoordinationCellValue(t, findingRow, "finding.confidence_band", "high")
	requireCoordinationCellValue(t, findingRow, "finding.closed_at", nil)
	requireCoordinationCollectionItemCount(t, findingRow, "finding.supporting_refs", 0)
	requireCoordinationCollectionItemCount(t, findingRow, "finding.contradictory_refs", 0)

	closed := mustPatch(t, store, actor, finding.RecordID, artifacts.FindingsViewSchemaID, 1, "txn-workbook_interaction-optional-finding-close",
		ValueChange("finding.state", "closed"),
		CollectionChange("finding.supporting_refs", Collection(addOptionalSurfaceRecordRef(support.RecordID))),
		CollectionChange("finding.contradictory_refs", Collection(addOptionalSurfaceRecordRef(contradiction.RecordID))),
	)
	closedRow := closed.Payload["row"].(map[string]any)
	requireCoordinationCellValue(t, closedRow, "finding.state", "closed")
	requireCellNonEmpty(t, closedRow, "finding.closed_at")
	requireCoordinationCollectionItemCount(t, closedRow, "finding.supporting_refs", 1)
	requireCoordinationCollectionItemCount(t, closedRow, "finding.contradictory_refs", 1)
	requireManualReferenceLink(t, harness, finding.RecordID, support.RecordID, "finding.supporting_refs", "supported_by")
	requireManualReferenceLink(t, harness, finding.RecordID, contradiction.RecordID, "finding.contradictory_refs", "references_record")

	reopened := mustPatch(t, store, actor, finding.RecordID, artifacts.FindingsViewSchemaID, 2, "txn-workbook_interaction-optional-finding-reopen",
		ValueChange("finding.state", "open"),
	)
	requireCoordinationCellValue(t, reopened.Payload["row"].(map[string]any), "finding.closed_at", nil)

	patchOutcome, err := Patch(t, store, actor, finding.RecordID, artifacts.FindingsViewSchemaID, 3, "txn-workbook_interaction-optional-finding-invalid-confidence",
		ValueChange("finding.confidence_score", 101),
	)
	requireMutationValidation(t, patchOutcome, err, "finding.confidence_score", "invalid_value")

	investigativeQuery := mustCreateRow(t, store, actor, incident.ID, artifacts.InvestigativeQueriesViewSchemaID, "txn-workbook_interaction-optional-query", map[string]any{"investigative_query.platform": "Kusto",
		"investigative_query.purpose":    "Find suspicious PowerShell",
		"investigative_query.query_text": "SecurityEvent | take 10"}, nil, Time(0))
	queryRow := investigativeQuery.Payload["row"].(map[string]any)
	requireArtifactType(t, harness, investigativeQuery.RecordID, "investigative_query")
	requireCellNonEmpty(t, queryRow, "investigative_query.query_id")
	requireCoordinationCellValue(t, queryRow, "investigative_query.platform", "Kusto")
	requireCoordinationCellValue(t, queryRow, "investigative_query.created_by_user_id", actor.ID.String())
	requireCellNonEmpty(t, queryRow, "investigative_query.created_at")
	requireCellNonEmpty(t, queryRow, "investigative_query.created_day")

	forensicKeyword := mustCreateRow(t, store, actor, incident.ID, artifacts.ForensicKeywordsViewSchemaID, "txn-workbook_interaction-optional-keyword", map[string]any{"forensic_keyword.pattern": "powershell.exe",
		"forensic_keyword.reason": "Interactive shell execution"}, nil, Time(0))
	keywordRow := forensicKeyword.Payload["row"].(map[string]any)
	requireArtifactType(t, harness, forensicKeyword.RecordID, "forensic_keyword")
	requireCellNonEmpty(t, keywordRow, "forensic_keyword.keyword_id")
	requireCoordinationCellValue(t, keywordRow, "forensic_keyword.match_mode", "literal")
	requireCoordinationCellValue(t, keywordRow, "forensic_keyword.case_sensitive", false)

	keywordPatched := mustPatch(t, store, actor, forensicKeyword.RecordID, artifacts.ForensicKeywordsViewSchemaID, 1, "txn-workbook_interaction-optional-keyword-patch",
		ValueChange("forensic_keyword.match_mode", "regex"),
		ValueChange("forensic_keyword.case_sensitive", true),
	)
	keywordPatchedRow := keywordPatched.Payload["row"].(map[string]any)
	requireCoordinationCellValue(t, keywordPatchedRow, "forensic_keyword.match_mode", "regex")
	requireCoordinationCellValue(t, keywordPatchedRow, "forensic_keyword.case_sensitive", true)

	patchOutcome, err = Patch(t, store, actor, forensicKeyword.RecordID, artifacts.ForensicKeywordsViewSchemaID, 2, "txn-workbook_interaction-optional-keyword-invalid-mode",
		ValueChange("forensic_keyword.match_mode", "glob"),
	)
	requireMutationValidation(t, patchOutcome, err, "forensic_keyword.match_mode", "invalid_value")
}

func TestOptionalStandardizedSurfacesProjectionQueryBehavior_Unit(t *testing.T) {
	harness := appsupport.StartStore(t, "workbook_interaction-optional-surfaces-query")
	store := appsupport.NewWorkbookCatalog(harness.DB, workbookTestConflictTokens())
	actor := authstoretest.SeedActor(t, harness.DB, authstoretest.ActorAttributes{Email: "optional-query@example.test", DisplayName: "Optional Query"})
	incident := appsupport.CreateIncidentInStore(t, harness.DB, actor, "txn-workbook_interaction-optional-query-incident", "IR-OPTIONAL-QUERY", "Workbook inspector optional query behavior")

	finding := mustCreateRow(t, store, actor, incident.ID, artifacts.FindingsViewSchemaID, "txn-workbook_interaction-optional-query-finding-high", map[string]any{"finding.statement": "Confirmed malware execution",
		"finding.kind":             "hypothesis",
		"finding.confidence_score": 91}, nil, Time(0))
	mustCreateRow(t, store, actor, incident.ID, artifacts.FindingsViewSchemaID, "txn-workbook_interaction-optional-query-finding-low", map[string]any{"finding.statement": "Possible benign admin action",
		"finding.kind":             "finding",
		"finding.confidence_score": 12}, nil, Time(0))
	requireProjectedRow(t, store, incident.ID, artifacts.FindingsViewSchemaID, viewschema.QueryMeta{
		Filters: []viewschema.Filter{{FieldKey: "finding.confidence_band", Op: "eq", Arg: map[string]any{"value": "high"}}},
		Sort:    []viewschema.SortEntry{{FieldKey: "finding.confidence_score", Direction: "desc"}},
		GroupBy: StringPtr("finding.kind"),
	}, finding.RecordID, "finding.kind", "hypothesis")

	query := mustCreateRow(t, store, actor, incident.ID, artifacts.InvestigativeQueriesViewSchemaID, "txn-workbook_interaction-optional-query-kusto", map[string]any{"investigative_query.platform": "Kusto",
		"investigative_query.purpose":    "Endpoint triage",
		"investigative_query.query_text": "DeviceProcessEvents | take 20"}, nil, Time(0))
	mustCreateRow(t, store, actor, incident.ID, artifacts.InvestigativeQueriesViewSchemaID, "txn-workbook_interaction-optional-query-splunk", map[string]any{"investigative_query.platform": "Splunk",
		"investigative_query.purpose":    "Network triage",
		"investigative_query.query_text": "index=proxy | head 20"}, nil, Time(0))
	requireProjectedRow(t, store, incident.ID, artifacts.InvestigativeQueriesViewSchemaID, viewschema.QueryMeta{
		Filters: []viewschema.Filter{{FieldKey: "investigative_query.platform", Op: "eq", Arg: map[string]any{"value": "KUSTO"}}},
		Sort:    []viewschema.SortEntry{{FieldKey: "investigative_query.created_day", Direction: "asc"}},
		GroupBy: StringPtr("investigative_query.platform"),
	}, query.RecordID, "investigative_query.platform", "Kusto")

	keyword := mustCreateRow(t, store, actor, incident.ID, artifacts.ForensicKeywordsViewSchemaID, "txn-workbook_interaction-optional-query-keyword-regex", map[string]any{"forensic_keyword.pattern": "(?i)powershell",
		"forensic_keyword.reason":         "Shell execution",
		"forensic_keyword.match_mode":     "regex",
		"forensic_keyword.case_sensitive": true}, nil, Time(0))
	mustCreateRow(t, store, actor, incident.ID, artifacts.ForensicKeywordsViewSchemaID, "txn-workbook_interaction-optional-query-keyword-literal", map[string]any{"forensic_keyword.pattern": "cmd.exe",
		"forensic_keyword.reason": "Command shell"}, nil, Time(0))
	requireProjectedRow(t, store, incident.ID, artifacts.ForensicKeywordsViewSchemaID, viewschema.QueryMeta{
		Filters: []viewschema.Filter{{FieldKey: "forensic_keyword.case_sensitive", Op: "eq", Arg: map[string]any{"value": true}}},
		Sort:    []viewschema.SortEntry{{FieldKey: "forensic_keyword.match_mode", Direction: "desc"}},
		GroupBy: StringPtr("forensic_keyword.match_mode"),
	}, keyword.RecordID, "forensic_keyword.match_mode", "regex")
}

func TestFindingsConfidenceBandBoundaries_Unit(t *testing.T) {
	harness := appsupport.StartStore(t, "workbook_interaction-optional-findings-confidence-boundaries")
	store := appsupport.NewWorkbookCatalog(harness.DB, workbookTestConflictTokens())
	actor := authstoretest.SeedActor(t, harness.DB, authstoretest.ActorAttributes{Email: "optional-boundaries@example.test", DisplayName: "Optional Boundaries"})
	incident := appsupport.CreateIncidentInStore(t, harness.DB, actor, "txn-workbook_interaction-optional-boundaries-incident", "IR-OPTIONAL-BAND", "Workbook inspector optional finding confidence bands")

	createdByBand := map[string][]uuid.UUID{}
	for index, tc := range []struct {
		key      string
		score    *int64
		wantBand string
	}{
		{key: "unset", score: nil, wantBand: "unset"},
		{key: "low-zero", score: optionalScorePtr(0), wantBand: "low"},
		{key: "low-upper", score: optionalScorePtr(39), wantBand: "low"},
		{key: "medium-lower", score: optionalScorePtr(40), wantBand: "medium"},
		{key: "medium-upper", score: optionalScorePtr(69), wantBand: "medium"},
		{key: "high-lower", score: optionalScorePtr(70), wantBand: "high"},
		{key: "high-upper", score: optionalScorePtr(100), wantBand: "high"},
	} {
		values := map[string]any{"finding.statement": "Confidence boundary " + tc.key}
		if tc.score != nil {
			values["finding.confidence_score"] = *tc.score
		}
		result := mustCreateRow(t, store, actor, incident.ID, artifacts.FindingsViewSchemaID, "txn-workbook_interaction-optional-boundary-"+tc.key, values, nil, Time(time.Duration(index)*time.Minute))
		row := result.Payload["row"].(map[string]any)
		if tc.score == nil {
			requireCoordinationCellValue(t, row, "finding.confidence_score", nil)
		} else {
			requireCellNumericValue(t, row, "finding.confidence_score", *tc.score)
		}
		requireCoordinationCellValue(t, row, "finding.confidence_band", tc.wantBand)
		createdByBand[tc.wantBand] = append(createdByBand[tc.wantBand], result.RecordID)
	}

	requireOptionalSurfaceBandQuery(t, store, incident.ID, "unset", createdByBand["unset"])
	requireOptionalSurfaceBandQuery(t, store, incident.ID, "low", createdByBand["low"])
	requireOptionalSurfaceBandQuery(t, store, incident.ID, "medium", createdByBand["medium"])
	requireOptionalSurfaceBandQuery(t, store, incident.ID, "high", createdByBand["high"])
}

func requireOptionalSurfaceResources(t testing.TB) {
	t.Helper()
	resources := viewschema.ListPublicResources()
	if len(resources) != 17 {
		t.Fatalf("current build discovery must expose fourteen required plus three optional surfaces, got %d", len(resources))
	}
	if _, ok := viewschema.LookupPublicResource("cartulary.view.hypotheses.v1"); ok {
		t.Fatal("cartulary.view.hypotheses.v1 must not be exposed")
	}
	tests := []struct {
		viewSchemaID string
		artifactType string
		wantFields   []string
	}{
		{
			viewSchemaID: artifacts.FindingsViewSchemaID,
			artifactType: "finding",
			wantFields: []string{
				"finding.statement",
				"finding.kind",
				"finding.state",
				"finding.owner_user_id",
				"finding.confidence_score",
				"finding.closed_at",
				"finding.updated_at",
				"finding.supporting_refs",
				"finding.contradictory_refs",
				"finding.confidence_band",
			},
		},
		{
			viewSchemaID: artifacts.InvestigativeQueriesViewSchemaID,
			artifactType: "investigative_query",
			wantFields: []string{
				"investigative_query.platform",
				"investigative_query.purpose",
				"investigative_query.query_text",
				"investigative_query.created_by_user_id",
				"investigative_query.created_at",
				"investigative_query.query_id",
				"investigative_query.created_day",
			},
		},
		{
			viewSchemaID: artifacts.ForensicKeywordsViewSchemaID,
			artifactType: "forensic_keyword",
			wantFields: []string{
				"forensic_keyword.pattern",
				"forensic_keyword.reason",
				"forensic_keyword.match_mode",
				"forensic_keyword.case_sensitive",
				"forensic_keyword.created_at",
				"forensic_keyword.keyword_id",
				"forensic_keyword.created_day",
			},
		},
	}
	for _, tc := range tests {
		schema, ok := viewschema.Lookup(tc.viewSchemaID)
		if !ok {
			t.Fatalf("missing optional schema %s", tc.viewSchemaID)
		}
		if schema.BaseProjection != "artifact_grid_projection" {
			t.Fatalf("%s base projection: got %q", tc.viewSchemaID, schema.BaseProjection)
		}
		filter, ok := schema.CanonicalSourceFilter()
		wantFilter := viewschema.CanonicalSourceFilter{Kind: "artifact_type", Field: "artifact_type", Value: tc.artifactType}
		if !ok || filter != wantFilter {
			t.Fatalf("%s filter: got %#v ok=%v want %#v", tc.viewSchemaID, filter, ok, wantFilter)
		}
		resource, ok := viewschema.LookupPublicResource(tc.viewSchemaID)
		if !ok {
			t.Fatalf("missing public resource %s", tc.viewSchemaID)
		}
		gotFields := make([]string, 0, len(resource.Fields))
		for _, field := range resource.Fields {
			gotFields = append(gotFields, field.FieldKey)
		}
		if !slices.Equal(gotFields, tc.wantFields) {
			t.Fatalf("%s fields:\ngot  %v\nwant %v", tc.viewSchemaID, gotFields, tc.wantFields)
		}
	}
}

func requireOptionalSurfaceBandQuery(t testing.TB, store *workbook.WorkbookContributionCatalog, incidentID uuid.UUID, band string, want []uuid.UUID) {
	t.Helper()
	queryProvider1, ok := store.QueryFor(artifacts.FindingsViewSchemaID)
	if !ok {
		t.Fatalf("query provider missing for %s", artifacts.FindingsViewSchemaID)
	}
	rowsPage, err := queryProvider1.QueryRowsPage(context.Background(), workbook.QueryCommand{IncidentID: incidentID, ViewSchemaID: artifacts.FindingsViewSchemaID, Query: viewschema.QueryMeta{
		Filters: []viewschema.Filter{{FieldKey: "finding.confidence_band", Op: "eq", Arg: map[string]any{"value": band}}},
		Sort: []viewschema.SortEntry{
			{FieldKey: "finding.confidence_score", Direction: "asc"},
			{FieldKey: "record_id", Direction: "asc"},
		},
		GroupBy: StringPtr("finding.confidence_band"),
	}, Window: querypage.Window{Limit: 100}})
	if rowsPage.HasMore {
		t.Fatal("fixture exceeds explicit 100-row query window")
	}
	rows := rowsPage.Rows
	if err != nil {
		t.Fatalf("query finding confidence band %s: %v", band, err)
	}
	requireRecordOrder(t, rows, want)
	for _, row := range rows {
		groupValues := row["group_values"].(map[string]any)
		if groupValues["finding.confidence_band"] != band {
			t.Fatalf("finding confidence group value: got %#v want %q", groupValues["finding.confidence_band"], band)
		}
	}
}

func optionalScorePtr(value int64) *int64 {
	return &value
}

func mustPatch(t testing.TB, store *workbook.WorkbookContributionCatalog, actor authn.UserRecord, recordID uuid.UUID, viewSchemaID string, baseRowVersion int64, clientTxnID string, changes ...map[string]any) workbook.MutationResult {
	t.Helper()
	outcome, err := Patch(t, store, actor, recordID, viewSchemaID, baseRowVersion, clientTxnID, changes...)
	return workbookroutetest.RequireMutationResult(t, outcome, err)
}

func Patch(t testing.TB, store *workbook.WorkbookContributionCatalog, actor authn.UserRecord, recordID uuid.UUID, viewSchemaID string, baseRowVersion int64, clientTxnID string, changes ...map[string]any) (workbook.MutationOutcome, error) {
	return workbookroutetest.PatchWorkbookRow(store, context.Background(), workbook.PatchCommand{
		Actor: actor, RecordID: recordID, AuthoritativeRecordType: "artifact", RequestID: "req-" + clientTxnID, Now: Time(0),
	}, httptestx.JSONReader(t, map[string]any{"view_schema_id": viewSchemaID, "base_row_version": baseRowVersion, "client_txn_id": clientTxnID, "changes": changes}))
}

func ValueChange(fieldKey string, value any) map[string]any {
	return map[string]any{"field_key": fieldKey, "value": value}
}

func CollectionChange(fieldKey string, value map[string]any) map[string]any {
	return map[string]any{"field_key": fieldKey, "action_payload": value}
}

func Collection(actions ...map[string]any) map[string]any {
	return map[string]any{"kind": "collection_actions_v1", "actions": actions}
}

func addOptionalSurfaceRecordRef(recordID uuid.UUID) map[string]any {
	return map[string]any{"op": "add_record_ref", "linked_record_id": recordID}
}

func requireCellNumericValue(t testing.TB, row map[string]any, fieldKey string, want int64) {
	t.Helper()
	got := row["cells"].(map[string]any)[fieldKey].(map[string]any)["value"]
	switch value := got.(type) {
	case int:
		if int64(value) == want {
			return
		}
	case int32:
		if int64(value) == want {
			return
		}
	case int64:
		if value == want {
			return
		}
	case float64:
		if int64(value) == want {
			return
		}
	}
	t.Fatalf("unexpected %s value: got %#v want %d", fieldKey, got, want)
}

type optionalSurfaceDurableState struct {
	Records              int
	Findings             int
	InvestigativeQueries int
	ForensicKeywords     int
}

func countOptionalSurfaceDurableState(t testing.TB, db postgres.DB, incidentID uuid.UUID) optionalSurfaceDurableState {
	t.Helper()
	var state optionalSurfaceDurableState
	if err := db.QueryRow(context.Background(), `
SELECT COUNT(*)
  FROM records r
  JOIN artifacts a
    ON a.incident_id = r.incident_id
   AND a.record_id = r.record_id
 WHERE r.incident_id = $1
   AND r.record_type = 'artifact'
   AND a.artifact_type IN ('finding', 'investigative_query', 'forensic_keyword')
`, incidentID).Scan(&state.Records); err != nil {
		t.Fatalf("count optional surface records: %v", err)
	}
	if err := db.QueryRow(context.Background(), `SELECT COUNT(*) FROM artifact_findings WHERE incident_id = $1`, incidentID).Scan(&state.Findings); err != nil {
		t.Fatalf("count findings: %v", err)
	}
	if err := db.QueryRow(context.Background(), `SELECT COUNT(*) FROM artifact_investigative_queries WHERE incident_id = $1`, incidentID).Scan(&state.InvestigativeQueries); err != nil {
		t.Fatalf("count investigative queries: %v", err)
	}
	if err := db.QueryRow(context.Background(), `SELECT COUNT(*) FROM artifact_forensic_keywords WHERE incident_id = $1`, incidentID).Scan(&state.ForensicKeywords); err != nil {
		t.Fatalf("count forensic keywords: %v", err)
	}
	return state
}

func requireOptionalSurfaceDurableState(t testing.TB, db postgres.DB, incidentID uuid.UUID, want optionalSurfaceDurableState, context string) {
	t.Helper()
	got := countOptionalSurfaceDurableState(t, db, incidentID)
	if got != want {
		t.Fatalf("%s changed optional durable state: got %+v want %+v", context, got, want)
	}
}

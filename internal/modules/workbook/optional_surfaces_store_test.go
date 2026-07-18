package workbook_test

import (
	"context"
	"slices"
	"testing"
	"time"

	"github.com/google/uuid"

	recordstoretest "github.com/JochiRaider/cartulary/internal/modules/records/testsupport/storetest"
	"github.com/JochiRaider/cartulary/internal/modules/workbook"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
	"github.com/JochiRaider/cartulary/internal/platform/viewschema"
)

func TestOptionalStandardizedSurfacesStoreBehavior_Unit(t *testing.T) {
	requireOptionalSurfaceResources(t)

	ctx := context.Background()
	harness := recordstoretest.StartStore(t, "phase9-optional-surfaces-store")
	store := workbook.NewStore(harness.DB)
	actor := recordstoretest.SeedLocalUserFlags(t, harness.DB, "optional-surfaces@example.test", "Optional Surfaces", "OptionalSurfaces1!", false, false, true)
	incident := recordstoretest.CreateIncidentInStore(t, harness.DB, actor, "txn-phase9-optional-incident", "IR-OPTIONAL", "Workbook inspector optional surfaces")

	before := countOptionalSurfaceDurableState(t, harness.DB, incident.ID)
	_, err := store.CreateWorkbookRow(ctx, actor, incident.ID, workbook.CreateRequest{
		ViewSchemaID: workbook.FindingsViewSchemaID,
		ClientTxnID:  "txn-phase9-optional-finding-reject",
		Values: map[string]workbook.ValueChange{
			"finding.kind": Text("finding"),
		},
	}, []byte("txn-phase9-optional-finding-reject"), "req-phase9-optional-finding-reject", Time(0))
	requireMutationValidation(t, err, "finding.statement", "missing_required_field")
	requireOptionalSurfaceDurableState(t, harness.DB, incident.ID, before, "rejected finding create")

	support := mustCreateRow(t, store, actor, incident.ID, workbook.NotesViewSchemaID, "txn-phase9-optional-support-note", map[string]workbook.ValueChange{
		"note.title": Text("Supporting note"),
	}, nil, Time(0))
	contradiction := mustCreateRow(t, store, actor, incident.ID, workbook.NotesViewSchemaID, "txn-phase9-optional-contradiction-note", map[string]workbook.ValueChange{
		"note.title": Text("Contradictory note"),
	}, nil, Time(0))

	finding := mustCreateRow(t, store, actor, incident.ID, workbook.FindingsViewSchemaID, "txn-phase9-optional-finding", map[string]workbook.ValueChange{
		"finding.statement":        Text("EDR shows suspicious process execution"),
		"finding.kind":             Text("hypothesis"),
		"finding.confidence_score": optionalNumber(72),
	}, nil, Time(0))
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

	closed := mustPatch(t, store, actor, finding.RecordID, workbook.FindingsViewSchemaID, 1, "txn-phase9-optional-finding-close",
		ValueChange("finding.state", Text("closed")),
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

	reopened := mustPatch(t, store, actor, finding.RecordID, workbook.FindingsViewSchemaID, 2, "txn-phase9-optional-finding-reopen",
		ValueChange("finding.state", Text("open")),
	)
	requireCoordinationCellValue(t, reopened.Payload["row"].(map[string]any), "finding.closed_at", nil)

	_, err = Patch(store, actor, finding.RecordID, workbook.FindingsViewSchemaID, 3, "txn-phase9-optional-finding-invalid-confidence",
		ValueChange("finding.confidence_score", optionalNumber(101)),
	)
	requireMutationValidation(t, err, "finding.confidence_score", "invalid_value")

	investigativeQuery := mustCreateRow(t, store, actor, incident.ID, workbook.InvestigativeQueriesViewSchemaID, "txn-phase9-optional-query", map[string]workbook.ValueChange{
		"investigative_query.platform":   Text("Kusto"),
		"investigative_query.purpose":    Text("Find suspicious PowerShell"),
		"investigative_query.query_text": Text("SecurityEvent | take 10"),
	}, nil, Time(0))
	queryRow := investigativeQuery.Payload["row"].(map[string]any)
	requireArtifactType(t, harness, investigativeQuery.RecordID, "investigative_query")
	requireCellNonEmpty(t, queryRow, "investigative_query.query_id")
	requireCoordinationCellValue(t, queryRow, "investigative_query.platform", "Kusto")
	requireCoordinationCellValue(t, queryRow, "investigative_query.created_by_user_id", actor.ID.String())
	requireCellNonEmpty(t, queryRow, "investigative_query.created_at")
	requireCellNonEmpty(t, queryRow, "investigative_query.created_day")

	forensicKeyword := mustCreateRow(t, store, actor, incident.ID, workbook.ForensicKeywordsViewSchemaID, "txn-phase9-optional-keyword", map[string]workbook.ValueChange{
		"forensic_keyword.pattern": Text("powershell.exe"),
		"forensic_keyword.reason":  Text("Interactive shell execution"),
	}, nil, Time(0))
	keywordRow := forensicKeyword.Payload["row"].(map[string]any)
	requireArtifactType(t, harness, forensicKeyword.RecordID, "forensic_keyword")
	requireCellNonEmpty(t, keywordRow, "forensic_keyword.keyword_id")
	requireCoordinationCellValue(t, keywordRow, "forensic_keyword.match_mode", "literal")
	requireCoordinationCellValue(t, keywordRow, "forensic_keyword.case_sensitive", false)

	keywordPatched := mustPatch(t, store, actor, forensicKeyword.RecordID, workbook.ForensicKeywordsViewSchemaID, 1, "txn-phase9-optional-keyword-patch",
		ValueChange("forensic_keyword.match_mode", Text("regex")),
		ValueChange("forensic_keyword.case_sensitive", optionalBool(true)),
	)
	keywordPatchedRow := keywordPatched.Payload["row"].(map[string]any)
	requireCoordinationCellValue(t, keywordPatchedRow, "forensic_keyword.match_mode", "regex")
	requireCoordinationCellValue(t, keywordPatchedRow, "forensic_keyword.case_sensitive", true)

	_, err = Patch(store, actor, forensicKeyword.RecordID, workbook.ForensicKeywordsViewSchemaID, 2, "txn-phase9-optional-keyword-invalid-mode",
		ValueChange("forensic_keyword.match_mode", Text("glob")),
	)
	requireMutationValidation(t, err, "forensic_keyword.match_mode", "invalid_value")
}

func TestOptionalStandardizedSurfacesProjectionQueryBehavior_Unit(t *testing.T) {
	harness := recordstoretest.StartStore(t, "phase9-optional-surfaces-query")
	store := workbook.NewStore(harness.DB)
	actor := recordstoretest.SeedLocalUserFlags(t, harness.DB, "optional-query@example.test", "Optional Query", "OptionalQuery1!", false, false, true)
	incident := recordstoretest.CreateIncidentInStore(t, harness.DB, actor, "txn-phase9-optional-query-incident", "IR-OPTIONAL-QUERY", "Workbook inspector optional query behavior")

	finding := mustCreateRow(t, store, actor, incident.ID, workbook.FindingsViewSchemaID, "txn-phase9-optional-query-finding-high", map[string]workbook.ValueChange{
		"finding.statement":        Text("Confirmed malware execution"),
		"finding.kind":             Text("hypothesis"),
		"finding.confidence_score": optionalNumber(91),
	}, nil, Time(0))
	mustCreateRow(t, store, actor, incident.ID, workbook.FindingsViewSchemaID, "txn-phase9-optional-query-finding-low", map[string]workbook.ValueChange{
		"finding.statement":        Text("Possible benign admin action"),
		"finding.kind":             Text("finding"),
		"finding.confidence_score": optionalNumber(12),
	}, nil, Time(0))
	requireProjectedRow(t, store, incident.ID, workbook.FindingsViewSchemaID, viewschema.QueryMeta{
		Filters: []viewschema.Filter{{FieldKey: "finding.confidence_band", Op: "eq", Arg: map[string]any{"value": "high"}}},
		Sort:    []viewschema.SortEntry{{FieldKey: "finding.confidence_score", Direction: "desc"}},
		GroupBy: StringPtr("finding.kind"),
	}, finding.RecordID, "finding.kind", "hypothesis")

	query := mustCreateRow(t, store, actor, incident.ID, workbook.InvestigativeQueriesViewSchemaID, "txn-phase9-optional-query-kusto", map[string]workbook.ValueChange{
		"investigative_query.platform":   Text("Kusto"),
		"investigative_query.purpose":    Text("Endpoint triage"),
		"investigative_query.query_text": Text("DeviceProcessEvents | take 20"),
	}, nil, Time(0))
	mustCreateRow(t, store, actor, incident.ID, workbook.InvestigativeQueriesViewSchemaID, "txn-phase9-optional-query-splunk", map[string]workbook.ValueChange{
		"investigative_query.platform":   Text("Splunk"),
		"investigative_query.purpose":    Text("Network triage"),
		"investigative_query.query_text": Text("index=proxy | head 20"),
	}, nil, Time(0))
	requireProjectedRow(t, store, incident.ID, workbook.InvestigativeQueriesViewSchemaID, viewschema.QueryMeta{
		Filters: []viewschema.Filter{{FieldKey: "investigative_query.platform", Op: "eq", Arg: map[string]any{"value": "KUSTO"}}},
		Sort:    []viewschema.SortEntry{{FieldKey: "investigative_query.created_day", Direction: "asc"}},
		GroupBy: StringPtr("investigative_query.platform"),
	}, query.RecordID, "investigative_query.platform", "Kusto")

	keyword := mustCreateRow(t, store, actor, incident.ID, workbook.ForensicKeywordsViewSchemaID, "txn-phase9-optional-query-keyword-regex", map[string]workbook.ValueChange{
		"forensic_keyword.pattern":        Text("(?i)powershell"),
		"forensic_keyword.reason":         Text("Shell execution"),
		"forensic_keyword.match_mode":     Text("regex"),
		"forensic_keyword.case_sensitive": optionalBool(true),
	}, nil, Time(0))
	mustCreateRow(t, store, actor, incident.ID, workbook.ForensicKeywordsViewSchemaID, "txn-phase9-optional-query-keyword-literal", map[string]workbook.ValueChange{
		"forensic_keyword.pattern": Text("cmd.exe"),
		"forensic_keyword.reason":  Text("Command shell"),
	}, nil, Time(0))
	requireProjectedRow(t, store, incident.ID, workbook.ForensicKeywordsViewSchemaID, viewschema.QueryMeta{
		Filters: []viewschema.Filter{{FieldKey: "forensic_keyword.case_sensitive", Op: "eq", Arg: map[string]any{"value": true}}},
		Sort:    []viewschema.SortEntry{{FieldKey: "forensic_keyword.match_mode", Direction: "desc"}},
		GroupBy: StringPtr("forensic_keyword.match_mode"),
	}, keyword.RecordID, "forensic_keyword.match_mode", "regex")
}

func TestFindingsConfidenceBandBoundaries_Unit(t *testing.T) {
	harness := recordstoretest.StartStore(t, "phase9-optional-findings-confidence-boundaries")
	store := workbook.NewStore(harness.DB)
	actor := recordstoretest.SeedLocalUserFlags(t, harness.DB, "optional-boundaries@example.test", "Optional Boundaries", "OptionalBoundaries1!", false, false, true)
	incident := recordstoretest.CreateIncidentInStore(t, harness.DB, actor, "txn-phase9-optional-boundaries-incident", "IR-OPTIONAL-BAND", "Workbook inspector optional finding confidence bands")

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
		values := map[string]workbook.ValueChange{
			"finding.statement": Text("Confidence boundary " + tc.key),
		}
		if tc.score != nil {
			values["finding.confidence_score"] = optionalNumber(*tc.score)
		}
		result := mustCreateRow(t, store, actor, incident.ID, workbook.FindingsViewSchemaID, "txn-phase9-optional-boundary-"+tc.key, values, nil, Time(time.Duration(index)*time.Minute))
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
			viewSchemaID: workbook.FindingsViewSchemaID,
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
			viewSchemaID: workbook.InvestigativeQueriesViewSchemaID,
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
			viewSchemaID: workbook.ForensicKeywordsViewSchemaID,
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

func requireOptionalSurfaceBandQuery(t testing.TB, store *workbook.Store, incidentID uuid.UUID, band string, want []uuid.UUID) {
	t.Helper()
	rows, err := store.QueryRows(context.Background(), incidentID, workbook.FindingsViewSchemaID, viewschema.QueryMeta{
		Filters: []viewschema.Filter{{FieldKey: "finding.confidence_band", Op: "eq", Arg: map[string]any{"value": band}}},
		Sort: []viewschema.SortEntry{
			{FieldKey: "finding.confidence_score", Direction: "asc"},
			{FieldKey: "record_id", Direction: "asc"},
		},
		GroupBy: StringPtr("finding.confidence_band"),
	})
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

func mustPatch(t testing.TB, store *workbook.Store, actor authn.UserRecord, recordID uuid.UUID, viewSchemaID string, baseRowVersion int64, clientTxnID string, changes ...workbook.PatchChange) workbook.MutationResult {
	t.Helper()
	result, err := Patch(store, actor, recordID, viewSchemaID, baseRowVersion, clientTxnID, changes...)
	if err != nil {
		t.Fatalf("patch %s: %v", clientTxnID, err)
	}
	return result
}

func Patch(store *workbook.Store, actor authn.UserRecord, recordID uuid.UUID, viewSchemaID string, baseRowVersion int64, clientTxnID string, changes ...workbook.PatchChange) (workbook.MutationResult, error) {
	return store.PatchWorkbookRow(context.Background(), actor, recordID, workbook.PatchRequest{
		ViewSchemaID:   viewSchemaID,
		BaseRowVersion: baseRowVersion,
		ClientTxnID:    clientTxnID,
		Changes:        changes,
	}, []byte(clientTxnID), "req-"+clientTxnID, Time(0))
}

func ValueChange(fieldKey string, value workbook.ValueChange) workbook.PatchChange {
	return workbook.PatchChange{FieldKey: fieldKey, Value: &value}
}

func CollectionChange(fieldKey string, value workbook.CollectionActionPayload) workbook.PatchChange {
	return workbook.PatchChange{FieldKey: fieldKey, Collection: &value}
}

func Collection(actions ...workbook.CollectionAction) workbook.CollectionActionPayload {
	return workbook.CollectionActionPayload{Actions: actions}
}

func addOptionalSurfaceRecordRef(recordID uuid.UUID) workbook.CollectionAction {
	return workbook.CollectionAction{Op: "add_record_ref", LinkedRecordID: &recordID}
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

func optionalNumber(value int64) workbook.ValueChange {
	return workbook.ValueChange{Kind: "number", Number: &value}
}

func optionalBool(value bool) workbook.ValueChange {
	return workbook.ValueChange{Kind: "bool", Bool: &value}
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

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

func TestPhase9_U_9_09_OptionalStandardizedSurfacesStoreBehavior(t *testing.T) {
	requirePhase9U909OptionalSurfaceResources(t)

	ctx := context.Background()
	harness := recordstoretest.StartStore(t, "phase9-optional-surfaces-store")
	store := workbook.NewStore(harness.DB)
	actor := recordstoretest.SeedLocalUserFlags(t, harness.DB, "optional-surfaces@example.test", "Optional Surfaces", "OptionalSurfaces1!", false, false, true)
	incident := recordstoretest.CreateIncidentInStore(t, harness.DB, actor, "txn-phase9-optional-incident", "IR-OPTIONAL", "Phase 9 optional surfaces")

	before := countOptionalSurfaceDurableState(t, harness.DB, incident.ID)
	_, err := store.CreateWorkbookRow(ctx, actor, incident.ID, workbook.CreateRequest{
		ViewSchemaID: workbook.FindingsViewSchemaID,
		ClientTxnID:  "txn-phase9-optional-finding-reject",
		Values: map[string]workbook.ValueChange{
			"finding.kind": sprint7Text("finding"),
		},
	}, []byte("txn-phase9-optional-finding-reject"), "req-phase9-optional-finding-reject", sprint7Time(0))
	requireSprint7MutationValidation(t, err, "finding.statement", "missing_required_field")
	requireOptionalSurfaceDurableState(t, harness.DB, incident.ID, before, "rejected finding create")

	support := mustCreateSprint7Row(t, store, actor, incident.ID, workbook.NotesViewSchemaID, "txn-phase9-optional-support-note", map[string]workbook.ValueChange{
		"note.title": sprint7Text("Supporting note"),
	}, nil, sprint7Time(0))
	contradiction := mustCreateSprint7Row(t, store, actor, incident.ID, workbook.NotesViewSchemaID, "txn-phase9-optional-contradiction-note", map[string]workbook.ValueChange{
		"note.title": sprint7Text("Contradictory note"),
	}, nil, sprint7Time(0))

	finding := mustCreateSprint7Row(t, store, actor, incident.ID, workbook.FindingsViewSchemaID, "txn-phase9-optional-finding", map[string]workbook.ValueChange{
		"finding.statement":        sprint7Text("EDR shows suspicious process execution"),
		"finding.kind":             sprint7Text("hypothesis"),
		"finding.confidence_score": optionalNumber(72),
	}, nil, sprint7Time(0))
	findingRow := finding.Payload["row"].(map[string]any)
	requireSprint7ArtifactType(t, harness, finding.RecordID, "finding")
	requireSprint7CellValue(t, findingRow, "finding.statement", "EDR shows suspicious process execution")
	requireSprint7CellValue(t, findingRow, "finding.kind", "hypothesis")
	requireSprint7CellValue(t, findingRow, "finding.state", "open")
	requireSprint7CellValue(t, findingRow, "finding.owner_user_id", actor.ID.String())
	requireSprint6CellNumericValue(t, findingRow, "finding.confidence_score", 72)
	requireSprint7CellValue(t, findingRow, "finding.confidence_band", "high")
	requireSprint7CellValue(t, findingRow, "finding.closed_at", nil)
	requireSprint7CollectionItemCount(t, findingRow, "finding.supporting_refs", 0)
	requireSprint7CollectionItemCount(t, findingRow, "finding.contradictory_refs", 0)

	closed := mustSprint6Patch(t, store, actor, finding.RecordID, workbook.FindingsViewSchemaID, 1, "txn-phase9-optional-finding-close",
		sprint6ValueChange("finding.state", sprint7Text("closed")),
		sprint6CollectionChange("finding.supporting_refs", sprint6Collection(addSprint6RecordRef(support.RecordID))),
		sprint6CollectionChange("finding.contradictory_refs", sprint6Collection(addSprint6RecordRef(contradiction.RecordID))),
	)
	closedRow := closed.Payload["row"].(map[string]any)
	requireSprint7CellValue(t, closedRow, "finding.state", "closed")
	requireSprint7CellNonEmpty(t, closedRow, "finding.closed_at")
	requireSprint7CollectionItemCount(t, closedRow, "finding.supporting_refs", 1)
	requireSprint7CollectionItemCount(t, closedRow, "finding.contradictory_refs", 1)
	requireSprint7ManualReferenceLink(t, harness, finding.RecordID, support.RecordID, "finding.supporting_refs", "supported_by")
	requireSprint7ManualReferenceLink(t, harness, finding.RecordID, contradiction.RecordID, "finding.contradictory_refs", "references_record")

	reopened := mustSprint6Patch(t, store, actor, finding.RecordID, workbook.FindingsViewSchemaID, 2, "txn-phase9-optional-finding-reopen",
		sprint6ValueChange("finding.state", sprint7Text("open")),
	)
	requireSprint7CellValue(t, reopened.Payload["row"].(map[string]any), "finding.closed_at", nil)

	_, err = sprint6Patch(store, actor, finding.RecordID, workbook.FindingsViewSchemaID, 3, "txn-phase9-optional-finding-invalid-confidence",
		sprint6ValueChange("finding.confidence_score", optionalNumber(101)),
	)
	requireSprint7MutationValidation(t, err, "finding.confidence_score", "invalid_value")

	investigativeQuery := mustCreateSprint7Row(t, store, actor, incident.ID, workbook.InvestigativeQueriesViewSchemaID, "txn-phase9-optional-query", map[string]workbook.ValueChange{
		"investigative_query.platform":   sprint7Text("Kusto"),
		"investigative_query.purpose":    sprint7Text("Find suspicious PowerShell"),
		"investigative_query.query_text": sprint7Text("SecurityEvent | take 10"),
	}, nil, sprint7Time(0))
	queryRow := investigativeQuery.Payload["row"].(map[string]any)
	requireSprint7ArtifactType(t, harness, investigativeQuery.RecordID, "investigative_query")
	requireSprint7CellNonEmpty(t, queryRow, "investigative_query.query_id")
	requireSprint7CellValue(t, queryRow, "investigative_query.platform", "Kusto")
	requireSprint7CellValue(t, queryRow, "investigative_query.created_by_user_id", actor.ID.String())
	requireSprint7CellNonEmpty(t, queryRow, "investigative_query.created_at")
	requireSprint7CellNonEmpty(t, queryRow, "investigative_query.created_day")

	forensicKeyword := mustCreateSprint7Row(t, store, actor, incident.ID, workbook.ForensicKeywordsViewSchemaID, "txn-phase9-optional-keyword", map[string]workbook.ValueChange{
		"forensic_keyword.pattern": sprint7Text("powershell.exe"),
		"forensic_keyword.reason":  sprint7Text("Interactive shell execution"),
	}, nil, sprint7Time(0))
	keywordRow := forensicKeyword.Payload["row"].(map[string]any)
	requireSprint7ArtifactType(t, harness, forensicKeyword.RecordID, "forensic_keyword")
	requireSprint7CellNonEmpty(t, keywordRow, "forensic_keyword.keyword_id")
	requireSprint7CellValue(t, keywordRow, "forensic_keyword.match_mode", "literal")
	requireSprint7CellValue(t, keywordRow, "forensic_keyword.case_sensitive", false)

	keywordPatched := mustSprint6Patch(t, store, actor, forensicKeyword.RecordID, workbook.ForensicKeywordsViewSchemaID, 1, "txn-phase9-optional-keyword-patch",
		sprint6ValueChange("forensic_keyword.match_mode", sprint7Text("regex")),
		sprint6ValueChange("forensic_keyword.case_sensitive", optionalBool(true)),
	)
	keywordPatchedRow := keywordPatched.Payload["row"].(map[string]any)
	requireSprint7CellValue(t, keywordPatchedRow, "forensic_keyword.match_mode", "regex")
	requireSprint7CellValue(t, keywordPatchedRow, "forensic_keyword.case_sensitive", true)

	_, err = sprint6Patch(store, actor, forensicKeyword.RecordID, workbook.ForensicKeywordsViewSchemaID, 2, "txn-phase9-optional-keyword-invalid-mode",
		sprint6ValueChange("forensic_keyword.match_mode", sprint7Text("glob")),
	)
	requireSprint7MutationValidation(t, err, "forensic_keyword.match_mode", "invalid_value")
}

func TestPhase9_U_9_09_OptionalStandardizedSurfacesProjectionQueryBehavior(t *testing.T) {
	harness := recordstoretest.StartStore(t, "phase9-optional-surfaces-query")
	store := workbook.NewStore(harness.DB)
	actor := recordstoretest.SeedLocalUserFlags(t, harness.DB, "optional-query@example.test", "Optional Query", "OptionalQuery1!", false, false, true)
	incident := recordstoretest.CreateIncidentInStore(t, harness.DB, actor, "txn-phase9-optional-query-incident", "IR-OPTIONAL-QUERY", "Phase 9 optional query behavior")

	finding := mustCreateSprint7Row(t, store, actor, incident.ID, workbook.FindingsViewSchemaID, "txn-phase9-optional-query-finding-high", map[string]workbook.ValueChange{
		"finding.statement":        sprint7Text("Confirmed malware execution"),
		"finding.kind":             sprint7Text("hypothesis"),
		"finding.confidence_score": optionalNumber(91),
	}, nil, sprint7Time(0))
	mustCreateSprint7Row(t, store, actor, incident.ID, workbook.FindingsViewSchemaID, "txn-phase9-optional-query-finding-low", map[string]workbook.ValueChange{
		"finding.statement":        sprint7Text("Possible benign admin action"),
		"finding.kind":             sprint7Text("finding"),
		"finding.confidence_score": optionalNumber(12),
	}, nil, sprint7Time(0))
	requireSprint7ProjectedRow(t, store, incident.ID, workbook.FindingsViewSchemaID, viewschema.QueryMeta{
		Filters: []viewschema.Filter{{FieldKey: "finding.confidence_band", Op: "eq", Arg: map[string]any{"value": "high"}}},
		Sort:    []viewschema.SortEntry{{FieldKey: "finding.confidence_score", Direction: "desc"}},
		GroupBy: sprint7StringPtr("finding.kind"),
	}, finding.RecordID, "finding.kind", "hypothesis")

	query := mustCreateSprint7Row(t, store, actor, incident.ID, workbook.InvestigativeQueriesViewSchemaID, "txn-phase9-optional-query-kusto", map[string]workbook.ValueChange{
		"investigative_query.platform":   sprint7Text("Kusto"),
		"investigative_query.purpose":    sprint7Text("Endpoint triage"),
		"investigative_query.query_text": sprint7Text("DeviceProcessEvents | take 20"),
	}, nil, sprint7Time(0))
	mustCreateSprint7Row(t, store, actor, incident.ID, workbook.InvestigativeQueriesViewSchemaID, "txn-phase9-optional-query-splunk", map[string]workbook.ValueChange{
		"investigative_query.platform":   sprint7Text("Splunk"),
		"investigative_query.purpose":    sprint7Text("Network triage"),
		"investigative_query.query_text": sprint7Text("index=proxy | head 20"),
	}, nil, sprint7Time(0))
	requireSprint7ProjectedRow(t, store, incident.ID, workbook.InvestigativeQueriesViewSchemaID, viewschema.QueryMeta{
		Filters: []viewschema.Filter{{FieldKey: "investigative_query.platform", Op: "eq", Arg: map[string]any{"value": "KUSTO"}}},
		Sort:    []viewschema.SortEntry{{FieldKey: "investigative_query.created_day", Direction: "asc"}},
		GroupBy: sprint7StringPtr("investigative_query.platform"),
	}, query.RecordID, "investigative_query.platform", "Kusto")

	keyword := mustCreateSprint7Row(t, store, actor, incident.ID, workbook.ForensicKeywordsViewSchemaID, "txn-phase9-optional-query-keyword-regex", map[string]workbook.ValueChange{
		"forensic_keyword.pattern":        sprint7Text("(?i)powershell"),
		"forensic_keyword.reason":         sprint7Text("Shell execution"),
		"forensic_keyword.match_mode":     sprint7Text("regex"),
		"forensic_keyword.case_sensitive": optionalBool(true),
	}, nil, sprint7Time(0))
	mustCreateSprint7Row(t, store, actor, incident.ID, workbook.ForensicKeywordsViewSchemaID, "txn-phase9-optional-query-keyword-literal", map[string]workbook.ValueChange{
		"forensic_keyword.pattern": sprint7Text("cmd.exe"),
		"forensic_keyword.reason":  sprint7Text("Command shell"),
	}, nil, sprint7Time(0))
	requireSprint7ProjectedRow(t, store, incident.ID, workbook.ForensicKeywordsViewSchemaID, viewschema.QueryMeta{
		Filters: []viewschema.Filter{{FieldKey: "forensic_keyword.case_sensitive", Op: "eq", Arg: map[string]any{"value": true}}},
		Sort:    []viewschema.SortEntry{{FieldKey: "forensic_keyword.match_mode", Direction: "desc"}},
		GroupBy: sprint7StringPtr("forensic_keyword.match_mode"),
	}, keyword.RecordID, "forensic_keyword.match_mode", "regex")
}

func TestPhase9_U_9_09_FindingsConfidenceBandBoundaries(t *testing.T) {
	harness := recordstoretest.StartStore(t, "phase9-optional-findings-confidence-boundaries")
	store := workbook.NewStore(harness.DB)
	actor := recordstoretest.SeedLocalUserFlags(t, harness.DB, "optional-boundaries@example.test", "Optional Boundaries", "OptionalBoundaries1!", false, false, true)
	incident := recordstoretest.CreateIncidentInStore(t, harness.DB, actor, "txn-phase9-optional-boundaries-incident", "IR-OPTIONAL-BAND", "Phase 9 optional finding confidence bands")

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
			"finding.statement": sprint7Text("Confidence boundary " + tc.key),
		}
		if tc.score != nil {
			values["finding.confidence_score"] = optionalNumber(*tc.score)
		}
		result := mustCreateSprint7Row(t, store, actor, incident.ID, workbook.FindingsViewSchemaID, "txn-phase9-optional-boundary-"+tc.key, values, nil, sprint7Time(time.Duration(index)*time.Minute))
		row := result.Payload["row"].(map[string]any)
		if tc.score == nil {
			requireSprint7CellValue(t, row, "finding.confidence_score", nil)
		} else {
			requireSprint6CellNumericValue(t, row, "finding.confidence_score", *tc.score)
		}
		requireSprint7CellValue(t, row, "finding.confidence_band", tc.wantBand)
		createdByBand[tc.wantBand] = append(createdByBand[tc.wantBand], result.RecordID)
	}

	requirePhase9U909BandQuery(t, store, incident.ID, "unset", createdByBand["unset"])
	requirePhase9U909BandQuery(t, store, incident.ID, "low", createdByBand["low"])
	requirePhase9U909BandQuery(t, store, incident.ID, "medium", createdByBand["medium"])
	requirePhase9U909BandQuery(t, store, incident.ID, "high", createdByBand["high"])
}

func requirePhase9U909OptionalSurfaceResources(t testing.TB) {
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

func requirePhase9U909BandQuery(t testing.TB, store *workbook.Store, incidentID uuid.UUID, band string, want []uuid.UUID) {
	t.Helper()
	rows, err := store.QueryRows(context.Background(), incidentID, workbook.FindingsViewSchemaID, viewschema.QueryMeta{
		Filters: []viewschema.Filter{{FieldKey: "finding.confidence_band", Op: "eq", Arg: map[string]any{"value": band}}},
		Sort: []viewschema.SortEntry{
			{FieldKey: "finding.confidence_score", Direction: "asc"},
			{FieldKey: "record_id", Direction: "asc"},
		},
		GroupBy: sprint7StringPtr("finding.confidence_band"),
	})
	if err != nil {
		t.Fatalf("query finding confidence band %s: %v", band, err)
	}
	requireSprint7RecordOrder(t, rows, want)
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

func mustSprint6Patch(t testing.TB, store *workbook.Store, actor authn.UserRecord, recordID uuid.UUID, viewSchemaID string, baseRowVersion int64, clientTxnID string, changes ...workbook.PatchChange) workbook.MutationResult {
	t.Helper()
	result, err := sprint6Patch(store, actor, recordID, viewSchemaID, baseRowVersion, clientTxnID, changes...)
	if err != nil {
		t.Fatalf("patch %s: %v", clientTxnID, err)
	}
	return result
}

func sprint6Patch(store *workbook.Store, actor authn.UserRecord, recordID uuid.UUID, viewSchemaID string, baseRowVersion int64, clientTxnID string, changes ...workbook.PatchChange) (workbook.MutationResult, error) {
	return store.PatchWorkbookRow(context.Background(), actor, recordID, workbook.PatchRequest{
		ViewSchemaID:   viewSchemaID,
		BaseRowVersion: baseRowVersion,
		ClientTxnID:    clientTxnID,
		Changes:        changes,
	}, []byte(clientTxnID), "req-"+clientTxnID, sprint7Time(0))
}

func sprint6ValueChange(fieldKey string, value workbook.ValueChange) workbook.PatchChange {
	return workbook.PatchChange{FieldKey: fieldKey, Value: &value}
}

func sprint6CollectionChange(fieldKey string, value workbook.CollectionActionPayload) workbook.PatchChange {
	return workbook.PatchChange{FieldKey: fieldKey, Collection: &value}
}

func sprint6Collection(actions ...workbook.CollectionAction) workbook.CollectionActionPayload {
	return workbook.CollectionActionPayload{Actions: actions}
}

func addSprint6RecordRef(recordID uuid.UUID) workbook.CollectionAction {
	return workbook.CollectionAction{Op: "add_record_ref", LinkedRecordID: &recordID}
}

func requireSprint6CellNumericValue(t testing.TB, row map[string]any, fieldKey string, want int64) {
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

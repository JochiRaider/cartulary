package reporting_test

import (
	"context"
	"encoding/json"
	"net/http"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/modules/auth/testsupport/flowtest"
	evidencereporting "github.com/JochiRaider/cartulary/internal/modules/evidence/reportingprovider"
	"github.com/JochiRaider/cartulary/internal/modules/incidents/testsupport/scenariotest"
	linkreporting "github.com/JochiRaider/cartulary/internal/modules/links/reportingprovider"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/testutil/appsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/httptestx"
)

func TestReportingEvidenceProviderRedaction_Integration(t *testing.T) {
	runtime := appsupport.StartRuntime(t)
	harness := runtime.StartDefaultServer(t, "reporting-evidence-provider-redaction")
	admin, actorID := flowtest.ProvisionBootstrapAdmin(t, harness.Server.HTTP.URL)
	incident := scenariotest.CreateIncident(t, harness.Server, admin, map[string]any{
		"client_txn_id": "txn-reporting-evidence-provider-incident",
		"incident_key":  "RPT-EVIDENCE-PROVIDER",
		"title":         "Reporting Evidence provider closure",
	})
	incidentID := uuid.MustParse(incident["incident_id"].(string))

	// This schema mutation is intentionally local to the dedicated test database.
	// It proves that future Evidence columns remain private by construction.
	if _, err := harness.Pool.Exec(context.Background(), `ALTER TABLE evidence ADD COLUMN future_private text`); err != nil {
		t.Fatalf("add future Evidence fixture column: %v", err)
	}

	eligibleID := uuid.MustParse("00000000-0000-0000-0000-000000700001")
	deletedID := uuid.MustParse("00000000-0000-0000-0000-000000700002")
	sourceID := uuid.MustParse("00000000-0000-0000-0000-000000700003")
	fixedAt := time.Date(2026, time.July, 7, 12, 34, 56, 0, time.UTC)
	ctx := context.Background()
	if _, err := harness.Pool.Exec(ctx, `
INSERT INTO records (
    record_id, incident_id, record_type, created_by_user_id, updated_by_user_id,
    created_at, updated_at, deleted_at, deleted_by_user_id
) VALUES
    ($1, $4, 'evidence', $5, $5, $6, $6, NULL, NULL),
    ($2, $4, 'evidence', $5, $5, $6, $6, $6, $5),
    ($3, $4, 'task_request', $5, $5, $6, $6, NULL, NULL)
`, eligibleID, deletedID, sourceID, incidentID, actorID, fixedAt); err != nil {
		t.Fatalf("seed Evidence reporting record envelopes: %v", err)
	}
	if _, err := harness.Pool.Exec(ctx, `
INSERT INTO evidence (
    record_id, incident_id, title, lifecycle_state, requested_at, received_at,
    storage_ref, blob_hash, collector_party_text, collector_party_id,
    source_party_text, source_party_id, upload_state, created_at, updated_at,
    future_private
) VALUES
    ($1, $3, NULL, 'received', NULL, $4,
     'forbidden-storage-ref', 'forbidden-blob-hash', NULL, NULL,
     'Allowed source party', NULL, 'complete', $4, $4,
     'future-private-value-must-not-export'),
    ($2, $3, 'Deleted Evidence must not export', 'received', $4, $4,
     'deleted-forbidden-storage-ref', 'deleted-forbidden-blob-hash', 'Deleted collector', NULL,
     'Deleted source', NULL, 'complete', $4, $4,
     'deleted-future-private-value')
`, eligibleID, deletedID, incidentID, fixedAt); err != nil {
		t.Fatalf("seed Evidence reporting rows: %v", err)
	}
	if _, err := harness.Pool.Exec(ctx, `
INSERT INTO record_links (
    incident_id, src_record_id, dst_record_id, link_type, provenance,
    owner_user_id, decided_at, created_at, created_by_user_id
) VALUES ($1, $2, $3, 'supported_by', 'manual', $4, $5, $5, $4)
`, incidentID, sourceID, eligibleID, actorID, fixedAt); err != nil {
		t.Fatalf("seed logical Evidence support link: %v", err)
	}

	tx, err := harness.Pool.Begin(ctx)
	if err != nil {
		t.Fatalf("begin Reporting provider snapshot: %v", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	logicalTargets, err := evidencereporting.CollectLogicalSupportTargetsTx(ctx, tx, incidentID)
	if err != nil {
		t.Fatalf("collect logical Evidence support targets: %v", err)
	}
	supportRefs, err := linkreporting.CollectSupportRefsTx(ctx, tx, incidentID, logicalTargets)
	if err != nil {
		t.Fatalf("collect Reporting support refs: %v", err)
	}
	wantSupportRef := "/evidence/" + eligibleID.String()
	if got := supportRefs[sourceID.String()]; !slices.Equal(got, []string{wantSupportRef}) {
		t.Fatalf("logical Evidence support refs = %v, want [%s]", got, wantSupportRef)
	}
	first, err := evidencereporting.CollectFactsTx(ctx, tx, incidentID, supportRefs)
	if err != nil {
		t.Fatalf("collect Evidence Reporting facts: %v", err)
	}
	second, err := evidencereporting.CollectFactsTx(ctx, tx, incidentID, supportRefs)
	if err != nil {
		t.Fatalf("collect Evidence Reporting facts again: %v", err)
	}
	firstJSON, err := json.Marshal(first)
	if err != nil {
		t.Fatalf("encode first Evidence Reporting facts: %v", err)
	}
	secondJSON, err := json.Marshal(second)
	if err != nil {
		t.Fatalf("encode second Evidence Reporting facts: %v", err)
	}
	if string(firstJSON) != string(secondJSON) {
		t.Fatalf("Evidence Reporting facts are not deterministic:\nfirst:  %s\nsecond: %s", firstJSON, secondJSON)
	}
	if len(first.FieldFacts) != 1 {
		t.Fatalf("Evidence Reporting field fact count = %d, want 1 eligible row: %#v", len(first.FieldFacts), first.FieldFacts)
	}
	fact := first.FieldFacts[0]
	if fact.Path != "/evidence/"+eligibleID.String() || fact.SourceFamily != "evidence" || fact.ContentClass != "source_evidence" {
		t.Fatalf("unexpected Evidence Reporting fact identity: %#v", fact)
	}
	value, ok := fact.Value.(map[string]any)
	if !ok {
		t.Fatalf("Evidence Reporting value type = %T, want map[string]any", fact.Value)
	}
	gotKeys := make([]string, 0, len(value))
	for key := range value {
		gotKeys = append(gotKeys, key)
	}
	slices.Sort(gotKeys)
	wantKeys := []string{
		"collector_party_id",
		"collector_party_text",
		"created_at",
		"lifecycle_state",
		"received_at",
		"record_id",
		"requested_at",
		"source_party_id",
		"source_party_text",
		"title",
		"updated_at",
		"upload_state",
	}
	if !slices.Equal(gotKeys, wantKeys) {
		t.Fatalf("Evidence Reporting value keys = %v, want exact allowlist %v", gotKeys, wantKeys)
	}
	if value["record_id"] != eligibleID.String() || value["title"] != nil || value["requested_at"] != nil || value["collector_party_text"] != nil {
		t.Fatalf("Evidence Reporting value did not preserve identity/null posture: %#v", value)
	}
	for _, timestampKey := range []string{"received_at", "created_at", "updated_at"} {
		got, ok := value[timestampKey].(string)
		if !ok {
			t.Fatalf("Evidence Reporting %s = %#v, want timestamp string", timestampKey, value[timestampKey])
		}
		parsed, err := time.Parse(time.RFC3339Nano, got)
		if err != nil || !parsed.Equal(fixedAt) {
			t.Fatalf("Evidence Reporting %s = %q (%v), want %s", timestampKey, got, err, fixedAt.Format(time.RFC3339Nano))
		}
	}
	encoded := string(firstJSON)
	for _, forbidden := range []string{
		"incident_id",
		"blob_hash",
		"storage_ref",
		"object_blob_id",
		"future_private",
		"forbidden-storage-ref",
		"forbidden-blob-hash",
		"future-private-value-must-not-export",
		"Deleted Evidence must not export",
	} {
		if strings.Contains(encoded, forbidden) {
			t.Fatalf("Evidence Reporting facts exposed forbidden value %q: %s", forbidden, encoded)
		}
	}
	if err := tx.Rollback(ctx); err != nil {
		t.Fatalf("close direct Reporting provider snapshot: %v", err)
	}

	createSnapshot := httptestx.DoJSON(
		t,
		http.MethodPost,
		harness.Server.HTTP.URL+"/api/v1/snapshots",
		map[string]any{
			"incident_id":   incidentID.String(),
			"client_txn_id": "txn-reporting-evidence-provider-snapshot",
		},
		httptestx.WithCookies(admin.SessionCookie, admin.CSRFCookie),
		httptestx.WithHeader(authn.CSRFHeaderName, admin.CSRFCookie.Value),
	)
	snapshotJob := httptestx.RequireSuccessEnvelope(t, createSnapshot, http.StatusAccepted)["data"].(map[string]any)
	snapshotID := requireSucceededJobResourceID(t, harness, admin, snapshotJob, "snapshot")
	exportModel := requireSnapshotExportModel(t, harness.DB, snapshotID)
	exportJSON, err := json.Marshal(exportModel)
	if err != nil {
		t.Fatalf("encode Evidence Reporting export model: %v", err)
	}
	exportText := string(exportJSON)
	if !strings.Contains(exportText, "evidence:"+eligibleID.String()) {
		t.Fatalf("Reporting export model omitted logical Evidence support identity: %s", exportText)
	}
	for _, forbidden := range []string{
		"future_private",
		"future-private-value-must-not-export",
		"forbidden-storage-ref",
		"forbidden-blob-hash",
		"Deleted Evidence must not export",
	} {
		if strings.Contains(exportText, forbidden) {
			t.Fatalf("Reporting export model exposed forbidden Evidence value %q: %s", forbidden, exportText)
		}
	}
}

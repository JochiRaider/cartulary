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

	"github.com/JochiRaider/cartulary/internal/app/reportingassembly"
	"github.com/JochiRaider/cartulary/internal/modules/auth/testsupport/flowtest"
	evidencereporting "github.com/JochiRaider/cartulary/internal/modules/evidence/reportingprovider"
	"github.com/JochiRaider/cartulary/internal/modules/incidents/testsupport/scenariotest"
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
	supportedLinkID := uuid.MustParse("00000000-0000-0000-0000-000000700004")
	attachedLinkID := uuid.MustParse("00000000-0000-0000-0000-000000700005")
	derivedLinkID := uuid.MustParse("00000000-0000-0000-0000-000000700006")
	tagID := uuid.MustParse("00000000-0000-0000-0000-000000700007")
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
    record_link_id, incident_id, src_record_id, dst_record_id, link_type,
    field_key, provenance, owner_user_id, decided_at, created_at,
    created_by_user_id
) VALUES
    ($1, $4, $5, $6, 'supported_by', 'fixture.support_refs', 'manual', $7, $8, $8, $7),
    ($2, $4, $5, $6, 'attached_evidence', NULL, 'manual', $7, $8, $8, $7),
    ($3, $4, $5, $6, 'derived_from', NULL, 'manual', $7, $8, $8, $7)
`, supportedLinkID, attachedLinkID, derivedLinkID, incidentID, sourceID, eligibleID, actorID, fixedAt); err != nil {
		t.Fatalf("seed logical Evidence support link: %v", err)
	}
	if _, err := harness.Pool.Exec(ctx, `
INSERT INTO record_tags (
    record_tag_id, incident_id, record_id, tag_name, normalized_tag_name,
    created_by_user_id, created_at, updated_at
) VALUES ($1, $2, $3, 'Priority', 'priority', $4, $5, $5)
`, tagID, incidentID, sourceID, actorID, fixedAt); err != nil {
		t.Fatalf("seed Links Reporting tag: %v", err)
	}

	tx, err := harness.Pool.Begin(ctx)
	if err != nil {
		t.Fatalf("begin Reporting provider snapshot: %v", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	linksProvider := reportingassembly.NewLinksProvider()
	supportRefs, err := linksProvider.CollectSupportRefsTx(ctx, tx, incidentID)
	if err != nil {
		t.Fatalf("collect adapted Reporting support refs: %v", err)
	}
	wantSupportRef := "/evidence/" + eligibleID.String()
	if got := supportRefs[sourceID.String()]; !slices.Equal(got, []string{wantSupportRef, wantSupportRef}) {
		t.Fatalf("logical Evidence support refs = %v, want repeated [%s %s]", got, wantSupportRef, wantSupportRef)
	}
	linkFacts, err := linksProvider.CollectFactsTx(ctx, tx, incidentID, supportRefs)
	if err != nil {
		t.Fatalf("collect adapted Links Reporting facts: %v", err)
	}
	if linkFacts.ProviderKey != "links" || len(linkFacts.FieldFacts) != 4 {
		t.Fatalf("Links Reporting output = provider %q fields %d, want links/4: %#v", linkFacts.ProviderKey, len(linkFacts.FieldFacts), linkFacts)
	}
	wantPaths := []string{
		"/relationships/" + supportedLinkID.String(),
		"/relationships/" + attachedLinkID.String(),
		"/relationships/" + derivedLinkID.String(),
		"/tags/" + tagID.String(),
	}
	slices.Sort(wantPaths)
	gotPaths := make([]string, 0, len(linkFacts.FieldFacts))
	for _, field := range linkFacts.FieldFacts {
		gotPaths = append(gotPaths, field.Path)
	}
	if !slices.Equal(gotPaths, wantPaths) {
		t.Fatalf("Links Reporting paths = %v, want %v", gotPaths, wantPaths)
	}
	for _, field := range linkFacts.FieldFacts {
		value, ok := field.Value.(map[string]any)
		if !ok {
			t.Fatalf("Links Reporting value type = %T, want map[string]any", field.Value)
		}
		if _, exists := value["incident_id"]; exists {
			t.Fatalf("Links Reporting value leaked incident_id: %#v", value)
		}
		for _, forbidden := range []string{"note", "description", "comment"} {
			if _, exists := value[forbidden]; exists {
				t.Fatalf("Links Reporting value exposed narrative member %q: %#v", forbidden, value)
			}
		}
		if value["deleted_at"] != nil || value["deleted_by_user_id"] != nil {
			t.Fatalf("active Links Reporting value lost explicit tombstone nulls: %#v", value)
		}
		if strings.HasPrefix(field.Path, "/relationships/") {
			if field.SourceFamily != "record_link" || field.ContentClass != "derived_analytic" {
				t.Fatalf("unexpected link Reporting classification: %#v", field)
			}
		} else if field.SourceFamily != "record_tag" || field.ContentClass != "derived_analytic" {
			t.Fatalf("unexpected tag Reporting classification: %#v", field)
		}
	}
	for _, field := range linkFacts.FieldFacts {
		if field.Path != "/relationships/"+supportedLinkID.String() {
			continue
		}
		value := field.Value.(map[string]any)
		if value["field_key"] != "fixture.support_refs" || value["record_link_id"] != supportedLinkID.String() {
			t.Fatalf("field-aware Links Reporting value = %#v", value)
		}
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
	if _, err := tx.Exec(ctx, `UPDATE records SET deleted_at = $2, deleted_by_user_id = $3 WHERE record_id = $1`, eligibleID, fixedAt.Add(time.Minute), actorID); err != nil {
		t.Fatalf("soft-delete Reporting link endpoint in characterization transaction: %v", err)
	}
	inactiveSupportRefs, err := linksProvider.CollectSupportRefsTx(ctx, tx, incidentID)
	if err != nil {
		t.Fatalf("collect support refs after destination deletion: %v", err)
	}
	if got := inactiveSupportRefs[sourceID.String()]; len(got) != 0 {
		t.Fatalf("destination-deleted support refs = %v, want none", got)
	}
	inactiveLinkFacts, err := linksProvider.CollectFactsTx(ctx, tx, incidentID, inactiveSupportRefs)
	if err != nil {
		t.Fatalf("collect link facts after destination deletion: %v", err)
	}
	if len(inactiveLinkFacts.FieldFacts) != 1 || inactiveLinkFacts.FieldFacts[0].Path != "/tags/"+tagID.String() {
		t.Fatalf("destination-deleted Links fields = %#v, want only active source tag", inactiveLinkFacts.FieldFacts)
	}
	if _, err := tx.Exec(ctx, `UPDATE records SET deleted_at = NULL, deleted_by_user_id = NULL WHERE record_id = $1`, eligibleID); err != nil {
		t.Fatalf("reactivate Reporting link destination: %v", err)
	}
	if _, err := tx.Exec(ctx, `UPDATE records SET deleted_at = $2, deleted_by_user_id = $3 WHERE record_id = $1`, sourceID, fixedAt.Add(2*time.Minute), actorID); err != nil {
		t.Fatalf("soft-delete Reporting link source: %v", err)
	}
	sourceDeletedSupportRefs, err := linksProvider.CollectSupportRefsTx(ctx, tx, incidentID)
	if err != nil {
		t.Fatalf("collect support refs after source deletion: %v", err)
	}
	if got := sourceDeletedSupportRefs[sourceID.String()]; len(got) != 0 {
		t.Fatalf("source-deleted support refs = %v, want none", got)
	}
	sourceDeletedFacts, err := linksProvider.CollectFactsTx(ctx, tx, incidentID, sourceDeletedSupportRefs)
	if err != nil {
		t.Fatalf("collect link facts after source deletion: %v", err)
	}
	if len(sourceDeletedFacts.FieldFacts) != 0 {
		t.Fatalf("source-deleted Links fields = %#v, want none", sourceDeletedFacts.FieldFacts)
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
	for _, activeLinkID := range []uuid.UUID{supportedLinkID, attachedLinkID, derivedLinkID} {
		if !strings.Contains(exportText, activeLinkID.String()) {
			t.Fatalf("Reporting export model omitted active relationship %s: %s", activeLinkID, exportText)
		}
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
	if _, err := harness.Pool.Exec(ctx, `UPDATE records SET deleted_at = $2, deleted_by_user_id = $3 WHERE record_id = $1`, eligibleID, fixedAt.Add(3*time.Minute), actorID); err != nil {
		t.Fatalf("soft-delete Reporting destination before corrected snapshot: %v", err)
	}
	retainedExportModel := requireSnapshotExportModel(t, harness.DB, snapshotID)
	retainedExportJSON, err := json.Marshal(retainedExportModel)
	if err != nil {
		t.Fatalf("encode retained immutable Reporting export model: %v", err)
	}
	if string(retainedExportJSON) != string(exportJSON) {
		t.Fatalf("endpoint deletion rewrote immutable Reporting snapshot:\nbefore: %s\nafter:  %s", exportJSON, retainedExportJSON)
	}
	correctedSnapshotResponse := httptestx.DoJSON(
		t,
		http.MethodPost,
		harness.Server.HTTP.URL+"/api/v1/snapshots",
		map[string]any{
			"incident_id":   incidentID.String(),
			"client_txn_id": "txn-reporting-active-endpoint-snapshot",
		},
		httptestx.WithCookies(admin.SessionCookie, admin.CSRFCookie),
		httptestx.WithHeader(authn.CSRFHeaderName, admin.CSRFCookie.Value),
	)
	correctedSnapshotJob := httptestx.RequireSuccessEnvelope(t, correctedSnapshotResponse, http.StatusAccepted)["data"].(map[string]any)
	correctedSnapshotID := requireSucceededJobResourceID(t, harness, admin, correctedSnapshotJob, "snapshot")
	correctedExportModel := requireSnapshotExportModel(t, harness.DB, correctedSnapshotID)
	correctedExportJSON, err := json.Marshal(correctedExportModel)
	if err != nil {
		t.Fatalf("encode active-endpoint Reporting export model: %v", err)
	}
	correctedExportText := string(correctedExportJSON)
	for _, inactiveLinkID := range []uuid.UUID{supportedLinkID, attachedLinkID, derivedLinkID} {
		if strings.Contains(correctedExportText, inactiveLinkID.String()) {
			t.Fatalf("new Reporting snapshot retained relationship with deleted endpoint %s: %s", inactiveLinkID, correctedExportText)
		}
	}
}

package evidence_test

// Shared full-stack lifecycle fixtures and contract assertions. Concern-specific
// integration tests live in the boundary, projection/collaboration,
// authorization, cleanup, and quarantine files.
import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/JochiRaider/cartulary/internal/modules/evidence/blobref"
	evidenceprojection "github.com/JochiRaider/cartulary/internal/modules/evidence/projectionports"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/testutil/appsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/collaborationsupport/incidentwstest"
	"github.com/JochiRaider/cartulary/internal/testutil/httptestx"
	workbookscenariotest "github.com/JochiRaider/cartulary/internal/testutil/workbookscenariotest"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"net/http"
	"reflect"
	"strings"
	"testing"
	"time"
)

type expectedSupportEffect struct {
	rowVersion   int64
	viewSchemaID string
	fieldKeys    []string
}

func requireExactSupportEffects(t testing.TB, result evidenceprojection.EvidenceAssociationEffectsResult, expected map[uuid.UUID]expectedSupportEffect) {
	t.Helper()
	if len(result.Changes) != len(expected) {
		t.Fatalf("support projection effect count got %d want %d: %#v", len(result.Changes), len(expected), result.Changes)
	}
	for index, change := range result.Changes {
		if index > 0 && strings.Compare(result.Changes[index-1].RecordID.String(), change.RecordID.String()) >= 0 {
			t.Fatalf("support projection effects are not uniquely ordered: %#v", result.Changes)
		}
		want, ok := expected[change.RecordID]
		if !ok {
			t.Fatalf("unexpected support projection effect: %#v", change)
		}
		if change.RowVersion != want.rowVersion || len(change.AffectedViews) != 1 {
			t.Fatalf("support projection effect got %#v want row_version=%d with one view", change, want.rowVersion)
		}
		view := change.AffectedViews[0]
		if view.ViewSchemaID != want.viewSchemaID || view.ChangeKind != evidenceprojection.SupportChangeInvalidate || view.Patch != nil || !reflect.DeepEqual(view.ChangedFieldKeys, want.fieldKeys) {
			t.Fatalf("support projection view got %#v want view=%s invalidate fields=%#v", view, want.viewSchemaID, want.fieldKeys)
		}
	}
}

func AwaitRecordChanges(t testing.TB, client *incidentwstest.Client, expected map[uuid.UUID]int64) map[uuid.UUID]map[string]any {
	t.Helper()
	changes := make(map[uuid.UUID]map[string]any, len(expected))
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) && len(changes) < len(expected) {
		message, err := client.AwaitNextMessage(time.Until(deadline))
		if err != nil {
			t.Fatalf("wait for record_changed set: %v", err)
		}
		if message.Type != "record_changed" {
			continue
		}
		var payload map[string]any
		if err := json.Unmarshal(message.Payload, &payload); err != nil {
			t.Fatalf("decode record_changed payload: %v", err)
		}
		payloadRowVersion, ok := payload["row_version"].(float64)
		if !ok {
			t.Fatalf("record_changed payload missing numeric row_version: %#v", payload)
		}
		recordID, err := uuid.Parse(payload["record_id"].(string))
		if err != nil {
			t.Fatalf("record_changed payload has invalid record_id: %#v", payload)
		}
		if rowVersion, ok := expected[recordID]; ok && int64(payloadRowVersion) == rowVersion {
			changes[recordID] = payload
		}
	}
	if len(changes) != len(expected) {
		t.Fatalf("timed out waiting for record_changed set: got=%d want=%d", len(changes), len(expected))
	}
	return changes
}

func requireEntityEvidenceProjectionCount(t testing.TB, harness *appsupport.ServerHarness, login appsupport.LoginResult, incidentID uuid.UUID, viewSchemaID string, recordID uuid.UUID, fieldKey string, want int) {
	t.Helper()
	row := workbookscenariotest.FindRow(t, workbookscenariotest.QueryViewRows(t, harness.Server.HTTP.URL, incidentID.String(), viewSchemaID, login), recordID.String())
	got := int(row["cells"].(map[string]any)[fieldKey].(map[string]any)["value"].(float64))
	if got != want {
		t.Fatalf("%s got %d want %d in row %#v", fieldKey, got, want, row)
	}
}

func recordTypeForTest(t testing.TB, harness *appsupport.ServerHarness, recordID uuid.UUID) string {
	t.Helper()
	var recordType string
	if err := harness.DB.QueryRowContext(context.Background(), `SELECT record_type FROM records WHERE record_id = $1`, recordID).Scan(&recordType); err != nil {
		t.Fatalf("load record type for %s: %v", recordID, err)
	}
	return recordType
}

func setProjectionEvidenceCount(t testing.TB, harness *appsupport.ServerHarness, tableName string, recordID uuid.UUID, evidenceCount int) {
	t.Helper()
	if _, err := harness.DB.ExecContext(context.Background(), "UPDATE "+tableName+" SET evidence_count = $1 WHERE record_id = $2", evidenceCount, recordID); err != nil {
		t.Fatalf("set %s Evidence count for %s: %v", tableName, recordID, err)
	}
}

func requireProjectionEvidenceCount(t testing.TB, harness *appsupport.ServerHarness, tableName string, recordID uuid.UUID, want int) {
	t.Helper()
	var got int
	if err := harness.DB.QueryRowContext(context.Background(), "SELECT evidence_count FROM "+tableName+" WHERE record_id = $1", recordID).Scan(&got); err != nil {
		t.Fatalf("load %s Evidence count for %s: %v", tableName, recordID, err)
	}
	if got != want {
		t.Fatalf("%s Evidence count for %s got %d want %d", tableName, recordID, got, want)
	}
}

func requireProjectionEvidenceCountTx(t testing.TB, tx pgx.Tx, tableName string, recordID uuid.UUID, want int) {
	t.Helper()
	var got int
	if err := tx.QueryRow(context.Background(), "SELECT evidence_count FROM "+tableName+" WHERE record_id = $1", recordID).Scan(&got); err != nil {
		t.Fatalf("load transactional %s Evidence count for %s: %v", tableName, recordID, err)
	}
	if got != want {
		t.Fatalf("transactional %s Evidence count for %s got %d want %d", tableName, recordID, got, want)
	}
}

func RequireAffectedView(t testing.TB, payload map[string]any, viewSchemaID string) {
	t.Helper()
	affectedViews, ok := payload["affected_views"].([]any)
	if !ok {
		t.Fatalf("record_changed payload missing affected_views: %#v", payload)
	}
	for _, rawView := range affectedViews {
		view, ok := rawView.(map[string]any)
		if !ok {
			t.Fatalf("record_changed affected view has invalid shape: %#v", rawView)
		}
		if view["view_schema_id"] == viewSchemaID {
			if view["change_kind"] == "" {
				t.Fatalf("affected view missing change_kind: %#v", view)
			}
			return
		}
	}
	t.Fatalf("record_changed payload missing affected view %s: %#v", viewSchemaID, payload)
}

func RequireExactInvalidation(t testing.TB, payload map[string]any, viewSchemaID string) {
	t.Helper()
	affectedViews, ok := payload["affected_views"].([]any)
	if !ok || len(affectedViews) != 1 {
		t.Fatalf("record_changed payload must contain exactly one affected view: %#v", payload)
	}
	view, ok := affectedViews[0].(map[string]any)
	if !ok || view["view_schema_id"] != viewSchemaID || view["change_kind"] != "invalidate" {
		t.Fatalf("record_changed affected view must be %s invalidate: %#v", viewSchemaID, affectedViews[0])
	}
	if _, present := view["patch_cells"]; present {
		t.Fatalf("record_changed invalidation must not carry patch_cells: %#v", view)
	}
}

func ChangedFieldKeys(t testing.TB, payload map[string]any) []string {
	t.Helper()
	rawKeys, ok := payload["changed_field_keys"].([]any)
	if !ok {
		t.Fatalf("record_changed payload missing changed_field_keys: %#v", payload)
	}
	keys := make([]string, 0, len(rawKeys))
	for _, rawKey := range rawKeys {
		key, ok := rawKey.(string)
		if !ok {
			t.Fatalf("record_changed changed field key is not a string: %#v", rawKey)
		}
		keys = append(keys, key)
	}
	return keys
}

func requireCreateExpiry(t testing.TB, data map[string]any, field string, want time.Time) {
	t.Helper()
	gotRaw, ok := data[field].(string)
	if !ok {
		t.Fatalf("%s got %T in %#v", field, data[field], data)
	}
	got := mustParseTime(t, gotRaw)
	if !got.Equal(want.UTC()) {
		t.Fatalf("%s got %s want %s", field, got, want.UTC())
	}
	if field == "target_expires_at" {
		target, ok := data["upload_target"].(map[string]any)
		if !ok {
			t.Fatalf("upload_target got %T", data["upload_target"])
		}
		if target["expires_at"] != gotRaw {
			t.Fatalf("upload_target.expires_at got %#v want %q", target["expires_at"], gotRaw)
		}
	}
}

func mustParseTime(t testing.TB, value string) time.Time {
	t.Helper()
	parsed, err := time.Parse(time.RFC3339Nano, value)
	if err != nil {
		t.Fatalf("parse time %q: %v", value, err)
	}
	return parsed.UTC()
}

func extendSessionForClockJump(t testing.TB, harness *appsupport.ServerHarness, userID any, expiresAt time.Time) {
	t.Helper()
	if _, err := harness.DB.ExecContext(context.Background(), `
UPDATE user_sessions
   SET idle_expires_at = $2,
       absolute_expires_at = $2,
       session_expires_at = $2,
       updated_at = now()
 WHERE user_id = $1
   AND revoked_at IS NULL
`, userID, expiresAt.UTC()); err != nil {
		t.Fatalf("extend test session for clock jump: %v", err)
	}
}

func requireHTTPWorkbookCreate(t testing.TB, harness *appsupport.ServerHarness, login appsupport.LoginResult, incidentID uuid.UUID, viewSchemaID string, body map[string]any) map[string]any {
	t.Helper()
	resp := appsupport.DoJSON(t, http.MethodPost, harness.Server.HTTP.URL+"/api/v1/incidents/"+incidentID.String()+"/views/"+viewSchemaID+"/rows", body, authOptions(login)...)
	return httptestx.RequireSuccessEnvelope(t, resp, http.StatusCreated)["data"].(map[string]any)
}

func requireHTTPWorkbookPatch(t testing.TB, harness *appsupport.ServerHarness, login appsupport.LoginResult, recordID uuid.UUID, body map[string]any) map[string]any {
	t.Helper()
	resp := appsupport.DoJSON(t, http.MethodPatch, harness.Server.HTTP.URL+"/api/v1/records/"+recordID.String(), body, authOptions(login)...)
	return httptestx.RequireSuccessEnvelope(t, resp, http.StatusOK)["data"].(map[string]any)
}

func requireTimelineEvidenceProjection(t testing.TB, harness *appsupport.ServerHarness, login appsupport.LoginResult, incidentID uuid.UUID, recordID uuid.UUID, wantCount int, wantHasEvidence bool) {
	t.Helper()
	resp := appsupport.DoJSON(t, http.MethodPost, harness.Server.HTTP.URL+"/api/v1/incidents/"+incidentID.String()+"/views/cartulary.view.timeline.v2/query", map[string]any{}, authOptions(login)...)
	data := httptestx.RequireSuccessEnvelope(t, resp, http.StatusOK)["data"].(map[string]any)
	for _, raw := range data["rows"].([]any) {
		row := raw.(map[string]any)
		if row["record_id"] != recordID.String() {
			continue
		}
		cells := row["cells"].(map[string]any)
		if got := int(cells["timeline.evidence_count"].(map[string]any)["value"].(float64)); got != wantCount {
			t.Fatalf("timeline.evidence_count got %d want %d in row %#v", got, wantCount, row)
		}
		if got := cells["timeline.has_evidence"].(map[string]any)["value"].(bool); got != wantHasEvidence {
			t.Fatalf("timeline.has_evidence got %v want %v in row %#v", got, wantHasEvidence, row)
		}
		return
	}
	t.Fatalf("timeline row %s not found in query %#v", recordID, data["rows"])
}

func requireEvidenceProjectionRow(t testing.TB, harness *appsupport.ServerHarness, login appsupport.LoginResult, incidentID uuid.UUID, recordID uuid.UUID) map[string]any {
	t.Helper()
	resp := appsupport.DoJSON(t, http.MethodPost, harness.Server.HTTP.URL+"/api/v1/incidents/"+incidentID.String()+"/views/cartulary.view.evidence.v1/query", map[string]any{}, authOptions(login)...)
	data := httptestx.RequireSuccessEnvelope(t, resp, http.StatusOK)["data"].(map[string]any)
	for _, raw := range data["rows"].([]any) {
		row := raw.(map[string]any)
		if row["record_id"] == recordID.String() {
			return row
		}
	}
	t.Fatalf("Evidence row %s not found in query %#v", recordID, data["rows"])
	return nil
}

func requireEvidenceProjectionLinkedCount(t testing.TB, harness *appsupport.ServerHarness, login appsupport.LoginResult, incidentID uuid.UUID, recordID uuid.UUID, want int) {
	t.Helper()
	row := requireEvidenceProjectionRow(t, harness, login, incidentID, recordID)
	cells := row["cells"].(map[string]any)
	if got := int(cells["evidence.linked_record_count"].(map[string]any)["value"].(float64)); got != want {
		t.Fatalf("evidence.linked_record_count got %d want %d in row %#v", got, want, row)
	}
}

type SourceHistoryCounts struct {
	ChangeSets   int
	Mutations    int
	Revisions    int
	RecordLinks  int
	TimelineRows int
}

func ProcessCounts(t testing.TB, harness *appsupport.ServerHarness, incidentID uuid.UUID, recordID uuid.UUID) SourceHistoryCounts {
	t.Helper()
	var counts SourceHistoryCounts
	if err := harness.DB.QueryRowContext(context.Background(), `
SELECT
    (SELECT COUNT(*) FROM change_sets WHERE incident_id = $1),
    (SELECT COUNT(*) FROM change_set_mutations m JOIN change_sets cs ON cs.change_set_id = m.change_set_id WHERE cs.incident_id = $1),
    (SELECT COUNT(*) FROM record_revisions WHERE record_id = $2),
    (SELECT COUNT(*) FROM record_links WHERE incident_id = $1 AND deleted_at IS NULL),
    (SELECT COUNT(*) FROM timeline_events WHERE incident_id = $1)
`, incidentID, recordID).Scan(&counts.ChangeSets, &counts.Mutations, &counts.Revisions, &counts.RecordLinks, &counts.TimelineRows); err != nil {
		t.Fatalf("count evidence_lifecycle source/history state: %v", err)
	}
	return counts
}

func requireTimelineProjectionStorage(t testing.TB, harness *appsupport.ServerHarness, recordID uuid.UUID, wantCount int, wantHasEvidence bool) {
	t.Helper()
	var count int
	var hasEvidence bool
	if err := harness.DB.QueryRowContext(context.Background(), `
SELECT evidence_count, has_evidence
  FROM timeline_grid_projection
 WHERE record_id = $1
`, recordID).Scan(&count, &hasEvidence); err != nil {
		t.Fatalf("load timeline projection storage: %v", err)
	}
	if count != wantCount || hasEvidence != wantHasEvidence {
		t.Fatalf("timeline projection storage got count=%d has_evidence=%v want count=%d has_evidence=%v", count, hasEvidence, wantCount, wantHasEvidence)
	}
}

func countEvidenceRevisions(t testing.TB, harness *appsupport.ServerHarness, recordID uuid.UUID) int {
	t.Helper()
	var count int
	if err := harness.DB.QueryRowContext(context.Background(), `SELECT COUNT(*) FROM record_revisions WHERE record_id = $1`, recordID).Scan(&count); err != nil {
		t.Fatalf("count evidence revisions: %v", err)
	}
	return count
}

func countEvidenceBlobLinks(t testing.TB, harness *appsupport.ServerHarness, recordID uuid.UUID) int {
	t.Helper()
	var count int
	if err := harness.DB.QueryRowContext(context.Background(), `SELECT COUNT(*) FROM evidence WHERE record_id = $1 AND object_blob_id IS NOT NULL`, recordID).Scan(&count); err != nil {
		t.Fatalf("count evidence blob links: %v", err)
	}
	return count
}

func countAttachedEvidenceLinks(t testing.TB, harness *appsupport.ServerHarness, incidentID uuid.UUID, srcRecordID uuid.UUID, dstRecordID uuid.UUID) int {
	t.Helper()
	var count int
	if err := harness.DB.QueryRowContext(context.Background(), `
SELECT COUNT(*)
  FROM record_links
 WHERE incident_id = $1
   AND src_record_id = $2
   AND dst_record_id = $3
   AND link_type = 'attached_evidence'
   AND field_key = 'timeline.attached_evidence_ids'
   AND deleted_at IS NULL
`, incidentID, srcRecordID, dstRecordID).Scan(&count); err != nil {
		t.Fatalf("count attached evidence links: %v", err)
	}
	return count
}

func insertRouteBlob(t testing.TB, harness *appsupport.ServerHarness, incidentID uuid.UUID, actorID uuid.UUID, uploadState string) uuid.UUID {
	t.Helper()
	objectBlobID := uuid.New()
	storageKey, err := blobref.ObjectBlobStorageKey(incidentID, objectBlobID)
	if err != nil {
		t.Fatalf("evidence_lifecycle route blob storage key: %v", err)
	}
	now := time.Now().UTC()
	var finalizedAt any
	var terminalReason any
	var failedAt any
	if uploadState == "available" || uploadState == "quarantined" {
		finalizedAt = now
	}
	if uploadState == "failed" {
		terminalReason = "pending_timeout"
		failedAt = now
	}
	if _, err := harness.DB.ExecContext(context.Background(), `
INSERT INTO object_blobs (
    object_blob_id, incident_id, created_by_user_id, storage_key, upload_state,
    byte_size, filename_hint, content_type_hint, observed_size, observed_content_type,
    observed_sha256_hex, target_expires_at, pending_expires_at, finalized_at,
    terminal_reason, failed_at, cleanup_due_at, created_at, updated_at
) VALUES (
    $1, $2, $3, $4, $5,
    11, 'route.txt', 'text/plain', 11, 'text/plain',
    '0000000000000000000000000000000000000000000000000000000000000000',
    $6, $7, $8, $9, $10, $11, $12, $12
)
`, objectBlobID, incidentID, actorID, storageKey, uploadState,
		now.Add(time.Hour), now.Add(24*time.Hour), finalizedAt, terminalReason, failedAt, nullableCleanupDue(uploadState, now), now); err != nil {
		t.Fatalf("insert evidence_lifecycle route blob: %v", err)
	}
	return objectBlobID
}

func nullableCleanupDue(uploadState string, now time.Time) any {
	if uploadState == "failed" {
		return now.Add(time.Hour)
	}
	return nil
}

func loginLocalUserNoMFA(t testing.TB, harness *appsupport.ServerHarness, username string, password string) appsupport.LoginResult {
	t.Helper()
	resp := appsupport.DoJSON(t, http.MethodPost, harness.Server.HTTP.URL+"/api/v1/auth/login", map[string]any{
		"username": username,
		"password": password,
	})
	httptestx.RequireSuccessEnvelope(t, resp, http.StatusOK)
	var sessionCookie *http.Cookie
	var csrfCookie *http.Cookie
	for _, cookie := range resp.Cookies() {
		if cookie.Name == authn.SessionCookieName {
			sessionCookie = cookie
		}
		if cookie.Name == authn.CSRFCookieName {
			csrfCookie = cookie
		}
	}
	if sessionCookie == nil || csrfCookie == nil {
		t.Fatalf("login did not set session and csrf cookies: %#v", resp.Cookies())
	}
	return appsupport.LoginResult{SessionCookie: sessionCookie, CSRFCookie: csrfCookie}
}

func updateRecordDeletedState(t testing.TB, harness *appsupport.ServerHarness, recordID uuid.UUID, deleted bool) {
	t.Helper()
	deletedAt := any(nil)
	if deleted {
		deletedAt = time.Now().UTC()
	}
	if _, err := harness.DB.ExecContext(context.Background(), `
UPDATE records
   SET deleted_at = $2,
       deleted_by_user_id = CASE WHEN $2::timestamptz IS NULL THEN NULL ELSE created_by_user_id END,
       row_version = row_version + 1,
       updated_at = now()
 WHERE record_id = $1
`, recordID, deletedAt); err != nil {
		t.Fatalf("update record deleted state: %v", err)
	}
}

func structuredTableText(t testing.TB, harness *appsupport.ServerHarness, tables ...string) string {
	t.Helper()
	var builder strings.Builder
	for _, table := range tables {
		query := fmt.Sprintf(`SELECT COALESCE(string_agg(to_jsonb(t)::text, E'\n'), '') FROM %s AS t`, quoteIdent(table))
		var text string
		if err := harness.DB.QueryRowContext(context.Background(), query).Scan(&text); err != nil {
			t.Fatalf("dump structured table %s: %v", table, err)
		}
		builder.WriteString(table)
		builder.WriteByte('\n')
		builder.WriteString(text)
		builder.WriteByte('\n')
	}
	return builder.String()
}

func quoteIdent(identifier string) string {
	return `"` + strings.ReplaceAll(identifier, `"`, `""`) + `"`
}

func requireEvidenceStates(t testing.TB, harness *appsupport.ServerHarness, recordID uuid.UUID, wantLifecycle string, wantUpload string) {
	t.Helper()
	var lifecycle string
	var upload string
	if err := harness.DB.QueryRowContext(context.Background(), `
SELECT e.lifecycle_state, COALESCE(b.upload_state, e.upload_state)
  FROM evidence e
  LEFT JOIN object_blobs b ON b.object_blob_id = e.object_blob_id
 WHERE e.record_id = $1
`, recordID).Scan(&lifecycle, &upload); err != nil {
		t.Fatalf("load evidence states: %v", err)
	}
	if lifecycle != wantLifecycle || upload != wantUpload {
		t.Fatalf("evidence states got lifecycle=%s upload=%s want %s/%s", lifecycle, upload, wantLifecycle, wantUpload)
	}
}

func requireObservedContentType(t testing.TB, harness *appsupport.ServerHarness, objectBlobID uuid.UUID, want string) {
	t.Helper()
	var got string
	if err := harness.DB.QueryRowContext(context.Background(), `
SELECT observed_content_type
  FROM object_blobs
 WHERE object_blob_id = $1
`, objectBlobID).Scan(&got); err != nil {
		t.Fatalf("load observed content type: %v", err)
	}
	if got != want {
		t.Fatalf("observed content type got %q want %q", got, want)
	}
}

func attachUploadedBlobWithHints(t *testing.T, harness *appsupport.ServerHarness, login appsupport.LoginResult, incidentID uuid.UUID, recordID uuid.UUID, payload []byte, filename string, hintContentType string, uploadContentType string, createTxn string, attachTxn string) map[string]any {
	t.Helper()
	createResp := appsupport.DoJSON(t, http.MethodPost, harness.Server.HTTP.URL+"/api/v1/object-blobs", map[string]any{
		"incident_id":       incidentID.String(),
		"client_txn_id":     createTxn,
		"byte_size":         len(payload),
		"filename_hint":     filename,
		"content_type_hint": hintContentType,
		"sha256_hex":        fmt.Sprintf("%x", sha256Sum(payload)),
	}, authOptions(login)...)
	createData := httptestx.RequireSuccessEnvelope(t, createResp, http.StatusCreated)["data"].(map[string]any)
	putObject(t, harness.Server.HTTP.URL, createData["upload_target"].(map[string]any)["href"].(string), payload, uploadContentType, login)
	attachResp := appsupport.DoJSON(t, http.MethodPost, harness.Server.HTTP.URL+"/api/v1/evidence-records/"+recordID.String()+"/attach-blob", map[string]any{
		"object_blob_id":   createData["object_blob_id"],
		"base_row_version": 1,
		"client_txn_id":    attachTxn,
	}, authOptions(login)...)
	attachData := httptestx.RequireSuccessEnvelope(t, attachResp, http.StatusOK)["data"].(map[string]any)
	row := attachData["row"].(map[string]any)
	available := requireHTTPWorkbookPatch(t, harness, login, recordID, map[string]any{
		"view_schema_id":   "cartulary.view.evidence.v1",
		"base_row_version": int(row["row_version"].(float64)),
		"client_txn_id":    attachTxn + "-available",
		"changes": []map[string]any{{
			"field_key": "evidence.lifecycle_state",
			"value":     "available",
		}},
	})
	attachData["row"] = available["row"]
	return attachData
}

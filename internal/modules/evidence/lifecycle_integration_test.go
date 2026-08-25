package evidence_test

// Shared lifecycle fixtures, Collaboration publication, and quarantine contracts.
import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	platformws "github.com/JochiRaider/cartulary/internal/modules/collaboration/protocol"
	"github.com/JochiRaider/cartulary/internal/modules/evidence"
	"github.com/JochiRaider/cartulary/internal/modules/evidence/blobref"
	evidenceprojection "github.com/JochiRaider/cartulary/internal/modules/evidence/workbookprojection"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/testutil/appsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/collaborationsupport/incidentwstest"
	"github.com/JochiRaider/cartulary/internal/testutil/httptestx"
	workbookscenariotest "github.com/JochiRaider/cartulary/internal/testutil/workbookscenariotest"
	"github.com/coder/websocket"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"net/http"
	"reflect"
	"slices"
	"strings"
	"testing"
	"time"
)

func TestAttachPublishesWorkbookWebSocketRefresh_Integration(t *testing.T) {
	harness := appsupport.StartServer(t, "evidence_lifecycle-attach-websocket-refresh")
	login, adminID := appsupport.ProvisionBootstrapAdmin(t, harness.Server)
	incident := appsupport.CreateIncident(t, harness.Server, login, map[string]any{
		"client_txn_id": "txn-evidence_lifecycle-i-07-incident",
		"incident_key":  "evidence_lifecycle-i-07",
		"title":         "Evidence attach websocket refresh",
	})
	incidentID := appsupport.MustUUID(t, incident["incident_id"].(string))

	timelineData := requireHTTPWorkbookCreate(t, harness, login, incidentID, "cartulary.view.timeline.v2", map[string]any{
		"client_txn_id":                   "txn-evidence_lifecycle-i-07-timeline",
		"timeline.activity_synopsis_text": "WebSocket evidence count target",
	})
	timelineRow := timelineData["row"].(map[string]any)
	timelineRecordID := appsupport.MustUUID(t, timelineRow["record_id"].(string))
	timelineRowVersion := int64(timelineRow["row_version"].(float64))
	hostData := requireHTTPWorkbookCreate(t, harness, login, incidentID, "cartulary.view.hosts.v1", map[string]any{
		"client_txn_id":     "txn-evidence_lifecycle-i-07-host",
		"host.display_name": "Evidence host",
		"host.hostname":     "EVIDENCE-HOST",
	})
	hostRecordID := appsupport.MustUUID(t, hostData["row"].(map[string]any)["record_id"].(string))
	hostRowVersion := int64(hostData["row"].(map[string]any)["row_version"].(float64))
	identityData := requireHTTPWorkbookCreate(t, harness, login, incidentID, "cartulary.view.identities.v1", map[string]any{
		"client_txn_id":         "txn-evidence_lifecycle-i-07-identity",
		"identity.display_name": "Evidence identity",
		"identity.email":        "evidence.identity@example.test",
	})
	identityRecordID := appsupport.MustUUID(t, identityData["row"].(map[string]any)["record_id"].(string))
	identityRowVersion := int64(identityData["row"].(map[string]any)["row_version"].(float64))
	unaffectedTimelineData := requireHTTPWorkbookCreate(t, harness, login, incidentID, "cartulary.view.timeline.v2", map[string]any{
		"client_txn_id":                   "txn-evidence_lifecycle-i-07-unaffected-timeline",
		"timeline.activity_synopsis_text": "Unrelated timeline row",
	})
	unaffectedTimelineID := appsupport.MustUUID(t, unaffectedTimelineData["row"].(map[string]any)["record_id"].(string))
	unaffectedHostData := requireHTTPWorkbookCreate(t, harness, login, incidentID, "cartulary.view.hosts.v1", map[string]any{
		"client_txn_id":     "txn-evidence_lifecycle-i-07-unaffected-host",
		"host.display_name": "Unrelated host",
		"host.hostname":     "UNRELATED-HOST",
	})
	unaffectedHostID := appsupport.MustUUID(t, unaffectedHostData["row"].(map[string]any)["record_id"].(string))
	unaffectedIdentityData := requireHTTPWorkbookCreate(t, harness, login, incidentID, "cartulary.view.identities.v1", map[string]any{
		"client_txn_id":         "txn-evidence_lifecycle-i-07-unaffected-identity",
		"identity.display_name": "Unrelated identity",
		"identity.email":        "unrelated.identity@example.test",
	})
	unaffectedIdentityID := appsupport.MustUUID(t, unaffectedIdentityData["row"].(map[string]any)["record_id"].(string))
	unregisteredData := requireHTTPWorkbookCreate(t, harness, login, incidentID, "cartulary.view.notes.v1", map[string]any{
		"client_txn_id": "txn-evidence_lifecycle-i-07-unregistered-note",
		"note.title":    "Unregistered projection subject",
	})
	unregisteredRecordID := appsupport.MustUUID(t, unregisteredData["row"].(map[string]any)["record_id"].(string))
	evidenceData := requireHTTPWorkbookCreate(t, harness, login, incidentID, "cartulary.view.evidence.v1", map[string]any{
		"client_txn_id":  "txn-evidence_lifecycle-i-07-evidence",
		"evidence.title": "WebSocket evidence",
	})
	evidenceRecordID := appsupport.MustUUID(t, evidenceData["row"].(map[string]any)["record_id"].(string))
	if _, err := harness.DB.ExecContext(context.Background(), `
INSERT INTO record_links (
    incident_id, src_record_id, dst_record_id, link_type, field_key,
    provenance, owner_user_id, created_by_user_id, decided_at, created_at
) VALUES ($1, $2, $3, 'attached_evidence', 'timeline.attached_evidence_ids', 'manual', $4, $4, now(), now())
`, incidentID, timelineRecordID, evidenceRecordID, adminID); err != nil {
		t.Fatalf("insert attached evidence link: %v", err)
	}
	for _, sourceRecordID := range []uuid.UUID{hostRecordID, identityRecordID} {
		if _, err := harness.DB.ExecContext(context.Background(), `
INSERT INTO record_links (
    incident_id, src_record_id, dst_record_id, link_type,
    provenance, owner_user_id, created_by_user_id, decided_at, created_at
) VALUES ($1, $2, $3, 'attached_evidence', 'manual', $4, $4, now(), now())
`, incidentID, sourceRecordID, evidenceRecordID, adminID); err != nil {
			t.Fatalf("insert entity attached evidence link for %s: %v", sourceRecordID, err)
		}
	}

	for tableName, values := range map[string]map[uuid.UUID]int{
		"timeline_grid_projection": {
			timelineRecordID:     41,
			unaffectedTimelineID: 77,
		},
		"host_grid_projection": {
			hostRecordID:     42,
			unaffectedHostID: 77,
		},
		"identity_grid_projection": {
			identityRecordID:     43,
			unaffectedIdentityID: 77,
		},
	} {
		for recordID, evidenceCount := range values {
			setProjectionEvidenceCount(t, harness, tableName, recordID, evidenceCount)
		}
	}

	subjects := []evidenceprojection.EvidenceAssociationSubject{
		{RecordID: timelineRecordID, RecordType: recordTypeForTest(t, harness, timelineRecordID)},
		{RecordID: hostRecordID, RecordType: recordTypeForTest(t, harness, hostRecordID)},
		{RecordID: identityRecordID, RecordType: recordTypeForTest(t, harness, identityRecordID)},
		{RecordID: unregisteredRecordID, RecordType: recordTypeForTest(t, harness, unregisteredRecordID)},
	}
	slices.SortFunc(subjects, func(left evidenceprojection.EvidenceAssociationSubject, right evidenceprojection.EvidenceAssociationSubject) int {
		return strings.Compare(left.RecordID.String(), right.RecordID.String())
	})
	tx, err := harness.Pool.BeginTx(context.Background(), pgx.TxOptions{})
	if err != nil {
		t.Fatalf("begin support projection rollback proof: %v", err)
	}
	effectsResult, err := harness.Projections.EvidenceSupportEffects().RefreshEvidenceAssociationEffects(
		context.Background(),
		tx,
		evidenceprojection.EvidenceAssociationEffectsInput{IncidentID: incidentID, Subjects: subjects},
	)
	if err != nil {
		_ = tx.Rollback(context.Background())
		t.Fatalf("refresh exact support projection effects: %v", err)
	}
	requireExactSupportEffects(t, effectsResult, map[uuid.UUID]expectedSupportEffect{
		timelineRecordID: {rowVersion: timelineRowVersion, viewSchemaID: "cartulary.view.timeline.v2", fieldKeys: []string{"timeline.attached_evidence_ids", "timeline.evidence_count", "timeline.has_evidence"}},
		hostRecordID:     {rowVersion: hostRowVersion, viewSchemaID: "cartulary.view.hosts.v1", fieldKeys: []string{"host.evidence_count"}},
		identityRecordID: {rowVersion: identityRowVersion, viewSchemaID: "cartulary.view.identities.v1", fieldKeys: []string{"identity.evidence_count"}},
	})
	for tableName, recordID := range map[string]uuid.UUID{
		"timeline_grid_projection": timelineRecordID,
		"host_grid_projection":     hostRecordID,
		"identity_grid_projection": identityRecordID,
	} {
		requireProjectionEvidenceCountTx(t, tx, tableName, recordID, 0)
	}
	if err := tx.Rollback(context.Background()); err != nil {
		t.Fatalf("rollback support projection proof: %v", err)
	}
	for tableName, values := range map[string]map[uuid.UUID]int{
		"timeline_grid_projection": {timelineRecordID: 41, unaffectedTimelineID: 77},
		"host_grid_projection":     {hostRecordID: 42, unaffectedHostID: 77},
		"identity_grid_projection": {identityRecordID: 43, unaffectedIdentityID: 77},
	} {
		for recordID, want := range values {
			requireProjectionEvidenceCount(t, harness, tableName, recordID, want)
		}
	}
	for tableName, recordID := range map[string]uuid.UUID{
		"timeline_grid_projection": timelineRecordID,
		"host_grid_projection":     hostRecordID,
		"identity_grid_projection": identityRecordID,
	} {
		setProjectionEvidenceCount(t, harness, tableName, recordID, 0)
	}

	socket := incidentwstest.ConnectAndHello(t, harness.Server.HTTP.URL, incidentID.String(), incidentwstest.ConnectOptions{
		SessionToken:     login.SessionCookie.Value,
		ClientInstanceID: "evidence_lifecycle-i-07-record-change-listener",
		Presence: platformws.PresenceInput{
			SheetRef: map[string]string{"kind": "view_schema", "id": "cartulary.view.timeline.v2"},
			Mode:     "viewing",
		},
	})
	defer socket.Close(websocket.StatusNormalClosure, "test_complete")

	attachData := attachUploadedBlobWithoutLifecyclePromotion(t, harness, login, incidentID, evidenceRecordID, []byte("evidence_lifecycle websocket attach"), "websocket.txt", "text/plain", "txn-evidence_lifecycle-i-07-blob", "txn-evidence_lifecycle-i-07-attach")
	rowVersion := int64(attachData["row"].(map[string]any)["row_version"].(float64))
	changes := AwaitRecordChanges(t, socket, map[uuid.UUID]int64{
		evidenceRecordID: rowVersion,
		timelineRecordID: timelineRowVersion,
		hostRecordID:     hostRowVersion,
		identityRecordID: identityRowVersion,
	})

	evidenceChange := changes[evidenceRecordID]
	RequireAffectedView(t, evidenceChange, "cartulary.view.evidence.v1")
	if !containsString(ChangedFieldKeys(t, evidenceChange), "evidence.upload_state") {
		t.Fatalf("evidence attach changed keys missing evidence.upload_state: %#v", evidenceChange)
	}
	timelineChange := changes[timelineRecordID]
	RequireExactInvalidation(t, timelineChange, "cartulary.view.timeline.v2")
	changedKeys := ChangedFieldKeys(t, timelineChange)
	for _, key := range []string{"timeline.attached_evidence_ids", "timeline.evidence_count", "timeline.has_evidence"} {
		if !containsString(changedKeys, key) {
			t.Fatalf("timeline websocket changed keys missing %s: %#v", key, timelineChange)
		}
	}
	hostChange := changes[hostRecordID]
	RequireExactInvalidation(t, hostChange, "cartulary.view.hosts.v1")
	if !containsString(ChangedFieldKeys(t, hostChange), "host.evidence_count") {
		t.Fatalf("host websocket changed keys missing host.evidence_count: %#v", hostChange)
	}
	identityChange := changes[identityRecordID]
	RequireExactInvalidation(t, identityChange, "cartulary.view.identities.v1")
	if !containsString(ChangedFieldKeys(t, identityChange), "identity.evidence_count") {
		t.Fatalf("identity websocket changed keys missing identity.evidence_count: %#v", identityChange)
	}
	requireEntityEvidenceProjectionCount(t, harness, login, incidentID, "cartulary.view.hosts.v1", hostRecordID, "host.evidence_count", 0)
	requireEntityEvidenceProjectionCount(t, harness, login, incidentID, "cartulary.view.identities.v1", identityRecordID, "identity.evidence_count", 0)
	availableData := requireHTTPWorkbookPatch(t, harness, login, evidenceRecordID, map[string]any{
		"view_schema_id":   "cartulary.view.evidence.v1",
		"base_row_version": rowVersion,
		"client_txn_id":    "txn-evidence_lifecycle-i-07-available",
		"changes": []map[string]any{{
			"field_key": "evidence.lifecycle_state",
			"value":     "available",
		}},
	})
	availableRowVersion := int64(availableData["row"].(map[string]any)["row_version"].(float64))
	availabilityChanges := AwaitRecordChanges(t, socket, map[uuid.UUID]int64{
		evidenceRecordID: availableRowVersion,
		timelineRecordID: timelineRowVersion,
		hostRecordID:     hostRowVersion,
		identityRecordID: identityRowVersion,
	})
	RequireExactInvalidation(t, availabilityChanges[timelineRecordID], "cartulary.view.timeline.v2")
	RequireExactInvalidation(t, availabilityChanges[hostRecordID], "cartulary.view.hosts.v1")
	RequireExactInvalidation(t, availabilityChanges[identityRecordID], "cartulary.view.identities.v1")
	requireEntityEvidenceProjectionCount(t, harness, login, incidentID, "cartulary.view.hosts.v1", hostRecordID, "host.evidence_count", 1)
	requireEntityEvidenceProjectionCount(t, harness, login, incidentID, "cartulary.view.identities.v1", identityRecordID, "identity.evidence_count", 1)
	for tableName, recordID := range map[string]uuid.UUID{
		"timeline_grid_projection": unaffectedTimelineID,
		"host_grid_projection":     unaffectedHostID,
		"identity_grid_projection": unaffectedIdentityID,
	} {
		requireProjectionEvidenceCount(t, harness, tableName, recordID, 77)
	}
	if _, err := harness.DB.ExecContext(context.Background(), `UPDATE host_grid_projection SET evidence_count = 0 WHERE record_id = $1`, hostRecordID); err != nil {
		t.Fatalf("corrupt host evidence count: %v", err)
	}
	if _, err := harness.DB.ExecContext(context.Background(), `UPDATE identity_grid_projection SET evidence_count = 0 WHERE record_id = $1`, identityRecordID); err != nil {
		t.Fatalf("corrupt identity evidence count: %v", err)
	}
	projectionRebuild := harness.Projections
	if err := projectionRebuild.RebuildHosts(context.Background(), incidentID); err != nil {
		t.Fatalf("rebuild host evidence projection: %v", err)
	}
	if err := projectionRebuild.RebuildIdentities(context.Background(), incidentID); err != nil {
		t.Fatalf("rebuild identity evidence projection: %v", err)
	}
	requireEntityEvidenceProjectionCount(t, harness, login, incidentID, "cartulary.view.hosts.v1", hostRecordID, "host.evidence_count", 1)
	requireEntityEvidenceProjectionCount(t, harness, login, incidentID, "cartulary.view.identities.v1", identityRecordID, "identity.evidence_count", 1)

	objectBlobID := appsupport.MustUUID(t, attachData["object_blob_id"].(string))
	quarantine, err := newEvidenceLifecycleTestService(harness).QuarantineBlob(context.Background(), adminID, objectBlobID, "content_inspection_quarantine", "req-evidence_lifecycle-i-07-quarantine", time.Now().UTC())
	if err != nil {
		t.Fatalf("quarantine entity-linked evidence: %v", err)
	}
	quarantineChanges := map[uuid.UUID]evidence.AttachRecordChange{}
	supportQuarantineChanges := make([]evidence.AttachRecordChange, 0, len(quarantine.ChangedEvidenceRows))
	for _, change := range quarantine.ChangedEvidenceRows {
		quarantineChanges[change.RecordID] = change
		if len(change.AffectedViews) > 0 {
			supportQuarantineChanges = append(supportQuarantineChanges, change)
		}
	}
	requireOrderedAttachRecordChanges(t, supportQuarantineChanges)
	if !containsString(quarantineChanges[hostRecordID].ChangedFieldKeys, "host.evidence_count") {
		t.Fatalf("quarantine host change missing host.evidence_count: %#v", quarantineChanges[hostRecordID])
	}
	if !containsString(quarantineChanges[identityRecordID].ChangedFieldKeys, "identity.evidence_count") {
		t.Fatalf("quarantine identity change missing identity.evidence_count: %#v", quarantineChanges[identityRecordID])
	}
	requireEntityEvidenceProjectionCount(t, harness, login, incidentID, "cartulary.view.hosts.v1", hostRecordID, "host.evidence_count", 0)
	requireEntityEvidenceProjectionCount(t, harness, login, incidentID, "cartulary.view.identities.v1", identityRecordID, "identity.evidence_count", 0)
}

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

func requireOrderedAttachRecordChanges(t testing.TB, changes []evidence.AttachRecordChange) {
	t.Helper()
	for index, change := range changes {
		if index > 0 && strings.Compare(changes[index-1].RecordID.String(), change.RecordID.String()) >= 0 {
			t.Fatalf("attach record changes are not uniquely ordered: %#v", changes)
		}
		if len(change.AffectedViews) != 1 || change.AffectedViews[0].ChangeKind != evidenceprojection.SupportChangeInvalidate || change.AffectedViews[0].Patch != nil {
			t.Fatalf("attach record change is not a neutral invalidation: %#v", change)
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

func newEvidenceLifecycleTestService(harness *appsupport.ServerHarness) *evidence.BlobLifecycleService {
	service, err := evidence.NewBlobLifecycleService(evidence.BlobLifecycleDependencies{
		Postgres:       harness.Pool,
		Revisions:      harness.Revisions.Appender(),
		Projections:    harness.Projections.EvidencePort(),
		SupportEffects: harness.Projections.EvidenceSupportEffects(),
		Collaboration:  harness.Collaboration.Publications(),
	})
	if err != nil {
		panic(err)
	}
	return service
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

func TestQuarantineBoundaryPreservesTwoStepAttach_Integration(t *testing.T) {
	t.Run("AC-405 object bytes stay outside structured state and loss fails closed", func(t *testing.T) {
		harness := appsupport.StartServer(t, "evidence_lifecycle-i-04-object-boundary")
		login, adminID := appsupport.ProvisionBootstrapAdmin(t, harness.Server)
		incident := appsupport.CreateIncident(t, harness.Server, login, map[string]any{
			"client_txn_id": "txn-evidence_lifecycle-i-04-boundary-incident",
			"incident_key":  "evidence_lifecycle-i-04-boundary",
			"title":         "Evidence object boundary",
		})
		incidentID := appsupport.MustUUID(t, incident["incident_id"].(string))
		recordID := uuid.New()
		seedEvidenceRecord(t, harness, incidentID, adminID, recordID)

		marker := "evidence_lifecycle-ac405-marker-" + uuid.NewString() + "-payload"
		payload := []byte("prefix-" + marker + "-suffix")
		attachData := attachUploadedBlobWithMetadata(t, harness, login, incidentID, recordID, payload, "boundary.txt", "text/plain", "txn-evidence_lifecycle-i-04-boundary-blob", "txn-evidence_lifecycle-i-04-boundary-attach")
		objectBlobID := appsupport.MustUUID(t, attachData["object_blob_id"].(string))

		preview := issueEvidenceHandle(t, harness, login, recordID, "preview-handle")
		if got := string(redeemHandle(t, harness.Server.HTTP.URL+preview["href"].(string), login)); got != string(payload) {
			t.Fatalf("preview before object loss got %q want %q", got, string(payload))
		}
		download := issueEvidenceHandle(t, harness, login, recordID, "download-handle")
		if got := string(redeemHandle(t, harness.Server.HTTP.URL+download["href"].(string), login)); got != string(payload) {
			t.Fatalf("download before object loss got %q want %q", got, string(payload))
		}

		structured := structuredTableText(t, harness,
			"records", "evidence", "object_blobs", "route_idempotency",
			"change_sets", "change_set_mutations", "record_revisions", "evidence_access_handles",
		)
		if strings.Contains(structured, marker) {
			t.Fatalf("structured Postgres state contains inline evidence payload marker %q", marker)
		}

		storageKey := blobStorageKey(t, harness, objectBlobID)
		if err := harness.ObjectStore.DeleteObject(context.Background(), storageKey); err != nil {
			t.Fatalf("delete object bytes: %v", err)
		}
		if got := countEvidenceBlobLinks(t, harness, recordID); got != 1 {
			t.Fatalf("object loss changed committed evidence blob link count: got %d want 1", got)
		}
		requireEvidenceStates(t, harness, recordID, "available", "available")
		for _, endpoint := range []string{"preview-handle", "download-handle"} {
			requireEvidenceAccessUnavailableReason(t,
				appsupport.DoJSON(t, http.MethodPost, harness.Server.HTTP.URL+"/api/v1/evidence-records/"+recordID.String()+"/"+endpoint, map[string]any{}, authOptions(login)...),
				"blob_missing",
			)
		}
	})

	t.Run("failed unattached cleanup deletes bytes and retains metadata", func(t *testing.T) {
		harness := appsupport.StartServer(t, "evidence_lifecycle-i-04-cleanup")
		login, adminID := appsupport.ProvisionBootstrapAdmin(t, harness.Server)
		incident := appsupport.CreateIncident(t, harness.Server, login, map[string]any{
			"client_txn_id": "txn-evidence_lifecycle-i-04-cleanup-incident",
			"incident_key":  "evidence_lifecycle-i-04-cleanup",
			"title":         "Evidence cleanup",
		})
		incidentID := appsupport.MustUUID(t, incident["incident_id"].(string))
		now := time.Now().UTC().Truncate(time.Second)

		cleanupBlobID := uuid.New()
		cleanupKey := "evidence_lifecycle/i-04/cleanup/" + cleanupBlobID.String()
		cleanupPayload := "evidence_lifecycle cleanup orphan bytes"
		if err := harness.ObjectStore.PutObject(context.Background(), cleanupKey, strings.NewReader(cleanupPayload), int64(len(cleanupPayload)), "text/plain"); err != nil {
			t.Fatalf("put cleanup candidate object: %v", err)
		}
		insertFailedCleanupBlob(t, harness, incidentID, adminID, cleanupBlobID, cleanupKey, now)

		expiredBlobID := uuid.New()
		insertExpiredPendingBlob(t, harness, incidentID, adminID, expiredBlobID, "evidence_lifecycle/i-04/expired/"+expiredBlobID.String(), now)

		cleanup, err := evidence.NewCleanupService(harness.Pool)
		if err != nil {
			t.Fatalf("compose cleanup service: %v", err)
		}
		result, err := cleanup.SweepFailedUnattachedBlobs(context.Background(), harness.ObjectStore, now)
		if err != nil {
			t.Fatalf("cleanup failed unattached blob bytes: %v", err)
		}
		if result.ExpiredPendingCount != 1 || result.CleanedBlobCount != 1 {
			t.Fatalf("cleanup result got expired=%d cleaned=%d want 1/1", result.ExpiredPendingCount, result.CleanedBlobCount)
		}
		if _, err := harness.ObjectStore.StatObject(context.Background(), cleanupKey); err == nil {
			t.Fatalf("cleanup candidate object bytes still exist at %s", cleanupKey)
		}
		requireCleanedFailedBlobMetadata(t, harness, cleanupBlobID)
		requireExpiredPendingFailed(t, harness, expiredBlobID)
		if got := countEvidenceRowsForBlob(t, harness, cleanupBlobID); got != 0 {
			t.Fatalf("cleanup created or retained evidence link rows: got %d want 0", got)
		}
	})

	t.Run("quarantine bridges evidence and blocks attach preview and download", func(t *testing.T) {
		harness := appsupport.StartServer(t, "evidence_lifecycle-i-04-quarantine")
		login, adminID := appsupport.ProvisionBootstrapAdmin(t, harness.Server)
		incident := appsupport.CreateIncident(t, harness.Server, login, map[string]any{
			"client_txn_id": "txn-evidence_lifecycle-i-04-quarantine-incident",
			"incident_key":  "evidence_lifecycle-i-04-quarantine",
			"title":         "Evidence quarantine",
		})
		incidentID := appsupport.MustUUID(t, incident["incident_id"].(string))
		recordID := uuid.New()
		seedEvidenceRecord(t, harness, incidentID, adminID, recordID)
		attachData := attachUploadedBlobWithMetadata(t, harness, login, incidentID, recordID, []byte("evidence_lifecycle quarantine body"), "quarantine.txt", "text/plain", "txn-evidence_lifecycle-i-04-quarantine-blob", "txn-evidence_lifecycle-i-04-quarantine-attach")
		objectBlobID := appsupport.MustUUID(t, attachData["object_blob_id"].(string))
		preview := issueEvidenceHandle(t, harness, login, recordID, "preview-handle")
		download := issueEvidenceHandle(t, harness, login, recordID, "download-handle")
		beforeRevisions := countEvidenceRevisions(t, harness, recordID)

		store := newEvidenceLifecycleTestService(harness)
		if _, err := store.QuarantineBlob(context.Background(), adminID, objectBlobID, "unsupported_trigger", "req-evidence_lifecycle-i-04-bad-trigger", time.Now().UTC()); !errors.Is(err, evidence.ErrIllegalBlobTransition) {
			t.Fatalf("unsupported quarantine trigger got %v want ErrIllegalBlobTransition", err)
		}
		result, err := store.QuarantineBlob(context.Background(), adminID, objectBlobID, "content_inspection_quarantine", "req-evidence_lifecycle-i-04-quarantine", time.Now().UTC())
		if err != nil {
			t.Fatalf("quarantine blob: %v", err)
		}
		if result.ChangedEvidenceRecord != 1 || result.ChangeSetID == uuid.Nil {
			t.Fatalf("quarantine result got changed=%d change_set=%s", result.ChangedEvidenceRecord, result.ChangeSetID)
		}
		requireEvidenceStates(t, harness, recordID, "quarantined", "quarantined")
		if got := countEvidenceRevisions(t, harness, recordID); got != beforeRevisions+1 {
			t.Fatalf("quarantine revision count got %d want %d", got, beforeRevisions+1)
		}
		requireChangeSetSource(t, harness, result.ChangeSetID, "evidence.blob.quarantine")

		for _, handle := range []map[string]any{preview, download} {
			requireEvidenceAccessUnavailableReason(t,
				appsupport.DoJSON(t, http.MethodGet, harness.Server.HTTP.URL+handle["href"].(string), nil, appsupport.WithCookies(login.SessionCookie)),
				"evidence_quarantined",
			)
		}
		for _, endpoint := range []string{"preview-handle", "download-handle"} {
			requireEvidenceAccessUnavailableReason(t,
				appsupport.DoJSON(t, http.MethodPost, harness.Server.HTTP.URL+"/api/v1/evidence-records/"+recordID.String()+"/"+endpoint, map[string]any{}, authOptions(login)...),
				"evidence_quarantined",
			)
		}
		secondRecordID := uuid.New()
		seedEvidenceRecord(t, harness, incidentID, adminID, secondRecordID)
		attachBlocked := httptestx.RequireErrorEnvelope(t,
			appsupport.DoJSON(t, http.MethodPost, harness.Server.HTTP.URL+"/api/v1/evidence-records/"+secondRecordID.String()+"/attach-blob", map[string]any{
				"object_blob_id":   objectBlobID.String(),
				"base_row_version": 1,
				"client_txn_id":    "txn-evidence_lifecycle-i-04-quarantine-attach-blocked",
			}, authOptions(login)...),
			http.StatusConflict,
			"evidence_attach_rejected",
		)
		attachDetails := attachBlocked["error"].(map[string]any)["details"].(map[string]any)
		if got := attachDetails["reason_code"]; got != evidence.AttachReasonBlobNotVisible {
			t.Fatalf("associated quarantined blob reason got %v want %s", got, evidence.AttachReasonBlobNotVisible)
		}

		pendingID := uuid.New()
		insertExpiredPendingBlob(t, harness, incidentID, adminID, pendingID, "evidence_lifecycle/i-04/pending-quarantine/"+pendingID.String(), time.Now().UTC().Add(2*time.Hour))
		if _, err := store.QuarantineBlob(context.Background(), adminID, pendingID, "admin_quarantine", "req-evidence_lifecycle-i-04-pending-quarantine", time.Now().UTC()); !errors.Is(err, evidence.ErrIllegalBlobTransition) {
			t.Fatalf("pending quarantine got %v want ErrIllegalBlobTransition", err)
		}
	})

	t.Run("active content uses observed media state for preview policy", func(t *testing.T) {
		for _, active := range []struct {
			name        string
			contentType string
			filename    string
		}{
			{name: "html", contentType: "text/html", filename: "pretend-image.png"},
			{name: "svg", contentType: "image/svg+xml", filename: "pretend-raster.png"},
		} {
			t.Run(active.name, func(t *testing.T) {
				harness := appsupport.StartServer(t, "evidence_lifecycle-i-04-active-"+active.name)
				login, adminID := appsupport.ProvisionBootstrapAdmin(t, harness.Server)
				incident := appsupport.CreateIncident(t, harness.Server, login, map[string]any{
					"client_txn_id": "txn-evidence_lifecycle-i-04-active-" + active.name + "-incident",
					"incident_key":  "evidence_lifecycle-i-04-active-" + active.name,
					"title":         "Evidence active content " + active.name,
				})
				incidentID := appsupport.MustUUID(t, incident["incident_id"].(string))
				recordID := uuid.New()
				seedEvidenceRecord(t, harness, incidentID, adminID, recordID)
				payload := []byte("<script>window.__cartulary_evidence_lifecycle_active_content = true</script>")
				attachData := attachUploadedBlobWithHints(t, harness, login, incidentID, recordID, payload, active.filename, active.contentType, active.contentType, "txn-evidence_lifecycle-i-04-active-"+active.name+"-blob", "txn-evidence_lifecycle-i-04-active-"+active.name+"-attach")
				objectBlobID := appsupport.MustUUID(t, attachData["object_blob_id"].(string))
				requireObservedContentType(t, harness, objectBlobID, active.contentType)

				requireEvidenceAccessUnavailableReason(t,
					appsupport.DoJSON(t, http.MethodPost, harness.Server.HTTP.URL+"/api/v1/evidence-records/"+recordID.String()+"/preview-handle", map[string]any{}, authOptions(login)...),
					"unsupported_preview",
				)
				download := issueEvidenceHandle(t, harness, login, recordID, "download-handle")
				if got := string(redeemHandle(t, harness.Server.HTTP.URL+download["href"].(string), login)); got != string(payload) {
					t.Fatalf("download active content got %q want %q", got, string(payload))
				}
			})
		}
	})
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

func insertFailedCleanupBlob(t testing.TB, harness *appsupport.ServerHarness, incidentID uuid.UUID, actorID uuid.UUID, objectBlobID uuid.UUID, storageKey string, now time.Time) {
	t.Helper()
	if _, err := harness.DB.ExecContext(context.Background(), `
INSERT INTO object_blobs (
    object_blob_id, incident_id, created_by_user_id, storage_key, upload_state,
    byte_size, filename_hint, content_type_hint, observed_size, observed_content_type,
    observed_sha256_hex, target_expires_at, pending_expires_at, finalized_at,
    terminal_reason, failed_at, cleanup_due_at, created_at, updated_at
) VALUES (
    $1, $2, $3, $4, 'failed',
    27, 'cleanup.txt', 'text/plain', 27, 'text/plain',
    '0000000000000000000000000000000000000000000000000000000000000000',
    $5::timestamptz - interval '3 hours', $5::timestamptz - interval '2 hours', NULL,
    'pending_timeout', $5::timestamptz - interval '2 hours', $5::timestamptz - interval '1 hour',
    $5::timestamptz - interval '3 hours', $5::timestamptz - interval '2 hours'
)
`, objectBlobID, incidentID, actorID, storageKey, now.UTC()); err != nil {
		t.Fatalf("insert failed cleanup blob: %v", err)
	}
}

func insertExpiredPendingBlob(t testing.TB, harness *appsupport.ServerHarness, incidentID uuid.UUID, actorID uuid.UUID, objectBlobID uuid.UUID, storageKey string, now time.Time) {
	t.Helper()
	if _, err := harness.DB.ExecContext(context.Background(), `
INSERT INTO object_blobs (
    object_blob_id, incident_id, created_by_user_id, storage_key, upload_state,
    byte_size, filename_hint, content_type_hint,
    target_expires_at, pending_expires_at, created_at, updated_at
) VALUES (
    $1, $2, $3, $4, 'pending',
    10, 'expired.txt', 'text/plain',
    $5::timestamptz - interval '2 hours', $5::timestamptz - interval '1 hour',
    $5::timestamptz - interval '25 hours', $5::timestamptz - interval '25 hours'
)
`, objectBlobID, incidentID, actorID, storageKey, now.UTC()); err != nil {
		t.Fatalf("insert expired pending blob: %v", err)
	}
}

func requireCleanedFailedBlobMetadata(t testing.TB, harness *appsupport.ServerHarness, objectBlobID uuid.UUID) {
	t.Helper()
	var uploadState string
	var cleaned bool
	if err := harness.DB.QueryRowContext(context.Background(), `
SELECT upload_state, cleaned_up_at IS NOT NULL
  FROM object_blobs
 WHERE object_blob_id = $1
`, objectBlobID).Scan(&uploadState, &cleaned); err != nil {
		t.Fatalf("load cleaned blob metadata: %v", err)
	}
	if uploadState != "failed" || !cleaned {
		t.Fatalf("cleaned failed blob metadata got state=%s cleaned=%v want failed/true", uploadState, cleaned)
	}
}

func requireExpiredPendingFailed(t testing.TB, harness *appsupport.ServerHarness, objectBlobID uuid.UUID) {
	t.Helper()
	var uploadState string
	var terminalReason string
	var failedAt time.Time
	var cleanupDueAt time.Time
	var cleaned bool
	if err := harness.DB.QueryRowContext(context.Background(), `
SELECT upload_state, terminal_reason, failed_at, cleanup_due_at, cleaned_up_at IS NOT NULL
  FROM object_blobs
 WHERE object_blob_id = $1
`, objectBlobID).Scan(&uploadState, &terminalReason, &failedAt, &cleanupDueAt, &cleaned); err != nil {
		t.Fatalf("load expired pending blob: %v", err)
	}
	if uploadState != "failed" || terminalReason != "pending_timeout" || cleaned {
		t.Fatalf("expired pending blob got state=%s reason=%s failed_at=%s cleanup_due_at=%s cleaned=%v", uploadState, terminalReason, failedAt, cleanupDueAt, cleaned)
	}
	if !cleanupDueAt.Equal(failedAt.Add(45 * time.Minute)) {
		t.Fatalf("expired pending cleanup_due_at = %s, want failed_at + 45 minutes (%s)", cleanupDueAt, failedAt.Add(45*time.Minute))
	}
}

func countEvidenceRowsForBlob(t testing.TB, harness *appsupport.ServerHarness, objectBlobID uuid.UUID) int {
	t.Helper()
	var count int
	if err := harness.DB.QueryRowContext(context.Background(), `
SELECT count(*)
  FROM evidence
 WHERE object_blob_id = $1
`, objectBlobID).Scan(&count); err != nil {
		t.Fatalf("count evidence rows for blob: %v", err)
	}
	return count
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

func requireChangeSetSource(t testing.TB, harness *appsupport.ServerHarness, changeSetID uuid.UUID, wantSource string) {
	t.Helper()
	var source string
	if err := harness.DB.QueryRowContext(context.Background(), `
SELECT source
  FROM change_sets
 WHERE change_set_id = $1
`, changeSetID).Scan(&source); err != nil {
		t.Fatalf("load change set source: %v", err)
	}
	if source != wantSource {
		t.Fatalf("change set source got %s want %s", source, wantSource)
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

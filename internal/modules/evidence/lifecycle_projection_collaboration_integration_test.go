package evidence_test

import (
	"context"
	"slices"
	"strings"
	"testing"

	"github.com/coder/websocket"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	platformws "github.com/JochiRaider/cartulary/internal/modules/collaboration/protocol"
	evidenceprojection "github.com/JochiRaider/cartulary/internal/modules/evidence/projectionports"
	"github.com/JochiRaider/cartulary/internal/testutil/appsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/collaborationsupport/incidentwstest"
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
	effectsResult, err := harness.Projections.EvidenceAssociationEffects().RefreshEvidenceAssociationEffects(
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
	if err := projectionRebuild.RebuildIncident(context.Background(), incidentID); err != nil {
		t.Fatalf("rebuild host evidence projection: %v", err)
	}
	if err := projectionRebuild.RebuildIncident(context.Background(), incidentID); err != nil {
		t.Fatalf("rebuild identity evidence projection: %v", err)
	}
	requireEntityEvidenceProjectionCount(t, harness, login, incidentID, "cartulary.view.hosts.v1", hostRecordID, "host.evidence_count", 1)
	requireEntityEvidenceProjectionCount(t, harness, login, incidentID, "cartulary.view.identities.v1", identityRecordID, "identity.evidence_count", 1)

}

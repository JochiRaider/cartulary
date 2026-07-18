package timeline_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/modules/entities/hostidentity"
	"github.com/JochiRaider/cartulary/internal/modules/records/testsupport/golden"
	recordstoretest "github.com/JochiRaider/cartulary/internal/modules/records/testsupport/storetest"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/platform/fieldnorm"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"

	. "github.com/JochiRaider/cartulary/internal/modules/timeline"
)

// U-4-01 / REQ-02-028..REQ-02-036 / AC-019, AC-020, AC-022.
func TestBindingMode_Unit(t *testing.T) {
	harness := recordstoretest.StartStore(t, "phase4-u-4-01")
	timelineStore := newResolutionTimelineCommands(harness.DB)
	entityStore := hostidentity.NewStore(harness.DB)
	actor := recordstoretest.SeedLocalUserFlags(t, harness.DB, "u401@example.test", "U401", "U401Phase4Pass1!", false, false, true)
	incident := recordstoretest.CreateIncidentInStore(t, harness.DB, actor, "txn-phase4-u-4-01-incident", "IR-U401", "Phase 4 U-4-01")

	recordstoretest.RequireViewFieldBindingMode(t, "U-4-01", golden.RecordTimelineViewSchemaID, golden.RecordFieldTimelineHostRefs, "mention_origin")
	recordstoretest.RequireViewFieldBindingMode(t, "U-4-01", golden.RecordTimelineViewSchemaID, golden.RecordFieldTimelineIdentityRefs, "mention_origin")
	recordstoretest.RequireViewFieldBindingMode(t, "U-4-01", golden.RecordHostsViewSchemaID, "host.display_name", "entity_origin")
	recordstoretest.RequireViewFieldBindingMode(t, "U-4-01", golden.RecordHostsViewSchemaID, "host.hostname", "entity_origin")
	recordstoretest.RequireViewFieldBindingMode(t, "U-4-01", golden.RecordHostsViewSchemaID, "host.aliases", "entity_origin")
	recordstoretest.RequireViewFieldBindingMode(t, "U-4-01", golden.RecordIdentitiesViewSchemaID, "identity.display_name", "entity_origin")
	recordstoretest.RequireViewFieldBindingMode(t, "U-4-01", golden.RecordIdentitiesViewSchemaID, "identity.email", "entity_origin")
	recordstoretest.RequireViewFieldBindingMode(t, "U-4-01", golden.RecordIdentitiesViewSchemaID, "identity.aliases", "entity_origin")

	normalizedHostToken, ok := fieldnorm.NormalizeMentionToken("WS-023")
	if !ok {
		t.Fatal("normalize mention token")
	}
	timelineResult, err := timelineStore.CreateRow(context.Background(), actor, incident.ID, CreateRequest{
		ClientTxnID:          "txn-phase4-u-4-01-timeline",
		ActivitySynopsisText: stringPtr("Mention origin row"),
		HostRefs: &CollectionActionPayload{
			Actions: []CollectionAction{
				{
					Op:             "add_token",
					RawText:        "WS-023",
					NormalizedText: normalizedHostToken,
				},
			},
		},
	}, []byte("txn-phase4-u-4-01-timeline"), "req-phase4-u-4-01-timeline", time.Now().UTC())
	if err != nil {
		t.Fatalf("create mention-origin row: %v", err)
	}
	if got := recordstoretest.QueryCount(t, harness.DB, `SELECT COUNT(*) FROM entity_mentions WHERE source_record_id = $1`, timelineResult.RecordID); got != 1 {
		t.Fatalf("expected mention_origin write to create one entity mention, got %d", got)
	}
	if got := recordstoretest.QueryCount(t, harness.DB, `SELECT COUNT(*) FROM hosts WHERE incident_id = $1`, incident.ID); got != 0 {
		t.Fatalf("mention_origin write must not synthesize hosts, got %d", got)
	}
	if got := recordstoretest.QueryCount(t, harness.DB, `SELECT COUNT(*) FROM identities WHERE incident_id = $1`, incident.ID); got != 0 {
		t.Fatalf("mention_origin write must not synthesize identities, got %d", got)
	}

	entityResult, err := entityStore.CreateHostRow(context.Background(), actor, incident.ID, hostidentity.CreateRequest{
		ClientTxnID: "txn-phase4-u-4-01-host",
		Values: map[string]string{
			"host.display_name": "WS-023",
			"host.hostname":     "WS-023",
		},
	}, []byte("txn-phase4-u-4-01-host"), "req-phase4-u-4-01-host", time.Now().UTC())
	if err != nil {
		t.Fatalf("create entity-origin host row: %v", err)
	}
	if entityResult.RecordID == timelineResult.RecordID {
		t.Fatalf("unexpected shared record id between timeline row and host row: %#v %#v", timelineResult, entityResult)
	}
	if got := recordstoretest.QueryCount(t, harness.DB, `SELECT COUNT(*) FROM hosts WHERE incident_id = $1`, incident.ID); got != 1 {
		t.Fatalf("expected entity_origin write to create one host, got %d", got)
	}
	if got := recordstoretest.QueryCount(t, harness.DB, `SELECT COUNT(*) FROM entity_mentions WHERE source_record_id = $1`, entityResult.RecordID); got != 0 {
		t.Fatalf("entity_origin write must not synthesize mentions, got %d", got)
	}

	identityResult, err := entityStore.CreateIdentityRow(context.Background(), actor, incident.ID, hostidentity.CreateRequest{
		ClientTxnID: "txn-phase4-u-4-01-identity",
		Values: map[string]string{
			"identity.display_name": "Alex Analyst",
			"identity.email":        "alex.analyst@example.test",
		},
	}, []byte("txn-phase4-u-4-01-identity"), "req-phase4-u-4-01-identity", time.Now().UTC())
	if err != nil {
		t.Fatalf("create entity-origin identity row: %v", err)
	}
	if identityResult.RecordID == timelineResult.RecordID || identityResult.RecordID == entityResult.RecordID {
		t.Fatalf("unexpected shared record id between timeline, host, and identity rows: timeline=%s host=%s identity=%s", timelineResult.RecordID, entityResult.RecordID, identityResult.RecordID)
	}
	if got := recordstoretest.QueryCount(t, harness.DB, `SELECT COUNT(*) FROM identities WHERE incident_id = $1`, incident.ID); got != 1 {
		t.Fatalf("expected entity_origin write to create one identity, got %d", got)
	}
	if got := recordstoretest.QueryCount(t, harness.DB, `SELECT COUNT(*) FROM entity_mentions WHERE source_record_id = $1`, identityResult.RecordID); got != 0 {
		t.Fatalf("identity entity_origin write must not synthesize mentions, got %d", got)
	}

	recordstoretest.SeedEntityAlias(t, harness.DB, incident.ID, actor.ID, entityResult.RecordID, "host", "Import Host Alias")
	recordstoretest.SeedEntityAlias(t, harness.DB, incident.ID, actor.ID, identityResult.RecordID, "identity", "Import Identity Alias")
	importHostNormalizedToken, ok := fieldnorm.NormalizeMentionToken(" import   host alias ")
	if !ok {
		t.Fatal("normalize import host mention token")
	}
	importIdentityNormalizedToken, ok := fieldnorm.NormalizeMentionToken(" import   identity alias ")
	if !ok {
		t.Fatal("normalize import identity mention token")
	}
	importRequest := CreateRequest{
		ClientTxnID:          "txn-phase4-u-4-01-import-row",
		ActivitySynopsisText: stringPtr("Import create preserves mention tokens"),
		HostRefs: &CollectionActionPayload{Actions: []CollectionAction{{
			Op:             "add_token",
			RawText:        " import   host alias ",
			NormalizedText: importHostNormalizedToken,
		}}},
		IdentityRefs: &CollectionActionPayload{Actions: []CollectionAction{{
			Op:             "add_token",
			RawText:        " import   identity alias ",
			NormalizedText: importIdentityNormalizedToken,
		}}},
	}
	importResult, err := timelineStore.CreateImportedRow(context.Background(), actor, incident.ID, importRequest, TimelineCreateRequestHash(importRequest), "req-phase4-u-4-01-import-row", time.Now().UTC())
	if err != nil {
		t.Fatalf("import create row: %v", err)
	}
	importRow := importResult.Payload["row"].(map[string]any)
	importHostItem := recordstoretest.RequireSingleCollectionItem(t, importRow, golden.RecordFieldTimelineHostRefs)
	if importHostItem["item_kind"] != "unresolved_mention" || importHostItem["resolved_record_id"] != nil {
		t.Fatalf("import host token must remain unresolved, got %#v", importHostItem)
	}
	if _, ok := importHostItem["provenance"]; ok {
		t.Fatalf("import host token must not surface auto-match provenance, got %#v", importHostItem)
	}
	importIdentityItem := recordstoretest.RequireSingleCollectionItem(t, importRow, golden.RecordFieldTimelineIdentityRefs)
	if importIdentityItem["item_kind"] != "unresolved_mention" || importIdentityItem["resolved_record_id"] != nil {
		t.Fatalf("import identity token must remain unresolved, got %#v", importIdentityItem)
	}
	if _, ok := importIdentityItem["provenance"]; ok {
		t.Fatalf("import identity token must not surface auto-match provenance, got %#v", importIdentityItem)
	}
	importHostMention := recordstoretest.LookupMention(t, harness.DB, recordstoretest.MentionIDFromItemRef(t, importHostItem["item_ref"].(string)))
	if importHostMention.ResolutionStatus != golden.RecordMentionStatusUnresolved || importHostMention.ResolvedRecordID != nil || importHostMention.ResolutionMethod != nil {
		t.Fatalf("import host mention must remain unresolved, got %#v", importHostMention)
	}
	importIdentityMention := recordstoretest.LookupMention(t, harness.DB, recordstoretest.MentionIDFromItemRef(t, importIdentityItem["item_ref"].(string)))
	if importIdentityMention.ResolutionStatus != golden.RecordMentionStatusUnresolved || importIdentityMention.ResolvedRecordID != nil || importIdentityMention.ResolutionMethod != nil {
		t.Fatalf("import identity mention must remain unresolved, got %#v", importIdentityMention)
	}
	if got := recordstoretest.QueryCount(t, harness.DB, `SELECT COUNT(*) FROM record_links WHERE src_record_id = $1 AND deleted_at IS NULL`, importResult.RecordID); got != 0 {
		t.Fatalf("import create must not create auto-match links, got %d", got)
	}
}

// U-4-02 / REQ-02-031..REQ-02-032, REQ-02-058 / AC-019, AC-021.
func TestDuplicateMentionProvenance_Unit(t *testing.T) {
	harness := recordstoretest.StartStore(t, "phase4-u-4-02")
	store := newResolutionTimelineCommands(harness.DB)
	actor := recordstoretest.SeedLocalUserFlags(t, harness.DB, "u402@example.test", "U402", "U402Phase4Pass1!", false, false, true)
	incident := recordstoretest.CreateIncidentInStore(t, harness.DB, actor, "txn-phase4-u-4-02-incident", "IR-U402", "Phase 4 U-4-02")

	normalizedToken, ok := fieldnorm.NormalizeMentionToken("WS-023")
	if !ok {
		t.Fatal("normalize mention token")
	}
	first, err := store.CreateRow(context.Background(), actor, incident.ID, CreateRequest{
		ClientTxnID:          "txn-phase4-u-4-02-first",
		ActivitySynopsisText: stringPtr("Duplicate provenance one"),
		HostRefs: &CollectionActionPayload{
			Actions: []CollectionAction{{Op: "add_token", RawText: "WS-023", NormalizedText: normalizedToken}},
		},
	}, []byte("txn-phase4-u-4-02-first"), "req-phase4-u-4-02-first", time.Now().UTC())
	if err != nil {
		t.Fatalf("create first row: %v", err)
	}
	second, err := store.CreateRow(context.Background(), actor, incident.ID, CreateRequest{
		ClientTxnID:          "txn-phase4-u-4-02-second",
		ActivitySynopsisText: stringPtr("Duplicate provenance two"),
		HostRefs: &CollectionActionPayload{
			Actions: []CollectionAction{{Op: "add_token", RawText: "WS-023", NormalizedText: normalizedToken}},
		},
	}, []byte("txn-phase4-u-4-02-second"), "req-phase4-u-4-02-second", time.Now().UTC())
	if err != nil {
		t.Fatalf("create second row: %v", err)
	}

	rows, err := harness.DB.Query(context.Background(), `
SELECT entity_mention_id::text, source_record_id::text, raw_text, origin_locator
  FROM entity_mentions
 WHERE source_record_id IN ($1, $2)
 ORDER BY created_at ASC, entity_mention_id ASC
`, first.RecordID, second.RecordID)
	if err != nil {
		t.Fatalf("query entity mentions: %v", err)
	}
	defer rows.Close()

	type mentionRow struct {
		id            string
		sourceRecord  string
		rawText       string
		originLocator string
	}
	var mentions []mentionRow
	for rows.Next() {
		var row mentionRow
		if err := rows.Scan(&row.id, &row.sourceRecord, &row.rawText, &row.originLocator); err != nil {
			t.Fatalf("scan entity mention: %v", err)
		}
		mentions = append(mentions, row)
	}
	if err := rows.Err(); err != nil {
		t.Fatalf("iterate entity mentions: %v", err)
	}
	if len(mentions) != 2 {
		t.Fatalf("expected two durable mentions for repeated identical tokens, got %#v", mentions)
	}
	if mentions[0].rawText != "WS-023" || mentions[1].rawText != "WS-023" {
		t.Fatalf("expected repeated raw text preservation, got %#v", mentions)
	}
	if mentions[0].id == mentions[1].id || mentions[0].originLocator == mentions[1].originLocator || mentions[0].sourceRecord == mentions[1].sourceRecord {
		t.Fatalf("expected repeated identical tokens to keep distinct provenance, got %#v", mentions)
	}
	if mentions[0].sourceRecord != first.RecordID.String() || mentions[1].sourceRecord != second.RecordID.String() {
		t.Fatalf("expected mention provenance to stay attached to each source row, got %#v", mentions)
	}
}

func TestAttachedEvidenceCreateAndPatch(t *testing.T) {
	harness := recordstoretest.StartStore(t, "phase5-attached-evidence")
	store := newResolutionTimelineCommands(harness.DB)
	actor := recordstoretest.SeedLocalUserFlags(t, harness.DB, "u5attach@example.test", "U5ATTACH", "U5AttachPass1!", false, false, true)
	incident := recordstoretest.CreateIncidentInStore(t, harness.DB, actor, "txn-phase5-attached-incident", "IR-U5ATTACH", "Phase 5 attached evidence")

	evidenceID := seedTimelineEvidence(t, harness, incident.ID, actor.ID, "Screenshot one", "available")
	create := CreateRequest{
		ClientTxnID: "txn-phase5-attached-create",
		AttachedEvidence: &CollectionActionPayload{Actions: []CollectionAction{{
			Op:             "add_record_ref",
			LinkedRecordID: &evidenceID,
		}}},
	}
	created, err := store.CreateRow(context.Background(), actor, incident.ID, create, TimelineCreateRequestHash(create), "req-phase5-attached-create", time.Now().UTC())
	if err != nil {
		t.Fatalf("create screenshot-only timeline row: %v", err)
	}
	cells := created.Payload["row"].(map[string]any)["cells"].(map[string]any)
	if got := cells["timeline.capture_state"].(map[string]any)["value"]; got != "rough" {
		t.Fatalf("screenshot-only create capture state got %#v want rough", got)
	}
	if got := cells["timeline.evidence_count"].(map[string]any)["value"]; got != 1 {
		t.Fatalf("screenshot-only create evidence_count got %#v want 1", got)
	}
	if got := recordstoretest.QueryCount(t, harness.DB, `SELECT COUNT(*) FROM record_links WHERE src_record_id = $1 AND dst_record_id = $2 AND link_type = 'attached_evidence' AND field_key = 'timeline.attached_evidence_ids' AND deleted_at IS NULL`, created.RecordID, evidenceID); got != 1 {
		t.Fatalf("expected one attached evidence link, got %d", got)
	}

	row := CreateRequest{
		ClientTxnID:          "txn-phase5-attached-existing",
		ActivitySynopsisText: stringPtr("Existing row"),
	}
	existing, err := store.CreateRow(context.Background(), actor, incident.ID, row, TimelineCreateRequestHash(row), "req-phase5-attached-existing", time.Now().UTC())
	if err != nil {
		t.Fatalf("create existing row: %v", err)
	}
	patchEvidenceID := seedTimelineEvidence(t, harness, incident.ID, actor.ID, "Screenshot two", "available")
	patch := PatchRequest{
		ViewSchemaID:   TimelineViewSchemaID,
		BaseRowVersion: existing.RowVersion,
		ClientTxnID:    "txn-phase5-attached-patch",
		CanonicalChange: []PatchChange{{
			FieldKey: "timeline.attached_evidence_ids",
			ActionPayload: &CollectionActionPayload{Actions: []CollectionAction{{
				Op:             "add_record_ref",
				LinkedRecordID: &patchEvidenceID,
			}}},
		}},
	}
	patched, err := store.PatchRow(context.Background(), actor, existing.RecordID, patch, TimelinePatchRequestHash(patch), "req-phase5-attached-patch", time.Now().UTC())
	if err != nil {
		t.Fatalf("patch existing row with attached evidence: %v", err)
	}
	patchedCells := patched.Payload["row"].(map[string]any)["cells"].(map[string]any)
	if got := patchedCells["timeline.capture_state"].(map[string]any)["value"]; got != "enriched" {
		t.Fatalf("attached evidence patch capture state got %#v want enriched", got)
	}
	if got := patchedCells["timeline.evidence_count"].(map[string]any)["value"]; got != 1 {
		t.Fatalf("attached evidence patch evidence_count got %#v want 1", got)
	}
}

type resolutionTimelineCommands struct {
	facade *Facade
}

func newResolutionTimelineCommands(pool postgres.DB) *resolutionTimelineCommands {
	return &resolutionTimelineCommands{facade: NewFacade(pool)}
}

func (c *resolutionTimelineCommands) CreateRow(ctx context.Context, actor authn.UserRecord, incidentID uuid.UUID, request CreateRequest, requestHash []byte, requestID string, now time.Time) (MutationResult, error) {
	return c.facade.CreateRow(ctx, CreateRowCommand{
		Actor:       actor,
		IncidentID:  incidentID,
		Request:     request,
		RequestHash: requestHash,
		RequestID:   requestID,
		Now:         now,
	})
}

func (c *resolutionTimelineCommands) CreateImportedRow(ctx context.Context, actor authn.UserRecord, incidentID uuid.UUID, request CreateRequest, requestHash []byte, requestID string, now time.Time) (MutationResult, error) {
	return c.facade.CreateImportedRow(ctx, CreateRowCommand{
		Actor:       actor,
		IncidentID:  incidentID,
		Request:     request,
		RequestHash: requestHash,
		RequestID:   requestID,
		Now:         now,
	})
}

func (c *resolutionTimelineCommands) PatchRow(ctx context.Context, actor authn.UserRecord, recordID uuid.UUID, request PatchRequest, requestHash []byte, requestID string, now time.Time) (MutationResult, error) {
	return c.facade.PatchRow(ctx, PatchRowCommand{
		Actor:       actor,
		RecordID:    recordID,
		Request:     request,
		RequestHash: requestHash,
		RequestID:   requestID,
		Now:         now,
	})
}

func seedTimelineEvidence(t testing.TB, harness *recordstoretest.StoreHarness, incidentID uuid.UUID, actorID uuid.UUID, title string, uploadState string) uuid.UUID {
	t.Helper()
	now := time.Now().UTC()
	recordID := uuid.New()
	blobID := uuid.New()
	if _, err := harness.DB.Exec(context.Background(), `
INSERT INTO records (record_id, incident_id, record_type, created_by_user_id, created_at, updated_by_user_id, updated_at, row_version)
VALUES ($1, $2, 'evidence', $3, $4, $3, $4, 1)
`, recordID, incidentID, actorID, now); err != nil {
		t.Fatalf("insert evidence envelope: %v", err)
	}
	if _, err := harness.DB.Exec(context.Background(), `
INSERT INTO object_blobs (
    object_blob_id, incident_id, created_by_user_id, storage_key, upload_state,
    byte_size, filename_hint, content_type_hint, observed_size, observed_content_type,
    target_expires_at, pending_expires_at, finalized_at, created_at, updated_at
) VALUES ($1, $2, $3, $4, $5, 8, 'screenshot.txt', 'text/plain', 8, 'text/plain', $6, $6, $7, $7, $7)
`, blobID, incidentID, actorID, "phase5/"+blobID.String(), uploadState, now.Add(time.Hour), now); err != nil {
		t.Fatalf("insert object blob: %v", err)
	}
	if _, err := harness.DB.Exec(context.Background(), `
INSERT INTO evidence (record_id, incident_id, title, lifecycle_state, upload_state, object_blob_id, requested_at, created_at, updated_at)
VALUES ($1, $2, $3, 'available', $4, $5, $6, $6, $6)
`, recordID, incidentID, title, uploadState, blobID, now); err != nil {
		t.Fatalf("insert evidence row: %v", err)
	}
	return recordID
}

func stringPtr(value string) *string {
	return &value
}

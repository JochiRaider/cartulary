package links_test

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"github.com/JochiRaider/cartulary/internal/modules/collaboration/testsupport/incidentwstest"
	entitytest "github.com/JochiRaider/cartulary/internal/modules/entities/testsupport"
	envelopetest "github.com/JochiRaider/cartulary/internal/modules/records/testsupport/envelopetest"
	timelinetest "github.com/JochiRaider/cartulary/internal/modules/timeline/testsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/appsupport"
	"net/http"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/coder/websocket"
	"github.com/google/uuid"

	authstoretest "github.com/JochiRaider/cartulary/internal/modules/auth/testsupport/storetest"

	"github.com/JochiRaider/cartulary/internal/app/projectionassembly"
	"github.com/JochiRaider/cartulary/internal/app/timelineassembly"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/incidentportability"
	"github.com/JochiRaider/cartulary/internal/modules/links"
	linktest "github.com/JochiRaider/cartulary/internal/modules/links/testsupport"
	"github.com/JochiRaider/cartulary/internal/modules/timeline"
	timelineadmission "github.com/JochiRaider/cartulary/internal/modules/timeline/admission"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
	"github.com/JochiRaider/cartulary/internal/testutil/conflicttest"
	"github.com/JochiRaider/cartulary/internal/testutil/httptestx"
	"github.com/JochiRaider/cartulary/internal/testutil/revisionsupport"
	workbookscenariotest "github.com/JochiRaider/cartulary/internal/testutil/workbookscenariotest"
)

const TimelineView = "cartulary.view.timeline.v2"

func TestActiveLinksAndTagsViewsV1Contract(t *testing.T) {
	harness := appsupport.StartStore(t, "saved_view_query-active-links-tags-v1")
	actor := authstoretest.SeedLocalUserRecord(t, harness.DB, "saved_view_query-views@example.test", "Workbook query Views", "SavedViewQueryViewsPass1!", false, true, true)
	incident := appsupport.CreateIncidentInStore(t, harness.DB, actor, "txn-saved_view_query-active-views-incident", "IR-P8-VIEWS", "Workbook query active view contracts")
	incidentID := incident.ID

	wantLinkColumns := []string{
		"record_link_id",
		"incident_id",
		"src_record_id",
		"src_record_type",
		"dst_record_id",
		"dst_record_type",
		"link_type",
		"provenance",
		"confidence",
		"owner_user_id",
		"decided_at",
		"created_at",
		"deleted_at",
		"deleted_by_user_id",
		"created_by_user_id",
		"field_key",
	}
	if got := columnNamesPG(t, harness.DB, "active_record_links_v1"); !slices.Equal(got, wantLinkColumns) {
		t.Fatalf("active_record_links_v1 columns got %v want %v", got, wantLinkColumns)
	}
	wantTagColumns := []string{
		"record_tag_id",
		"incident_id",
		"record_id",
		"record_type",
		"tag_name",
		"normalized_tag_name",
		"created_by_user_id",
		"created_at",
		"updated_at",
		"deleted_at",
		"deleted_by_user_id",
	}
	if got := columnNamesPG(t, harness.DB, "active_record_tags_v1"); !slices.Equal(got, wantTagColumns) {
		t.Fatalf("active_record_tags_v1 columns got %v want %v", got, wantTagColumns)
	}
	for _, forbidden := range []string{"note", "description", "comment"} {
		if slices.Contains(columnNamesPG(t, harness.DB, "record_links"), forbidden) {
			t.Fatalf("record_links unexpectedly exposes link-local narrative column %q", forbidden)
		}
	}

	src := uuid.New()
	dst := uuid.New()
	fieldDst := uuid.New()
	deletedLinkDst := uuid.New()
	deletedEndpoint := uuid.New()
	taggedRecord := uuid.New()
	deletedTagRecord := uuid.New()
	deletedTagEndpoint := uuid.New()
	for _, recordID := range []uuid.UUID{src, dst, fieldDst, deletedLinkDst, deletedEndpoint, taggedRecord, deletedTagRecord, deletedTagEndpoint} {
		timelinetest.SeedTimelineRecord(t, harness.DB, incidentID, actor.ID, recordID)
	}

	activeLinkID := uuid.New()
	fieldLinkID := uuid.New()
	deletedLinkID := uuid.New()
	endpointDeletedLinkID := uuid.New()
	if _, err := harness.DB.Exec(context.Background(), `
INSERT INTO record_links (record_link_id, incident_id, src_record_id, dst_record_id, link_type, field_key, provenance, owner_user_id, created_by_user_id)
VALUES
    ($1, $2, $3, $4, 'references_record', NULL, 'manual', $11, $11),
    ($5, $2, $3, $6, 'references_record', 'fixture.field_ref', 'manual', $11, $11),
    ($7, $2, $3, $8, 'references_record', NULL, 'manual', $11, $11),
    ($9, $2, $3, $10, 'references_record', NULL, 'manual', $11, $11)
`, activeLinkID, incidentID, src, dst, fieldLinkID, fieldDst, deletedLinkID, deletedLinkDst, endpointDeletedLinkID, deletedEndpoint, actor.ID); err != nil {
		t.Fatalf("seed active view record_links: %v", err)
	}
	if _, err := harness.DB.Exec(context.Background(), `UPDATE record_links SET deleted_at = now(), deleted_by_user_id = $2 WHERE record_link_id = $1`, deletedLinkID, actor.ID); err != nil {
		t.Fatalf("soft-delete link fixture: %v", err)
	}
	if _, err := harness.DB.Exec(context.Background(), `UPDATE records SET deleted_at = now(), deleted_by_user_id = $2 WHERE record_id = $1`, deletedEndpoint, actor.ID); err != nil {
		t.Fatalf("soft-delete link endpoint fixture: %v", err)
	}
	if got := appsupport.QueryCount(t, harness.DB, `SELECT count(*) FROM active_record_links_v1 WHERE record_link_id IN ($1, $2)`, activeLinkID, fieldLinkID); got != 2 {
		t.Fatalf("active_record_links_v1 did not expose active unfielded and field-key links, got %d", got)
	}
	if got := appsupport.QueryCount(t, harness.DB, `SELECT count(*) FROM active_record_links_v1 WHERE record_link_id IN ($1, $2)`, deletedLinkID, endpointDeletedLinkID); got != 0 {
		t.Fatalf("active_record_links_v1 exposed deleted link or deleted endpoint link, got %d", got)
	}
	var fieldKey string
	if err := harness.DB.QueryRow(context.Background(), `SELECT field_key FROM active_record_links_v1 WHERE record_link_id = $1`, fieldLinkID).Scan(&fieldKey); err != nil {
		t.Fatalf("query active link field_key: %v", err)
	}
	if fieldKey != "fixture.field_ref" {
		t.Fatalf("active_record_links_v1 lost field_key: got %q", fieldKey)
	}

	activeTagID := uuid.New()
	deletedTagID := uuid.New()
	endpointDeletedTagID := uuid.New()
	if _, err := harness.DB.Exec(context.Background(), `
INSERT INTO record_tags (record_tag_id, incident_id, record_id, tag_name, normalized_tag_name, created_by_user_id)
VALUES
    ($1, $2, $3, 'Active Tag', 'active tag', $8),
    ($4, $2, $5, 'Deleted Tag', 'deleted tag', $8),
    ($6, $2, $7, 'Endpoint Deleted Tag', 'endpoint deleted tag', $8)
`, activeTagID, incidentID, taggedRecord, deletedTagID, deletedTagRecord, endpointDeletedTagID, deletedTagEndpoint, actor.ID); err != nil {
		t.Fatalf("seed active view record_tags: %v", err)
	}
	if _, err := harness.DB.Exec(context.Background(), `UPDATE record_tags SET deleted_at = now(), deleted_by_user_id = $2 WHERE record_tag_id = $1`, deletedTagID, actor.ID); err != nil {
		t.Fatalf("soft-delete tag fixture: %v", err)
	}
	if _, err := harness.DB.Exec(context.Background(), `UPDATE records SET deleted_at = now(), deleted_by_user_id = $2 WHERE record_id = $1`, deletedTagEndpoint, actor.ID); err != nil {
		t.Fatalf("soft-delete tag endpoint fixture: %v", err)
	}
	if got := appsupport.QueryCount(t, harness.DB, `SELECT count(*) FROM active_record_tags_v1 WHERE record_tag_id = $1`, activeTagID); got != 1 {
		t.Fatalf("active_record_tags_v1 did not expose active tag, got %d", got)
	}
	if got := appsupport.QueryCount(t, harness.DB, `SELECT count(*) FROM active_record_tags_v1 WHERE record_tag_id IN ($1, $2)`, deletedTagID, endpointDeletedTagID); got != 0 {
		t.Fatalf("active_record_tags_v1 exposed deleted tag or deleted endpoint tag, got %d", got)
	}
}

func TestRecordLinkOwnerValidation(t *testing.T) {
	harness := appsupport.StartStore(t, "saved_view_query-link-owner-validation")
	actor := authstoretest.SeedLocalUserRecord(t, harness.DB, "saved_view_query-validation@example.test", "Workbook query Validation", "SavedViewQueryValidationPass1!", false, true, true)
	incident := appsupport.CreateIncidentInStore(t, harness.DB, actor, "txn-saved_view_query-link-validation-incident", "IR-P8-VALIDATE", "Workbook query link validation")
	src := uuid.New()
	dst := uuid.New()
	replacement := uuid.New()
	superseded := uuid.New()
	host := uuid.New()
	for _, recordID := range []uuid.UUID{src, dst, replacement, superseded} {
		timelinetest.SeedTimelineRecord(t, harness.DB, incident.ID, actor.ID, recordID)
	}
	entitytest.SeedHostRecord(t, harness.DB, incident.ID, actor.ID, host, "Validation Host", "validation-host", "", "")

	tx, err := harness.DB.BeginTx(context.Background(), pgx.TxOptions{})
	if err != nil {
		t.Fatalf("begin validation tx: %v", err)
	}
	defer func() { _ = tx.Rollback(context.Background()) }()

	store := links.NewStore()
	confidence := 100
	if _, _, err := store.UpsertLinkCommandTx(context.Background(), tx, links.UpsertLinkCommand{
		IncidentID:  incident.ID,
		SrcRecordID: src,
		DstRecordID: dst,
		LinkType:    links.LinkType(links.LinkTypeReferencesRecord),
		Provenance:  links.LinkProvenance(links.LinkProvenanceManual),
		Confidence:  &confidence,
		OwnerUserID: actor.ID,
		Now:         time.Now().UTC(),
	}); !errors.Is(err, links.ErrInvalidRecordLink) {
		t.Fatalf("manual link confidence should be rejected, got %v", err)
	}
	if _, _, err := store.UpsertLinkCommandTx(context.Background(), tx, links.UpsertLinkCommand{
		IncidentID:  incident.ID,
		SrcRecordID: src,
		DstRecordID: dst,
		LinkType:    links.LinkType(links.LinkTypeSupportedBy),
		Provenance:  links.LinkProvenance(links.LinkProvenanceAutoMatch),
		Confidence:  &confidence,
		OwnerUserID: actor.ID,
		Now:         time.Now().UTC(),
	}); !errors.Is(err, links.ErrInvalidRecordLink) {
		t.Fatalf("auto_match on unsupported link type should be rejected, got %v", err)
	}
	if _, _, err := store.UpsertLinkCommandTx(context.Background(), tx, links.UpsertLinkCommand{
		IncidentID:  incident.ID,
		SrcRecordID: src,
		DstRecordID: src,
		LinkType:    links.LinkType(links.LinkTypeReferencesRecord),
		Provenance:  links.LinkProvenance(links.LinkProvenanceManual),
		OwnerUserID: actor.ID,
		Now:         time.Now().UTC(),
	}); !errors.Is(err, links.ErrInvalidRecordLink) {
		t.Fatalf("self-link should be rejected, got %v", err)
	}
	if _, _, err := store.UpsertLinkCommandTx(context.Background(), tx, links.UpsertLinkCommand{
		IncidentID:  incident.ID,
		SrcRecordID: src,
		DstRecordID: dst,
		LinkType:    links.LinkType(links.LinkTypeObservedOnHost),
		Provenance:  links.LinkProvenance(links.LinkProvenanceAutoMatch),
		Confidence:  &confidence,
		OwnerUserID: actor.ID,
		Now:         time.Now().UTC(),
	}); err != nil {
		t.Fatalf("valid auto_match observation rejected: %v", err)
	}
	if _, err := store.InsertSupersedesCommandTx(context.Background(), tx, links.InsertSupersedesCommand{
		IncidentID:          incident.ID,
		ReplacementRecordID: replacement,
		SupersededRecordID:  superseded,
		OwnerUserID:         actor.ID,
		Now:                 time.Now().UTC(),
	}); err != nil {
		t.Fatalf("valid timeline supersedes rejected: %v", err)
	}
	if _, err := store.InsertSupersedesCommandTx(context.Background(), tx, links.InsertSupersedesCommand{
		IncidentID:          incident.ID,
		ReplacementRecordID: replacement,
		SupersededRecordID:  host,
		OwnerUserID:         actor.ID,
		Now:                 time.Now().UTC(),
	}); !errors.Is(err, links.ErrInvalidRecordLink) {
		t.Fatalf("mixed-type supersedes should be rejected, got %v", err)
	}
	if _, err := tx.Exec(context.Background(), `UPDATE records SET deleted_at = now(), deleted_by_user_id = $2 WHERE record_id = $1`, dst, actor.ID); err != nil {
		t.Fatalf("soft-delete validation endpoint: %v", err)
	}
	if _, _, err := store.UpsertLinkCommandTx(context.Background(), tx, links.UpsertLinkCommand{
		IncidentID:  incident.ID,
		SrcRecordID: src,
		DstRecordID: dst,
		LinkType:    links.LinkType(links.LinkTypeReferencesRecord),
		Provenance:  links.LinkProvenance(links.LinkProvenanceManual),
		OwnerUserID: actor.ID,
		Now:         time.Now().UTC(),
	}); !errors.Is(err, links.ErrInvalidRecordLink) {
		t.Fatalf("deleted endpoint should be rejected, got %v", err)
	}
}

func TestFieldAwareLinkIdentityAndMergeCharacterization(t *testing.T) {
	harness := appsupport.StartStore(t, "links-field-aware-characterization")
	actor := authstoretest.SeedLocalUserRecord(t, harness.DB, "links-field-aware@example.test", "Links Field Aware", "LinksFieldAwarePass1!", false, true, true)
	incident := appsupport.CreateIncidentInStore(t, harness.DB, actor, "txn-links-field-aware-incident", "IR-LINK-FIELD", "Links field identity")
	ctx := context.Background()
	now := time.Date(2026, time.August, 21, 18, 40, 0, 0, time.UTC)
	src := uuid.New()
	dst := uuid.New()
	for _, recordID := range []uuid.UUID{src, dst} {
		timelinetest.SeedTimelineRecord(t, harness.DB, incident.ID, actor.ID, recordID)
	}
	tx, err := harness.DB.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		t.Fatalf("begin field-aware characterization: %v", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	store := links.NewStore()
	fieldA := "fixture.links.field_a"
	fieldB := "fixture.links.field_b"
	linkA, inserted, err := store.UpsertFieldReferenceRecordTx(ctx, tx, incident.ID, src, dst, fieldA, links.LinkTypeReferencesRecord, actor.ID, now)
	if err != nil || !inserted {
		t.Fatalf("insert field A link: inserted=%t err=%v", inserted, err)
	}
	replayedA, inserted, err := store.UpsertFieldReferenceRecordTx(ctx, tx, incident.ID, src, dst, fieldA, links.LinkTypeReferencesRecord, actor.ID, now.Add(time.Second))
	if err != nil || inserted || replayedA.RecordLinkID != linkA.RecordLinkID {
		t.Fatalf("same-field replay = (%s, %t, %v), want original %s without insert", replayedA.RecordLinkID, inserted, err, linkA.RecordLinkID)
	}
	linkB, inserted, err := store.UpsertFieldReferenceRecordTx(ctx, tx, incident.ID, src, dst, fieldB, links.LinkTypeReferencesRecord, actor.ID, now)
	if err != nil || !inserted || linkB.RecordLinkID == linkA.RecordLinkID {
		t.Fatalf("different-field insert = (%s, %t, %v), want distinct active binding", linkB.RecordLinkID, inserted, err)
	}
	unfielded, inserted, err := store.UpsertLinkCommandTx(ctx, tx, links.UpsertLinkCommand{
		IncidentID: incident.ID, SrcRecordID: src, DstRecordID: dst,
		LinkType: links.LinkType(links.LinkTypeReferencesRecord), Provenance: links.LinkProvenance(links.LinkProvenanceManual),
		OwnerUserID: actor.ID, Now: now,
	})
	if err != nil || !inserted {
		t.Fatalf("insert null-field link: inserted=%t err=%v", inserted, err)
	}
	replayedUnfielded, inserted, err := store.UpsertLinkCommandTx(ctx, tx, links.UpsertLinkCommand{
		IncidentID: incident.ID, SrcRecordID: src, DstRecordID: dst,
		LinkType: links.LinkType(links.LinkTypeReferencesRecord), Provenance: links.LinkProvenance(links.LinkProvenanceManual),
		OwnerUserID: actor.ID, Now: now.Add(time.Second),
	})
	if err != nil || inserted || replayedUnfielded.RecordLinkID != unfielded.RecordLinkID {
		t.Fatalf("null-field replay = (%s, %t, %v), want original %s without insert", replayedUnfielded.RecordLinkID, inserted, err, unfielded.RecordLinkID)
	}

	if _, err := store.TombstoneFieldReferenceRecordTx(ctx, tx, incident.ID, src, dst, fieldA, links.LinkTypeReferencesRecord, "timeline_event", actor.ID, now.Add(2*time.Second)); err != nil {
		t.Fatalf("tombstone exact field A binding: %v", err)
	}
	if _, err := store.TombstoneFieldReferenceRecordTx(ctx, tx, incident.ID, src, dst, "fixture.links.foreign", links.LinkTypeReferencesRecord, "timeline_event", actor.ID, now.Add(3*time.Second)); !errors.Is(err, links.ErrFieldReferenceNotFound) {
		t.Fatalf("foreign field removal error = %v, want ErrFieldReferenceNotFound", err)
	}
	var activeA, activeB, activeNull int
	if err := tx.QueryRow(ctx, `
SELECT
    count(*) FILTER (WHERE field_key = $4),
    count(*) FILTER (WHERE field_key = $5),
    count(*) FILTER (WHERE field_key IS NULL)
  FROM record_links
 WHERE incident_id = $1 AND src_record_id = $2 AND dst_record_id = $3
   AND link_type = 'references_record' AND deleted_at IS NULL
`, incident.ID, src, dst, fieldA, fieldB).Scan(&activeA, &activeB, &activeNull); err != nil {
		t.Fatalf("query field-scoped bindings: %v", err)
	}
	if activeA != 0 || activeB != 1 || activeNull != 1 {
		t.Fatalf("active field counts = A:%d B:%d null:%d, want 0/1/1", activeA, activeB, activeNull)
	}
}

func TestMergeDeduplicatesOnlyFullFieldAwareTupleCharacterization(t *testing.T) {
	harness := appsupport.StartStore(t, "links-field-aware-merge-characterization")
	actor := authstoretest.SeedLocalUserRecord(t, harness.DB, "links-field-merge@example.test", "Links Field Merge", "LinksFieldMergePass1!", false, true, true)
	incident := appsupport.CreateIncidentInStore(t, harness.DB, actor, "txn-links-field-merge-incident", "IR-LINK-MERGE", "Links field merge")
	ctx := context.Background()
	now := time.Date(2026, time.August, 21, 18, 41, 0, 0, time.UTC)
	survivor := uuid.New()
	loser := uuid.New()
	target := uuid.New()
	for _, recordID := range []uuid.UUID{survivor, loser, target} {
		timelinetest.SeedTimelineRecord(t, harness.DB, incident.ID, actor.ID, recordID)
	}
	tx, err := harness.DB.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		t.Fatalf("begin merge characterization: %v", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()
	if _, err := tx.Exec(ctx, `
INSERT INTO record_links (
    record_link_id, incident_id, src_record_id, dst_record_id, link_type,
    field_key, provenance, owner_user_id, created_by_user_id, decided_at, created_at
) VALUES
    ('00000000-0000-0000-0000-000000000101', $1, $2, $4, 'references_record', 'fixture.links.field_a', 'manual', $5, $5, $6, $6),
    ('00000000-0000-0000-0000-000000000102', $1, $3, $4, 'references_record', 'fixture.links.field_a', 'manual', $5, $5, $6, $6),
    ('00000000-0000-0000-0000-000000000103', $1, $3, $4, 'references_record', 'fixture.links.field_b', 'manual', $5, $5, $6, $6)
`, incident.ID, survivor, loser, target, actor.ID, now); err != nil {
		t.Fatalf("seed merge links: %v", err)
	}
	result, err := links.NewStore().RepointMergedLinksTx(ctx, tx, links.RepointMergedLinksCommand{
		IncidentID: incident.ID, SurvivorRecordID: survivor, LoserRecordID: loser,
		ActorUserID: actor.ID, Now: now.Add(time.Minute),
	})
	if err != nil {
		t.Fatalf("repoint merged links: %v", err)
	}
	if result.DedupedCount != 1 || result.RepointedCount != 1 || len(result.Mutations) != 3 {
		t.Fatalf("merge result = deduped:%d repointed:%d mutations:%d, want 1/1/3", result.DedupedCount, result.RepointedCount, len(result.Mutations))
	}
	var activeA, activeB, loserActive int
	if err := tx.QueryRow(ctx, `
SELECT
    count(*) FILTER (WHERE src_record_id = $2 AND field_key = 'fixture.links.field_a'),
    count(*) FILTER (WHERE src_record_id = $2 AND field_key = 'fixture.links.field_b'),
    count(*) FILTER (WHERE src_record_id = $3)
  FROM record_links
 WHERE incident_id = $1 AND dst_record_id = $4
   AND link_type = 'references_record' AND deleted_at IS NULL
`, incident.ID, survivor, loser, target).Scan(&activeA, &activeB, &loserActive); err != nil {
		t.Fatalf("query merged field-aware links: %v", err)
	}
	if activeA != 1 || activeB != 1 || loserActive != 0 {
		t.Fatalf("post-merge active counts = fieldA:%d fieldB:%d loser:%d, want 1/1/0", activeA, activeB, loserActive)
	}
}

func TestLinksIncidentBundleRejectsUnknownLinkMembersBeforeMutation(t *testing.T) {
	descriptor := links.NewIncidentBundleSourcePort().Descriptor()
	if !slices.Contains(descriptor.InvariantIDs, "links_tags.row_shape_exact") {
		t.Fatalf("Links source descriptor omits exact row-shape invariant: %#v", descriptor.InvariantIDs)
	}
	harness := appsupport.StartStore(t, "links-bundle-unknown-member-characterization")
	actor := authstoretest.SeedLocalUserRecord(t, harness.DB, "links-bundle-shape@example.test", "Links Bundle Shape", "LinksBundleShapePass1!", false, true, true)
	incident := appsupport.CreateIncidentInStore(t, harness.DB, actor, "txn-links-bundle-shape-incident", "IR-LINK-BUNDLE", "Links bundle shape")
	src := uuid.New()
	dst := uuid.New()
	for _, recordID := range []uuid.UUID{src, dst} {
		timelinetest.SeedTimelineRecord(t, harness.DB, incident.ID, actor.ID, recordID)
	}
	now := time.Date(2026, time.August, 21, 18, 42, 0, 0, time.UTC)
	tx, err := harness.DB.BeginTx(context.Background(), pgx.TxOptions{})
	if err != nil {
		t.Fatalf("begin unknown-member import: %v", err)
	}
	defer func() { _ = tx.Rollback(context.Background()) }()
	for _, member := range []string{"note", "description", "comment", "future_undeclared_member"} {
		t.Run(member, func(t *testing.T) {
			linkID := uuid.New()
			row := map[string]any{
				"record_link_id": linkID.String(), "incident_id": incident.ID.String(),
				"src_record_id": src.String(), "dst_record_id": dst.String(),
				"link_type": "references_record", "field_key": "fixture.links.bundle",
				"provenance": "manual", "confidence": nil,
				"owner_user_id": actor.ID.String(), "created_by_user_id": actor.ID.String(),
				"decided_at": now.Format(time.RFC3339Nano), "created_at": now.Format(time.RFC3339Nano),
				"deleted_at": nil, "deleted_by_user_id": nil,
				member: "unsupported value",
			}
			payload, err := json.Marshal(row)
			if err != nil {
				t.Fatalf("encode unknown-member link row: %v", err)
			}
			attributions := &countingLinkAttributions{}
			err = linktest.ImportIncidentBundleFilesTx(context.Background(), tx, map[string][]byte{
				"data/record_links.ndjson": append(payload, '\n'),
				"data/record_tags.ndjson":  {},
			}, actor.ID, attributions)
			var unexpectedColumns *incidentportability.UnexpectedColumnsError
			if !errors.As(err, &unexpectedColumns) {
				t.Fatalf("unknown member error = %T %[1]v, want UnexpectedColumnsError", err)
			}
			if attributions.calls != 0 {
				t.Fatalf("unknown member recorded %d attributions before rejection", attributions.calls)
			}
			var count int
			if err := tx.QueryRow(context.Background(), `SELECT count(*) FROM record_links WHERE record_link_id = $1`, linkID).Scan(&count); err != nil {
				t.Fatalf("query rejected link row: %v", err)
			}
			if count != 0 {
				t.Fatalf("unknown-member import inserted %d rows, want 0", count)
			}
		})
	}
	facts, err := (links.ActiveFactReader{}).LoadTx(context.Background(), tx, incident.ID)
	if err != nil {
		t.Fatalf("load empty active fact set: %v", err)
	}
	if facts.RecordLinks == nil || facts.RecordTags == nil || len(facts.RecordLinks) != 0 || len(facts.RecordTags) != 0 {
		t.Fatalf("empty active facts lost non-nil collection posture: %#v", facts)
	}
	if err := tx.Rollback(context.Background()); err != nil {
		t.Fatalf("close active fact reader transaction: %v", err)
	}
	_, err = (links.ActiveFactReader{}).LoadTx(context.Background(), tx, incident.ID)
	var readFailure *links.ActiveFactsReadError
	if !errors.As(err, &readFailure) {
		t.Fatalf("active fact read error = %T %[1]v, want ActiveFactsReadError", err)
	}
}

type countingLinkAttributions struct {
	calls int
}

func (recorder *countingLinkAttributions) RecordImportedAttribution(string, string, string, string) error {
	recorder.calls++
	return nil
}

func (*countingLinkAttributions) ImportedAttributions() []incidentportability.ImportedAttribution {
	return []incidentportability.ImportedAttribution{}
}

func TestTypedLinksAndTags_Unit(t *testing.T) {
	harness := appsupport.StartStore(t, "saved_view_query-u-8-01-links-tags")
	actor := authstoretest.SeedLocalUserRecord(t, harness.DB, "saved_view_query-u801@example.test", "Workbook query U801", "SavedViewQueryU801Pass1!", false, true, true)
	incident := appsupport.CreateIncidentInStore(t, harness.DB, actor, "txn-saved_view_query-u-8-01-incident", "IR-P8-U801", "Workbook query typed links and tags")
	incidentID := incident.ID
	revisionComposition := revisionsupport.MustComposition(t)
	projections, err := projectionassembly.Build(harness.DB)
	if err != nil {
		t.Fatalf("compose Projections: %v", err)
	}
	conflictTokens := conflicttest.NewCodec("timeline")
	evidenceOwner := appsupport.NewEvidenceOwnerRuntimeForTimeline(
		harness.DB,
		conflictTokens,
		revisionComposition.Runtime.Appender(),
		revisionComposition.Intents,
		projections,
	)
	timelineBundle, err := timelineassembly.NewBundle(timelineassembly.Dependencies{
		Postgres:            harness.DB,
		ConflictTokens:      conflictTokens,
		Revisions:           revisionComposition.Runtime.Appender(),
		Collaboration:       revisionComposition.Intents,
		EvidenceAttachments: evidenceOwner.TimelineAttachmentContribution(),
		TimelineProjection:  projections.TimelinePorts().Writer,
		EntityProjection:    projections.EntityPorts().Writer,
		AssessmentRows:      projections.AssessmentPorts().Rows,
	})
	if err != nil {
		t.Fatalf("compose Timeline: %v", err)
	}
	timelineFacade := timelineBundle.Facade

	t.Run("closed base relationship vocabulary is enforced by structured rows", func(t *testing.T) {
		baseTokens := []string{
			"observed_on_host",
			"observed_as_identity",
			"references_indicator",
			"attached_evidence",
			"references_artifact",
			"derived_from",
			"merged_into",
			"supported_by",
			"references_record",
			"supersedes",
		}
		for _, token := range baseTokens {
			src := uuid.New()
			dst := uuid.New()
			timelinetest.SeedTimelineRecord(t, harness.DB, incidentID, actor.ID, src)
			timelinetest.SeedTimelineRecord(t, harness.DB, incidentID, actor.ID, dst)
			if _, err := harness.DB.Exec(context.Background(), `
INSERT INTO record_links (incident_id, src_record_id, dst_record_id, link_type, provenance, owner_user_id, created_by_user_id)
VALUES ($1, $2, $3, $4, 'manual', $5, $5)
`, incidentID, src, dst, token, actor.ID); err != nil {
				t.Fatalf("base link_type %q rejected: %v", token, err)
			}
		}
		src := uuid.New()
		dst := uuid.New()
		timelinetest.SeedTimelineRecord(t, harness.DB, incidentID, actor.ID, src)
		timelinetest.SeedTimelineRecord(t, harness.DB, incidentID, actor.ID, dst)
		if _, err := harness.DB.Exec(context.Background(), `SAVEPOINT saved_view_query_invalid_link_type`); err != nil {
			t.Fatalf("create invalid link savepoint: %v", err)
		}
		if _, err := harness.DB.Exec(context.Background(), `
INSERT INTO record_links (incident_id, src_record_id, dst_record_id, link_type, provenance, owner_user_id, created_by_user_id)
VALUES ($1, $2, $3, 'free_text_relation', 'manual', $4, $4)
`, incidentID, src, dst, actor.ID); err == nil {
			t.Fatalf("free-text link_type was accepted")
		}
		if _, err := harness.DB.Exec(context.Background(), `ROLLBACK TO SAVEPOINT saved_view_query_invalid_link_type`); err != nil {
			t.Fatalf("rollback invalid link savepoint: %v", err)
		}
	})

	t.Run("timeline tags use add_tag remove_tag and composite mutation targets", func(t *testing.T) {
		request, apiErr := timelineadmission.DecodeTimelineCreateRequest(strings.NewReader(`{
			"client_txn_id": "txn-saved_view_query-u-8-01-create-tags",
			"timeline.activity_synopsis_text": "saved_view_query tags",
			"timeline.tags": {
				"kind": "collection_actions_v1",
				"actions": [
					{ "op": "add_tag", "tag_name": "  Rough  " },
					{ "op": "add_tag", "tag_name": "rough" }
				]
			}
		}`))
		if apiErr != nil {
			t.Fatalf("decode add_tag create request: %#v", apiErr)
		}
		if got := request.Tags.Actions[0].RawText; got != "Rough" {
			t.Fatalf("add_tag did not preserve trimmed display label: got %q", got)
		}
		if got := request.Tags.Actions[0].NormalizedText; got != "rough" {
			t.Fatalf("add_tag did not store folded dedupe key: got %q", got)
		}
		result, err := timelineFacade.CreateRow(context.Background(), timeline.CreateRowCommand{
			Actor:       actor,
			IncidentID:  incidentID,
			Request:     request,
			RequestHash: []byte("txn-saved_view_query-u-8-01-create-tags"),
			RequestID:   "req-saved_view_query-u-8-01-create-tags",
			Now:         time.Now().UTC(),
		})
		if err != nil {
			t.Fatalf("create timeline row: %v", err)
		}
		row := result.Payload["row"].(map[string]any)
		recordID := result.RecordID
		if got := appsupport.QueryCount(t, harness.DB, `
SELECT COUNT(*)
  FROM record_tags
 WHERE incident_id = $1
   AND record_id = $2
   AND tag_name = 'Rough'
   AND normalized_tag_name = 'rough'
   AND deleted_at IS NULL
`, incidentID, recordID); got != 1 {
			t.Fatalf("duplicate add_tag did not coalesce into one normalized structured tag, got %d", got)
		}
		tagID := stringScalarPG(t, harness.DB, `SELECT record_tag_id::text FROM record_tags WHERE record_id = $1 AND deleted_at IS NULL`, recordID)
		item := singleCollectionItem(t, row, "timeline.tags")
		wantRef := "record_tag:" + recordID.String() + ":" + tagID
		if item["item_ref"] != wantRef || item["item_kind"] != "tag" || item["display_text"] != "Rough" || item["tag_id"] != tagID {
			t.Fatalf("unexpected tag collection item: got %#v want item_ref=%s tag_id=%s", item, wantRef, tagID)
		}
		if got := appsupport.QueryCount(t, harness.DB, `
SELECT COUNT(*)
  FROM change_set_mutations
 WHERE target_kind = 'record_tag'
   AND target_id = $1
   AND operation_kind = 'create'
   AND after_value ->> 'record_tag_id' = $2
   AND after_value ->> 'record_id' = $3
   AND after_value ->> 'tag_name' = 'Rough'
   AND after_value ->> 'normalized_tag_name' = 'rough'
`, wantRef, tagID, recordID.String()); got != 1 {
			t.Fatalf("tag add did not persist deterministic record_tag mutation detail, got %d", got)
		}

		patchRequest := timeline.PatchRequest{
			ViewSchemaID:   TimelineView,
			BaseRowVersion: 1,
			ClientTxnID:    "txn-saved_view_query-u-8-01-remove-tag",
			CanonicalChange: []timeline.PatchChange{{
				FieldKey: "timeline.tags",
				ActionPayload: &timeline.CollectionActionPayload{Actions: []timeline.CollectionAction{{
					Op:      "remove_tag",
					ItemRef: wantRef,
				}}},
			}},
		}
		patchedResult, err := timelineFacade.PatchRow(context.Background(), timeline.PatchRowCommand{
			Actor:       actor,
			RecordID:    recordID,
			Request:     patchRequest,
			RequestHash: []byte("txn-saved_view_query-u-8-01-remove-tag"),
			RequestID:   "req-saved_view_query-u-8-01-remove-tag",
			Now:         time.Now().UTC(),
		})
		if err != nil {
			t.Fatalf("patch timeline row: %v", err)
		}
		patched := patchedResult.Payload["row"].(map[string]any)
		if items := collectionItems(t, patched, "timeline.tags"); len(items) != 0 {
			t.Fatalf("remove_tag left active tag items: %#v", items)
		}
		if got := appsupport.QueryCount(t, harness.DB, `
SELECT COUNT(*)
  FROM change_set_mutations
 WHERE target_kind = 'record_tag'
   AND target_id = $1
   AND operation_kind = 'delete'
   AND before_value ->> 'record_tag_id' = $2
   AND after_value ->> 'deleted_at' IS NOT NULL
`, wantRef, tagID); got != 1 {
			t.Fatalf("tag remove did not persist deterministic record_tag delete mutation, got %d", got)
		}
	})

	t.Run("obsolete tag actions and invalid labels fail closed", func(t *testing.T) {
		cases := []struct {
			name string
			tag  string
		}{
			{
				name: "obsolete add_token",
				tag:  `{ "op": "add_token", "raw_text": "obsolete" }`,
			},
			{
				name: "normalized empty",
				tag:  `{ "op": "add_tag", "tag_name": " \t " }`,
			},
			{
				name: "control character",
				tag:  "{ \"op\": \"add_tag\", \"tag_name\": \"bad\\u0001tag\" }",
			},
			{
				name: "too long",
				tag:  `{ "op": "add_tag", "tag_name": "` + strings.Repeat("a", 65) + `" }`,
			},
		}
		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				_, apiErr := timelineadmission.DecodeTimelineCreateRequest(strings.NewReader(`{
					"client_txn_id": "txn-saved_view_query-u-8-01-invalid-tag",
					"timeline.activity_synopsis_text": "invalid",
					"timeline.tags": {
						"kind": "collection_actions_v1",
						"actions": [` + tc.tag + `]
					}
				}`))
				if apiErr == nil || apiErr.Code != "invalid_mutation_payload" {
					t.Fatalf("expected invalid_mutation_payload, got %#v", apiErr)
				}
			})
		}
	})
}

func TestLinkTagProjectionHistoryQuery_Integration(t *testing.T) {
	harness := appsupport.StartServer(t, "saved_view_query-i-8-03-link-tag-atomic")
	login, actorID := appsupport.ProvisionBootstrapAdmin(t, harness.Server)
	incident := appsupport.CreateIncident(t, harness.Server, login, map[string]any{
		"client_txn_id": "txn-saved_view_query-i-8-03-incident",
		"incident_key":  "IR-P8-I803",
		"title":         "Workbook query link tag atomicity",
	})
	incidentID := mustUUID(t, incident["incident_id"].(string))

	row := createTimelineRow(t, harness, login, incidentID, map[string]any{
		"client_txn_id":                   "txn-saved_view_query-i-8-03-create",
		"timeline.activity_synopsis_text": "saved_view_query atomic row",
	})
	recordID := mustUUID(t, row["record_id"].(string))
	evidenceID := uuid.New()
	envelopetest.SeedRecordEnvelope(t, harness.DB, incidentID, actorID, evidenceID, "evidence")
	if _, err := harness.DB.Exec(`
INSERT INTO evidence (record_id, incident_id, title, lifecycle_state, upload_state)
VALUES ($1, $2, 'saved_view_query evidence', 'available', 'available')
`, evidenceID, incidentID); err != nil {
		t.Fatalf("seed evidence: %v", err)
	}

	socket := incidentwstest.ConnectViewSocket(t, harness.Server, incidentID.String(), TimelineView, login.SessionCookie.Value)
	defer socket.Close(websocket.StatusNormalClosure, "test_complete")

	patched := patchTimelineRow(t, harness, login, recordID, map[string]any{
		"view_schema_id":   TimelineView,
		"base_row_version": row["row_version"],
		"client_txn_id":    "txn-saved_view_query-i-8-03-link-tag",
		"changes": []map[string]any{
			{
				"field_key":      "timeline.tags",
				"action_payload": tagActions(map[string]any{"op": "add_tag", "tag_name": " Needs Review "}),
			},
			{
				"field_key": "timeline.attached_evidence_ids",
				"action_payload": collectionActions(map[string]any{
					"op":               "add_record_ref",
					"linked_record_id": evidenceID.String(),
				}),
			},
		},
	})
	if patched["row_version"] != float64(2) {
		t.Fatalf("patch did not advance source row version to 2: %#v", patched)
	}
	tagID := stringScalar(t, harness.DB, `SELECT record_tag_id::text FROM record_tags WHERE record_id = $1 AND deleted_at IS NULL`, recordID)
	tagRef := "record_tag:" + recordID.String() + ":" + tagID
	if got := countRows(t, harness.DB, `SELECT COUNT(*) FROM timeline_grid_projection WHERE record_id = $1 AND row_version = 2`, recordID); got != 1 {
		t.Fatalf("timeline projection did not update atomically, got %d", got)
	}
	if got := countRows(t, harness.DB, `
SELECT COUNT(*)
  FROM change_set_mutations
 WHERE change_set_id::text = $1
   AND target_kind = 'record_tag'
   AND target_id = $2
`, patched["change_set_id"], tagRef); got != 1 {
		t.Fatalf("change metadata missing record_tag target, got %d", got)
	}
	if got := countRows(t, harness.DB, `
SELECT COUNT(*)
  FROM change_set_mutations
 WHERE change_set_id::text = $1
   AND target_kind = 'record_link'
   AND operation_kind = 'create'
   AND after_value ->> 'link_type' = 'attached_evidence'
`, patched["change_set_id"]); got != 1 {
		t.Fatalf("change metadata missing record_link target, got %d", got)
	}

	queryRows := workbookscenariotest.QueryViewRows(t, harness.Server.HTTP.URL, incidentID.String(), TimelineView, login)
	queryRow := workbookscenariotest.FindRow(t, queryRows, recordID.String())
	if got := singleCollectionItem(t, queryRow, "timeline.tags")["item_ref"]; got != tagRef {
		t.Fatalf("workbook query returned stale tag ref: got %#v want %s", got, tagRef)
	}
	if got := singleCollectionItem(t, queryRow, "timeline.attached_evidence_ids")["linked_record_id"]; got != evidenceID.String() {
		t.Fatalf("workbook query returned stale evidence ref: got %#v want %s", got, evidenceID)
	}
	change := incidentwstest.RequireRecordChanged(t, socket, recordID.String(), 2)
	if change.ChangeSetID != patched["change_set_id"].(string) {
		t.Fatalf("record_changed change_set_id got %s want %s", change.ChangeSetID, patched["change_set_id"])
	}
	gotChangedKeys := append([]string(nil), change.ChangedFieldKeys...)
	for _, key := range []string{"timeline.attached_evidence_ids", "timeline.evidence_count", "timeline.has_evidence", "timeline.tags"} {
		if !slices.Contains(gotChangedKeys, key) {
			t.Fatalf("record_changed changed keys missing %s: %v", key, gotChangedKeys)
		}
	}

	history := historyItems(t, getHistory(t, harness, login, recordID))
	tagHistory := historyItemForTarget(t, history, "record_tag", tagRef)
	requireRollbackActionContains(t, tagHistory, "history_entry")
	ref := tagHistory["history_entry_ref"].(string)
	beforeRefRows := countRows(t, harness.DB, `SELECT COUNT(*) FROM record_history_entry_refs WHERE record_id = $1`, recordID)
	rollbackData := rollbackRecord(t, harness, login, recordID, map[string]any{
		"base_row_version": 2,
		"client_txn_id":    "txn-saved_view_query-i-8-03-tag-rollback",
		"target":           map[string]any{"kind": "history_entry", "history_entry_ref": ref},
	})
	if rollbackData["row_version"] != float64(3) {
		t.Fatalf("tag rollback did not advance row version to 3: %#v", rollbackData)
	}
	if got := countRows(t, harness.DB, `SELECT COUNT(*) FROM record_tags WHERE record_tag_id::text = $1 AND deleted_at IS NOT NULL`, tagID); got != 1 {
		t.Fatalf("tag rollback did not tombstone active tag, got %d", got)
	}
	rollbackChangeSetID := rollbackData["rollback_change_set_id"].(string)
	if got := countRows(t, harness.DB, `SELECT COUNT(*) FROM change_sets WHERE change_set_id::text = $1 AND source = 'rollback' AND client_txn_id = 'txn-saved_view_query-i-8-03-tag-rollback'`, rollbackChangeSetID); got != 1 {
		t.Fatalf("tag rollback did not append attributed rollback change_set, got %d", got)
	}
	if got := countRows(t, harness.DB, `SELECT COUNT(*) FROM change_set_mutations WHERE change_set_id::text = $1 AND target_kind = 'record_tag' AND target_id = $2 AND operation_kind = 'rollback'`, rollbackChangeSetID, tagRef); got != 1 {
		t.Fatalf("tag rollback did not append record_tag inverse mutation, got %d", got)
	}
	if got := countRows(t, harness.DB, `SELECT COUNT(*) FROM change_set_mutations WHERE change_set_id::text = $1 AND target_kind = 'record_tag' AND target_id = $2 AND operation_kind = 'create'`, patched["change_set_id"], tagRef); got != 1 {
		t.Fatalf("tag rollback rewrote original mutation, got %d", got)
	}
	if got := countRows(t, harness.DB, `SELECT COUNT(*) FROM record_history_entry_refs WHERE record_id = $1`, recordID); got != beforeRefRows {
		t.Fatalf("tag rollback rewrote prior history refs: before=%d after=%d", beforeRefRows, got)
	}
	afterRollbackRows := workbookscenariotest.QueryViewRows(t, harness.Server.HTTP.URL, incidentID.String(), TimelineView, login)
	afterRollbackRow := workbookscenariotest.FindRow(t, afterRollbackRows, recordID.String())
	if items := collectionItems(t, afterRollbackRow, "timeline.tags"); len(items) != 0 {
		t.Fatalf("query still shows rolled-back tag: %#v", items)
	}
	if got := singleCollectionItem(t, afterRollbackRow, "timeline.attached_evidence_ids")["linked_record_id"]; got != evidenceID.String() {
		t.Fatalf("tag rollback changed unrelated link: got %#v want %s", got, evidenceID)
	}
}

func createTimelineRow(t testing.TB, harness *appsupport.ServerHarness, login appsupport.LoginResult, incidentID uuid.UUID, body map[string]any) map[string]any {
	t.Helper()
	resp := appsupport.DoJSON(t, http.MethodPost, harness.Server.HTTP.URL+"/api/v1/incidents/"+incidentID.String()+"/views/"+TimelineView+"/rows", body, appsupport.WithCookies(login.SessionCookie, login.CSRFCookie), appsupport.WithHeader(authn.CSRFHeaderName, login.CSRFCookie.Value))
	data := httptestx.RequireSuccessEnvelope(t, resp, http.StatusCreated)["data"].(map[string]any)
	row := data["row"].(map[string]any)
	row["change_set_id"] = data["change_set_id"]
	return row
}

func patchTimelineRow(t testing.TB, harness *appsupport.ServerHarness, login appsupport.LoginResult, recordID uuid.UUID, body map[string]any) map[string]any {
	t.Helper()
	resp := appsupport.DoJSON(t, http.MethodPatch, harness.Server.HTTP.URL+"/api/v1/records/"+recordID.String(), body, appsupport.WithCookies(login.SessionCookie, login.CSRFCookie), appsupport.WithHeader(authn.CSRFHeaderName, login.CSRFCookie.Value))
	data := httptestx.RequireSuccessEnvelope(t, resp, http.StatusOK)["data"].(map[string]any)
	row := data["row"].(map[string]any)
	row["change_set_id"] = data["change_set_id"]
	return row
}

func rollbackRecord(t testing.TB, harness *appsupport.ServerHarness, login appsupport.LoginResult, recordID uuid.UUID, body map[string]any) map[string]any {
	t.Helper()
	resp := appsupport.DoJSON(t, http.MethodPost, harness.Server.HTTP.URL+"/api/v1/records/"+recordID.String()+"/rollback", body, appsupport.WithCookies(login.SessionCookie, login.CSRFCookie), appsupport.WithHeader(authn.CSRFHeaderName, login.CSRFCookie.Value))
	if resp.StatusCode != http.StatusOK {
		var envelope map[string]any
		if err := json.NewDecoder(resp.Body).Decode(&envelope); err == nil {
			t.Fatalf("rollback status got %d want 200: %#v", resp.StatusCode, envelope)
		}
	}
	return httptestx.RequireSuccessEnvelope(t, resp, http.StatusOK)["data"].(map[string]any)
}

func getHistory(t testing.TB, harness *appsupport.ServerHarness, login appsupport.LoginResult, recordID uuid.UUID) map[string]any {
	t.Helper()
	resp := appsupport.DoJSON(t, http.MethodGet, harness.Server.HTTP.URL+"/api/v1/records/"+recordID.String()+"/history", nil, appsupport.WithCookies(login.SessionCookie))
	return httptestx.RequireSuccessEnvelope(t, resp, http.StatusOK)
}

func tagActions(actions ...map[string]any) map[string]any {
	return collectionActions(actions...)
}

func collectionActions(actions ...map[string]any) map[string]any {
	return map[string]any{"kind": "collection_actions_v1", "actions": actions}
}

func collectionItems(t testing.TB, row map[string]any, fieldKey string) []map[string]any {
	t.Helper()
	value := row["cells"].(map[string]any)[fieldKey].(map[string]any)["value"].(map[string]any)
	items := make([]map[string]any, 0)
	switch rawItems := value["items"].(type) {
	case []any:
		for _, rawItem := range rawItems {
			items = append(items, rawItem.(map[string]any))
		}
	case []map[string]any:
		items = append(items, rawItems...)
	default:
		t.Fatalf("unexpected %s items type %T", fieldKey, value["items"])
	}
	return items
}

func singleCollectionItem(t testing.TB, row map[string]any, fieldKey string) map[string]any {
	t.Helper()
	items := collectionItems(t, row, fieldKey)
	if len(items) != 1 {
		t.Fatalf("expected one %s item, got %#v", fieldKey, items)
	}
	return items[0]
}

func historyItems(t testing.TB, envelope map[string]any) []map[string]any {
	t.Helper()
	raw := envelope["data"].(map[string]any)["items"].([]any)
	items := make([]map[string]any, 0, len(raw))
	for _, rawItem := range raw {
		items = append(items, rawItem.(map[string]any))
	}
	return items
}

func historyItemForTarget(t testing.TB, items []map[string]any, targetKind string, targetID string) map[string]any {
	t.Helper()
	for _, item := range items {
		summary := item["diff_summary"].(map[string]any)
		units := summary["units"].([]any)
		for _, rawUnit := range units {
			unit := rawUnit.(map[string]any)
			if unit["target_kind"] == targetKind && unit["target_id"] == targetID {
				return item
			}
		}
	}
	t.Fatalf("missing history item target_kind=%s target_id=%s in %#v", targetKind, targetID, items)
	return nil
}

func requireRollbackActionContains(t testing.TB, item map[string]any, expected string) {
	t.Helper()
	raw := item["available_rollback_actions"].([]any)
	for _, value := range raw {
		if value.(string) == expected {
			if _, ok := item["history_entry_ref"].(string); expected == "history_entry" && !ok {
				t.Fatalf("history_entry rollback item missing history_entry_ref: %#v", item)
			}
			return
		}
	}
	t.Fatalf("rollback actions missing %s in %#v", expected, item)
}

func countRows(t testing.TB, db *sql.DB, query string, args ...any) int {
	t.Helper()
	var count int
	if err := db.QueryRow(query, args...).Scan(&count); err != nil {
		t.Fatalf("count rows: %v", err)
	}
	return count
}

func stringScalar(t testing.TB, db *sql.DB, query string, args ...any) string {
	t.Helper()
	var value string
	if err := db.QueryRow(query, args...).Scan(&value); err != nil {
		t.Fatalf("query scalar: %v", err)
	}
	return value
}

func stringScalarPG(t testing.TB, db postgres.DB, query string, args ...any) string {
	t.Helper()
	var value string
	if err := db.QueryRow(context.Background(), query, args...).Scan(&value); err != nil {
		t.Fatalf("query scalar: %v", err)
	}
	return value
}

func columnNamesPG(t testing.TB, db postgres.DB, tableName string) []string {
	t.Helper()
	rows, err := db.Query(context.Background(), `
SELECT column_name
  FROM information_schema.columns
 WHERE table_schema = 'public'
   AND table_name = $1
 ORDER BY ordinal_position
`, tableName)
	if err != nil {
		t.Fatalf("query %s columns: %v", tableName, err)
	}
	defer rows.Close()
	var columns []string
	for rows.Next() {
		var column string
		if err := rows.Scan(&column); err != nil {
			t.Fatalf("scan %s column: %v", tableName, err)
		}
		columns = append(columns, column)
	}
	if err := rows.Err(); err != nil {
		t.Fatalf("iterate %s columns: %v", tableName, err)
	}
	return columns
}

func mustUUID(t testing.TB, value string) uuid.UUID {
	t.Helper()
	parsed, err := uuid.Parse(value)
	if err != nil {
		t.Fatalf("parse uuid %q: %v", value, err)
	}
	return parsed
}

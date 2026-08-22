package links_test

import (
	"context"
	"slices"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	authstoretest "github.com/JochiRaider/cartulary/internal/modules/auth/testsupport/storetest"
	entitytest "github.com/JochiRaider/cartulary/internal/modules/entities/testsupport"
	"github.com/JochiRaider/cartulary/internal/modules/links"
	timelinetest "github.com/JochiRaider/cartulary/internal/modules/timeline/testsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/appsupport"
)

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
	nonCanonicalConfidence := 99
	if _, err := store.UpsertLinkCommandTx(context.Background(), tx, links.UpsertLinkCommand{
		IncidentID:  incident.ID,
		SrcRecordID: src,
		DstRecordID: dst,
		LinkType:    links.LinkType(links.LinkTypeReferencesRecord),
		Provenance:  links.LinkProvenance(links.LinkProvenanceManual),
		Confidence:  &confidence,
		OwnerUserID: actor.ID,
		Now:         time.Now().UTC(),
	}); err == nil {
		t.Fatalf("manual link confidence should be rejected, got %v", err)
	}
	if _, err := store.UpsertLinkCommandTx(context.Background(), tx, links.UpsertLinkCommand{
		IncidentID:  incident.ID,
		SrcRecordID: src,
		DstRecordID: dst,
		LinkType:    links.LinkType(links.LinkTypeSupportedBy),
		Provenance:  links.LinkProvenance(links.LinkProvenanceAutoMatch),
		Confidence:  &confidence,
		OwnerUserID: actor.ID,
		Now:         time.Now().UTC(),
	}); err == nil {
		t.Fatalf("auto_match on unsupported link type should be rejected, got %v", err)
	}
	if _, err := store.UpsertLinkCommandTx(context.Background(), tx, links.UpsertLinkCommand{
		IncidentID:  incident.ID,
		SrcRecordID: src,
		DstRecordID: src,
		LinkType:    links.LinkType(links.LinkTypeReferencesRecord),
		Provenance:  links.LinkProvenance(links.LinkProvenanceManual),
		OwnerUserID: actor.ID,
		Now:         time.Now().UTC(),
	}); err == nil {
		t.Fatalf("self-link should be rejected, got %v", err)
	}
	if _, err := store.UpsertLinkCommandTx(context.Background(), tx, links.UpsertLinkCommand{
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
	if _, err := store.UpsertLinkCommandTx(context.Background(), tx, links.UpsertLinkCommand{
		IncidentID:  incident.ID,
		SrcRecordID: src,
		DstRecordID: host,
		LinkType:    links.LinkType(links.LinkTypeObservedOnHost),
		Provenance:  links.LinkProvenance(links.LinkProvenanceAutoMatch),
		Confidence:  &nonCanonicalConfidence,
		OwnerUserID: actor.ID,
		Now:         time.Now().UTC(),
	}); err == nil {
		t.Fatalf("auto_match confidence other than 100 should be rejected, got %v", err)
	}
	supersedesResult, err := store.InsertSupersedesCommandTx(context.Background(), tx, links.InsertSupersedesCommand{
		IncidentID:          incident.ID,
		ReplacementRecordID: replacement,
		SupersededRecordID:  superseded,
		OwnerUserID:         actor.ID,
		Now:                 time.Now().UTC(),
	})
	if err != nil {
		t.Fatalf("valid timeline supersedes rejected: %v", err)
	}
	if supersedesResult.Mutation == nil || supersedesResult.Mutation.Operation != "create" || supersedesResult.LinkType != links.LinkTypeSupersedes {
		t.Fatalf("supersedes result missing canonical create mutation: %#v", supersedesResult)
	}
	requireCanonicalRecordLinkMutationValue(t, supersedesResult.Mutation.AfterValue)
	if _, err := store.InsertSupersedesCommandTx(context.Background(), tx, links.InsertSupersedesCommand{
		IncidentID:          incident.ID,
		ReplacementRecordID: replacement,
		SupersededRecordID:  host,
		OwnerUserID:         actor.ID,
		Now:                 time.Now().UTC(),
	}); err == nil {
		t.Fatalf("mixed-type supersedes should be rejected, got %v", err)
	}
	if _, err := tx.Exec(context.Background(), `UPDATE records SET deleted_at = now(), deleted_by_user_id = $2 WHERE record_id = $1`, dst, actor.ID); err != nil {
		t.Fatalf("soft-delete validation endpoint: %v", err)
	}
	if _, err := store.UpsertLinkCommandTx(context.Background(), tx, links.UpsertLinkCommand{
		IncidentID:  incident.ID,
		SrcRecordID: src,
		DstRecordID: dst,
		LinkType:    links.LinkType(links.LinkTypeReferencesRecord),
		Provenance:  links.LinkProvenance(links.LinkProvenanceManual),
		OwnerUserID: actor.ID,
		Now:         time.Now().UTC(),
	}); err == nil {
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
	host := uuid.New()
	entitytest.SeedHostRecord(t, harness.DB, incident.ID, actor.ID, host, "Links command host", "links-command-host", "", "")
	tx, err := harness.DB.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		t.Fatalf("begin field-aware characterization: %v", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	store := links.NewStore()
	fieldA := "fixture.links.field_a"
	fieldB := "fixture.links.field_b"
	fieldAResult, err := store.ApplyRecordRefCollectionWithMutationValuesTx(ctx, tx, links.RecordRefCollectionCommand{
		IncidentID: incident.ID, SourceRecordID: src, ActorUserID: actor.ID,
		FieldKey: fieldA, LinkType: links.LinkType(links.LinkTypeReferencesRecord), ExpectedTargetType: "timeline_event",
		AddRecordIDs: []uuid.UUID{dst}, Now: now,
	})
	if err != nil || len(fieldAResult.RecordLinks) != 1 {
		t.Fatalf("insert field A link: result=%#v err=%v", fieldAResult, err)
	}
	linkAID := fieldAResult.RecordLinks[0].RecordLinkID
	replayedA, err := store.ApplyRecordRefCollectionWithMutationValuesTx(ctx, tx, links.RecordRefCollectionCommand{
		IncidentID: incident.ID, SourceRecordID: src, ActorUserID: actor.ID,
		FieldKey: fieldA, LinkType: links.LinkType(links.LinkTypeReferencesRecord), ExpectedTargetType: "timeline_event",
		AddRecordIDs: []uuid.UUID{dst}, Now: now.Add(time.Second),
	})
	if err != nil || len(replayedA.RecordLinks) != 0 {
		t.Fatalf("same-field replay = (%#v, %v), want no mutation for %s", replayedA, err, linkAID)
	}
	fieldBResult, err := store.ApplyRecordRefCollectionWithMutationValuesTx(ctx, tx, links.RecordRefCollectionCommand{
		IncidentID: incident.ID, SourceRecordID: src, ActorUserID: actor.ID,
		FieldKey: fieldB, LinkType: links.LinkType(links.LinkTypeReferencesRecord), ExpectedTargetType: "timeline_event",
		AddRecordIDs: []uuid.UUID{dst}, Now: now,
	})
	if err != nil || len(fieldBResult.RecordLinks) != 1 || fieldBResult.RecordLinks[0].RecordLinkID == linkAID {
		t.Fatalf("different-field insert = (%#v, %v), want distinct active binding", fieldBResult, err)
	}
	unfielded, err := store.UpsertLinkCommandTx(ctx, tx, links.UpsertLinkCommand{
		IncidentID: incident.ID, SrcRecordID: src, DstRecordID: dst,
		LinkType: links.LinkType(links.LinkTypeReferencesRecord), Provenance: links.LinkProvenance(links.LinkProvenanceManual),
		OwnerUserID: actor.ID, Now: now,
	})
	if err != nil || unfielded.Mutation == nil || unfielded.Mutation.Operation != "create" {
		t.Fatalf("insert null-field link: result=%#v err=%v", unfielded, err)
	}
	replayedUnfielded, err := store.UpsertLinkCommandTx(ctx, tx, links.UpsertLinkCommand{
		IncidentID: incident.ID, SrcRecordID: src, DstRecordID: dst,
		LinkType: links.LinkType(links.LinkTypeReferencesRecord), Provenance: links.LinkProvenance(links.LinkProvenanceManual),
		OwnerUserID: actor.ID, Now: now.Add(time.Second),
	})
	if err != nil || replayedUnfielded.Mutation != nil || replayedUnfielded.RecordLinkID != unfielded.RecordLinkID {
		t.Fatalf("null-field replay = (%#v, %v), want original %s without mutation", replayedUnfielded, err, unfielded.RecordLinkID)
	}
	requireCanonicalRecordLinkMutationValue(t, unfielded.Mutation.AfterValue)
	if unfielded.Mutation.BeforeValue != nil || unfielded.Mutation.AfterValue["field_key"] != nil || unfielded.Mutation.AfterValue["confidence"] != nil {
		t.Fatalf("unexpected manual create mutation: %#v", unfielded.Mutation)
	}

	metadataCreate, err := store.UpsertLinkCommandTx(ctx, tx, links.UpsertLinkCommand{
		IncidentID: incident.ID, SrcRecordID: src, DstRecordID: host,
		LinkType: links.LinkType(links.LinkTypeObservedOnHost), Provenance: links.LinkProvenance(links.LinkProvenanceManual),
		OwnerUserID: actor.ID, Now: now,
	})
	if err != nil || metadataCreate.Mutation == nil || metadataCreate.Mutation.Operation != "create" {
		t.Fatalf("create metadata-test link: result=%#v err=%v", metadataCreate, err)
	}
	metadataCreate.Mutation.AfterValue["provenance"] = "caller_corruption"
	autoConfidence := 100
	metadataPatch, err := store.UpsertLinkCommandTx(ctx, tx, links.UpsertLinkCommand{
		IncidentID: incident.ID, SrcRecordID: src, DstRecordID: host,
		LinkType: links.LinkType(links.LinkTypeObservedOnHost), Provenance: links.LinkProvenance(links.LinkProvenanceAutoMatch),
		Confidence: &autoConfidence, OwnerUserID: actor.ID, Now: now.Add(time.Second),
	})
	if err != nil || metadataPatch.Mutation == nil || metadataPatch.Mutation.Operation != "patch" {
		t.Fatalf("patch link metadata: result=%#v err=%v", metadataPatch, err)
	}
	requireCanonicalRecordLinkMutationValue(t, metadataPatch.Mutation.BeforeValue)
	requireCanonicalRecordLinkMutationValue(t, metadataPatch.Mutation.AfterValue)
	if metadataPatch.Mutation.BeforeValue["provenance"] != links.LinkProvenanceManual || metadataPatch.Mutation.AfterValue["provenance"] != links.LinkProvenanceAutoMatch || metadataPatch.Mutation.AfterValue["confidence"] != 100 {
		t.Fatalf("unexpected metadata patch mutation: %#v", metadataPatch.Mutation)
	}
	metadataPatch.Mutation.AfterValue["provenance"] = "caller_corruption"
	tombstoned, found, err := store.TombstoneActiveLinkCommandTx(ctx, tx, links.TombstoneActiveLinkCommand{
		IncidentID: incident.ID, SrcRecordID: src, DstRecordID: host,
		LinkType: links.LinkType(links.LinkTypeObservedOnHost), ActorUserID: actor.ID, Now: now.Add(2 * time.Second),
	})
	if err != nil || !found || tombstoned.Mutation == nil || tombstoned.Mutation.Operation != "delete" {
		t.Fatalf("tombstone active link: result=%#v found=%t err=%v", tombstoned, found, err)
	}
	requireCanonicalRecordLinkMutationValue(t, tombstoned.Mutation.BeforeValue)
	requireCanonicalRecordLinkMutationValue(t, tombstoned.Mutation.AfterValue)
	if tombstoned.Mutation.BeforeValue["provenance"] != links.LinkProvenanceAutoMatch || tombstoned.Mutation.AfterValue["deleted_by_user_id"] != actor.ID.String() {
		t.Fatalf("tombstone did not preserve canonical before/after state: %#v", tombstoned.Mutation)
	}
	absent, found, err := store.TombstoneActiveLinkCommandTx(ctx, tx, links.TombstoneActiveLinkCommand{
		IncidentID: incident.ID, SrcRecordID: src, DstRecordID: host,
		LinkType: links.LinkType(links.LinkTypeObservedOnHost), ActorUserID: actor.ID, Now: now.Add(3 * time.Second),
	})
	if err != nil || found || absent.Mutation != nil {
		t.Fatalf("second tuple tombstone = (%#v, %t, %v), want absent no-op", absent, found, err)
	}

	collectionField := "fixture.links.collection"
	collectionCreate, err := store.ApplyRecordRefCollectionWithMutationValuesTx(ctx, tx, links.RecordRefCollectionCommand{
		IncidentID: incident.ID, SourceRecordID: src, ActorUserID: actor.ID,
		FieldKey: collectionField, LinkType: links.LinkType(links.LinkTypeReferencesRecord), ExpectedTargetType: "timeline_event",
		AddRecordIDs: []uuid.UUID{dst}, Now: now.Add(4 * time.Second),
	})
	if err != nil || len(collectionCreate.RecordLinks) != 1 || collectionCreate.RecordTags == nil {
		t.Fatalf("create collection mutation: result=%#v err=%v", collectionCreate, err)
	}
	requireCanonicalRecordLinkMutationValue(t, collectionCreate.RecordLinks[0].AfterValue)
	collectionCreate.RecordLinks[0].AfterValue["field_key"] = "caller_corruption"
	collectionDelete, err := store.ApplyRecordRefCollectionWithMutationValuesTx(ctx, tx, links.RecordRefCollectionCommand{
		IncidentID: incident.ID, SourceRecordID: src, ActorUserID: actor.ID,
		FieldKey: collectionField, LinkType: links.LinkType(links.LinkTypeReferencesRecord), ExpectedTargetType: "timeline_event",
		RemoveRecordIDs: []uuid.UUID{dst}, Now: now.Add(5 * time.Second),
	})
	if err != nil || len(collectionDelete.RecordLinks) != 1 || collectionDelete.RecordLinks[0].BeforeValue["field_key"] != collectionField {
		t.Fatalf("delete collection mutation: result=%#v err=%v", collectionDelete, err)
	}
	requireCanonicalRecordLinkMutationValue(t, collectionDelete.RecordLinks[0].BeforeValue)
	requireCanonicalRecordLinkMutationValue(t, collectionDelete.RecordLinks[0].AfterValue)

	tagCreate, err := store.ApplyTagCollectionWithMutationValuesTx(ctx, tx, links.TagCollectionCommand{
		IncidentID: incident.ID, RecordID: src, ActorUserID: actor.ID, FieldKey: "timeline.tags",
		AddTags: []links.TagCollectionAdd{{RawText: "Durable", NormalizedText: "durable"}}, Now: now.Add(6 * time.Second),
	})
	if err != nil || len(tagCreate.RecordTags) != 1 || tagCreate.RecordLinks == nil {
		t.Fatalf("create tag mutation: result=%#v err=%v", tagCreate, err)
	}
	requireCanonicalRecordTagMutationValue(t, tagCreate.RecordTags[0].AfterValue)
	tagID := tagCreate.RecordTags[0].RecordTagID
	tagCreate.RecordTags[0].AfterValue["tag_name"] = "caller_corruption"
	tagDelete, err := store.ApplyTagCollectionWithMutationValuesTx(ctx, tx, links.TagCollectionCommand{
		IncidentID: incident.ID, RecordID: src, ActorUserID: actor.ID, FieldKey: "timeline.tags",
		RemoveTags: []links.RecordTagRef{{RecordID: src, RecordTagID: tagID}}, Now: now.Add(7 * time.Second),
	})
	if err != nil || len(tagDelete.RecordTags) != 1 || tagDelete.RecordTags[0].BeforeValue["tag_name"] != "Durable" {
		t.Fatalf("delete tag mutation: result=%#v err=%v", tagDelete, err)
	}
	requireCanonicalRecordTagMutationValue(t, tagDelete.RecordTags[0].BeforeValue)
	requireCanonicalRecordTagMutationValue(t, tagDelete.RecordTags[0].AfterValue)

	factReader := links.FactReader{}
	recordFacts, err := factReader.LoadRecordTx(ctx, tx, incident.ID, src)
	if err != nil {
		t.Fatalf("load record-scoped facts: %v", err)
	}
	if recordFacts.RecordLinks == nil || recordFacts.RecordTags == nil || len(recordFacts.RecordLinks) != 3 || len(recordFacts.RecordTags) != 0 {
		t.Fatalf("record-scoped facts = %#v, want three active outbound links and non-nil empty tags", recordFacts)
	}
	for _, fact := range recordFacts.RecordLinks {
		if fact.LinkType.String() != links.LinkTypeReferencesRecord || fact.Provenance.String() != links.LinkProvenanceManual {
			t.Fatalf("record fact lost typed link vocabulary: %#v", fact)
		}
	}
	linkChanges, err := factReader.LoadCollectionChangesTx(ctx, tx, incident.ID, src, now)
	if err != nil {
		t.Fatalf("load link collection changes: %v", err)
	}
	if !slices.Equal(linkChanges.LinkFieldKeys, []string{fieldA, fieldB}) || linkChanges.TagsChanged {
		t.Fatalf("link collection changes = %#v, want sorted unique field keys without tags", linkChanges)
	}
	tagChanges, err := factReader.LoadCollectionChangesTx(ctx, tx, incident.ID, src, now.Add(6*time.Second))
	if err != nil {
		t.Fatalf("load tag collection changes: %v", err)
	}
	if tagChanges.LinkFieldKeys == nil || len(tagChanges.LinkFieldKeys) != 0 || !tagChanges.TagsChanged {
		t.Fatalf("tag collection changes = %#v, want non-nil empty link fields and TagsChanged", tagChanges)
	}
	foreignIncidentChanges, err := factReader.LoadCollectionChangesTx(ctx, tx, uuid.New(), src, now)
	if err != nil {
		t.Fatalf("load foreign-incident collection changes: %v", err)
	}
	if foreignIncidentChanges.LinkFieldKeys == nil || len(foreignIncidentChanges.LinkFieldKeys) != 0 || foreignIncidentChanges.TagsChanged {
		t.Fatalf("foreign incident leaked collection changes: %#v", foreignIncidentChanges)
	}

	fieldADelete, err := store.ApplyRecordRefCollectionWithMutationValuesTx(ctx, tx, links.RecordRefCollectionCommand{
		IncidentID: incident.ID, SourceRecordID: src, ActorUserID: actor.ID,
		FieldKey: fieldA, LinkType: links.LinkType(links.LinkTypeReferencesRecord), ExpectedTargetType: "timeline_event",
		RemoveRecordIDs: []uuid.UUID{dst}, Now: now.Add(2 * time.Second),
	})
	if err != nil || len(fieldADelete.RecordLinks) != 1 || fieldADelete.RecordLinks[0].RecordLinkID != linkAID {
		t.Fatalf("tombstone exact field A binding: result=%#v err=%v", fieldADelete, err)
	}
	if _, err := store.ApplyRecordRefCollectionWithMutationValuesTx(ctx, tx, links.RecordRefCollectionCommand{
		IncidentID: incident.ID, SourceRecordID: src, ActorUserID: actor.ID,
		FieldKey: "fixture.links.foreign", LinkType: links.LinkType(links.LinkTypeReferencesRecord), ExpectedTargetType: "timeline_event",
		RemoveRecordIDs: []uuid.UUID{dst}, Now: now.Add(3 * time.Second),
	}); err == nil {
		t.Fatal("foreign field removal unexpectedly succeeded")
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
